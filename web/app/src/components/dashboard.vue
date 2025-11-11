<template>
<header class="navbar">
    <UserNav />
</header>
<div>
    <ServerSummary />
</div>
</template>

<script setup>
import UserNav from './UserNav.vue'
import ServerSummary from './ServerSummary.vue'
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
