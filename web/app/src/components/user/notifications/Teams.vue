<template>
  <div class="channel-config">
    <label for="Webhook">Webhook URL</label>
    <input id="Webhook" v-model="config.webhook"/>
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
