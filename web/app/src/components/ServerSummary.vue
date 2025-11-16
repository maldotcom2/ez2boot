<template>
  <div class="summary-container">
    <!-- Refresh button -->
    <button @click="loadServerSessions" style="margin-bottom: 1rem;">Refresh</button>

    <table class="server-table">
      <thead>
        <tr>
          <th>Group Name</th>
          <th>Server Count</th>
          <th>Time Remaining</th>
          <th>Expiry</th>
          <th>Current User</th>
          <th>Action</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="server in servers" :key="server.server_group">
          <td>{{ server.server_group }}</td>
          <td><span :title="server.server_names">{{ server.server_count }}</span></td>
          <td>{{ server.expiry ? formatTimeRemaining(Math.floor((server.expiry - Math.floor(Date.now() / 1000)) / 60)): '-' }}</td>
          <td>{{ server.expiry ? new Date(server.expiry * 1000).toLocaleString() : '-' }}</td>
          <td>{{ server.current_user || '-' }}</td>
          <td>
            <div class="controls-container">
              <input v-model="duration[server.server_group]" placeholder="eg 3h" :disabled="server.current_user && server.current_user !== userState.email" />
                <!-- Start Session enabled if nobody is using it -->
              <button @click="startServerSession(server.server_group)"
              :disabled="!!server.current_user || !duration[server.server_group]">Start Session</button>

              <!-- Extend Session enabled for current user -->
              <button @click="updateServerSession(server.server_group)"
              :disabled="server.current_user !== userState.email || !duration[server.server_group]">Update Session</button>

              <!-- End Session enabled for current user-->
              <button @click="endServerSession(server.server_group)"
              :disabled="server.current_user !== userState.email">End Session</button>
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

const servers = ref([])
const duration = ref({})

// Load table data from specialised endpoint
async function loadServerSessions() {
  try {
    const response = await axios.get('/ui/session/summary')
    if (response.data.success) {
        servers.value = response.data.data
    }
    duration.value = {} // empty the input

    } catch (err) {
    console.error('Error loading server sessions:', err)
  }
}

// Start a new server session
async function startServerSession(serverGroup) {
  if (!validateDuration(duration.value[serverGroup])) {
    return
  }

  try {
    const response = await axios.post('/ui/session/new', {
      server_group: serverGroup,
      duration: duration.value[serverGroup]
    })

    if (response.data.success) {
      duration.value[serverGroup] = ''
      loadServerSessions() // refresh table after creating session
    }
  } catch (err) {
    console.error('Error starting server session:', err)
  }
}

// Update server session
async function updateServerSession(serverGroup) {
  if (!validateDuration(duration.value[serverGroup])) {
    return
  }

  try {
    const response = await axios.put('/ui/session/update', {
      server_group: serverGroup,
      duration: duration.value[serverGroup]
    })

    if (response.data.success) {
      loadServerSessions() // refresh table after creating session
    }
  } catch (err) {
    console.error('Error updating server session:', err)
  }
}

async function endServerSession(serverGroup) {
  try {
    const response = await axios.put('/ui/session/update', {
      server_group: serverGroup,
      duration: '0m'
    })

    if (response.data.success) {
      loadServerSessions() // refresh table after creating session
    }
  } catch (err) {
    console.error('Error updating server session:', err)
  }
}

function formatTimeRemaining(minutesRemaining) {
  if (minutesRemaining <= 1) {
    return '< 1 minute'
  }

  else if (minutesRemaining < 0) {
    return 'expired'
  }

  else return `${minutesRemaining} minutes`
}

function validateDuration(duration) {
  const regex = /^\d+(m|h)$/
  return regex.test(duration) // true
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

.controls-container button {
  width: 140px;
}

.controls-container input {
  width: 50px;
}

</style>
