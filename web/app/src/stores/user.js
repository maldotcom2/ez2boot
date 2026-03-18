import { defineStore } from 'pinia'
import axios from 'axios'

export const useUserStore = defineStore('user', {
  state: () => ({
    userID: null,
    email: null,
    isAdmin: false,
    identityProvider: null,
    hasMFA: false,
    loaded: false, // Stops re-fetch v
    error: null,
  }),

  actions: {
    async loadUser() {
      if (this.loaded) return // Stops re-fetch ^

      try {
        const response = await axios.get('ui/user/auth', { withCredentials: true })
        this.userID = response.data.data.user_id
        this.email = response.data.data.email
        this.isAdmin = response.data.data.is_admin
        this.identityProvider = response.data.data.identity_provider
        this.hasMFA = response.data.data.mfa_confirmed
        this.loaded = true
      } catch (err) {
        this.error = err?.response?.data?.error || err.message
        throw err
      }
    }
  }
})
