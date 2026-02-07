<template>
  <div class="user-notifications">
    <aside class="sidebar">
      <select class="notification-selector" v-model="selectedType">
        <option v-for="t in supportedTypes" :key="t.type" :value="t.type">
          {{ t.label }}
        </option>
      </select>
      <div class="actions">
        <button type="button" @click="saveUserNotification">Save</button>
        <button type="button" @click="deleteUserNotification">Delete</button>
      </div>
    </aside>
    <main class="config-panel">
      <component v-if="selectedType" :is="formComponents[selectedType]" v-model="notificationData[selectedType]"/>
    </main>
  </div>
</template>


<script setup>
import { ref, reactive, onMounted, watch } from 'vue'
import axios from 'axios'
import EmailForm from './notifications/Email.vue'
import TelegramForm from './notifications/Telegram.vue'

const selectedType = ref('')
const supportedTypes = ref([])
const notificationData = reactive({})
const error = ref('')

// Gracefully handle case of no user notification settings from load, pass valid object to child
watch(selectedType, (newType) => {
  if (newType && !notificationData[newType]) {
    notificationData[newType] = {}
  }
})

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
      // Populates fields with existing config if available
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
  error.value = ''
  try {
    // Build the payload
    const payload = {
      type: selectedType.value,
      channel_config: notificationData[selectedType.value] || {}
    }

    const response = await axios.post('/ui/user/notification', payload)
    if (response.data.success) {
    console.log("Notification settings saved")
    }
  } catch (err) {
    if (err.response) {
      error.value = `Failed to save: ${err.response.data.error || err.response.statusText}`
    } else if (err.request) {
      error.value = 'No response from server'
    } else {
      error.value = err.message
    }
  }
}

// Delete settings - user will have no notifications
async function deleteUserNotification() {
  if (!confirm("Are you sure you want to delete this notification?")) {
    return
  }

  error.value = ''  // Reset error
  try {
    const response = await axios.delete('/ui/user/notification')
    if (response.data.success) {
      notificationData[selectedType.value] = {}
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
.user-notifications {
  flex: 1;
  display: grid;
  grid-template-columns: 250px 1fr;
  gap: 1rem;
  width: 100%;
  background-color: var(--container-modal);
  border-radius: var(--small-radius);
  height: 100%;
}

.notification-selector {
  background-color: var(--low-glare);
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
  padding: 1rem; /* instead of 40px */
  box-sizing: border-box;
  color: var(--low-glare);
  background: var(--container-modal);
  border-radius: var(--small-radius);
  overflow: auto;
  height: 100%;
}

.actions button {
  display: block;
  width: 100%;
  margin-bottom: 0.5rem;
}

select {
  border-radius: var(--small-radius);
}

</style>
