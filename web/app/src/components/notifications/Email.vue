<template>
  <div class="channel-config">
    <label for="host">Host</label>
    <input id="host" v-model="config.host"/>
    <label for="port">Port</label>
    <input id="port" v-model="config.port"/>
    <label for="to">To</label>
    <input id="to" v-model="config.to"/>
    <label for="from">From</label>
    <input id="from" v-model="config.from"/>
    <div class="checkbox-row">
      <label for="auth">Authentication</label>
      <input id="auth" class="checkbox" type="checkbox" v-model="config.auth"/>
    </div>
    <label for="user">User</label>
    <input id="user" v-model="config.user" :disabled="!config.auth"/>
    <label for="password">Password</label>
    <input id="password" type="password" v-model="config.password" :disabled="!config.auth"/>
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