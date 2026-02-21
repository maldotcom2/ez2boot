<template>
  <div class="user-notifications">
    <aside class="sidebar">
      <select class="notification-selector" v-model="selectedType">
        <option v-for="t in supportedTypes" :key="t.type" :value="t.type">
          {{ t.label }}
        </option>
      </select>
      <div class="actions">
        <button type="button" @click="saveUserNotification" :disabled="!isDirty">Save</button>
        <button type="button" @click="deleteUserNotification" :disabled="!canDelete">Delete</button>
      </div>
      <p class="result" :class="messageType">{{ message || '\u00A0' }}</p>
    </aside>
    <main class="config-panel">
      <component v-if="selectedType" :is="formComponents[selectedType]" v-model="notificationData[selectedType]"/>
    </main>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted, watch, computed } from 'vue'
import axios from 'axios'
import EmailForm from './notifications/Email.vue'
import TeamsForm from './notifications/Teams.vue'
import TelegramForm from './notifications/Telegram.vue'

const selectedType = ref('')
const supportedTypes = ref([])
const notificationData = reactive({})
const message = ref('')
const messageType = ref('')
const savedType = ref(null)
const originalData = reactive({})

// For enabling delete button
const canDelete = computed(() => {
  return (
    savedType.value !== null &&
    selectedType.value === savedType.value
  )
})

// Track changes to saved notifications
const isDirty = computed(() => {
  if (!selectedType.value) return false
  const current = notificationData[selectedType.value] || {}
  const original = originalData[selectedType.value] || {}
  return JSON.stringify(current) !== JSON.stringify(original)
})



// Gracefully handle case of no user notification settings from load, pass valid object to child
watch(selectedType, (newType) => {
  if (newType && !notificationData[newType]) {
    notificationData[newType] = {}
  }

  // Clear message when notification channel is changed
  message.value = ''
  messageType.value = ''
})

const formComponents = {
  email: EmailForm,
  teams: TeamsForm,
  telegram: TelegramForm
}

// Get supported channels
async function getNotificationTypes() {
  message.value = ''
  messageType.value = ''
  try {
    const response = await axios.get('ui/notification/types')
    if (response.data.success) {
    supportedTypes.value = response.data.data
      if (supportedTypes.value.length > 0) {
        selectedType.value = supportedTypes.value[0].type // default
      } else {
        message.value = 'Failed to get notification types'
      }
    }
  } catch (err) {
    messageType.value = 'error'
    if (err.response) {
      // Get server response
      message.value = `Failed to get notification types: ${err.response.data.error || err.response.statusText}`
    } else if (err.request) {
      // No response
      message.value = 'No response from server'
    } else {
      // other errors
      message.value = err.message
    }
  }
}

// Populate UI with currently stored user settings
async function loadUserNotification() {
  message.value = ''
  messageType.value = ''
  try {
    const response = await axios.get('/ui/user/notification')
    if (response.data.success && response.data.data) {
      const userNotif = response.data.data
      selectedType.value = userNotif.type || selectedType.value

      // Populates fields with existing config if available
      notificationData[userNotif.type] = userNotif.channel_config || {}

      // Snapshot data
      originalData[userNotif.type] = JSON.parse(JSON.stringify(notificationData[userNotif.type]))
      savedType.value = userNotif.type
    } else {
      savedType.value = null
    }
  } catch (err) {
    savedType.value = null
    messageType.value = 'error'
    if (err.response) {
      // Get server response
      message.value = `Failed to load settings: ${err.response.data.error || err.response.statusText}`
    } else if (err.request) {
      // No response
      message.value = 'No response from server'
    } else {
      // other errors
      message.value = err.message
    }
  }
}

// Save new settings
async function saveUserNotification() {
  message.value = ''
  messageType.value = ''
  try {
    // Build the payload
    const payload = {
      type: selectedType.value,
      channel_config: notificationData[selectedType.value] || {}
    }

    const response = await axios.post('/ui/user/notification', payload)
    if (response.data.success) {
      console.log("Notification settings saved")
      message.value = 'Notification settings saved'
      messageType.value = 'success'
      savedType.value = selectedType.value

      // Snapshot data
      originalData[selectedType.value] = JSON.parse(JSON.stringify(notificationData[selectedType.value]))

      // Purge values of non-selected notification forms
      Object.keys(notificationData).forEach((type) => {
        if (type !== selectedType.value) {
          delete notificationData[type]
          delete originalData[type]
        }
      })
    }

  } catch (err) {
    messageType.value = 'error'
    if (err.response) {
      message.value = `Failed to save settings: ${err.response.data.error || err.response.statusText}`
    } else if (err.request) {
      message.value = 'No response from server'
    } else {
      message.value = err.message
    }
  }
}

// Delete settings - user will have no notifications
async function deleteUserNotification() {
  message.value = ''
  messageType.value = ''
  if (!confirm("Are you sure you want to delete notification settings?")) {
    return
  }

  try {
    const response = await axios.delete('/ui/user/notification')
    if (response.data.success) {
      notificationData[selectedType.value] = {}
      message.value = 'Notification settings deleted'
      messageType.value = 'success'
      savedType.value = null

      // Snapshot data
      originalData[selectedType.value] = JSON.parse(JSON.stringify(notificationData[selectedType.value]))
    }
  } catch (err) {
    messageType.value = 'error'
    if (err.response) {
      // Get server response
      message.value = `Failed to delete settings: ${err.response.data.error || err.response.statusText}`
    } else if (err.request) {
      // No response
      message.value = 'No response from server'
    } else {
      // other errors
      message.value = err.message
    }
  }
}

onMounted(async () => {
  await getNotificationTypes()
  await loadUserNotification()
})
</script>

<style scoped>
.user-notifications {
  flex: 1;
  display: grid;
  grid-template-columns: 250px 1fr;
  padding: 1rem;
  gap: 1rem;
  width: 100%;
  background-color: var(--container-modal);
  border-radius: var(--small-radius);
  height: 100%;
}

.notification-selector {
  background-color: var(--low-glare);
  outline: none;
}

.sidebar {
  display: flex;
  flex-direction: column;
  border-radius: var(--small-radius);
  padding: 1rem;
  background: var(--container-modal);
  gap: 1rem;
}

.config-panel {
  display: flex;
  flex-direction: column;
  padding: 1rem;
  box-sizing: border-box;
  color: var(--low-glare);
  background: var(--container-modal);
  border-radius: var(--small-radius);
  overflow: auto;
  height: 100%;
}

.actions {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.actions button {
  display: block;
  width: 100%;
}

select {
  border-radius: var(--small-radius);
}

.result {
  min-height: 1.2rem;
  font-size: 1rem;
  text-align: center;
}

.result.error {
  color: var(--error-msg);
}

.result.success {
  color: var(--success-msg);
}


</style>
