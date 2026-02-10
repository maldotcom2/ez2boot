<template>
  <div class="page">
    <header>
      <UserNav />
    </header>
    <div class="admin-panel">
      <AdminMenu v-model="currentComponent" />
      <main class="content">
        <component :is="currentComponent" @switch-pane="currentComponent = $event" />
      </main>
    </div>
  </div>
</template>

<script setup>
import UserNav from '../user/UserNav.vue'
import { shallowRef } from 'vue'
import AdminMenu from './AdminMenu.vue'
import UserManagement from './AdminUserMgmt.vue'

const currentComponent = shallowRef(UserManagement) // default view on load
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

.admin-panel {
  flex: 1;
  display: flex;
  align-items: stretch;
  overflow: hidden;
}

.content {
  flex: 1;
  padding: 1rem;
  overflow-y: auto;
}

</style>