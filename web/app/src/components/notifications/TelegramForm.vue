<template>
  <div class="telegram-form">
      <input v-model="config.token" placeholder="Token" />
      <input v-model="config.chat_id" placeholder="Chat ID" />
      <button type="button" :disabled="!config.token || !config.chat_id" @click="$emit('save')">Save</button>
      <button type="button" @click="$emit('delete')">Delete</button>
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

defineEmits(['update:modelValue', 'save', 'delete'])

const config = toRef(props, 'modelValue')
</script>

<style scoped>
.telegram-form input {
  display: block;
  margin-bottom: 0.5rem;
}
</style>
