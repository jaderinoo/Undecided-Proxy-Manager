<template>
  <AppLayout @refresh="loadProxies" @logout="handleLogout">
    <v-container>
      <!-- Quick Stats -->
      <v-row class="mb-4">
        <v-col cols="12" sm="6" md="4">
          <v-card color="primary" variant="outlined">
            <v-card-text class="text-center">
              <v-icon color="primary" size="large">mdi-server-network</v-icon>
              <div class="text-h6">{{ proxies?.length || 0 }}</div>
              <div class="text-caption">Proxies</div>
            </v-card-text>
          </v-card>
        </v-col>
        <v-col cols="12" sm="6" md="4">
          <v-card color="success" variant="outlined">
            <v-card-text class="text-center">
              <v-icon color="success" size="large">mdi-docker</v-icon>
              <div class="text-h6">{{ containers?.length || 0 }}</div>
              <div class="text-caption">Containers</div>
            </v-card-text>
          </v-card>
        </v-col>
        <v-col cols="12" sm="6" md="4">
          <v-card color="warning" variant="outlined">
            <v-card-text class="text-center">
              <v-icon color="warning" size="large">mdi-shield-check</v-icon>
              <div class="text-h6">{{ sslCount }}</div>
              <div class="text-caption">SSL Enabled</div>
            </v-card-text>
          </v-card>
        </v-col>
      </v-row>

      <v-row>
        <v-col cols="12" md="6">
          <v-card>
            <v-card-title>
              <v-icon left>mdi-server-network</v-icon>
              Proxy Management
              <v-spacer></v-spacer>
              <v-btn
                color="success"
                variant="outlined"
                size="small"
                @click="openCreateProxyDialog"
                class="mr-2"
              >
                <v-icon left>mdi-plus</v-icon>
                Add Proxy
              </v-btn>
              <v-btn
                color="primary"
                variant="outlined"
                size="small"
                @click="loadProxies"
                :loading="loading"
              >
                <v-icon left>mdi-refresh</v-icon>
                Refresh
              </v-btn>
            </v-card-title>
            <v-card-text>
              <ErrorAlert :error="error" @clear="error = null" />

              <LoadingSpinner v-if="loading" />

              <div v-else>
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

        <v-col cols="12" md="6">
          <v-card>
            <v-card-title>
              <v-icon left>mdi-docker</v-icon>
              Container Overview
              <v-spacer></v-spacer>
              <v-btn
                color="success"
                variant="outlined"
                size="small"
                @click="loadContainers"
                :loading="loadingContainers"
              >
                <v-icon left>mdi-refresh</v-icon>
                Refresh
              </v-btn>
            </v-card-title>
            <v-card-text>
              <ErrorAlert :error="containerError" @clear="containerError = null" />

              <LoadingSpinner v-if="loadingContainers" />

              <div v-else>
                <div v-if="containers && containers.length > 0">
                  <div class="d-flex flex-wrap gap-2 mb-3">
                    <v-chip color="green" size="small">
                      {{ runningContainers }} Running
                    </v-chip>
                    <v-chip color="red" size="small">
                      {{ stoppedContainers }} Stopped
                    </v-chip>
                    <v-chip color="blue" size="small">
                      {{ createdContainers }} Created
                    </v-chip>
                  </div>
                  
                  <v-list density="compact">
                    <v-list-item
                      v-for="container in displayedContainers"
                      :key="container.id"
                      class="px-0"
                    >
                      <template v-slot:prepend>
                        <v-icon 
                          :color="getContainerStatusColor(container.state)"
                          size="small"
                        >
                          {{ getContainerStatusIcon(container.state) }}
                        </v-icon>
                      </template>
                      <v-list-item-title class="text-body-2">
                        {{ container.name || 'Unnamed' }}
                      </v-list-item-title>
                      <v-list-item-subtitle class="text-caption">
                        {{ container.image }} • {{ container.state }}
                      </v-list-item-subtitle>
                      <template v-slot:append>
                        <v-chip
                          :color="getContainerStatusColor(container.state)"
                          size="x-small"
                          variant="outlined"
                        >
                          {{ container.state }}
                        </v-chip>
                      </template>
                    </v-list-item>
                  </v-list>
                  
                  <v-btn
                    v-if="hasMoreContainers"
                    color="primary"
                    variant="text"
                    size="small"
                    class="mt-2"
                    @click="toggleContainerDisplay"
                  >
                    <v-icon left>
                      {{ showAllContainers ? 'mdi-chevron-up' : 'mdi-chevron-down' }}
                    </v-icon>
                    {{ showAllContainers ? 'Show Less' : `Show All ${containers.length} Containers` }}
                  </v-btn>
                </div>

                <v-empty-state
                  v-else
                  title="No containers found"
                  text="No Docker containers are currently available"
                >
                  <template v-slot:image>
                    <v-icon size="100" color="grey-lighten-1">mdi-docker</v-icon>
                  </template>
                </v-empty-state>
              </div>
            </v-card-text>
          </v-card>
        </v-col>
      </v-row>
    </v-container>

    <!-- Create Proxy Dialog -->
    <v-dialog v-model="createProxyDialog" max-width="600px">
      <v-card>
        <v-card-title>
          <v-icon left>mdi-plus</v-icon>
          Add New Proxy
        </v-card-title>
        <v-card-text>
          <v-form ref="createProxyForm" v-model="createProxyFormValid">
            <v-text-field
              v-model="newProxy.name"
              label="Proxy Name"
              :rules="[v => !!v || 'Name is required']"
              required
              class="mb-2"
            ></v-text-field>
            
            <v-text-field
              v-model="newProxy.domain"
              label="Domain"
              :rules="[v => !!v || 'Domain is required']"
              required
              class="mb-2"
              hint="e.g., example.com"
            ></v-text-field>
            
            <v-text-field
              v-model="newProxy.target_url"
              label="Target URL"
              :rules="[v => !!v || 'Target URL is required']"
              required
              class="mb-2"
              hint="e.g., http://localhost:3000"
            ></v-text-field>
            
            <v-switch
              v-model="newProxy.ssl_enabled"
              label="Enable SSL"
              color="success"
              class="mb-2"
            ></v-switch>
          </v-form>
        </v-card-text>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn
            color="grey"
            variant="text"
            @click="closeCreateProxyDialog"
          >
            Cancel
          </v-btn>
          <v-btn
            color="primary"
            @click="createProxy"
            :loading="creatingProxy"
            :disabled="!createProxyFormValid"
          >
            Create Proxy
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
  </AppLayout>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { apiService } from '../services/api'
