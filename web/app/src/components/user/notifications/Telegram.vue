<template>
  <div class="channel-config">
    <label for="token">Token</label>
    <input id="token" v-model="config.token"/>
    <label for="chat-id">Chat ID</label>
    <input id="chat-id" v-model="config.chat_id"/>
  </div>
</template>

<script setup>
import { toRef } from 'vue'

const props = defineProps({
  modelValue: {
    type: Object,
    required: true
  }
})

defineEmits(['update:modelValue'])

const config = toRef(props, 'modelValue')
</script>

<style scoped>

</style>
