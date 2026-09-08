<script setup lang="ts">
/* eslint-disable vue/no-v-html */
import { ref, nextTick, computed, onMounted, onBeforeUnmount, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import {
  PhChatCircleText,
  PhCheck,
  PhX,
  PhPaperPlaneRight,
  PhStop,
  PhSpinner,
  PhClockCounterClockwise,
  PhPlus,
  PhTrash,
  PhPencil,
  PhCopy,
  PhGear,
} from '@phosphor-icons/vue';
import type { Article } from '@/types/models';
import { readAIError } from '@/utils/aiError';
import { copyToClipboard } from '@/utils/clipboard';
import { useAIProfiles } from '@/composables/ai/useAIProfiles';
import BaseSelect from '@/components/common/BaseSelect.vue';

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
  settings: {
    ai_chat_enabled: boolean;
    ai_chat_profile_id: string;
    ai_chat_quick_prompts: string;
  };
}

const props = defineProps<Props>();

const emit = defineEmits<{
  close: [];
}>();

const { t } = useI18n();
const { profiles, defaultProfile, fetchProfiles } = useAIProfiles();

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
const selectedProfileId = ref(props.settings.ai_chat_profile_id || '');
const boundArticle = ref<Article>({ ...props.article });
const boundArticleContent = ref(props.articleContent);
const rebindSession = ref(false);
const articleMismatch = computed(() => props.article.id !== boundArticle.value.id);

watch(
  () => props.articleContent,
  (content) => {
    if (!articleMismatch.value) boundArticleContent.value = content;
  }
);

watch([showSessions, currentSessionId, () => props.article.id], cancelEditSession);

interface ActiveChatRequest {
  id: string;
  articleId: number;
  draftVersion: number;
  sessionId: number | null;
  controller: AbortController;
  stopped: boolean;
}
let activeRequest: ActiveChatRequest | null = null;
let creatingSession: {
  articleId: number;
  draftVersion: number;
  promise: Promise<ChatSession>;
} | null = null;
let draftVersion = 0;
let disposed = false;
let viewVersion = 0;

function isCurrentRequest(run: ActiveChatRequest) {
  return (
    !disposed &&
    !run.stopped &&
    activeRequest === run &&
    boundArticle.value.id === run.articleId &&
    draftVersion === run.draftVersion
  );
}

async function cancelRequest(run: ActiveChatRequest) {
  if (!run.sessionId) return;
  try {
    await fetch('/api/ai-chat/cancel', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ session_id: run.sessionId, request_id: run.id }),
      keepalive: true,
    });
  } catch (error) {
    console.error('Failed to cancel chat generation:', error);
  }
}

async function stopGeneration() {
  const run = activeRequest;
  if (!run) return;
  run.stopped = true;
  activeRequest = null;
  isLoading.value = false;
  const version = ++viewVersion;
  run.controller.abort();
  await cancelRequest(run);
  const stillStopped = () =>
    !disposed &&
    viewVersion === version &&
    !activeRequest &&
    boundArticle.value.id === run.articleId &&
    draftVersion === run.draftVersion;
  if (run.sessionId && stillStopped()) {
    await loadSessions(stillStopped);
    if (stillStopped()) await selectSession(run.sessionId, true, stillStopped);
  }
}

function closePanel() {
  disposed = true;
  void stopGeneration();
  emit('close');
}

onBeforeUnmount(() => {
  disposed = true;
  void stopGeneration();
  stopResize();
});

async function ensureSession(
  articleId: number,
  title: string,
  draft: number
): Promise<ChatSession> {
  if (
    !creatingSession ||
    creatingSession.articleId !== articleId ||
    creatingSession.draftVersion !== draft
  ) {
    const promise = (async () => {
      const response = await fetch('/api/ai/chat/session/create', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ article_id: articleId, title }),
      });
      if (!response.ok) throw new Error('Failed to create chat session');
      const session = (await response.json()) as ChatSession;
      if (
        !Number.isSafeInteger(session.id) ||
        session.id <= 0 ||
        session.article_id !== articleId
      ) {
        throw new Error('Chat session ID is missing');
      }
      return session;
    })();
    creatingSession = { articleId, draftVersion: draft, promise };
  }
  const pending = creatingSession;
  try {
    return await pending.promise;
  } catch (error) {
    if (creatingSession === pending) creatingSession = null;
    throw error;
  }
}

