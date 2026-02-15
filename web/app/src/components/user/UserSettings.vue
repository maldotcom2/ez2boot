<template>
  <div class="page">
    <header>
      <UserNav />
    </header>
    <div class="user-settings">
      <UserMenu v-model="currentComponent" />
      <main class="content">
        <component :is="currentComponent" @switch-pane="currentComponent = $event" />
      </main>
    </div>
  </div>
</template>

<script setup>
import UserNav from './UserNav.vue'
import { shallowRef } from 'vue'
import UserMenu from './UserMenu.vue'
import UserNotifications from './UserNotifications.vue'

const currentComponent = shallowRef(UserNotifications) // default view on load
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

.user-settings {
  flex: 1;
  display: flex;
  align-items: stretch;
  overflow: hidden;
}

.content {
  flex: 1;
  display: flex;
  padding: 1rem;
  overflow-y: auto;
}

</style>