<template>
  <div class="centre-container" >
    <div class="setup-form" >
      <p class="prompt">Create initial user</p>
      <input v-model="email" placeholder="Email" />
      <input v-model="password" type="password" placeholder="Password" />
      <input v-model="confirmPassword" type="password" placeholder="Confirm Password" />
      <button @click="createFirstUser">Create</button>
      <div class="message-container" >
        <p class="message" :class="{error: error}" v-if="error">{{ error }}</p>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import axios from 'axios'
import { useRouter } from 'vue-router'

const router = useRouter()
const email = ref('')
const password = ref('')
const confirmPassword = ref('')
const error = ref('')

async function createFirstUser() {
  error.value = ''  // Reset error
  if (password.value !== confirmPassword.value) {
    console.error("password and confirm password do not match")
    return
  }
  try {
    const response = await axios.post('ui/setup',
      {
        email: email.value,
        password: password.value
      }
    )
    console.log('User creation successful:', response.data)

    router.push('/login')

  } catch (err) {
    if (err.response) {
      // Get server response
      error.value = `User creation failed: ${err.response.data.error || err.response.statusText}`
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

.setup-form {
  display: flex;
  flex-direction: column;
  background-color: var(--container-modal);
  justify-content: center;
  align-items: center;
  padding: 40px;
  width: 25%;
  gap: 15px;
  border-radius:15px;
  outline: none;
}

.message-container {
  width: 100%;
  outline: auto;
}

.prompt {
  color: var(--low-glare);
}

input {
  height: 30px;
  width: 50%;
}

button {
  width: 50%;
}
</style>
