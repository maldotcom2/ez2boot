<template>
<header class="navbar">
    <UserNav />
</header>
</template>

<script setup>
import UserNav from './UserNav.vue'
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
