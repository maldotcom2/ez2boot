<template>
<div class="user-nav">
    <p>{{ userState.email }}</p>
    <button @click="toggleUserDropdown">Menu</button>
    <div v-if="isOpen" class="dropdown">
      <button v-if="userState.isAdmin" @click="$router.push('/adminpanel')">Admin Panel</button>
      <button @click="settings">Settings</button>
      <button @click="logout">Logout</button>
    </div>
</div>
</template>

<script setup>
import { ref } from 'vue'
import axios from 'axios'
import { useRouter } from 'vue-router'
import { userState } from '@/user.js'

const isOpen = ref(false)
const router = useRouter()
const email = ref('')
const password = ref('')
const error = ref('')

function toggleUserDropdown() {
  isOpen.value = !isOpen.value
}

async function logout() {
    try {
        const response = await axios.post(
            'ui/user/logout',
            {
                withCredentials: true // Cookies
            }
        )
        console.log('logout successful', response.data)
        router.push('/login')

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
    console.log(error.value)
  }
}
</script>

<style scoped>
.user-nav {
    display: flex;
    position: relative;
    align-items: center;
    gap: 10px;
    color: var(--low-glare);
}

.dropdown {
  position: absolute;
  top: 100%; /* below the button */
  right: 0;
  margin-top: 5px;
  background: var(--container-modal);
  display: flex;
  flex-direction: column;
  min-width: 150px;
  border-radius: 12px;
}

.dropdown button {
  background-color: var(--container-modal);
  color: var(--low-glare);
  padding: 8px 12px;
  text-align: left;
}
</style>