import { createRouter, createWebHistory } from 'vue-router'
import Login from '@/components/Login.vue'
import Dashboard from '@/components/Dashboard.vue'
import AdminPanel from '@/components/AdminPanel.vue'
import axios from 'axios'



const routes = [
  { path: '/', redirect: '/dashboard'}, // default route
  { path: '/login', component: Login},
  { path: '/dashboard', component: Dashboard, meta: { requiresAuth: true }}, // Protected route
  { path: '/adminpanel', component: AdminPanel, meta: {requiresAdmin: true, requiresAuth: true }} // Protected and Admin only
]

// Create router
const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes,
})

async function checkSession() {
  await axios.get('/ui/user/session', { withCredentials: true })
}

async function checkAdmin() {
  const response = await axios.get('/ui/user/auth', { withCredentials: true })
  return response.data.data.is_admin
}

router.beforeEach(async (to, from, next) => {
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
