<script setup lang="ts">
/* eslint-disable vue/no-v-html */
import { ref, nextTick, computed, onMounted } from 'vue';
import { useI18n } from 'vue-i18n';
import {
  PhChatCircleText,
  PhX,
  PhPaperPlaneRight,
  PhSpinner,
  PhClockCounterClockwise,
  PhPlus,
  PhTrash,
  PhPencil,
} from '@phosphor-icons/vue';
import type { Article } from '@/types/models';

interface ChatMessage {
  id: number;
  role: 'user' | 'assistant';
  content: string;
  html?: string; // Pre-rendered HTML from backend
  thinking?: string;
  created_at: string;
}

interface ChatSession {
  id: number;
  article_id: number;
  title: string;
  created_at: string;
  updated_at: string;
  message_count: number;
}

interface Props {
  article: Article;
  articleContent: string;
  settings: { ai_chat_enabled: boolean };
}

const props = defineProps<Props>();

const emit = defineEmits<{
  close: [];
}>();

const { t } = useI18n();

const isOpen = ref(true);
const isLoading = ref(false);
const inputMessage = ref('');
const messages = ref<ChatMessage[]>([]);
const chatContainer = ref<HTMLElement | null>(null);
const isFirstMessage = ref(true);
const currentSessionId = ref<number | null>(null);
const sessions = ref<ChatSession[]>([]);
const showSessions = ref(false);
const editingSessionId = ref<number | null>(null);
const editingSessionTitle = ref('');

// Resize functionality
const isResizing = ref(false);
const startX = ref(0);
const startY = ref(0);
const startWidth = ref(0);
const startHeight = ref(0);
const panelElement = ref<HTMLElement | null>(null);

// Initialize: load sessions for this article
onMounted(async () => {
  await loadSessions();
  // Auto-select the most recent session if available
  if (sessions.value.length > 0) {
    await selectSession(sessions.value[0].id);
  }
});

async function loadSessions() {
  try {
    const response = await fetch(`/api/ai/chat/sessions?article_id=${props.article.id}`);
    if (response.ok) {
      sessions.value = await response.json();
    }
  } catch (e) {
    console.error('Failed to load sessions:', e);
  }
}

async function selectSession(sessionId: number) {
  try {
    const response = await fetch(`/api/ai/chat/messages?session_id=${sessionId}`);
    if (response.ok) {
      const loadedMessages = await response.json();
      messages.value = loadedMessages;
      currentSessionId.value = sessionId;
      // Set isFirstMessage based on whether the session has any messages
      // New sessions (no messages) should have isFirstMessage = true
      isFirstMessage.value = loadedMessages.length === 0;
      showSessions.value = false;
      await nextTick();
      scrollToBottom();
    }
  } catch (e) {
    console.error('Failed to load session messages:', e);
  }
}

async function createNewSession() {
  try {
    const response = await fetch('/api/ai/chat/session/create', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        article_id: props.article.id,
        title: t('article.chat.newChat'),
      }),
    });

    if (response.ok) {
      const newSession = await response.json();
      sessions.value.unshift(newSession);
      // Select the new session and reset for first message
      await selectSession(newSession.id);
      // Ensure isFirstMessage is true for new sessions
      isFirstMessage.value = true;
    }
  } catch (e) {
    console.error('Failed to create session:', e);
  }
}

async function deleteSession(sessionId: number, e: Event) {
  e.stopPropagation();
  const confirmed = await window.showConfirm({
    title: t('common.confirm'),
    message: t('article.chat.confirmDeleteSession'),
    isDanger: true,
  });
  if (!confirmed) return;

  try {
    await fetch(`/api/ai/chat/session?session_id=${sessionId}`, {
      method: 'DELETE',
    });

    sessions.value = sessions.value.filter((s) => s.id !== sessionId);
    if (currentSessionId.value === sessionId) {
      currentSessionId.value = null;
      messages.value = [];
      if (sessions.value.length > 0) {
        await selectSession(sessions.value[0].id);
      }
    }
  } catch (e) {
    console.error('Failed to delete session:', e);
  }
}

function startEditSession(session: ChatSession, e: Event) {
  e.stopPropagation();
  editingSessionId.value = session.id;
  editingSessionTitle.value = session.title;
}

