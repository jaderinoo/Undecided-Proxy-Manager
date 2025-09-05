import { createApp } from 'vue'
import { createPinia } from 'pinia'
import { createRouter, createWebHistory } from 'vue-router'
import { createVuetify } from 'vuetify'
import * as components from 'vuetify/components'
import * as directives from 'vuetify/directives'
import { aliases, mdi } from 'vuetify/iconsets/mdi'
import '@mdi/font/css/materialdesignicons.css'
import 'vuetify/styles'
import './styles/main.scss'
import App from './App.vue'
import LoginPage from './pages/LoginPage.vue'
import Dashboard from './pages/Dashboard.vue'
import DNSPage from './pages/DNSPage.vue'
import ContainersPage from './pages/ContainersPage.vue'
import ProxiesPage from './pages/ProxiesPage.vue'
import CertificatesPage from './pages/CertificatesPage.vue'
import SettingsPage from './pages/SettingsPage.vue'
import { useAuthStore } from './stores/auth'

// Vuetify configuration
const vuetify = createVuetify({
  components,
  directives,
  icons: {
    defaultSet: 'mdi',
    aliases,
    sets: {
      mdi,
    },
  },
  theme: {
    defaultTheme: 'light',
    themes: {
      light: {
        colors: {
          primary: '#1976D2',
          secondary: '#424242',
          accent: '#82B1FF',
          error: '#FF5252',
          info: '#2196F3',
          success: '#4CAF50',
          warning: '#FFC107',
          surface: '#FFFFFF',
          background: '#FAFAFA',
          'on-primary': '#FFFFFF',
          'on-secondary': '#FFFFFF',
          'on-surface': '#212121',
          'on-background': '#212121',
        },
      },
      dark: {
        colors: {
          primary: '#64B5F6',
          secondary: '#B0BEC5',
          accent: '#1976D2',
          error: '#EF5350',
          info: '#29B6F6',
          success: '#66BB6A',
          warning: '#FFB74D',
          surface: '#303030',
          background: '#121212',
          'on-primary': '#000000',
          'on-secondary': '#000000',
          'on-surface': '#FFFFFF',
          'on-background': '#FFFFFF',
        },
      },
    },
  },
})

// Router configuration
const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: '/login',
      name: 'Login',
      component: LoginPage,
      meta: { requiresGuest: true }
    },
    {
      path: '/',
      name: 'Dashboard',
      component: Dashboard,
      meta: { requiresAuth: true }
    },
    {
      path: '/dashboard',
      redirect: '/'
    },
    {
      path: '/proxies',
      name: 'Proxies',
      component: ProxiesPage,
      meta: { requiresAuth: true }
    },
    {
      path: '/certificates',
      name: 'Certificates',
      component: CertificatesPage,
      meta: { requiresAuth: true }
    },
    {
      path: '/containers',
      name: 'Containers',
      component: ContainersPage,
      meta: { requiresAuth: true }
    },
    {
      path: '/dns',
      name: 'DNS',
      component: DNSPage,
      meta: { requiresAuth: true }
    },
    {
      path: '/settings',
      name: 'Settings',
      component: SettingsPage,
      meta: { requiresAuth: true }
    }
  ]
})

// Navigation guards
router.beforeEach((to, from, next) => {
  const authStore = useAuthStore()
  
  if (to.meta.requiresAuth && !authStore.isAuthenticated) {
    next('/login')
  } else if (to.meta.requiresGuest && authStore.isAuthenticated) {
    next('/')
  } else {
    next()
  }
})

const app = createApp(App)

app.use(createPinia())
app.use(vuetify)
app.use(router)
app.mount('#app')
