<template>
  <div class="centre-container" >
    <form class="login-form" @submit.prevent="login">
      <p class="prompt">Login</p>
      <input v-model="email" placeholder="Email" />
      <input v-model="password" type="password" placeholder="Password" />
      <button type="submit" :disabled="!email || !password">Login</button>
      <div class="message-container" >
        <p class="message" :class="{error: error}" v-if="error">{{ error }}</p>
      </div>
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
const error = ref('')

// async login function
async function login() {
  error.value = ''  // Reset error
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

    console.log('Login successful:', response.data)
    router.push('/dashboard')

  } catch (err) {
    if (err.response) {
      // Get server response
      error.value = `Login failed: ${err.response.data.error || err.response.statusText}`
    } else if (err.request) {
      // No response
      error.value = 'No response from server'
    } else {
      // other errors
      error.value = err.message
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
  padding: 40px;
  width: 300px;
  gap: 15px;
  border-radius: var(--big-radius);
  outline: auto;
}

.message-container {
  width: 100%;
}

.prompt {
  color: var(--low-glare);
}

input {
  height: 30px;
  width: 100%;
  padding: 5px;
}

button {
  width: 50%;
}
</style>