async function saveSessionTitle(sessionId: number) {
  try {
    await fetch(`/api/ai/chat/session?session_id=${sessionId}`, {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ title: editingSessionTitle.value }),
    });

    const session = sessions.value.find((s) => s.id === sessionId);
    if (session) {
      session.title = editingSessionTitle.value;
    }
    editingSessionId.value = null;
  } catch (e) {
    console.error('Failed to update session title:', e);
  }
}

function cancelEditSession() {
  editingSessionId.value = null;
  editingSessionTitle.value = '';
}

function startResize(e: MouseEvent) {
  isResizing.value = true;
  startX.value = e.clientX;
  startY.value = e.clientY;

  const panel = panelElement.value;
  if (panel) {
    const rect = panel.getBoundingClientRect();
    startWidth.value = rect.width;
    startHeight.value = rect.height;
  }

  document.addEventListener('mousemove', resize);
  document.addEventListener('mouseup', stopResize);
  e.preventDefault();
  e.stopPropagation();
}

function resize(e: MouseEvent) {
  if (!isResizing.value) return;

  const deltaX = startX.value - e.clientX;
  const deltaY = startY.value - e.clientY;

  const newWidth = Math.max(300, startWidth.value + deltaX);
  const newHeight = Math.max(200, startHeight.value + deltaY);

  const panel = panelElement.value;
  if (panel) {
    panel.classList.remove('w-[500px]', 'h-[600px]', 'w-[calc(100%-2rem)]', 'md:w-96');
    panel.style.width = `${newWidth}px`;
    panel.style.height = `${newHeight}px`;
    panel.style.maxWidth = 'none';
    panel.style.maxHeight = 'none';
  }
}

function stopResize() {
  isResizing.value = false;
  document.removeEventListener('mousemove', resize);
  document.removeEventListener('mouseup', stopResize);
}

async function sendMessage() {
  const message = inputMessage.value.trim();
  if (!message || isLoading.value) return;

  messages.value.push({
    id: 0,
    role: 'user',
    content: message,
    created_at: new Date().toISOString(),
  });
  inputMessage.value = '';
  isLoading.value = true;

  await nextTick();
  scrollToBottom();

  try {
    // Prepare article content for AI context
    // Use up to 50000 characters for better context while staying reasonable
    const articleContent = props.articleContent ? props.articleContent.slice(0, 50000) : '';

    const requestBody: any = {
      session_id: currentSessionId.value,
      article_id: props.article.id,
      messages: messages.value.slice(-10),
      is_first_message: isFirstMessage.value,
      article_title: props.article.title,
      article_url: props.article.url,
      // Include article content to ensure AI has context
      article_content: articleContent,
    };

    const response = await fetch('/api/ai-chat', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(requestBody),
    });

    if (response.ok) {
      const data = await response.json();
      messages.value.push({
        id: 0,
        role: 'assistant',
        content: data.response,
        html: data.html, // Use pre-rendered HTML from backend
        thinking: data.thinking,
        created_at: new Date().toISOString(),
      });

      if (data.session_id && data.session_id !== currentSessionId.value) {
        currentSessionId.value = data.session_id;
        await loadSessions();
      }

      isFirstMessage.value = false;
    } else {
      const errorText = await response.text();
      console.error('AI chat error response:', response.status, errorText);

      let errorMessage = t('article.chat.aiChatError');
      try {
        const errorData = JSON.parse(errorText);
        // Extract error message from various possible formats
        if (typeof errorData.error === 'string') {
          errorMessage = errorData.error;
        } else if (errorData.error?.message) {
          errorMessage = errorData.error.message;
        } else if (errorData.message) {
          errorMessage = errorData.message;
        } else if (typeof errorData === 'string') {
          errorMessage = errorData;
        } else {
          errorMessage = t('article.chat.aiChatError');
        }
      } catch {
        errorMessage = errorText || t('article.chat.aiChatError');
      }

      messages.value.push({
        id: 0,
        role: 'assistant',
        content: errorMessage,
        created_at: new Date().toISOString(),
      });
    }
  } catch (e) {
    console.error('AI chat error:', e);
    messages.value.push({
      id: 0,
      role: 'assistant',
      content: t('article.chat.aiChatError'),
      created_at: new Date().toISOString(),
    });
  } finally {
    isLoading.value = false;
    await nextTick();
    scrollToBottom();
  }
}

function scrollToBottom() {
  if (chatContainer.value) {
    chatContainer.value.scrollTop = chatContainer.value.scrollHeight;
  }
}

