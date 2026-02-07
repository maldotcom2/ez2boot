import { defineStore } from 'pinia'
import axios from 'axios'

export const useVersionStore = defineStore('version', {
  state: () => ({
    version: null,
    buildDate: null,
    updateAvailable: false,
    latestVersion: null,
    checkedAt: null,
    releaseURL: null,
    loaded: false, // Stops re-fetch v
    error: null,
  }),

  actions: {
    async getVersion() {
      if (this.loaded) return // Stops re-fetch ^

      try {
        const response = await axios.get('ui/version', { withCredentials: true })
        this.version = response.data.data.version
        this.buildDate = response.data.data.build_date
        this.updateAvailable = response.data.data.update_available
        this.latestVersion = response.data.data.latestVersion
        this.checkedAt = response.data.data.checked_at
        this.releaseURL = response.data.data.release_url
        this.loaded = true
      } catch (err) {
        this.error = err?.response?.data?.error || err.message
        throw err
      }
    }
  }
})
