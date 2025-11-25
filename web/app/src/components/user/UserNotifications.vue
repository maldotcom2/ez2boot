<template>
  <div class="notification-form">
    <select v-model="selectedType">
      <option v-for="t in supportedTypes" :key="t.type" :value="t.type">
        {{ t.label }}
      </option>
    </select>
    <component :is="formComponents[selectedType]" v-model="notificationData[selectedType]" />
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

async function getNotificationTypes() {
  error.value = ''  // Reset error
  try {
    const response = await axios.get('ui/notification/types')
    if (response.data.success) {
    supportedTypes.value = response.data.data
    console.log('Notification types:', response.data)
      if (supportedTypes.value.length > 0) {
        selectedType.value = supportedTypes.value[0].type // default
      } else {
        error.value = 'Failed to load notification types'
      }
    }
  } catch (err) {
    if (err.response) {
      // Get server response
      error.value = `Create user failed: ${err.response.data.error || err.response.statusText}`
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
})
</script>

<style scoped>
.notification-form {
    background-color: var(--container-modal);
}

</style>