const profileOptions = computed(() =>
  profiles.value.map((profile) => ({ value: String(profile.id), label: profile.name }))
);

const summaryPrompts = computed(() => [
  t('article.chat.promptConciseSummary'),
  t('article.chat.promptKeyPoints'),
  t('article.chat.promptDetailedSummary'),
]);

const questionPrompts = computed(() => [
  t('article.chat.promptMainContent'),
  t('article.chat.promptKeyPeople'),
  t('article.chat.promptMainViews'),
  t('article.chat.promptKeyInformation'),
  t('article.chat.promptExplain'),
  t('article.chat.promptAnalyze'),
  t('article.chat.promptVerify'),
]);

const customPrompts = computed<string[]>(() => {
  try {
    const parsed = JSON.parse(props.settings.ai_chat_quick_prompts || '[]');
    return Array.isArray(parsed)
      ? parsed.filter((item): item is string => typeof item === 'string')
      : [];
  } catch {
    return [];
  }
});

// Resize functionality
const isResizing = ref(false);
const startX = ref(0);
const startY = ref(0);
const startWidth = ref(0);
const startHeight = ref(0);
const panelElement = ref<HTMLElement | null>(null);
const panelSizeStorageKey = 'mrrssChatPanelSize';
let preferredPanelSize = { width: 500, height: 600 };

function applyPanelSize() {
  const panel = panelElement.value;
  if (!panel) return;
  const desktop = window.innerWidth >= 768;
  const availableWidth = Math.max(1, window.innerWidth - (desktop ? 24 : 16) - 16);
  const availableHeight = Math.max(1, window.innerHeight - (desktop ? 56 : 40) - 16);
  panel.style.width = `${Math.min(preferredPanelSize.width, availableWidth)}px`;
  panel.style.height = `${Math.min(preferredPanelSize.height, availableHeight)}px`;
}

onMounted(() => {
  try {
    const saved = JSON.parse(localStorage.getItem(panelSizeStorageKey) || 'null');
    if (
      saved &&
      Number.isFinite(saved.width) &&
      Number.isFinite(saved.height) &&
      saved.width >= 300 &&
      saved.height >= 200
    ) {
      preferredPanelSize = { width: Math.max(420, saved.width), height: saved.height };
    }
  } catch {
    // Invalid or unavailable local storage should not prevent opening chat.
  }
  applyPanelSize();
  window.addEventListener('resize', applyPanelSize);
});

onBeforeUnmount(() => {
  stopResize();
  window.removeEventListener('resize', applyPanelSize);
});

// Initialize: load sessions for this article
onMounted(async () => {
  const version = viewVersion;
  const isCurrentView = () => !disposed && version === viewVersion && !activeRequest;
  await fetchProfiles();
  if (!isCurrentView()) return;
  if (!selectedProfileId.value && defaultProfile.value) {
    selectedProfileId.value = String(defaultProfile.value.id);
  }
  await loadSessions(isCurrentView);
  // Auto-select the most recent session if available
  if (isCurrentView() && sessions.value.length > 0) {
    await selectSession(sessions.value[0].id, false, isCurrentView);
  }
});

async function loadSessions(isCurrentView = () => !disposed) {
  try {
    const articleId = boundArticle.value.id;
    const response = await fetch(`/api/ai/chat/sessions?article_id=${articleId}`);
    if (response.ok) {
      const loadedSessions: ChatSession[] = await response.json();
      if (isCurrentView() && boundArticle.value.id === articleId) {
        sessions.value = loadedSessions.filter((session) => session.article_id === articleId);
      }
    }
  } catch (e) {
    console.error('Failed to load sessions:', e);
  }
}

