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
import { useUserStore } from '@/stores/user'

const user = useUserStore()
const error = ref('')
const versionInfo = ref({ version: '', buildDate: '' })

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
      await user.loadUser()
      console.log('Current user id is', user.userID)
      console.log('Current user is', user.email)
      console.log('User is admin', user.isAdmin)
    } catch (err) {
      console.error("Failed to load user", user.error)
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
