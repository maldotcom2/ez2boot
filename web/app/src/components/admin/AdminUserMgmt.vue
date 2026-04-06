<template>
  <div class="user-mgmt-container">
    <div class="user-btn-container">
      <button @click="createUser()">Add User</button>
      <button @click="saveChanges" :disabled="changedUsers.size === 0">Save Changes</button>
    </div>
    <p class="result" :class="messageType">{{ message || '\u00A0' }}</p>
    <table class="user-mgmt-table">
      <thead>
        <tr>
          <th>Email</th>
          <th>Active</th>
          <th>Admin</th>
          <th>API</th>
          <th>UI</th>
          <th>IDP</th>
          <th>Last Login</th>
          <th>Action</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="user in users" :key="user.email">
          <td>{{ user.email }}</td>
          <td>
            <input
              class="checkbox"
              type="checkbox"
              v-model="user.is_active"
              @change="markChanged(user.user_id)"
              :disabled="user.user_id === currentUserId"
            />
          </td>
          <td>
            <input
              class="checkbox"
              type="checkbox"
              v-model="user.is_admin"
              @change="markChanged(user.user_id)"
              :disabled="user.user_id === currentUserId"
            />
          </td>
          <td>
            <input
              v-if="user.identity_provider === 'local'"
              class="checkbox"
              type="checkbox"
              v-model="user.api_enabled"
              @change="markChanged(user.user_id)"
              :disabled="user.user_id === currentUserId"
            />
          </td>
          <td>
            <input
              class="checkbox"
              type="checkbox"
              v-model="user.ui_enabled"
              @change="markChanged(user.user_id)"
              :disabled="user.user_id === currentUserId"
            />
          </td>
          <td>{{ user.identity_provider }}</td>
          <td>{{ user.last_login ? new Date(user.last_login * 1000).toLocaleString() : '-' }}</td>
          <td>
            <button @click="deleteUser(user.user_id)" :disabled="user.user_id === currentUserId">
              Delete User
            </button>
          </td>
        </tr>
      </tbody>
    </table>
  </div>
</template>

<script setup>
import { computed, ref, onMounted } from 'vue'
import axios from 'axios'
import { useUserStore } from '@/stores/user'
import AdminCreateUser from './AdminCreateUser.vue'

const emit = defineEmits(['switch-pane'])

const user = useUserStore()
const users = ref([])
const currentUserId = computed(() => user.userID)
const changedUsers = ref(new Set())
const message = ref('')
const messageType = ref('')

// Load table data from specialised endpoint
async function getUsers() {
  message.value = ''
  messageType.value = ''

  try {
    const response = await axios.get('/ui/users')
    if (response.data.success) {
      users.value = response.data.data
    }
  } catch (err) {
    messageType.value = 'error'
    if (err.response) {
      // Get server response
      message.value = `Failed to get users: ${err.response.data.error || err.response.statusText}`
    } else if (err.request) {
      // No response
      message.value = 'No response from server'
    } else {
      // other errors
      message.value = err.message
    }
  }
}

async function deleteUser(userID) {
  message.value = ''
  messageType.value = ''

  if (!confirm('Are you sure you want to delete this user?')) {
    return
  }

  try {
    await axios.delete('/ui/user', {
      data: { user_id: userID },
    })

    getUsers()

    message.value = 'User deleted'
    messageType.value = 'success'
  } catch (err) {
    messageType.value = 'error'
    if (err.response) {
      // Get server response
      message.value = `Error: ${err.response.data.error || err.response.statusText}`
    } else if (err.request) {
      // No response
      message.value = 'No response from server'
    } else {
      // other errors
      message.value = err.message
    }
  }
}

onMounted(async () => {
  await getUsers()
  await user.loadUser()
})

// Called on checkbox change
function markChanged(userId) {
  changedUsers.value.add(userId)
}

// Save only changed users
async function saveChanges() {
  message.value = ''
  messageType.value = ''

  const payload = users.value
    .filter((u) => changedUsers.value.has(u.user_id)) // only changed users
    .map((u) => ({
      user_id: u.user_id,
      is_active: u.is_active,
      is_admin: u.is_admin,
      api_enabled: u.api_enabled,
      ui_enabled: u.ui_enabled,
    }))
  try {
    await axios.put('/ui/user/auth', payload)
    changedUsers.value.clear()

    message.value = 'Auth changes saved'
    messageType.value = 'success'
  } catch (err) {
    messageType.value = 'error'
    if (err.response) {
      // Get server response
      message.value = `Failed to update user auth: ${err.response.data.error || err.response.statusText}`
    } else if (err.request) {
      // No response
      message.value = 'No response from server'
    } else {
      // other errors
      message.value = err.message
    }
  }
}

function createUser() {
  message.value = ''
  messageType.value = ''

  // Swap the right pane to Create User
  emit('switch-pane', AdminCreateUser)
}
</script>

<style scoped>
p {
  color: var(--low-glare);
}

.user-mgmt-container {
  background-color: var(--container-modal);
  padding: 1rem;
  border-radius: var(--small-radius);
}

.user-btn-container {
  display: flex;
  margin-bottom: 1rem;
  justify-content: right;
  gap: 1rem;
}

.user-btn-container button {
  width: 130px;
}

.user-mgmt-table {
  color: var(--low-glare);
  border-collapse: collapse;
  width: 100%;
  table-layout: fixed;
}

.user-mgmt-table th,
.user-mgmt-table td {
  border: 1px solid var(--low-glare);
  padding: 0.5rem;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.user-mgmt-table th {
  text-align: left;
}

.user-mgmt-table th:nth-child(1) {
  width: 24%;
} /* Email */
.user-mgmt-table th:nth-child(2) {
  width: 8%;
} /* Active */
.user-mgmt-table th:nth-child(3) {
  width: 8%;
} /* Admin */
.user-mgmt-table th:nth-child(4) {
  width: 8%;
} /* API */
.user-mgmt-table th:nth-child(5) {
  width: 8%;
} /* UI */
.user-mgmt-table th:nth-child(6) {
  width: 8%;
} /* IDP */
.user-mgmt-table th:nth-child(7) {
  width: 21%;
} /* Last Login */
.user-mgmt-table th:nth-child(8) {
  width: 15%;
} /* Actions */

.result {
  min-height: 1.2rem;
  font-size: 1rem;
  text-align: right;
}

.result.error {
  color: var(--error-msg);
}

.result.success {
  color: var(--success-msg);
}
</style>