async function selectSession(sessionId: number, force = false, isCurrentView?: () => boolean) {
  if (isLoading.value && !force) return;
  if (!isCurrentView) {
    const version = ++viewVersion;
    isCurrentView = () => !disposed && version === viewVersion && !activeRequest;
  }
  const session = sessions.value.find((item) => item.id === sessionId);
  if (!session || session.article_id !== boundArticle.value.id) return;
  try {
    const response = await fetch(`/api/ai/chat/messages?session_id=${sessionId}`);
    if (response.ok) {
      const loadedMessages = await response.json();
      if (!isCurrentView() || session.article_id !== boundArticle.value.id) return;
      messages.value = loadedMessages;
      currentSessionId.value = sessionId;
      rebindSession.value = false;
      // A cancelled first attempt may have saved only the user question.
      // Keep article context until an assistant response has been saved.
      isFirstMessage.value = !loadedMessages.some(
        (message: ChatMessage) => message.role === 'assistant'
      );
      showSessions.value = false;
      await nextTick();
      scrollToBottom();
    }
  } catch (e) {
    console.error('Failed to load session messages:', e);
  }
}

function createNewSession() {
  if (isLoading.value) return;
  ++viewVersion;
  ++draftVersion;
  creatingSession = null;
  const changedArticle = boundArticle.value.id !== props.article.id;
  boundArticle.value = { ...props.article };
  boundArticleContent.value = props.articleContent;
  if (changedArticle) sessions.value = [];
  currentSessionId.value = null;
  rebindSession.value = false;
  messages.value = [];
  inputMessage.value = '';
  isFirstMessage.value = true;
  showSessions.value = false;
  cancelEditSession();
  if (changedArticle) void loadSessions();
}

function continueWithCurrentArticle() {
  if (isLoading.value || !articleMismatch.value) return;
  ++viewVersion;
  ++draftVersion;
  creatingSession = null;
  boundArticle.value = { ...props.article };
  boundArticleContent.value = props.articleContent;
  isFirstMessage.value = true;
  rebindSession.value = true;
  const currentSession = sessions.value.find((session) => session.id === currentSessionId.value);
  if (currentSession) {
    currentSession.article_id = props.article.id;
    sessions.value = [currentSession];
  } else {
    sessions.value = [];
  }
  showSessions.value = false;
  cancelEditSession();
}

async function deleteSession(sessionId: number, e: Event) {
  e.stopPropagation();
  if (isLoading.value) return;
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
  if (isLoading.value) return;
  editingSessionId.value = session.id;
  editingSessionTitle.value = session.title;
}

async function saveSessionTitle(sessionId: number) {
  if (editingSessionId.value !== sessionId) return;
  const title = editingSessionTitle.value;
  try {
    const response = await fetch(`/api/ai/chat/session?session_id=${sessionId}`, {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ title }),
    });

    if (!response.ok) throw new Error(`Failed to update session title: ${response.status}`);
    const session = sessions.value.find((s) => s.id === sessionId);
    if (session) {
      session.title = title;
    }
    if (editingSessionId.value === sessionId && editingSessionTitle.value === title) {
      cancelEditSession();
    }
  } catch (e) {
    console.error('Failed to update session title:', e);
    window.showToast(t('article.chat.titleSaveFailed'), 'error');
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

  const newWidth = Math.max(420, startWidth.value + deltaX);
  const newHeight = Math.max(200, startHeight.value + deltaY);

  const panel = panelElement.value;
  if (panel) {
    preferredPanelSize = { width: newWidth, height: newHeight };
    applyPanelSize();
  }
}

