<template>
  <div class="centre-container" >
    <form class="setup-form" @submit.prevent="createFirstUser">
      <h1>Create initial user</h1>
      <label>
        Email
        <input v-model="email" placeholder="example@example.com" />
      </label>
      <label>
        Password
        <input v-model="password" type="password" />
      </label>
      <label>
        Confirm Password
        <input v-model="confirmPassword" type="password" />
      </label>
      <button type=submit :disabled="!email || !password || !confirmPassword">Create</button>
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
const confirmPassword = ref('')
const message = ref('')
const messageType = ref('')

async function createFirstUser() {
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

    message.value = 'User created'
    messageType.value = 'success'
    console.log('User creation successful:', response.data)
    setTimeout(() => {
        router.push({
        path: '/login',
        query: { message: 'user-created' }
      })
    }, 2000)

  } catch (err) {
    messageType.value = 'error'
    if (err.response) {
      // Get server response
      message.value = `User creation failed: ${err.response.data.error || err.response.statusText}`
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

.setup-form {
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
