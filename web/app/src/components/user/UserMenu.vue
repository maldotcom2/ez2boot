<template>
  <aside class="menu">
    <button @click="showUserNotifications">Notifications</button>
    <button @click="showChangePassword">Change Password</button>
    <button @click="showMFA">MFA</button>
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
import UserChangePassword from './UserChangePassword.vue'
import UserMFA from './UserMFA.vue'

// Update parent when button clicked
function showUserNotifications() {
  emit('update:modelValue', UserNotifications)
}

function showChangePassword() {
  emit('update:modelValue', UserChangePassword)
}

function showMFA() {
  emit('update:modelValue', UserMFA)
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