function stopResize() {
  if (isResizing.value) {
    try {
      localStorage.setItem(panelSizeStorageKey, JSON.stringify(preferredPanelSize));
    } catch {
      // Resizing remains available when local storage cannot be written.
    }
  }
  isResizing.value = false;
  document.removeEventListener('mousemove', resize);
  document.removeEventListener('mouseup', stopResize);
}

async function sendMessage() {
  const message = inputMessage.value.trim();
  if (!message || disposed || isLoading.value || showSessions.value || articleMismatch.value)
    return;
  const run: ActiveChatRequest = {
    id: globalThis.crypto?.randomUUID?.() || `${Date.now()}-${Math.random().toString(36).slice(2)}`,
    articleId: boundArticle.value.id,
    draftVersion,
    sessionId: currentSessionId.value,
    controller: new AbortController(),
    stopped: false,
  };
  activeRequest = run;
  ++viewVersion;
  const article = { ...boundArticle.value };
  const articleContent = boundArticleContent.value.slice(0, 50000);
  isLoading.value = true;

  try {
    // A short, single-flight create gives Stop a stable session ID before the
    // long provider request starts. Do not abort creation and lose its ID.
    if (!run.sessionId) {
      const session = await ensureSession(
        run.articleId,
        Array.from(message).slice(0, 60).join(''),
        run.draftVersion
      );
      run.sessionId = session.id;
      if (!isCurrentRequest(run)) return;
      // Consume the cached ID only when a live request takes ownership. A
      // stopped create can finish before the user retries without duplicating it.
      creatingSession = null;
      currentSessionId.value = session.id;
      sessions.value.unshift(session);
    }
    if (!isCurrentRequest(run)) return;
    // Preserve the input if session creation fails or is stopped before sending.
    messages.value.push({
      id: 0,
      role: 'user',
      content: message,
      created_at: new Date().toISOString(),
    });
    inputMessage.value = '';
    await nextTick();
    if (!isCurrentRequest(run)) return;
    scrollToBottom();
    const response = await fetch('/api/ai-chat', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      signal: run.controller.signal,
      body: JSON.stringify({
        request_id: run.id,
        session_id: run.sessionId,
        article_id: run.articleId,
        messages: messages.value.slice(-10),
        is_first_message: isFirstMessage.value,
        article_title: article.title,
        article_url: article.url,
        article_content: articleContent,
        profile_id: Number(selectedProfileId.value) || undefined,
        rebind_session: rebindSession.value,
      }),
    });
    if (!isCurrentRequest(run)) return;
    if (response.ok) {
      const data = await response.json();
      if (!isCurrentRequest(run)) return;
      messages.value.push({
        id: 0,
        role: 'assistant',
        content: data.response,
        html: data.html,
        thinking: data.thinking,
        created_at: new Date().toISOString(),
      });
      isFirstMessage.value = false;
      rebindSession.value = false;
      if (data.session_id) {
        currentSessionId.value = data.session_id;
        await loadSessions(() => isCurrentRequest(run));
      }
      if (isCurrentRequest(run) && data.history_saved === false) {
        window.showToast(t('article.chat.historySaveFailed'), 'warning');
      }
    } else {
      const chatError = await readAIError(response);
      if (!isCurrentRequest(run)) return;
      const errorPayload =
        chatError.payload !== null && typeof chatError.payload === 'object'
          ? (chatError.payload as Record<string, unknown>)
          : null;
      const persistedSessionID = Number(errorPayload?.session_id || 0);
      if (persistedSessionID > 0) {
        currentSessionId.value = persistedSessionID;
        await loadSessions(() => isCurrentRequest(run));
        if (!isCurrentRequest(run)) return;
        await selectSession(persistedSessionID, true, () => isCurrentRequest(run));
      }
      if (isCurrentRequest(run)) window.showToast(chatError.message, 'error');
    }
  } catch (error) {
    if (isCurrentRequest(run)) {
      console.error('AI chat error:', error);
      window.showToast(t('article.chat.aiChatError'), 'error');
    }
  } finally {
    if (isCurrentRequest(run)) {
      activeRequest = null;
      isLoading.value = false;
      await nextTick();
      scrollToBottom();
    }
  }
}

