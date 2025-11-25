<template>
<div class="user-nav">
    <p>{{ user.email }}</p>
    <button @click="toggleUserDropdown">Menu</button>
    <div v-if="isOpen" class="dropdown">
      <button v-if="user.isAdmin" @click="admin">Admin Panel</button>
      <button @click="dashboard">Dashboard</button>
      <button @click="settings">Settings</button>
      <button @click="logout">Logout</button>
    </div>
</div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import axios from 'axios'
import { useRouter } from 'vue-router'
import { useUserStore } from '@/stores/user'

const user = useUserStore()
const isOpen = ref(false)
const router = useRouter()
const error = ref('')

function toggleUserDropdown() {
  isOpen.value = !isOpen.value
}

function dashboard() {
  router.push("/dashboard")
}

function settings() {
  router.push("/settings")
}

function admin() {
  router.push("/adminpanel")
}

async function logout() {
    try {
        const response = await axios.post('ui/user/logout',{withCredentials: true})
        user.$reset() // purge Pinia store
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

onMounted(async () => {
    try {
      await user.loadUser()
    } catch (err) {
      console.error("Failed to load user", user.error)
    }
})
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
  border-radius: var(--small-radius);
}

.dropdown button {
  background-color: var(--container-modal);
  color: var(--low-glare);
  padding: 8px 12px;
  text-align: left;
}
</style>