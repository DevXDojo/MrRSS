<script setup lang="ts">
import { computed } from 'vue';
import type { XPathPreviewNode } from '@/utils/xpathPicker';
const props = defineProps<{ node: XPathPreviewNode; selected?: string }>();
const emit = defineEmits<{ pick: [node: XPathPreviewNode] }>();
// Source tags/attributes are data. Never bind source URLs, styles, events or HTML.
const safeTag = computed(() => ['p', 'div', 'article', 'section', 'ul', 'ol', 'li', 'h1', 'h2', 'h3', 'h4', 'h5', 'h6', 'span', 'strong', 'em', 'pre', 'blockquote'].includes(props.node.tag ?? '') ? props.node.tag : 'div');
</script>

<template>
  <span v-if="!node.tag">{{ node.text }}<XPathPreviewNode v-for="(child, index) in node.children" :key="index" :node="child" :selected="selected" @pick="emit('pick', $event)" /></span>
  <component :is="safeTag" v-else class="picker-node" :class="{ 'picker-selected': node.path === selected, 'picker-image': node.tag === 'img' }" @click.stop="emit('pick', node)">
    <span v-if="node.tag === 'img'">▧ {{ node.text || 'img' }}</span>
    <template v-else>{{ node.text }}</template>
    <XPathPreviewNode v-for="(child, index) in node.children" :key="index" :node="child" :selected="selected" @pick="emit('pick', $event)" />
  </component>
</template>

<style scoped>
@reference "../../../../style.css";
.picker-node { @apply cursor-crosshair; }
.picker-node:hover { outline: 1px dashed var(--color-accent); }
.picker-selected { outline: 2px solid var(--color-accent); }
li, article, section, p, blockquote { @apply my-2 p-2 border border-border rounded; }
h1, h2, h3, h4, h5, h6 { @apply font-semibold my-2; }
.picker-image { @apply inline-block p-2 bg-bg-tertiary text-text-secondary rounded; }
</style>