function handleKeydown(e: KeyboardEvent) {
  if (e.key === 'Enter' && !e.shiftKey) {
    e.preventDefault();
    sendMessage();
  }
}

const currentSessionTitle = computed(() => {
  if (currentSessionId.value) {
    const session = sessions.value.find((s) => s.id === currentSessionId.value);
    return session?.title || t('article.chat.aiChat');
  }
  return t('article.chat.aiChat');
});
</script>

<template>
  <Teleport to="body">
    <Transition name="chat-panel">
      <div
        v-if="isOpen"
        ref="panelElement"
        class="chat-panel fixed bottom-10 right-4 md:bottom-14 md:right-6 w-[500px] h-[600px] bg-bg-primary border border-border rounded-xl shadow-2xl flex flex-col z-50"
        :class="{ 'select-none': isResizing }"
      >
        <!-- Header -->
        <div
          class="flex items-center justify-between p-3 border-b border-border bg-bg-secondary rounded-t-xl relative"
        >
          <div class="flex items-center gap-2 flex-1">
            <PhChatCircleText :size="20" class="text-accent" />
            <button
              class="flex items-center gap-1 text-sm font-medium hover:text-accent transition-colors"
              :title="t('article.chat.switchSession')"
              @click="showSessions = !showSessions"
            >
              <span>{{ currentSessionTitle }}</span>
              <PhClockCounterClockwise :size="16" />
            </button>
          </div>
          <div class="flex items-center gap-1">
            <button
              class="p-1 hover:bg-bg-tertiary rounded-lg transition-colors"
              :title="t('article.chat.newChat')"
              @click="createNewSession"
            >
              <PhPlus :size="18" class="text-text-secondary" />
            </button>
            <button
              class="p-1 hover:bg-bg-tertiary rounded-lg transition-colors"
              :title="t('common.close')"
              @click="emit('close')"
            >
              <PhX :size="18" class="text-text-secondary" />
            </button>
          </div>

          <!-- Resize handle -->
          <div
            class="absolute -top-1 -left-1 w-3 h-3 cursor-nw-resize opacity-0 hover:opacity-100 transition-opacity"
            :class="isResizing ? 'opacity-100' : ''"
            @mousedown="startResize"
          >
            <div class="w-full h-full bg-accent rounded-full border border-white shadow-sm"></div>
          </div>
        </div>

        <!-- Session List Sidebar -->
        <Transition name="slide-in">
          <div
            v-if="showSessions"
            class="absolute top-12 left-0 right-0 bottom-12 bg-bg-secondary border-b border-border rounded-b-xl overflow-y-auto scroll-smooth"
          >
            <div class="p-2 space-y-1">
              <div
                v-for="session in sessions"
                :key="session.id"
                class="flex items-center gap-2 p-2 rounded-lg hover:bg-bg-tertiary cursor-pointer group"
                :class="{ 'bg-bg-tertiary': session.id === currentSessionId }"
                @click="selectSession(session.id)"
              >
                <PhChatCircleText :size="16" class="text-text-secondary" />
                <div v-if="editingSessionId === session.id" class="flex-1 flex items-center gap-1">
                  <input
                    v-model="editingSessionTitle"
                    class="flex-1 px-2 py-1 text-sm bg-bg-primary border border-border rounded focus:outline-none focus:border-accent"
                    @keyup.enter="saveSessionTitle(session.id)"
                    @keyup.esc="cancelEditSession"
                    @click.stop
                  />
                  <button
                    class="p-1 hover:bg-bg-primary rounded"
                    @click="saveSessionTitle(session.id)"
                  >
                    <PhPaperPlaneRight :size="14" />
                  </button>
                </div>
                <span v-else class="flex-1 text-sm truncate">{{ session.title }}</span>
                <div class="flex items-center gap-1 opacity-0 group-hover:opacity-100">
                  <button
                    class="p-1 hover:bg-bg-primary rounded"
                    @click="startEditSession(session, $event)"
                  >
                    <PhPencil :size="14" />
                  </button>
                  <button
                    class="p-1 hover:bg-bg-primary rounded text-red-500"
                    @click="deleteSession(session.id, $event)"
                  >
                    <PhTrash :size="14" />
                  </button>
                </div>
              </div>
              <div
                v-if="sessions.length === 0"
                class="text-center text-text-secondary text-sm py-4"
              >
                {{ t('article.chat.noSessions') }}
              </div>
            </div>
          </div>
        </Transition>

        <!-- Messages -->
        <div ref="chatContainer" class="flex-1 overflow-y-auto p-3 space-y-3 scroll-smooth">
          <div
            v-if="messages.length === 0"
            class="flex items-center justify-center h-full text-text-secondary text-sm"
          >
            {{ t('article.chat.aiChatWelcome') }}
          </div>
          <div
            v-for="(msg, index) in messages"
            :key="index"
            class="flex"
            :class="msg.role === 'user' ? 'justify-end' : 'justify-start'"
          >
            <div
              class="max-w-[80%] rounded-lg px-3 py-2 text-sm select-text cursor-text"
              :class="
                msg.role === 'user' ? 'bg-accent text-white' : 'bg-bg-secondary text-text-primary'
              "
            >
              <!-- Thinking section -->
              <div
                v-if="msg.thinking"
                class="mb-2 p-2 bg-bg-tertiary border-l-2 border-accent rounded text-xs text-text-secondary"
              >
                <div class="font-bold mb-1 flex items-center gap-1">
                  <PhSpinner :size="12" class="animate-spin" />
                  {{ t('article.chat.thinking') }}
                </div>
                <div class="whitespace-pre-wrap">{{ msg.thinking }}</div>
              </div>
              <!-- Message content with pre-rendered HTML from backend -->
              <div
                v-if="msg.role === 'assistant'"
                class="prose prose-sm max-w-none"
                v-html="msg.html || msg.content"
              ></div>
              <div v-else class="whitespace-pre-wrap break-words">{{ msg.content }}</div>
            </div>
          </div>
          <div v-if="isLoading" class="flex justify-start">
            <div class="bg-bg-secondary rounded-lg px-3 py-2 text-sm">
              <PhSpinner :size="16" class="animate-spin" />
            </div>
          </div>
        </div>

        <!-- Input -->
        <div class="p-3 border-t border-border bg-bg-secondary rounded-b-xl">
          <div class="flex gap-2">
            <input
              v-model="inputMessage"
              type="text"
              :placeholder="t('article.chat.aiChatInputPlaceholder')"
              class="flex-1 px-3 py-2 bg-bg-tertiary border border-border rounded-lg text-sm focus:outline-none focus:border-accent"
              :disabled="isLoading"
              @keydown="handleKeydown"
            />
            <button
              :disabled="isLoading || !inputMessage.trim()"
              class="px-3 py-2 bg-accent text-white rounded-lg hover:bg-accent-hover disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
              @click="sendMessage"
            >
              <PhPaperPlaneRight :size="18" />
            </button>
          </div>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<style>
