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
          <td><span :title="server.server_names">{{ server.server_count }}</span></td>
          <td>{{ server.expiry ? new Date(server.expiry * 1000).toLocaleString() : '-' }}</td>
          <td>{{ server.current_user || '-' }}</td>
          <td>
            <div class="controls-container">
              <input v-model="duration" placeholder="eg 3h" :disabled="server.current_user && server.current_user !== userState.email" />
                <!-- Nobody using session -->
                <button v-if="!server.current_user" @click="startServerSession(server.server_group, duration)">Start Session</button>

                <!-- Current session is mine -->
                <template v-else-if="server.current_user === userState.email">
                  <button @click="updateServerSession(server.server_group, duration)">Extend Session</button>
                  <button @click="updateServerSession(server.server_group, '0m')">End Session</button>
                </template>

                <!-- Session is someone else -->
                <button v-else disabled>In Use</button>
            </div>
          </td>
        </tr>
      </tbody>
    </table>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import axios from 'axios'
import { userState } from '@/user.js'

const servers = ref([]) // Makes 'servers' reactive

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
async function startServerSession(serverGroup, duration) {
  try {
    const response = await axios.post('/ui/session/new', {
      server_group: serverGroup,
      duration: duration
    })

    if (response.data.success) {
      loadServerSessions() // refresh table after creating session
    }
  } catch (err) {
    console.error('Error starting server session:', err)
  }
}

// Update server session
async function updateServerSession(serverGroup, duration) {
  try {
    const response = await axios.put('/ui/session/update', {
      server_group: serverGroup,
      duration: duration
    })

    if (response.data.success) {
      loadServerSessions() // refresh table after creating session
    }
  } catch (err) {
    console.error('Error updating server session:', err)
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

.controls-container {
  display: flex;
  gap: 10px;
}
</style>
