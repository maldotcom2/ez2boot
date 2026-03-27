<template>
  <div class="auth-config">
    <label for="issuer-url">Issuer URL</label> 
    <input id="issuer-url" placeholder="https://login.microsoftonline.com/.../v2.0" v-model="config.issuer_url"/>
    <label for="client-id">Client ID</label>
    <input id="client-id" v-model="config.client_id"/>
    <label for="client-secret">Client Secret</label>
    <input id="client-secret" type=password v-model="config.client_secret"/>
    <label for="App URL">App URL</label>
    <input id="App URL" placeholder="ez2boot.mycompany.com" v-model="config.app_url"/>
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
.issuer-container {
  display: flex;
  align-items: center;
  gap: 1rem;
  margin-bottom: 1rem;
  width: 500px;
}

.issuer-container input {
  flex: 1;
  width: auto !important;
  margin-bottom: 0;
}

.issuer-container button {
  width: auto;
  white-space: nowrap;
}
</style>