<template>
  <AppLayout @refresh="loadProxies" @logout="handleLogout">
    <v-container>
      <v-row>
        <v-col cols="12">
          <v-card>
            <v-card-title>
              <v-icon left>mdi-server-network</v-icon>
              Proxy Management
            </v-card-title>
            <v-card-text>
              <ErrorAlert :error="error" @clear="error = null" />

              <LoadingSpinner v-if="loading" />

              <div v-else>
                <v-chip color="primary" class="mb-4">
                  {{ proxies?.length || 0 }} Proxies
                </v-chip>

                <v-list>
                  <ProxyCard
                    v-for="proxy in proxies"
                    :key="proxy.id"
                    :proxy="proxy"
                  />
                </v-list>

                <v-empty-state
                  v-if="proxies && proxies.length === 0"
                  title="No proxies found"
                  text="Create your first proxy to get started"
                >
                  <template v-slot:image>
                    <v-icon size="100" color="grey-lighten-1">mdi-server-network</v-icon>
                  </template>
                </v-empty-state>
              </div>
            </v-card-text>
          </v-card>
        </v-col>
      </v-row>
    </v-container>
  </AppLayout>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { apiService } from '../services/api'
import { useAuthStore } from '../stores/auth'
import AppLayout from '../components/AppLayout.vue'
import ErrorAlert from '../components/ErrorAlert.vue'
import LoadingSpinner from '../components/LoadingSpinner.vue'
import ProxyCard from '../components/ProxyCard.vue'
import type { Proxy } from '../types/api'

const router = useRouter()
const authStore = useAuthStore()

const proxies = ref<Proxy[]>([])
const loading = ref(true)
const error = ref<string | null>(null)

const loadProxies = async () => {
  try {
    loading.value = true
    error.value = null
    const response = await apiService.getProxies()
    proxies.value = response.data || []
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Failed to load proxies'
    proxies.value = [] // Ensure proxies is always an array
  } finally {
    loading.value = false
  }
}


const handleLogout = () => {
  authStore.logout()
  router.push('/login')
}

onMounted(() => {
  // Load proxies when component mounts
  loadProxies()
})
</script>

<style>
/* Vuetify handles most styling, but we can add custom styles here if needed */
</style>
