<template>
  <div class="user-mfa">
    <div class="mfa-container">
      <h1>Enrol MFA</h1>
      <p v-if="!qrCode">This account is not enrolled in MFA, click enrol to begin.</p>
      <button v-if="!qrCode" type="button" @click="enrolMFA">Enrol</button>
        <template v-if="qrCode">
          <img class="qr-code" :src="`data:image/png;base64,${qrCode}`" />
          <label>
            6-digit Code
            <input v-model="confirmCode" :disabled="confirmed" />
          </label>
          <button type="button" :disabled="confirmed || confirmCode.length !== 6" @click="confirmMFA">Confirm</button>
      </template>
      <p class="result" :class="messageType">{{ message || '\u00A0' }}</p>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import axios from 'axios'
import { useRouter } from 'vue-router'

const router = useRouter()
const qrCode = ref(null)
const confirmCode = ref('')
const confirmed = ref(false)
const message = ref('')
const messageType = ref('')

async function enrolMFA() {
  message.value = ''
  messageType.value = ''

  try {
    const response = await axios.post('ui/user/mfa',
      {},
      {
        withCredentials: true // Cookies
      }
    )

    if (response.data.success) {
      qrCode.value = response.data.data
    }

  } catch (err) {
    messageType.value = 'error'
    if (err.response) {
      // Get server response
      message.value = `MFA enrolment failed: ${err.response.data.error || err.response.statusText}`
    } else if (err.request) {
      // No response
      message.value = 'No response from server'
    } else {
      // other errors
      message.value = err.message
    }
  }
}

async function confirmMFA() {
  message.value = ''
  messageType.value = ''

  try {
    const response = await axios.post('ui/user/mfa/confirm',
      {
        code: confirmCode.value
      },
      {
        withCredentials: true // Cookies
      }
    )

    if (response.data.success) {
      confirmed.value = true
      message.value = 'Successfully enrolled MFA'
      messageType.value = 'success'
    }

  } catch (err) {
    messageType.value = 'error'
    if (err.response) {
      // Get server response
      message.value = `MFA enrolment failed: ${err.response.data.error || err.response.statusText}`
    } else if (err.request) {
      // No response
      message.value = 'No response from server'
    } else {
      // other errors
      message.value = err.message
    }
  }
}

</script>

<style scoped>

.user-mfa {
  display: flex;
  width: 100%;
  background-color: var(--container-modal);
  overflow-x: auto;
  border-radius: var(--small-radius);
  justify-content: center;
}

.mfa-container {
  display: flex;
  flex-direction: column;
  color: var(--low-glare);
  background-color: var(--container-modal);
  border-radius: var(--small-radius);
  padding: 2rem;
  width: 400px;
  gap: 1rem;
}

.qr-code {
  max-width: 200px;
  align-self: center;
}

h1 {
  display: flex;
  justify-content: center;
}

input {
  width: 130px;
}

label {
  display: flex;
  flex-direction: column;
  align-items: center;
}

button {
  width: 130px;
  align-self: center;
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
