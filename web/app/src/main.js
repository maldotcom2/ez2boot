import { createPinia } from 'pinia'
import { createApp } from 'vue'
import App from './App.vue'
import router from './router'
import './assets/styles/global.css' // Global colours etc

const app = createApp(App)

app.use(router)
app.use(createPinia())
app.mount('#app')
