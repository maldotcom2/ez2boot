import { reactive } from 'vue'

// Store for currently logged in user parameters
export const userState = reactive({
  userID: 0,
  email: null,
  isAdmin: false,
})
