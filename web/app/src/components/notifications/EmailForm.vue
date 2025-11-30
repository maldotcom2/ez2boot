<template>
  <div class="email-config">
      <input v-model="config.host" placeholder="Host" />
      <input v-model="config.port" placeholder="Port" />
      <input v-model="config.to" placeholder="To" />
      <input v-model="config.from" placeholder="From" />
      <label>
        <input type="checkbox" v-model="config.auth" placeholder="Auth" />
        Authentication
      </label>
      <input v-model="config.user" :disabled="!config.auth" placeholder="Username" />
      <input type="password" v-model="config.password" :disabled="!config.auth" placeholder="Password" />
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
.email-config input {
  display: block;
  margin-bottom: 0.5rem;
}

.email-config label {
  color: var(--low-glare);
  display: flex;
  align-items: center;
}
</style>