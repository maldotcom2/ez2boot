<template>
  <div class="summary-container">
    <div class="user-btn-container">
      <button @click="loadServerSessions">Refresh</button>
    </div>
    <table class="server-table">
      <thead>
        <tr>
          <th>Group Name</th>
          <th>State</th>
          <th>Time Remaining</th>
          <th>Expiry</th>
          <th>Current User</th>
          <th>Action</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="server in servers" :key="server.server_group">
          <td>{{ server.server_group }}</td>
          <td>
            <div class="status-container">
              <span :class="'status-dot ' + getGroupState(server.servers)"></span>
              <button @click="openServerModal($event, server.servers, server.server_group)">Details</button>
            </div>
          </td>
          <td>{{ server.expiry ? formatTimeRemaining(Math.floor((server.expiry - Math.floor(Date.now() / 1000)) / 60)): '-' }}</td>
          <td>{{ server.expiry ? new Date(server.expiry * 1000).toLocaleString() : '-' }}</td>
          <td>{{ server.current_user || '-' }}</td>
          <td>
            <div class="controls-container">
              <input v-model.number="duration[server.server_group]" type="number" min="1" max="24" step="1" placeholder="hours" :disabled="server.current_user && server.current_user !== user.email" />
                <!-- Start Session enabled if nobody is using it -->
              <button @click="startServerSession(server.server_group)"
              :disabled="!!server.current_user || !duration[server.server_group]">Start Session</button>

              <!-- Extend Session enabled for current user -->
              <button @click="updateServerSession(server.server_group)"
              :disabled="server.current_user !== user.email || !duration[server.server_group]">Update Session</button>

              <!-- End Session enabled for current user-->
              <button @click="endServerSession(server.server_group)"
              :disabled="server.current_user !== user.email">End Session</button>
            </div>
          </td>
        </tr>
      </tbody>
    </table>
  </div>

  <!-- Server details modal -->
<div v-if="showModal" class="modal-backdrop" @click.self="closeModal">
  <div class="detail-modal" :style="{ top: modalPosition.top, left: modalPosition.left }">
    <span>Servers in {{ modalGroup }}</span>
    <ul>
      <li v-for="s in modalServers" :key="s.name" class="detail-modal-item">
        <span :class="['status-dot', s.state]"></span>
        {{ s.name }}
      </li>
    </ul>
    <button @click="closeModal">Close</button>
  </div>
</div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import axios from 'axios'
import { useUserStore } from '@/stores/user'

const user = useUserStore()
const servers = ref([])
const duration = ref({})
const showModal = ref(false)
const modalPosition = ref({ top: '0px', left: '0px' });
const modalServers = ref([])
const modalGroup = ref('')

// Load table data from specialised endpoint
async function loadServerSessions() {
  try {
    const response = await axios.get('/ui/session/summary')
    if (response.data.success) {
        servers.value = response.data.data
        console.log(response.data);
    }
    duration.value = {} // empty the input

    } catch (err) {
    console.error('Error loading server sessions:', err)
  }
}

// Start a new server session
async function startServerSession(serverGroup) {
  if (!validateDuration(duration.value[serverGroup])) {
    console.error("duration input invalid");
    return
  }

  try {
    const response = await axios.post('/ui/session/new', {
      server_group: serverGroup,
      duration: `${duration.value[serverGroup]}h`
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
      duration: `${duration.value[serverGroup]}h`
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
      duration: '0h'
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

function validateDuration(value) {
  return Number.isInteger(value) // true
}

function openServerModal(event, servers, groupName) {
  modalServers.value = servers || []
  modalGroup.value = groupName || (servers.length > 0 ? servers[0].server_group : '')
  showModal.value = true

  // Calculate position
  const rect = event.target.getBoundingClientRect();
  modalPosition.value.top = `${rect.top}px`; // top relative to viewport
  modalPosition.value.left = `${rect.right + 10}px`; // 10px to the right
}

function closeModal() {
  showModal.value = false
  modalServers.value = []
  modalGroup.value = ''
}

function getGroupState(serversList) {
  if (serversList.every(s => s.state === 'on')) 
    return 'on';
  else if (serversList.every(s => s.state === 'off')) 
    return 'off';
  else
    return 'transitioning';
}

// Load table on page load
onMounted(async () => {
  try {
    await user.loadUser()
  } catch (err) {
    console.error("Failed to load user", user.error)
  }

  try {
    await loadServerSessions()
  } catch (err) {
    console.error("Failed to load server sessions")
  }
  
})
</script>

<style>
.summary-container {
  margin: 1rem;
  padding: 1rem;
  background-color: var(--container-modal);
  border-radius: var(--small-radius);
}

.server-table {
  color: var(--low-glare);
  border-collapse: collapse;
  width: 100%;
  table-layout: fixed;
}

.server-table th,
.server-table td {
  border: 1px solid var(--low-glare);
  padding: 0.5rem;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.server-table th {
  text-align: left;
}

.server-table th:nth-child(1) { width: 10%; } /* Group Name */
.server-table th:nth-child(2) { width: 10%; } /* State */
.server-table th:nth-child(3) { width: 10%; } /* Time Remaining */
.server-table th:nth-child(4) { width: 15%; } /* Expiry */
.server-table th:nth-child(5) { width: 15%; } /* Current User */
.server-table th:nth-child(6) { width: 50%; } /* Actions */


.user-btn-container {
  display: flex;
  justify-content: right;
  margin-bottom: 1rem;
}

.user-btn-container button {
  width: 130px;
}

.controls-container {
  display: flex;
  gap: 1rem;
}

.controls-container button {
  width: 140px;
}

.controls-container input {
  width: 50px;
}

.modal-backdrop {
  position: fixed;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  z-index: 999; /* below modal */
}

.detail-modal {
  position: fixed;
  color: var(--low-glare);
  background: var(--container-modal);
  border: 1px var(--low-glare);
  padding: 1rem;
  border-radius: var(--small-radius);
  z-index: 1000;
  min-width: 100px;
}

.detail-modal button {
  display: block;
  margin: 1rem auto 0;
  padding: 0rem 1rem;
  height: 18px;
}

.detail-modal ul {
  padding: 0;
  margin: 0;
  margin-top: 10px;
  list-style: none; 
  text-align: left;
}

.detail-modal-item {
  display: flex;
  align-items: center;
  justify-content: start;
  margin-bottom: 4px;
}

.status-container {
  display: inline-flex;
  align-items: center;
}

.status-dot {
  display: inline-block;
  width: 12px;
  height: 12px;
  border-radius: 50%;
  border: 2px solid #666;
  margin-right: 16px;
}

.status-dot.on {
  background-color: #4caf50;
}

.status-dot.off {
  background-color: #f44336;
}

.status-dot.transitioning {
  background-color: #ffc107;
}

</style>
