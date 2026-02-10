<template>
  <div class="page" >
    <header>
      <UserNav />
    </header>
    <div class="centre-container" >
      <div class="legal-block"> 
        Copyright © 2026 maldotcom2<br />
        <span>
          Version: {{ version.version }} ({{ version.buildDate }})
           <span class="update-nag" v-if="version.updateAvailable"><a :href="version.releaseURL" target="_blank">Update Available! - {{ version.latestVersion }}</a></span>
           <span v-else>Up to date</span>
           <span> - last checked: {{ version.checkedAt ? new Date(version.checkedAt * 1000).toLocaleString() : 'never' }}</span>
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

header {
  box-shadow: 0 1px 0 0 var(--low-glare);
  z-index: 10;
}

.centre-container {
  flex: 1;
  display: flex;
  background-color: var(--container-modal);
  justify-content: center;
  align-items: flex-start;
  margin: 1rem 0.5rem;
  border-radius: var(--small-radius);
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