.chat-panel {
  user-select: text !important;
  -webkit-user-select: text !important;
  -moz-user-select: text !important;
  -ms-user-select: text !important;
}

.chat-panel.select-none {
  user-select: none !important;
  -webkit-user-select: none !important;
  -moz-user-select: none !important;
  -ms-user-select: none !important;
}

.chat-panel .select-text {
  user-select: text !important;
  -webkit-user-select: text !important;
  -moz-user-select: text !important;
  -ms-user-select: text !important;
  cursor: text !important;
}

.chat-panel .select-text * {
  user-select: text !important;
  -webkit-user-select: text !important;
  -moz-user-select: text !important;
  -ms-user-select: text !important;
}

.chat-panel-enter-active,
.chat-panel-leave-active {
  transition: all 0.3s ease;
}

.chat-panel-enter-from,
.chat-panel-leave-to {
  opacity: 0;
  transform: translateY(20px) scale(0.95);
}

.chat-panel-enter-to,
.chat-panel-leave-from {
  opacity: 1;
  transform: translateY(0) scale(1);
}

.slide-in-enter-active,
.slide-in-leave-active {
  transition: all 0.2s ease;
}

.slide-in-enter-from,
.slide-in-leave-to {
  opacity: 0;
  transform: translateX(-20px);
}

.slide-in-enter-to,
.slide-in-leave-from {
  opacity: 1;
  transform: translateX(0);
}

/* Markdown prose styles */
.prose {
  color: inherit;
}

.prose pre {
  margin: 0.5rem 0;
  white-space: pre-wrap;
  word-wrap: break-word;
}

.prose code {
  font-family: 'Courier New', Courier, monospace;
}

.prose ul {
  list-style-type: disc;
}

.prose ol {
  list-style-type: decimal;
}
</style>
