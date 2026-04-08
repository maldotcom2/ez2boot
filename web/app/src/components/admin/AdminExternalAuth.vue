<template>
  <div class="external-auth">
    <aside class="sidebar">
      <select class="auth-selector" v-model="selectedType">
        <option v-for="t in authTypes" :key="t.type" :value="t.type">
          {{ t.label }}
        </option>
      </select>
      <div class="actions">
        <button type="button" @click="saveConfig" :disabled="!isDirty">Save</button>
        <button type="button" @click="deleteConfig" :disabled="!canDelete">Delete</button>
        <button type="button" @click="testConfig" :disabled="!canDelete || isDirty">Test</button>
      </div>
      <p class="result" :class="messageType">{{ message || '\u00A0' }}</p>
    </aside>
    <main class="config-panel">
      <component
        v-if="selectedType"
        :is="formComponents[selectedType]"
        v-model="authConfig[selectedType]"
      />
    </main>
  </div>
</template>

<script setup>
import { computed, ref, reactive, onMounted, watch } from 'vue'
import axios from 'axios'
import LdapForm from './auth/Ldap.vue'
import OidcForm from './auth/Oidc.vue'

const selectedType = ref('')
const authConfig = reactive({})
const message = ref('')
const messageType = ref('')
const originalData = reactive({})
const loadedTypes = reactive(new Set()) // Track auth types with configs

const authTypes = [
  { type: 'ldap', label: 'LDAP' },
  { type: 'oidc', label: 'OIDC' },
]

const formComponents = {
  ldap: LdapForm,
  oidc: OidcForm,
}

// For enabling delete button
const canDelete = computed(() => {
  return loadedTypes.has(selectedType.value)
})

// Track changes to saved auth config
const isDirty = computed(() => {
  if (!selectedType.value) return false
  const current = authConfig[selectedType.value] || {}
  const original = originalData[selectedType.value] || {}
  return JSON.stringify(current) !== JSON.stringify(original)
})

// Gracefully handle load
watch(selectedType, async (newType) => {
  if (newType && !authConfig[newType]) {
    authConfig[newType] = {}
  }
  message.value = ''
  messageType.value = ''
  await loadConfig()
})

const apiRoutes = {
  ldap: '/ui/auth/ldap',
  oidc: '/ui/auth/oidc',
  test: {
    ldap: '/ui/auth/ldap/users/search',
    oidc: '/ui/auth/oidc/test',
  },
}

async function loadConfig() {
  message.value = ''
  messageType.value = ''
  try {
    const response = await axios.get(apiRoutes[selectedType.value])
    if (response.data.success && response.data.data) {
      authConfig[selectedType.value] = response.data.data
      originalData[selectedType.value] = JSON.parse(JSON.stringify(response.data.data))
      loadedTypes.add(selectedType.value)
    }
  } catch (err) {
    messageType.value = 'error'
    if (err.response) {
      message.value = `Failed to load config: ${err.response.data.error || err.response.statusText}`
    } else if (err.request) {
      message.value = 'No response from server'
    } else {
      message.value = err.message
    }
  }
}

async function saveConfig() {
  message.value = ''
  messageType.value = ''
  try {
    await axios.post(apiRoutes[selectedType.value], authConfig[selectedType.value])
    message.value = 'Config saved'
    messageType.value = 'success'
    originalData[selectedType.value] = JSON.parse(JSON.stringify(authConfig[selectedType.value]))
  } catch (err) {
    messageType.value = 'error'
    if (err.response) {
      message.value = `Failed to save config: ${err.response.data.error || err.response.statusText}`
    } else if (err.request) {
      message.value = 'No response from server'
    } else {
      message.value = err.message
    }
  }
}

async function deleteConfig() {
  message.value = ''
  messageType.value = ''
  if (
    !confirm(
      `Are you sure you want to delete the ${selectedType.value.toUpperCase()} configuration?`,
    )
  )
    return
  try {
    await axios.delete(apiRoutes[selectedType.value])
    authConfig[selectedType.value] = {}
    originalData[selectedType.value] = {}
    message.value = 'Config deleted'
    messageType.value = 'success'
    loadedTypes.delete(selectedType.value)
  } catch (err) {
    messageType.value = 'error'
    if (err.response) {
      message.value = `Failed to delete config: ${err.response.data.error || err.response.statusText}`
    } else if (err.request) {
      message.value = 'No response from server'
    } else {
      message.value = err.message
    }
  }
}

async function testConfig() {
  if (selectedType.value === 'ldap') return testLdap()
  if (selectedType.value === 'oidc') return testOidc()
}

async function testLdap() {
  message.value = ''
  messageType.value = ''
  try {
    await axios.post('/ui/auth/ldap/users/search', {
      query: 'test',
    })
    message.value = 'LDAP Connection successful'
    messageType.value = 'success'
  } catch (err) {
    if (err.response?.status === 404) {
      // Not found means connected successfully, just no results
      message.value = 'LDAP Connection successful'
      messageType.value = 'success'
    } else if (err.response) {
      message.value = `LDAP Connection failed: ${err.response.data.error || err.response.statusText}`
      messageType.value = 'error'
    } else if (err.request) {
      message.value = 'No response from server'
      messageType.value = 'error'
    } else {
      message.value = err.message
      messageType.value = 'error'
    }
  }
}

async function testOidc() {
  message.value = ''
  messageType.value = ''
  try {
    await axios.post('/ui/auth/oidc/test')
    message.value = 'Issuer URL reachable'
    messageType.value = 'success'
  } catch (err) {
    if (err.response) {
      message.value = `OIDC check failed: ${err.response.data.error || err.response.statusText}`
      messageType.value = 'error'
    } else if (err.request) {
      message.value = 'No response from server'
      messageType.value = 'error'
    } else {
      message.value = err.message
      messageType.value = 'error'
    }
  }
}

onMounted(async () => {
  selectedType.value = 'ldap'
})
</script>

<style scoped>
.external-auth {
  flex: 1;
  display: grid;
  grid-template-columns: 250px 1fr;
  padding: 1rem;
  gap: 1rem;
  width: 100%;
  background-color: var(--container-modal);
  border-radius: var(--small-radius);
  height: 100%;
}

.auth-selector {
  background-color: var(--low-glare);
  outline: none;
}

.sidebar {
  display: flex;
  flex-direction: column;
  border-radius: var(--small-radius);
  padding: 1rem;
  background: var(--container-modal);
  gap: 1rem;
}

.config-panel {
  display: flex;
  flex-direction: column;
  padding: 1rem;
  box-sizing: border-box;
  color: var(--low-glare);
  background: var(--container-modal);
  border-radius: var(--small-radius);
  overflow: auto;
  height: 100%;
}

.actions {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.actions button {
  display: block;
  width: 100%;
}

select {
  border-radius: var(--small-radius);
}

.result {
  min-height: 1.2rem;
  font-size: 1rem;
  text-align: center;
}

.result.error {
  color: var(--error-msg);
}

.result.success {
  color: var(--success-msg);
}

.config-panel input[type='checkbox'] {
  width: var(--input-height);
  margin: 0;
  display: inline-block;
  margin-bottom: 1rem;
}

.checkbox-row input {
  width: auto; /* override global width: 100% */
}

.config-panel input {
  width: 500px;
  display: block;
  margin-bottom: 1rem;
}
</style>
