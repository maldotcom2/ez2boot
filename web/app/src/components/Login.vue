<template>
  <div class="centre-container" >
    <form class="login-form" @submit.prevent="login">
      <h1>Login</h1>
      <template v-if="!mfaRequired">
        <label>
          Email
          <input v-model="email" />
        </label>
        <label>
          Password
          <input v-model="password" type="password" />
        </label>
        <button type="submit" :disabled="!email || !password">Login</button>
      </template>
      <template v-else>
        <p>MFA required: Open your authenticator app and enter the 6-digit code</p>
        <input v-model="mfaCode" maxlength="6"/>
        <button type="button" :disabled="mfaCode.length !== 6" @click="verifyMFA">Verify</button>
      </template>
      <p class="result" :class="messageType">{{ message || '\u00A0' }}</p>
    </form>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import axios from 'axios'
import { useRouter, useRoute } from 'vue-router'

const router = useRouter()
const route = useRoute()
const email = ref('')
const password = ref('')
const mfaCode = ref('')
const mfaRequired = ref(false)
const message = ref('')
const messageType = ref('')

if (route.query.message === 'password-changed') {
  messageType.value = 'success'
  message.value = 'Your password was changed. Please log in again.'
}

if (route.query.message === 'user-created') {
  messageType.value = 'success'
  message.value = 'Initial user created. Please login.'
}

// async login function
async function login() {
  message.value = ''
  messageType.value = ''

  try {
    const response = await axios.post('ui/auth/login',
      {
        email: email.value,
        password: password.value
      },
      {
        withCredentials: true // Cookies
      }
    )

    // Intercept if MFA is required for this user
    if (response.data.data?.mfa_required) {
      mfaRequired.value = true
      message.value = ''
      return
    }

    message.value = 'Login successful'
    messageType.value = 'success'
    setTimeout(() => {
      router.push({
      path: '/dashboard',
    })
    }, 1000)

  } catch (err) {
    messageType.value = 'error'
    if (err.response) {
      // Get server response
      message.value = `Login failed: ${err.response.data.error || err.response.statusText}`
    } else if (err.request) {
      // No response
      message.value = 'No response from server'
    } else {
      // other errors
      message.value = err.message
    }
  }
}

async function verifyMFA() {
  message.value = ''
  messageType.value = ''

  try {
    const response = await axios.post('ui/user/mfa/verify',
      { 
        code: mfaCode.value 
      },
      { 
        withCredentials: true 
      }
    )

    if (response.data.success) {
      message.value = 'Login successful'
      messageType.value = 'success'
      setTimeout(() => router.push('/dashboard'), 1000)
    }

  } catch (err) {
    messageType.value = 'error'
    if (err.response) {
      message.value = err.response.data.error || err.response.statusText
    } else if (err.request) {
      message.value = 'No response from server'
    } else {
      message.value = err.message
    }
  }
}
</script>

<style scoped>
.centre-container {
  display: flex;
  justify-content: center;
  align-items: center;
  height: 100vh; /* Full screen */
}

.login-form {
  display: flex;
  flex-direction: column;
  color: var(--low-glare);
  background-color: var(--container-modal);
  justify-content: center;
  align-items: center;
  padding: 3rem;
  width: 300px;
  gap: 1rem;
  border-radius: var(--big-radius);
  outline: auto;
}

h1 {
  display: flex;
  justify-content: center;
}

p {
  text-align: center;
}

input {
  width: 100%;
}

label {
  display: flex;
  flex-direction: column;
  width: 100%;
}

button {
  width: 100%;
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
