<template>
  <div class="telegram-config">
      <input v-model="config.token" placeholder="Token" />
      <input v-model="config.chat_id" placeholder="Chat ID" />
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
.telegram-config input {
  display: block;
  margin-bottom: 0.5rem;
}
</style>
