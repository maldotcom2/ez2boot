<template>
<header class="navbar">
    <UserNav />
</header>
<div>
    <ServerSummary />
    <p class="version-info">Version: {{ versionInfo.version }} ({{ versionInfo.build_date }})</p>
</div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import axios from 'axios'
import UserNav from './UserNav.vue'
import ServerSummary from './ServerSummary.vue'
import { userState } from '@/user.js'

const error = ref('')
const versionInfo = ref({ version: '', buildDate: '' })

// Get user authorisation values
async function getUserAuth() {
    try {
        const response = await axios.get('ui/user/auth', {withCredentials: true})
        console.log('got user auth', response.data)
        return response.data   
  } catch (err) {
    if (err.response) {
        // Get server response
        error.value = `User auth fetch failed: ${err.response.data.error || err.response.statusText}`
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

async function getVersion() {
  try {
    const response  = await axios.get('ui/version')
    if (response.data.success) {
      versionInfo.value = response.data.data
    }

  } catch (err) {
    if (err.response) {
        // Get server response
        error.value = `Get version failed: ${err.response.data.error || err.response.statusText}`
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
      const response = await getUserAuth()
      userState.email = response.data.email
      userState.isAdmin = response.data.is_admin
      console.log('Current user is', userState.email)
      console.log('User is admin', userState.isAdmin)
    } catch (err) {
      console.error('Could not fetch user on page load', err)
    }

    try {
      await getVersion()
    } catch (err) {
      console.error('Could not fetch user on page load', err)
    }
})

</script>

<style scoped>
.navbar {
  display: flex;
  justify-content: flex-end;
  align-items: center; /* vertically */
  padding: 10px 20px;
  background-color: var(--container-header);
  position: relative; /*for dropdown positioning */
  height: 60px;
  outline: auto;
}

.version-info {
  text-align: center;
  color: var(--low-glare);
  font-size: 0.8rem;
}

</style>
