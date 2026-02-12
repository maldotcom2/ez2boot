<template>
  <div class="audit-log-container">
    <div class="user-btn-container">
      <button @click="applyFilters">Apply</button>
      <button @click="resetFilters">Reset</button>
    </div>
    <p class="result" :class="messageType">{{ message || '\u00A0' }}</p>
    <table class="audit-table" v-if="events.length">
      <thead>
        <tr>
          <th>Time</th>
          <th>Actor</th>
          <th>Target</th>
          <th>Action</th>
          <th>Resource</th>
          <th>Success</th>
          <th>Reason</th>
          <th>Metadata</th>
        </tr>

        <tr class="filter-row">
          <th>
            <div class="time-filter">
              <input
                type="date"
                v-model="fromDate"
                @change="setFromDate($event.target.value)"
                placeholder="From"
              />
              <input
                type="date"
                v-model="toDate"
                @change="setToDate($event.target.value)"
                placeholder="To"
              />
            </div>
          </th>

          <th>
            <input
              v-model="filters.actor_email"
              placeholder="Actor email"
            />
          </th>

          <th>
            <input
              v-model="filters.target_email"
              placeholder="Target email"
            />
          </th>

          <th>
            <input
              v-model="filters.action"
              placeholder="Action"
            />
          </th>

          <th>
            <input
              v-model="filters.resource"
              placeholder="Resource"
            />
          </th>

          <th>
            <select v-model="filters.success">
              <option :value="null">All</option>
              <option :value="true">✓</option>
              <option :value="false">✗</option>
            </select>
          </th>

          <th>
            <input
              v-model="filters.reason"
              placeholder="Reason"
            />
          </th>

          <th>
            <input
              v-model="filters.metadata"
              placeholder="Metadata"
            />
          </th>
        </tr>
      </thead>

      <tbody>
        <tr v-for="event in events" :key="event.TimeStamp">
          <td>{{ formatTime(event.TimeStamp) }}</td>
          <td>{{ event.ActorEmail }}</td>
          <td>{{ event.TargetEmail }}</td>
          <td>{{ event.Action }}</td>
          <td>{{ event.Resource }}</td>
          <td>
            <span :class="event.Success ? 'ok' : 'fail'">
              {{ event.Success ? 'Yes' : 'No' }}
            </span>
          </td>
          <td>{{ event.Reason }}</td>
          <td>
            <div v-if="hasMetadata(event.Metadata)" class="metadata-lines">
              <div
                v-for="(value, key) in event.Metadata"
                :key="key"
              >
                {{ key }}: {{ value }}
              </div>
            </div>
            <span v-else>-</span>
          </td>
        </tr>
      </tbody>
    </table>

    <div class="pagination-btn-container">
      <button @click="prevPage" :disabled="!canGoBack">Back</button>
      <button @click="nextPage" :disabled="!nextCursor">Next</button>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import axios from 'axios'

const events = ref([])
const nextCursor = ref(null)
const cursorStack = ref([])
const fromDate = ref('')
const toDate = ref('')
const canGoBack = computed(() => cursorStack.value.length > 0)
const message = ref('')
const messageType = ref('')

const filters = reactive({
  // Pagination
  limit: 20,

  // Filters
  actor_email: '',
  target_email: '',
  action: '',
  resource: '',
  success: null, // true | false | null
  reason: '',

  // Time range (unix seconds)
  from: null,
  to: null,
})

async function fetchAuditEvents(cursor = null) {
  message.value = ''
  messageType.value = ''

  try {
    const params = {
      limit: filters.limit,
      actor_email: filters.actor_email || undefined,
      target_email: filters.target_email || undefined,
      action: filters.action || undefined,
      resource: filters.resource || undefined,
      success: filters.success,
      reason: filters.reason || undefined,
      metadata: filters.metadata,
      from: filters.from || undefined,
      to: filters.to || undefined,
    }

    if (cursor) {
      params.before = cursor
    }

    const response = await axios.get('/ui/audit/events', {
      params,
      withCredentials: true
    })

    if (response.data.success) {
        events.value = response.data.data?.events || []
        nextCursor.value = response.data.data?.next_cursor || null
    }

    if (events.value.length === 0) {
      message.value = 'No audit events found'
      messageType.value = 'info'
    }

    if (cursor) {
      cursorStack.value.push(cursor)
    }

  } catch (err) {
    messageType.value = 'error'
    if (err.response) {
      // Get server response
      message.value = `Failed to get audit events: ${err.response.data.error || err.response.statusText}`
    } else if (err.request) {
      // No response
      message.value = 'No response from server'
    } else {
      // other errors
      message.value = err.message
    }
  }
}

