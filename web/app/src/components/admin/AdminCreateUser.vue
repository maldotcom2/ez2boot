<template>
  <div class="centre-container">
      <form class="create-user-form" @submit.prevent="createUser">
        <p class="prompt">Create User</p>
        <input v-model="email" placeholder="Email" />
        <input v-model="password" type="password" placeholder="Password" />
        <input v-model="confirmPassword" type="password" placeholder="Confirm Password" />
        <button type="submit" :disabled="!passwordsMatch || !email || !password || !confirmPassword">Create</button>
        <p class="result" :class="messageType">{{ validationMessage || message || '\u00A0' }}</p>
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
const message = ref('')
const messageType = ref('')
const passwordsMatch = computed(() => password.value === confirmPassword.value)
const validationMessage = computed(() => {
  if (confirmPassword.value.length > 0 && !passwordsMatch.value) {
    messageType.value = 'error'
    return 'Passwords do not match'
  }
  return ''
})

const emit = defineEmits(['switch-pane'])

async function createUser() {
  message.value = ''
  messageType.value = ''

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

    message.value = 'User created'
    messageType.value = 'success'
    console.log('User created:', response.data)
    emit('switch-pane', AdminUserMgmt)

  } catch (err) {
    messageType.value = 'error'
    if (err.response) {
      // Get server response
      message.value = `Create user failed: ${err.response.data.error || err.response.statusText}`
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
  height: 100%;
  outline: none;
}

.create-user-form {
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

.create-user-form button{
  width: 100%;
}

.prompt {
  color: var(--low-glare);
  font-size: 1.2rem;
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
