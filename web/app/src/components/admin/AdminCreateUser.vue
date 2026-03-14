<template>
  <div class="centre-container">
    <div class="add-user-container">
      <p class="prompt">Create User</p>
      <div class="toggle">
        <button @click="mode = 'local', resetState()">Local</button>
        <button @click="mode = 'ldap', resetState()">LDAP</button>
      </div>
      <form v-if="mode === 'local'" @submit.prevent="createUser">
        <input v-model="email" placeholder="Email" />
        <input v-model="password" type="password" placeholder="Password" />
        <input v-model="confirmPassword" type="password" placeholder="Confirm Password" />
        <button type="submit" :disabled="!passwordsMatch || !email || !password || !confirmPassword">Create</button>
      </form>
      <form v-else @submit.prevent="searchLdap">
        <input v-model="ldapQuery" placeholder="Search by email" />
        <button type="submit" :disabled="!ldapQuery">Search</button>
        <div v-if="ldapResult" class="ldap-result">
          <p class="ldap-email">{{ ldapResult.Email }}</p>
        </div>
        <button v-if="ldapResult" @click="provisionLdapUser">Add User</button>
      </form>
      <p class="result" :class="messageType">{{ validationMessage || message || '\u00A0' }}</p>
    </div>
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'
import axios from 'axios'
import AdminUserMgmt from './AdminUserMgmt.vue'

const emit = defineEmits(['switch-pane'])
const message = ref('')
const messageType = ref('')
const mode = ref('local')

// Local
const email = ref('')
const password = ref('')
const confirmPassword = ref('')
const passwordsMatch = computed(() => password.value === confirmPassword.value)
const validationMessage = computed(() => {
  if (confirmPassword.value.length > 0 && !passwordsMatch.value) {
    messageType.value = 'error'
    return 'Passwords do not match'
  }
  return ''
})

// LDAP
const ldapQuery = ref('')
const ldapResult = ref(null)

function resetState() {
  message.value = ''
  messageType.value = ''
  email.value = ''
  password.value = ''
  confirmPassword.value = ''
  ldapQuery.value = ''
  ldapResult.value = null
}

async function createUser() {
  message.value = ''
  messageType.value = ''

  try {
    const response = await axios.post('ui/user',
      {
        email: email.value,
        password: password.value
      },
      {
        withCredentials: true // Cookies
      }
    )

    if (response.data.success) {
      message.value = 'User created'
      messageType.value = 'success'
      emit('switch-pane', AdminUserMgmt)
    }
  } catch (err) {
    messageType.value = 'error'
    if (err.response) {
      // Get server response
      message.value = `Create user failed: ${err.response.data.error || err.response.statusText}`
    } else if (err.request) {
      // No response
      message.value = 'No response from server'
    } else {
      // other errors
      message.value = err.message
    }
  }
}

async function searchLdap() {
  message.value = ''
  messageType.value = ''
  ldapResult.value = null

  try {
    const response = await axios.post('ui/auth/ldap/users/search',
      {
        query: ldapQuery.value
      },
      {
        withCredentials: true // Cookies
      })

    if (response.data.success) {
      ldapResult.value = response.data.data
    }
  } catch (err) {
    messageType.value = 'error'
      if (err.response?.status === 404) {
      message.value = 'User not found in directory'
    } else if (err.response) {
      // Get server response
      message.value = `Search failed: ${err.response.data.error || err.response.statusText}`
    } else if (err.request) {
      // No response
      message.value = 'No response from server'
    } else {
      // other errors
      message.value = err.message
    }
  }
}

async function provisionLdapUser() {
  message.value = ''
  messageType.value = ''
  try {
    await axios.post('/ui/user/ldap',
    { 
      email: ldapResult.value.Email 
    },
    { 
      withCredentials: true // Cookies
    })

    message.value = 'User added'
    messageType.value = 'success'
    emit('switch-pane', AdminUserMgmt)
  } catch (err) {
    messageType.value = 'error'
    if (err.response?.status === 409) {
      message.value = 'User already exists'
      // Get server response
    } else if (err.response) {
      message.value = `Failed to add user: ${err.response.data.error || err.response.statusText}`
      // No response
    } else if (err.request) {
      message.value = 'No response from server'
      // other errors
    } else {
      message.value = err.message
    }
  }
}
</script>

<style scoped>
.centre-container {
  display: flex;
  justify-content: center;
  align-items: center;
  height: 100%;
  outline: none;
}

.add-user-container {
  display: flex;
  flex-direction: column;
  background-color: var(--container-modal);
  justify-content: center;
  align-items: center;
  padding: 3rem;
  height: 400px;
  width: 300px;
  gap: 1rem;
  border-radius: var(--big-radius);
  outline: auto;
}

.add-user-container button{
  width: 100%;
}

form {
  display: flex;
  flex-direction: column;
  width: 100%;
  flex: 1;
  gap: 1rem;
}

.toggle {
  display: flex;
  width: 100%;
  gap: 0;
}

.toggle button:first-child {
  border-radius: var(--small-radius) 0 0 var(--small-radius);
}

.toggle button:last-child {
  border-radius: 0 var(--small-radius) var(--small-radius) 0;
}

.ldap-result {
  width: 100%;
  border: 1px solid var(--low-glare);
  border-radius: var(--small-radius);
  box-sizing: border-box;
  padding: 0.5rem;
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
  color: var(--low-glare);
}

.ldap-email {
  font-size: 0.85rem;
  opacity: 0.7;
}

.prompt {
  color: var(--low-glare);
  font-size: 1.2rem;
}

input {
  width: 100%;
}

button {
  width: 100%;
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

</style>
