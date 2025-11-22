<template>
<p>User Management</p>
  <div class="user-mgmt-container">
    <div class="user-mgmt-container">
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
          <th>Delete User</th>
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
import { ref, onMounted } from 'vue'
import axios from 'axios'
import { useUserStore } from '@/stores/user'
const user = useUserStore()

const users = ref([])
const currentUserId = ref(user.userID)
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

</script>

<style scoped>
p {
    color: var(--low-glare)
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