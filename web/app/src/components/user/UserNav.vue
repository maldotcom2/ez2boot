<template>
  <div class="user-nav">
    <span class="update-nag" v-if="version.updateAvailable"><a :href="version.releaseURL" target="_blank">Update Available!</a></span>
    <p>{{ user.email }}</p>
    <button @click="toggleUserDropdown">Menu</button>
    <div v-if="isOpen" class="dropdown">
      <button v-if="user.isAdmin" @click="admin">Admin Panel</button>
      <button @click="dashboard">Dashboard</button>
      <button @click="settings">Settings</button>
      <button @click="about">About</button>
      <button @click="logout">Logout</button>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import axios from 'axios'
import { useRouter } from 'vue-router'
import { useUserStore } from '@/stores/user'
import { useVersionStore } from '@/stores/version'

const user = useUserStore()
const version = useVersionStore()
const isOpen = ref(false)
const router = useRouter()
const error = ref('')

function toggleUserDropdown() {
  isOpen.value = !isOpen.value
}

function admin() {
  router.push("/adminpanel")
}

function dashboard() {
  router.push("/dashboard")
}

function settings() {
  router.push("/settings")
}

function about() {
  router.push("/about")
}

async function logout() {
  try {
    const response = await axios.post('ui/user/logout',{withCredentials: true})
    user.$reset() // purge User store
    version.$reset() // purge Version store
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
      console.error("Failed to load user store", user.error)
    }

    try {
      await version.getVersion()
    } catch (err) {
      console.error("Failed to load version store", version.error)
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

.update-nag {
  margin-right: 1rem;
}

.update-nag a {
  color: var(--warn-amber);
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