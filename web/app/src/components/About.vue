<template>
  <div class="page" >
    <header class="navbar">
      <UserNav />
    </header>
    <div class="centre-container" >
      <div class="legal-block"> 
        Copyright © 2026 maldotcom2<br />
        <span>
          Version: {{ version.version }} ({{ version.buildDate }})
           <span class="update-nag" v-if="version.updateAvailable"><a :href="version.releaseURL" target="_blank">Update Available! - {{ version.latestVersion }}</a></span>
           <span v-else>Up to date</span>
           <span> - Last checked {{ version.checkedAt ? new Date(version.checkedAt * 1000).toLocaleString() : '-' }}</span>
        </span>
        <a href="/LICENSE.txt" target="_blank">AGPLv3</a>
        <a href="https://github.com/maldotcom2/ez2boot/" target="_blank">Source</a>
        No warranty 
      </div>
    </div>
  </div>
</template>

<script setup>
import UserNav from './user/UserNav.vue'
import { useVersionStore } from '@/stores/version'

const version = useVersionStore()

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

.centre-container {
  flex: 1;
  display: flex;
  justify-content: center;
  align-items: flex-start;
}

.legal-link {
  color: var(--low-glare);
}

.legal-block {
  display: flex;
  flex-direction: column;
  color: var(--low-glare);
  margin-top: 2rem;
  padding: 1rem;
}

.legal-block a {
  color: var(--low-glare);
}

.update-nag a {
  color: var(--warn-amber);
}

</style>
