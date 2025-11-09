import { createRouter, createWebHistory } from 'vue-router'
import Login from '@/components/Login.vue'
import Dashboard from '@/components/Dashboard.vue'
import axios from 'axios'

const routes = [
  { path: '/login', component: Login },
  { 
    path: '/dashboard', 
    component: Dashboard,
    meta: { requiresAuth: true } // protect this route
  },
]

// Create router
const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes,
})

// Routing rules
router.beforeEach(async (to, from, next) => {
  if (!to.meta.requiresAuth) {
    return next() // Unprotected routes
  }

  try {
    // Check backend for valid session
    await axios.get('/ui/user/session', { withCredentials: true })  // Pass the cookie
    next() // Valid session
  } catch (err) {
    if (err.response && err.response.status === 401) {
      next('/login') // not logged in, redirect
    } else {
      console.error('Auth check failed', err)
      next('/login')
    }
  }
})

export default router
