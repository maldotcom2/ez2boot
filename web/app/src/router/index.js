import { createRouter, createWebHistory } from 'vue-router'
import Setup from '@/components/Setup.vue'
import Login from '@/components/login.vue'
import AdminPanel from '@/components/admin/AdminPanel.vue'
import Dashboard from '@/components/Dashboard.vue'
import UserSettings from '@/components/user/UserSettings.vue'
import About from '@/components/About.vue'
import axios from 'axios'

const routes = [
  { path: '/', redirect: '/dashboard'}, // default route
  { path: '/setup', component: Setup}, // only for setup bootstrap
  { path: '/login', component: Login},
  { path: '/adminpanel', component: AdminPanel, meta: {requiresAdmin: true, requiresAuth: true }}, // Protected and Admin only
  { path: '/dashboard', component: Dashboard, meta: { requiresAuth: true }}, // Protected route
  { path: '/settings', component: UserSettings, meta: { requiresAuth: true }}, // Protected route
  { path: '/about', component: About, meta: { requiresAuth: true }}, // Protected route
]

// Create router
const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes,
})

// User session validity
async function checkSession() {
  await axios.get('/ui/user/session', { withCredentials: true })
}

// Check if user is admin
async function checkAdmin() {
  const response = await axios.get('/ui/user/auth', { withCredentials: true })
  return response.data.data.is_admin
}

// Check if the app requires first user bootstrap
async function checkMode() {
  const response = await axios.get('/ui/mode')
  return response.data.data.setup_mode
}

router.beforeEach(async (to, from, next) => {
   const setupMode = await checkMode()

   // Setup only
  if (setupMode && to.path !== '/setup') {
    return next('/setup')
  }

  // Block access
  if (!setupMode && to.path === '/setup') {
    return next('/dashboard')
  }
  
  // Skip for unprotected routes
  if (!to.meta.requiresAuth) {
    return next()
  }

  try {

    // Check if session is still valid
    await checkSession()

    // Check authorisation
    if (to.meta.requiresAdmin) {
      const isAdmin = await checkAdmin()
      if (!isAdmin) {
        return next('/dashboard') // Redirect home
      }
    }

    next() // All checks passed
  } catch (err) {
    if (err.response?.status === 401) return next('/login')
    console.error('Auth check failed', err)
    next('/login')
  }
})

export default router
