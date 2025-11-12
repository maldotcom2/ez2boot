import { reactive } from 'vue'

// Store for currently logged in user parameters
export const userState = reactive({
  email: null,
  isAdmin: false,
})
