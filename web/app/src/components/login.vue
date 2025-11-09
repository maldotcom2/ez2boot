<template>
  <div class="login-form" >
    <input v-model="email" placeholder="Email" />
    <input v-model="password" type="password" placeholder="Password" />
    <button @click="login">Login</button>
    <p class="message" :class="{error: error}" v-if="error">{{ error }}</p>
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
    const response = await axios.post(
      'ui/user/login', // Login endpoint
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
.login-form {
  display: flex;
  flex-direction: column;
  justify-content: center;
  align-items: center;
  height: 100vh;
  width: 30%;
  gap: 15px;
  margin: 0 auto; /* centers the form horizontally */
}

button {
  padding: 10px 20px;
  background-color: var(--main-blue);
  color: white;
  border: none;
  border-radius: 4px;
  cursor: pointer;
}
</style>
