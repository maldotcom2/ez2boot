<template>
<header class="navbar">
    <UserNav />
</header>
<div>
    <ServerSummary />
</div>
</template>

<script setup>
import { ref } from 'vue'
import axios from 'axios'
import UserNav from './UserNav.vue'
import ServerSummary from './ServerSummary.vue'
import { onMounted } from 'vue'
import { userState } from '@/user.js'

const error = ref('')

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
</style>
