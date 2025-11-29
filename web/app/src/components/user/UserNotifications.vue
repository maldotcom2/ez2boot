<template>
  <div class="notification-form">
    <select v-model="selectedType">
      <option v-for="t in supportedTypes" :key="t.type" :value="t.type">
        {{ t.label }}
      </option>
    </select>
    <component v-if="selectedType" :is="formComponents[selectedType]" v-model="notificationData[selectedType]" @save="saveUserNotification" @delete="deleteUserNotification"/>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import axios from 'axios'
import EmailForm from '../notifications/EmailForm.vue'
import TelegramForm from '../notifications/TelegramForm.vue'

const selectedType = ref('')
const supportedTypes = ref([])
const notificationData = reactive({})
const error = ref('')

const formComponents = {
  email: EmailForm,
  telegram: TelegramForm
}

// Get supported channels
async function getNotificationTypes() {
  error.value = ''  // Reset error
  try {
    const response = await axios.get('ui/notification/types')
    if (response.data.success) {
    supportedTypes.value = response.data.data
      if (supportedTypes.value.length > 0) {
        selectedType.value = supportedTypes.value[0].type // default
      } else {
        error.value = 'Failed to get notification types'
      }
    }
  } catch (err) {
    if (err.response) {
      // Get server response
      error.value = `Failed to get notification types: ${err.response.data.error || err.response.statusText}`
    } else if (err.request) {
      // No response
      error.value = 'No response from server'
    } else {
      // other errors
      error.value = err.message
    }
  }
}

// Populate UI with currently stored user settings
async function loadUserNotification() {
  error.value = ''  // Reset error
  try {
    const response = await axios.get('/ui/user/notification')
    if (response.data.success && response.data.data) {
      const userNotif = response.data.data
      selectedType.value = userNotif.type || selectedType.value
      notificationData[userNotif.type] = userNotif.channel_config || {}
    }
  } catch (err) {
    if (err.response) {
      // Get server response
      error.value = `Failed to load user notification settings: ${err.response.data.error || err.response.statusText}`
    } else if (err.request) {
      // No response
      error.value = 'No response from server'
    } else {
      // other errors
      error.value = err.message
    }
  }
}

// Save new settings
async function saveUserNotification() {
  // TODO
  console.log('Save notifications')
}

// Delete settings - user will have no notifications
async function deleteUserNotification() {
  error.value = ''  // Reset error
  try {
    const response = await axios.delete('/ui/user/notification')
    console.log(response.data.data)
    if (response.data.success) {
      console.log("Notification deleted")
    }
  } catch (err) {
    if (err.response) {
      // Get server response
      error.value = `Failed to delete user notification: ${err.response.data.error || err.response.statusText}`
    } else if (err.request) {
      // No response
      error.value = 'No response from server'
    } else {
      // other errors
      error.value = err.message
    }
  }
}

onMounted(async () => {
  await getNotificationTypes()
  await loadUserNotification()
})
</script>

<style scoped>
.notification-form {
    background-color: var(--container-modal);
    border-radius: var(--small-radius);
}

</style>
