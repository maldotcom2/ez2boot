<template>
  <aside class="menu">
    <button @click="showUserManagement">User Management</button>
  </aside>
</template>

<script setup>
import { defineEmits } from 'vue'

// Props
const props = defineProps({
  modelValue: Object
})

// Emits for v-model
const emit = defineEmits(['update:modelValue'])

import UserManagement from './AdminUserMgmt.vue'

// Update parent when button clicked
function showUserManagement() {
  emit('update:modelValue', UserManagement)
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
