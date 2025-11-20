<template>
<p>User Management</p>
  <div class="user-mgmt-container">
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
          <td>{{ user.is_active }}</td>
          <td>{{ user.is_admin }}</td>
          <td>{{ user.api_enabled }}</td>
          <td>{{ user.ui_enabled }}</td>
          <td>{{ user.last_login ? new Date(user.last_login * 1000).toLocaleString() : '-' }}</td>
          <td>"Delete"</td>
        </tr>
      </tbody>
    </table>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import axios from 'axios'

const users = ref([])

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

onMounted(() => {
  getUsers()
})
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