async function sendSuggestedPrompt(prompt: string) {
  if (isLoading.value || showSessions.value || articleMismatch.value) return;
  inputMessage.value = prompt;
  await sendMessage();
}

async function copyMessage(content: string) {
  const copied = await copyToClipboard(content);
  window.showToast(
    copied ? t('common.toast.copiedToClipboard') : t('common.errors.failedToCopy'),
    copied ? 'success' : 'error'
  );
}

function openAISettings() {
  window.dispatchEvent(new CustomEvent('show-settings', { detail: { tab: 'ai' } }));
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
        class="chat-panel fixed bottom-10 right-4 md:bottom-14 md:right-6 w-[500px] h-[600px] bg-bg-primary text-text-primary border border-border rounded-xl shadow-2xl grid grid-cols-1 grid-rows-[auto_minmax(0,auto)_minmax(0,1fr)_auto] z-50"
        :class="{ 'select-none': isResizing }"
      >
        <!-- Header -->
        <div
          class="flex items-center justify-between p-3 border-b border-border bg-bg-secondary rounded-t-xl relative"
        >
          <div class="flex min-w-0 items-center gap-2 flex-1">
            <PhChatCircleText :size="20" class="shrink-0 text-accent" />
            <button
              class="flex min-w-0 items-center gap-1 text-sm font-medium hover:text-accent transition-colors"
              :disabled="isLoading"
              :title="t('article.chat.switchSession')"
              data-testid="chat-session-switcher"
              @click.stop="showSessions = !showSessions"
            >
              <span class="truncate">{{ currentSessionTitle }}</span>
              <PhClockCounterClockwise :size="16" class="shrink-0" />
            </button>
          </div>
          <div class="flex shrink-0 items-center gap-1">
            <BaseSelect
              v-if="profileOptions.length > 0"
              v-model="selectedProfileId"
              class="chat-profile-selector"
              :options="profileOptions"
              width="w-28 sm:w-36"
              size="xs"
              :disabled="isLoading"
              :title="t('article.chat.selectProfile')"
            />
            <button
              class="p-1 hover:bg-bg-tertiary rounded-lg transition-colors"
              :title="t('article.chat.openAISettings')"
              @click.stop="openAISettings"
            >
              <PhGear :size="18" class="text-text-secondary" />
            </button>
            <button
              class="p-1 hover:bg-bg-tertiary rounded-lg transition-colors"
              :disabled="isLoading"
              :title="t('article.chat.newChat')"
              data-testid="chat-new-session"
              @click.stop="createNewSession"
            >
              <PhPlus :size="18" class="text-text-secondary" />
            </button>
            <button
              class="p-1 hover:bg-bg-tertiary rounded-lg transition-colors"
              :title="t('common.close')"
              @click="closePanel"
              data-testid="chat-close"
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

        <div
          class="min-h-0 max-h-40 overflow-y-auto border-b border-border px-3 py-2"
          data-testid="chat-context-article"
          :data-context-article-id="boundArticle.id"
        >
          <p class="text-xs text-text-secondary">{{ t('article.chat.linkedArticle') }}</p>
          <p class="line-clamp-2 text-sm font-medium" :title="boundArticle.title">
            {{ boundArticle.title }}
          </p>
          <p class="truncate text-xs text-text-secondary">
            {{ boundArticle.feed_title || boundArticle.feed_name || boundArticle.url }}
          </p>
          <div v-if="articleMismatch" class="mt-2 space-y-2" role="status">
            <p class="text-xs text-text-secondary">
              {{ t('article.chat.articleMismatch', { title: boundArticle.title }) }}
            </p>
            <div class="flex flex-wrap gap-x-3 gap-y-1">
              <button
                type="button"
                class="text-xs text-accent hover:underline disabled:opacity-50"
                data-testid="chat-continue-current-article"
                :disabled="isLoading"
                @click.stop="continueWithCurrentArticle"
              >
                {{ t('article.chat.continueWithCurrentArticle') }}
              </button>
              <button
                type="button"
                class="text-xs text-text-secondary hover:text-accent hover:underline disabled:opacity-50"
                data-testid="chat-new-context"
                :disabled="isLoading"
                @click.stop="createNewSession"
              >
                {{ t('article.chat.newChatForCurrentArticle') }}
              </button>
            </div>
          </div>
        </div>

        <!-- Session List Sidebar -->
        <Transition name="slide-in">
          <div
            v-if="showSessions"
            class="col-start-1 row-start-3 z-10 min-h-0 bg-bg-secondary border-b border-border rounded-b-xl overflow-y-auto scroll-smooth"
          >
            <div class="p-2 space-y-1">
              <div
                v-for="session in sessions"
                :key="session.id"
                class="flex items-center gap-2 p-2 rounded-lg hover:bg-bg-tertiary cursor-pointer group"
                :class="{
                  'bg-bg-tertiary': session.id === currentSessionId,
                  'pointer-events-none opacity-60': isLoading,
                }"
                :data-session-id="session.id"
                @click.stop="selectSession(session.id)"
              >
                <PhChatCircleText :size="16" class="text-text-secondary" />
                <div
                  v-if="editingSessionId === session.id"
                  class="flex-1 min-w-0 flex items-center gap-1"
                >
                  <input
                    v-model="editingSessionTitle"
                    class="flex-1 min-w-0 px-2 py-1 text-sm bg-bg-primary border border-border rounded focus:outline-none focus:border-accent"
                    @keyup.enter="saveSessionTitle(session.id)"
                    @keyup.esc="cancelEditSession"
                    @click.stop
                  />
                  <button
                    class="p-1 hover:bg-bg-primary rounded"
                    :title="t('common.save')"
                    :aria-label="t('common.save')"
                    @click.stop="saveSessionTitle(session.id)"
                  >
                    <PhCheck :size="14" />
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
        <div
          ref="chatContainer"
          class="col-start-1 row-start-3 min-h-0 overflow-y-auto p-3 space-y-3 scroll-smooth"
          :class="{ invisible: showSessions }"
        >
          <div v-if="messages.length === 0" class="space-y-4 py-2 text-sm">
            <p class="text-center text-text-secondary">{{ t('article.chat.aiChatWelcome') }}</p>

            <section class="space-y-2">
              <h3 class="text-xs font-semibold uppercase tracking-wide text-text-tertiary">
                {{ t('article.chat.summarySuggestions') }}
              </h3>
              <div class="grid gap-2 sm:grid-cols-3">
                <button
                  v-for="prompt in summaryPrompts"
                  :key="prompt"
                  type="button"
                  class="cursor-pointer rounded-lg border border-border bg-bg-secondary px-3 py-2 text-left text-text-primary transition-colors hover:border-accent hover:bg-bg-tertiary"
                  :disabled="isLoading || showSessions || articleMismatch"
                  @click="sendSuggestedPrompt(prompt)"
                >
                  {{ prompt }}
                </button>
              </div>
            </section>

            <section class="space-y-2">
              <h3 class="text-xs font-semibold uppercase tracking-wide text-text-tertiary">
                {{ t('article.chat.questionSuggestions') }}
              </h3>
              <div class="grid gap-2 sm:grid-cols-2">
                <button
                  v-for="prompt in questionPrompts"
                  :key="prompt"
                  type="button"
                  class="cursor-pointer rounded-lg border border-border bg-bg-secondary px-3 py-2 text-left text-text-primary transition-colors hover:border-accent hover:bg-bg-tertiary"
                  :disabled="isLoading || showSessions || articleMismatch"
                  @click="sendSuggestedPrompt(prompt)"
                >
                  {{ prompt }}
                </button>
              </div>
            </section>

            <section v-if="customPrompts.length > 0" class="space-y-2">
              <h3 class="text-xs font-semibold uppercase tracking-wide text-text-tertiary">
                {{ t('article.chat.customSuggestions') }}
              </h3>
              <div class="grid gap-2 sm:grid-cols-2">
                <button
                  v-for="prompt in customPrompts"
                  :key="prompt"
                  type="button"
                  class="cursor-pointer rounded-lg border border-border bg-bg-secondary px-3 py-2 text-left text-text-primary transition-colors hover:border-accent hover:bg-bg-tertiary"
                  :disabled="isLoading || showSessions || articleMismatch"
                  @click="sendSuggestedPrompt(prompt)"
                >
                  {{ prompt }}
                </button>
              </div>
            </section>
          </div>
          <div
            v-for="(msg, index) in messages"
            :key="index"
            class="flex group"
            :class="msg.role === 'user' ? 'justify-end' : 'justify-start'"
          >
            <div
              class="flex items-start gap-1"
              :class="msg.role === 'user' ? 'flex-row-reverse' : 'w-full min-w-0'"
            >
              <div
                class="py-2 text-sm select-text cursor-text"
                :class="
                  msg.role === 'user'
                    ? 'max-w-[80%] rounded-lg px-3 bg-accent text-white'
                    : 'min-w-0 flex-1 text-text-primary'
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
                  v-if="msg.role === 'assistant' && msg.html"
                  class="prose prose-sm max-w-none"
                  v-html="msg.html"
                ></div>
                <div v-else class="whitespace-pre-wrap break-words">{{ msg.content }}</div>
              </div>
              <button
                class="shrink-0 p-1 rounded text-text-secondary opacity-0 group-hover:opacity-100 hover:bg-bg-tertiary hover:text-text-primary transition-all"
                :title="t('article.chat.copyMessage')"
                @click="copyMessage(msg.content)"
              >
                <PhCopy :size="14" />
              </button>
            </div>
          </div>
          <div v-if="isLoading" class="flex justify-start">
            <div class="bg-bg-secondary rounded-lg px-3 py-2 text-sm">
              <PhSpinner :size="16" class="animate-spin" />
            </div>
          </div>
        </div>

        <!-- Input -->
        <div class="row-start-4 p-3 border-t border-border bg-bg-secondary rounded-b-xl">
          <div class="flex gap-2">
            <input
              v-model="inputMessage"
              type="text"
              :placeholder="t('article.chat.aiChatInputPlaceholder')"
              class="flex-1 px-3 py-2 bg-bg-tertiary border border-border rounded-lg text-sm focus:outline-none focus:border-accent"
              :disabled="isLoading || showSessions || articleMismatch"
              @keydown="handleKeydown"
            />
            <button
              v-if="isLoading"
              :title="t('article.chat.stopGenerating')"
              :aria-label="t('article.chat.stopGenerating')"
              data-testid="chat-stop-generation"
              class="px-3 py-2 bg-accent text-white rounded-lg hover:bg-accent-hover transition-colors"
              @click="stopGeneration"
            >
              <PhStop :size="18" weight="fill" />
            </button>
            <button
              v-else
              data-testid="chat-send-message"
              :disabled="showSessions || articleMismatch || !inputMessage.trim()"
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
  min-width: min(420px, calc(100vw - 2rem));
  max-width: calc(100vw - 2rem);
  max-height: calc(100vh - 3.5rem);
  user-select: text !important;
  -webkit-user-select: text !important;
  -moz-user-select: text !important;
  -ms-user-select: text !important;
}

@media (min-width: 768px) {
  .chat-panel {
    min-width: min(420px, calc(100vw - 2.5rem));
    max-width: calc(100vw - 2.5rem);
    max-height: calc(100vh - 4.5rem);
  }
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

.chat-panel .chat-profile-selector,
.chat-panel .chat-profile-selector * {
  user-select: none !important;
  -webkit-user-select: none !important;
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
