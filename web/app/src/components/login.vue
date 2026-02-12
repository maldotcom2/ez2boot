<template>
  <div class="centre-container" >
    <form class="login-form" @submit.prevent="login">
      <p class="prompt">Login</p>
      <input v-model="email" placeholder="Email" />
      <input v-model="password" type="password" placeholder="Password" />
      <button type="submit" :disabled="!email || !password">Login</button>
      <p class="result" :class="messageType">{{ message || '\u00A0' }}</p>
    </form>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import axios from 'axios'
import { useRouter } from 'vue-router'

const router = useRouter()
const email = ref('')
const password = ref('')
const message = ref('')
const messageType = ref('')

// async login function
async function login() {
  message.value = ''
  messageType.value = ''

  try {
    const response = await axios.post('ui/user/login', // Login endpoint
      {
        email: email.value,
        password: password.value
      },
      {
        withCredentials: true // Cookies
      }
    )

    message.value = 'Login successful'
    messageType.value = 'success'
    console.log('Login successful:', response.data)
    router.push('/dashboard')

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
  background-color: var(--container-modal);
  justify-content: center;
  align-items: center;
  padding: 3rem;
  width: 300px;
  gap: 1rem;
  border-radius: var(--big-radius);
  outline: auto;
}

.prompt {
  color: var(--low-glare);
}

input {
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
