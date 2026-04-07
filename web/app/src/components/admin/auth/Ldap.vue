<template>
  <div class="auth-config">
    <label for="host">Host</label>
    <input id="host" placeholder="ad01" v-model="config.host" />
    <label for="port">Port</label>
    <input id="port" type="number" placeholder="389" v-model="config.port" />
    <label for="base-dn">Base DN</label>
    <input id="base-dn" placeholder="DC=example,DC=org" v-model="config.base_dn" />
    <label for="bind-dn">Bind DN</label>
    <input
      id="bind-dn"
      placeholder="CN=ldap.svc,CN=Users,DC=example,DC=org"
      v-model="config.bind_dn"
    />
    <label for="bind-password">Bind Password</label>
    <input id="bind-password" type="password" v-model="config.bind_password" />
    <div class="checkbox-row">
      <label for="use-ssl">Use SSL</label>
      <input id="use-ssl" class="checkbox" type="checkbox" v-model="config.use_ssl" />
    </div>
    <div class="checkbox-row">
      <label for="skip-tls-verify">Skip TLS Verify</label>
      <input
        id="skip-tls-verify"
        class="checkbox"
        type="checkbox"
        v-model="config.skip_tls_verify"
      />
    </div>
  </div>
</template>

<script setup>
import { toRef } from 'vue'

const props = defineProps({
  modelValue: {
    type: Object,
    required: true,
  },
})

defineEmits(['update:modelValue'])

const config = toRef(props, 'modelValue')
</script>

<style scoped></style>
