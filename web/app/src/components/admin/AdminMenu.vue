<template>
  <aside class="menu">
    <button @click="showUserManagement">User Management</button>
    <button @click="showAuditLog">Audit Log</button>
  </aside>
</template>

<script setup>

// Props
const props = defineProps({
  modelValue: Object
})

// Emits for v-model
const emit = defineEmits(['update:modelValue'])

import UserManagement from './AdminUserMgmt.vue'
import AuditLog from './AdminAuditLog.vue'

// Update parent when button clicked
function showUserManagement() {
  emit('update:modelValue', UserManagement)
}

function showAuditLog() {
  emit('update:modelValue', AuditLog)
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
