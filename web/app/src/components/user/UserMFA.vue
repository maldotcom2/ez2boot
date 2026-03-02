<template>
  <div class="user-mfa">
    <div class="mfa-container">
      <h1 v-if="!user.hasMFA">Enrol MFA</h1>
      <h1 v-else>Delete MFA</h1>
      <p v-if="!qrCode && !user.hasMFA">This account is not enrolled in MFA</p>
      <p v-if="user.hasMFA">This account is enrolled in MFA. Enter the 6 digit code and click delete to remove it.</p>
      <button v-if="!qrCode && !user.hasMFA" type="button" @click="enrolMFA">Enrol</button>
        <template v-if="qrCode || user.hasMFA">
          <img v-if="qrCode" class="qr-code" :src="`data:image/png;base64,${qrCode}`" />
          <label>
            6-digit Code
            <input v-model="mfaCode" />
          </label>
          <button v-if="qrCode" type="button" :disabled="mfaCode.length !== 6" @click="confirmMFA">Confirm</button>
          <button v-if="user.hasMFA" type="button" :disabled="mfaCode.length !== 6" @click="deleteMFA">Delete</button>
      </template>
      <p class="result" :class="messageType">{{ message || '\u00A0' }}</p>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import axios from 'axios'
import { useUserStore } from '@/stores/user'

const user = useUserStore()
const qrCode = ref(null)
const mfaCode = ref('')
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
        code: mfaCode.value
      },
      {
        withCredentials: true // Cookies
      }
    )

    if (response.data.success) {
      message.value = 'Successfully enrolled MFA'
      messageType.value = 'success'
      qrCode.value = null
      mfaCode.value = ''
      user.hasMFA = true
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

async function deleteMFA() {
  message.value = ''
  messageType.value = ''

  try {
    const response = await axios.post('ui/user/mfa/delete',
      {
        code: mfaCode.value
      },
      {
        withCredentials: true // Cookies
      }
    )

    if (response.data.success) {
      mfaCode.value = ''
      qrCode.value = null
      message.value = 'Successfully deleted MFA'
      messageType.value = 'success'
      user.hasMFA = false
    }

  } catch (err) {
    messageType.value = 'error'
    if (err.response) {
      // Get server response
      message.value = `MFA deletion failed: ${err.response.data.error || err.response.statusText}`
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

p {
  text-align: center;
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