async function applyFilters() {
  // Reset state
  cursorStack.value = []
  nextCursor.value = null
  events.value = []

  // Explicitly remove cursor
  filters.before = null

  // Fetch fresh
  await fetchAuditEvents()
}

async function resetFilters() {
  // Reset filters to defaults
  filters.actor_email = ''
  filters.target_email = ''
  filters.action = ''
  filters.resource = ''
  filters.success = null
  filters.reason = ''
  filters.metadata = ''
  filters.from = null
  filters.to = null

  // Reset date inputs
  fromDate.value = ''
  toDate.value = ''

  // Reset state
  cursorStack.value = []
  nextCursor.value = null
  events.value = []

  // Explicitly remove cursor
  filters.before = null

  // Fetch fresh
  await fetchAuditEvents()
}

async function nextPage() {
  if (!nextCursor.value) return
  await fetchAuditEvents(nextCursor.value)
}

async function prevPage() {
  cursorStack.value.pop()
  const prev = cursorStack.value.pop() || null
  await fetchAuditEvents(prev)
}

function setFromDate(dateStr) {
  if (!dateStr) {
    filters.from = null
    return
  }
  const date = new Date(dateStr)
  date.setHours(0, 0, 0, 0) // start of day local time
  filters.from = Math.floor(date.getTime() / 1000)
}

function setToDate(dateStr) {
  if (!dateStr) {
    filters.to = null
    return
  }
  const date = new Date(dateStr)
  date.setHours(23, 59, 59, 999) // end of day local time
  filters.to = Math.floor(date.getTime() / 1000)
}


function formatTime(ts) {
  return new Date(ts * 1000).toLocaleString()
}

function hasMetadata(metadata) {
  return (
    metadata &&
    typeof metadata === 'object' &&
    Object.keys(metadata).length > 0
  )
}

onMounted(() => {
  fetchAuditEvents()
})

</script>

<style scoped>
.audit-log-container {
  background-color: var(--container-modal);
  max-width: 100%;
  overflow-x: auto;
  font-size: small;
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

.pagination-btn-container {
  display: flex;
  margin-top: 1rem;
  justify-content: right;
  gap: 1rem;
}

.pagination-btn-container button {
  width: 130px;
}

.audit-table {
  color: var(--low-glare);
  border-collapse: collapse;
  width: 100%;
  table-layout: fixed;
}

.audit-table th,
.audit-table td {
  border: 1px solid var(--low-glare);
  padding: 0.5rem;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.audit-table th {
  text-align: left;
}

.filter-row input {
  max-width: 90%;
}

.audit-table th:nth-child(1) { width: 15%; } /* Time */
.audit-table th:nth-child(2) { width: 15%; } /* Author */
.audit-table th:nth-child(3) { width: 15%; } /* Target */
.audit-table th:nth-child(4) { width: 10%; } /* Action */
.audit-table th:nth-child(5) { width: 10%; } /* Resource */
.audit-table th:nth-child(6) { width: 5%; } /* Success */
.audit-table th:nth-child(7) { width: 15%; } /* Reason */
.audit-table th:nth-child(8) { width: 15%; } /* Metadata */

.time-filter {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.metadata-lines {
  white-space: normal;
  word-break: break-word;
  max-width: 100%;
}

.result {
  min-height: 1.2rem;
  font-size: 1rem;
  text-align: center;
}

.result.info {
  color: var(--low-glare);
}

.result.error {
  color: var(--error-msg);
}

.result.success {
  color: var(--success-msg);
}


</style>