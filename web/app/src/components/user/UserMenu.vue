<template>
  <aside class="menu">
    <button @click="showUserNotifications">Notifications</button>
  </aside>
</template>

<script setup>

// Props
const props = defineProps({
  modelValue: Object
})

// Emits for v-model
const emit = defineEmits(['update:modelValue'])

import UserNotifications from './UserNotifications.vue'

// Update parent when button clicked
function showUserNotifications() {
  emit('update:modelValue', UserNotifications)
}
</script>

<style scoped>
.menu {
  width: 220px;
  background-color: var(--container-modal);
  color: var(--low-glare);
  display: flex;
  flex-direction: column;
  gap: 1rem;
  padding: 1rem;
}
</style>
