<template>
  <div class="page" >
    <header>
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
      console.log('MFA status is', user.hasMFA)
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

header {
  box-shadow: 0 1px 0 0 var(--low-glare);
  z-index: 10;
}

</style>
