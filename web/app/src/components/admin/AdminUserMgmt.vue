<template>
  <div class="user-mgmt-container">
    <div class="user-btn-container">
      <button @click="createUser()">Create User</button>
      <button @click="saveChanges" :disabled="changedUsers.size === 0">Save Changes</button>
    </div>
    <table class="user-mgmt-table">
      <thead>
        <tr>
          <th>Email</th>
          <th>Active</th>
          <th>Admin</th>
          <th>API</th>
          <th>UI</th>
          <th>Last Login</th>
          <th>Action</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="user in users" :key="user.email">
          <td>{{ user.email }}</td>
          <td><input type="checkbox" v-model="user.is_active" @change="markChanged(user.user_id)" :disabled="user.user_id === currentUserId"/></td>
          <td><input type="checkbox" v-model="user.is_admin" @change="markChanged(user.user_id)" :disabled="user.user_id === currentUserId"/></td>
          <td><input type="checkbox" v-model="user.api_enabled" @change="markChanged(user.user_id)" :disabled="user.user_id === currentUserId"/></td>
          <td><input type="checkbox" v-model="user.ui_enabled" @change="markChanged(user.user_id)" :disabled="user.user_id === currentUserId"/></td>
          <td>{{ user.last_login ? new Date(user.last_login * 1000).toLocaleString() : '-' }}</td>
          <td><button @click="deleteUser(user.user_id)" :disabled="user.user_id === currentUserId">Delete User</button></td>
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

// Load table data from specialised endpoint
async function getUsers() {
  try {
    const response = await axios.get('/ui/users')
    if (response.data.success) {
        users.value = response.data.data
    }

  } catch (err) {
    console.error('Error loading users:', err)
  }
}

async function deleteUser(userID) {
  if (!confirm("Are you sure you want to delete this notification?")) {
    return
  }
  
  try {
    await axios.delete('/ui/user/delete',
      {
        data: { user_id: userID },
        withCredentials: true
      }
    )
    getUsers()
  } catch (err) {
    console.error('Error deleting user:', err)
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
  const payload = users.value
    .filter(u => changedUsers.value.has(u.user_id)) // only changed users
    .map(u => ({
      user_id: u.user_id,
      is_active: u.is_active,
      is_admin: u.is_admin,
      api_enabled: u.api_enabled,
      ui_enabled: u.ui_enabled
    }))
  try {
    await axios.post('/ui/user/auth/update', payload)
    changedUsers.value.clear()
  } catch (err) {
    console.error('Error saving users:', err)
  }
}

function createUser() {
  // Swap the right pane to Create User
  emit('switch-pane', AdminCreateUser)
}

</script>

<style scoped>
p {
    color: var(--low-glare)
}

.user-mgmt-container {
  background-color: var(--container-modal);
  padding: 12px;
  border-radius: var(--small-radius);
}

.user-btn-container {
  display: flex;
  margin-bottom: 5px;
  justify-content: right;
  gap: 5px;
}

.user-btn-container button {
  width: 130px;
}

.user-mgmt-table {
  color: var(--low-glare);
  border-collapse: collapse;
  width: 100%;
}

.user-mgmt-table th,
.user-mgmt-table td {
  border: 1px solid var(--low-glare);
  padding: 8px;
}

.user-mgmt-table th {
  text-align: left;
}

</style>