<template>
  <header class="navbar">
    <UserNav />
  </header>
  <div class="admin-panel">
    <<AdminMenu v-model="currentComponent" />
    <main class="content">
      <component :is="currentComponent" />
    </main>
  </div>
</template>

<script setup>
import UserNav from './UserNav.vue'
import { shallowRef } from 'vue'
import AdminMenu from './AdminMenu.vue'

const currentComponent = shallowRef(null) // default view on load
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

.admin-panel {
  display: flex;
  height: 100vh;
}

.content {
  flex: 1;
  padding: 1rem;
  overflow-y: auto;
}

</style>