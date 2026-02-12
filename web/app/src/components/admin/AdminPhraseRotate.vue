<template>
  <div class="rotate-phrase-container">
      <form class="rotate-phrase-form" @submit.prevent="rotatePhrase">
        <p class="warning">Warning! This feature will re-encrypt all user notification settings using the passphrase entered below.
            The environment variable must then be updated to this new passphrase and ez2boot restarted before notifications will work.
            If the new passphrase is lost before updating the environment variable, user notification settings cannot be recovered
            and users must re-apply their configurations.
            This is a one-shot. The passphrase can only be changed once before the environment variable must be updated. 
        </p>
        <div class="checkbox-row">
            <label for="acknowledge">I Understand</label>
            <input id="acknowledge" class="checkbox" type="checkbox" v-model="acknowledge"/>
        </div>
        <input type="text" :disabled="!acknowledge || messageType === 'success'" placeholder="New phrase" v-model="phrase"/>
        <button type="submit" :disabled="!phrase || !acknowledge || messageType === 'success'">Rotate</button>
        <p class="result" :class="messageType">{{ message || '\u00A0' }}</p>
      </form>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import axios from 'axios'

const phrase = ref('')
const acknowledge = ref(false)
const message = ref('')
const messageType = ref('')

async function rotatePhrase() {
  message.value = ''
  messageType.value = ''

  try {
    const response = await axios.post('ui/notification/rotate',
      {
        phrase: phrase.value,
      },
      {
        withCredentials: true // Cookies
      }
    )
    
    message.value = 'Encryption phrase rotated - set environment variable now!'
    messageType.value = 'success'
    console.log('Encryption phrase rotated:', response.data)

  } catch (err) {
    messageType.value = 'error'
    if (err.response) {
      // Get server response
      message.value = `Encryption phrase rotation failed: ${err.response.data.error || err.response.statusText}`
    } else if (err.request) {
      // No response
      message.value = 'No response from server'
    } else {
      // other errors
      message.value = err.message
    }
  }
}
</script>

<style scoped>
    
.rotate-phrase-container {
  background-color: var(--container-modal);
  width: 100%;
  min-width: 700px;
  overflow-x: auto;
  padding: 1rem;
  border-radius: var(--small-radius);
}

.rotate-phrase-form {
  color: var(--low-glare);
  display: flex;
  flex-direction: column;
  background-color: var(--container-modal);
  justify-content: center;
  align-items: center;
  padding: 3rem;
  gap: 1rem;
  border-radius: var(--big-radius);
}

.rotate-phrase-form input[type="text"],
.rotate-phrase-form button {
  width: 50%;
}

.checkbox-row {
  display: flex;
  flex-direction: row;
  width: 100%;
  align-items: center;
  justify-content: flex-start;
  gap: 1rem;
}

.checkbox-row input[type="checkbox"] {
  width: var(--input-height);
}

.warning {
    color: var(--warn-amber);
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
