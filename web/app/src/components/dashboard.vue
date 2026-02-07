<template>
  <div class="page" >
    <header class="navbar">
        <UserNav />
    </header>
    <div>
      <ServerSummary />
    </div>
  </div>
</template>

<script setup>
import { onMounted } from 'vue'
import UserNav from './user/UserNav.vue'
import ServerSummary from './ServerSummary.vue'
import { useUserStore } from '@/stores/user'

const user = useUserStore()

onMounted(async () => {
    try {
      await user.loadUser()
      console.log('Current user id is', user.userID)
      console.log('Current user is', user.email)
      console.log('User is admin', user.isAdmin)
    } catch (err) {
      console.error("Failed to load user", user.error)
    }
})

</script>

<style scoped>
.page {
  display: flex;
  flex-direction: column;
  min-height: 100vh;
}

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
