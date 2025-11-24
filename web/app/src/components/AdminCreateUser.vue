<template>
  <div class="centre-container">
      <form class="create-user-form" @submit.prevent="createUser">
        <p class="prompt">Create User</p>
        <input v-model="email" placeholder="Email" />
        <input v-model="password" type="password" placeholder="Password" />
        <input v-model="confirmPassword" type="password" placeholder="Confirm Password" />
        <p v-if="!passwordsMatch && confirmPassword.length > 0" class="error">Passwords do not match</p>
        <button type="submit" :disabled="!passwordsMatch || !email || !password || !confirmPassword">Create</button>
        <p v-if="error" class="error">{{ error }}</p>
      </form>
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'
import axios from 'axios'
import AdminUserMgmt from './AdminUserMgmt.vue'

const email = ref('')
const password = ref('')
const confirmPassword = ref('')
const error = ref('')
const passwordsMatch = computed(() => password.value === confirmPassword.value)

const emit = defineEmits(['switch-pane'])

async function createUser() {
  error.value = ''  // Reset error
  try {
    const response = await axios.post('ui/user/new',
      {
        email: email.value,
        password: password.value
      },
      {
        withCredentials: true // Cookies
      }
    )

    console.log('User created:', response.data)
    emit('switch-pane', AdminUserMgmt)

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
</script>

<style scoped>
.centre-container {
  display: flex;
  justify-content: center;
  align-items: center;
  height: 100%;
  outline: none;
}

.create-user-form {
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

.create-user-form button{
  width: 150px;
}

.prompt {
  color: var(--low-glare);
  font-size: 1.2rem;
}

input {
  height: 30px;
  width: 100%;
  padding: 5px;
}

button {
  width: 100%;
  padding: 8px 0;
  margin-top: 5px;
}

.error {
  color: red;
  font-size: 0.9rem;
}

</style>
