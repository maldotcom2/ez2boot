<template>
  <div class="summary-container">
    <!-- Refresh button -->
    <button @click="loadServerSessions" style="margin-bottom: 1rem;">Refresh</button>

    <table class="server-table">
      <thead>
        <tr>
          <th>Group Name</th>
          <th>Server Count</th>
          <th>Expiry</th>
          <th>Current User</th>
          <th>Action</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="server in servers" :key="server.server_group">
          <td>{{ server.server_group }}</td>
          <td>
            <span :title="server.server_names">{{ server.server_count }}</span>
          </td>
          <td>{{ server.expiry ? new Date(server.expiry * 1000).toLocaleString() : '-' }}</td>
          <td>{{ server.current_user || '-' }}</td>
          <td>
            <button @click="startServerSession(server.server_group)">Start Session</button>
          </td>
        </tr>
      </tbody>
    </table>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import axios from 'axios'

const servers = ref([])

// Load table data from specialised endpoint
async function loadServerSessions() {
  try {
    const response = await axios.get('/ui/session/summary')
    if (response.data.success) {
        servers.value = response.data.data
    }
    } catch (err) {
    console.error('Error loading server sessions:', err)
  }
}

// Start a new server session
async function startServerSession(serverGroup) {
  try {
    const response = await axios.post('/ui/session/new', {
      server_group: serverGroup,
      duration: '3h'
    })

    if (response.data.success) {
      loadServerSessions() // refresh table after creating session
    }
  } catch (err) {
    console.error('Error starting server session:', err)
  }
}

// Load table on page load
onMounted(() => {
  loadServerSessions()
})
</script>

<style>
.summary-container {
  background-color: var(--container-modal)
}

.server-table {
  color: var(--low-glare);
  border-collapse: collapse;
  width: 100%;
}

.server-table th,
.server-table td {
  border: 1px solid var(--low-glare);
  padding: 8px;
}

.server-table th {
  text-align: left;
}
</style>
