<script setup lang="ts">
import FeedFormModal from './FeedFormModal.vue';

const emit = defineEmits<{
  close: [];
  added: [feedId?: number];
}>();

function handleClose() {
  emit('close');
}

function handleAdded(feedId?: number) {
  emit('added', feedId);
}
</script>

<template>
  <FeedFormModal mode="add" @close="handleClose" @added="handleAdded" />
</template>
