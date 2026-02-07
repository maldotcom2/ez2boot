<template>
  <div class="page">
    <header class="navbar">
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