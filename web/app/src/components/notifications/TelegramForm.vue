<template>
<p>Telegram</p>
<div class="telegram-form">
  <form>
    <input v-model="token" placeholder="Token" />
    <input v-model="chatID" placeholder="Chat ID" />
    <button type="submit" :disabled="!token || !chatID">Save</button>
    <button @click="deleteNotification">Delete</button>
  </form>
</div>
</template>

<script setup>

</script>

<style scoped>

</style>
