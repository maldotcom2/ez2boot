<template>
  <div class="user-change-password">
    <form class="change-password-form" @submit.prevent="changePassword">
      <h1>Change Password</h1>
      <label>
        Current Password
        <input v-model="currentPassword" id="current-password" type="password" />
      </label>
      <label>
        New Password
        <input v-model="newPassword" id="new-password" type="password" />
      </label>
      <button type="submit" :disabled="!currentPassword || !newPassword">Change</button>
      <p class="result" :class="messageType">{{ message || '\u00A0' }}</p>
    </form>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import axios from 'axios'
import { useRouter } from 'vue-router'

const router = useRouter()
const currentPassword = ref('')
const newPassword = ref('')
const message = ref('')
const messageType = ref('')

async function changePassword() {
  message.value = ''
  messageType.value = ''

  try {
    const response = await axios.put('ui/user/password',
      {
        current_password: currentPassword.value,
        new_password: newPassword.value
      },
      {
        withCredentials: true // Cookies
      }
    )

    message.value = 'Password change successful'
    messageType.value = 'success'
    console.log('Password change successful:')
    setTimeout(() => {
        router.push({
        path: '/login',
        query: { message: 'password-changed' }
      })
    }, 2000)

  } catch (err) {
    messageType.value = 'error'
    if (err.response) {
      // Get server response
      message.value = `Password change failed: ${err.response.data.error || err.response.statusText}`
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

.user-change-password {
  display: flex;
  width: 100%;
  background-color: var(--container-modal);
  overflow-x: auto;
  border-radius: var(--small-radius);
  justify-content: center;
}

.change-password-form {
  display: flex;
  flex-direction: column;
  color: var(--low-glare);
  background-color: var(--container-modal);
  border-radius: var(--small-radius);
  padding: 2rem;
  width: 400px;
  gap: 1rem;
}

h1 {
  display: flex;
  justify-content: center;
}

input {
  width: 100%;
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