import { useAuthStore } from '../stores/auth'
import AppLayout from '../components/AppLayout.vue'
import ErrorAlert from '../components/ErrorAlert.vue'
import LoadingSpinner from '../components/LoadingSpinner.vue'
import ProxyCard from '../components/ProxyCard.vue'
import type { Proxy, Container, ProxyCreateRequest } from '../types/api'

const router = useRouter()
const authStore = useAuthStore()

const proxies = ref<Proxy[]>([])
const containers = ref<Container[]>([])
const loading = ref(true)
const loadingContainers = ref(false)
const error = ref<string | null>(null)
const containerError = ref<string | null>(null)
const showAllContainers = ref(false)

// Create proxy dialog state
const createProxyDialog = ref(false)
const createProxyFormValid = ref(false)
const creatingProxy = ref(false)
const newProxy = ref<ProxyCreateRequest>({
  name: '',
  domain: '',
  target_url: '',
  ssl_enabled: false
})

// Computed properties
const sslCount = computed(() => 
  proxies.value.filter(p => p.ssl_enabled).length
)

const runningContainers = computed(() => 
  containers.value.filter(c => c.state === 'running').length
)

const stoppedContainers = computed(() => 
  containers.value.filter(c => c.state === 'exited').length
)

const createdContainers = computed(() => 
  containers.value.filter(c => c.state === 'created').length
)

const displayedContainers = computed(() => {
  if (showAllContainers.value) {
    return containers.value
  }
  return containers.value.slice(0, 5)
})

const hasMoreContainers = computed(() => 
  containers.value.length > 5
)

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

const loadContainers = async () => {
  try {
    loadingContainers.value = true
    containerError.value = null
    const response = await apiService.getContainers()
    containers.value = response.containers || []
  } catch (err) {
    containerError.value = err instanceof Error ? err.message : 'Failed to load containers'
    containers.value = []
  } finally {
    loadingContainers.value = false
  }
}


const getContainerStatusColor = (state: string) => {
  switch (state) {
    case 'running': return 'green'
    case 'exited': return 'red'
    case 'created': return 'blue'
    case 'paused': return 'orange'
    default: return 'grey'
  }
}

const getContainerStatusIcon = (state: string) => {
  switch (state) {
    case 'running': return 'mdi-play-circle'
    case 'exited': return 'mdi-stop-circle'
    case 'created': return 'mdi-plus-circle'
    case 'paused': return 'mdi-pause-circle'
    default: return 'mdi-help-circle'
  }
}

const toggleContainerDisplay = () => {
  showAllContainers.value = !showAllContainers.value
}

// Create proxy dialog methods
const openCreateProxyDialog = () => {
  createProxyDialog.value = true
  // Reset form
  newProxy.value = {
    name: '',
    domain: '',
    target_url: '',
    ssl_enabled: false
  }
  createProxyFormValid.value = false
}

const closeCreateProxyDialog = () => {
  createProxyDialog.value = false
  error.value = null
}

const createProxy = async () => {
  try {
    creatingProxy.value = true
    error.value = null
    
    const response = await apiService.createProxy(newProxy.value)
    
    // Add the new proxy to the list
    if (response.data) {
      proxies.value.unshift(response.data)
    }
    
    // Close dialog and reset form
    closeCreateProxyDialog()
    
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Failed to create proxy'
  } finally {
    creatingProxy.value = false
  }
}

const handleLogout = () => {
  authStore.logout()
  router.push('/login')
}

onMounted(async () => {
  // Load all data when component mounts
  await Promise.all([
    loadProxies(),
    loadContainers()
  ])
})
</script>

<style>
/* Vuetify handles most styling, but we can add custom styles here if needed */
</style>
