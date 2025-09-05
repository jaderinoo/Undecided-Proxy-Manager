<template>
  <AppLayout @refresh="loadContainers" @logout="handleLogout">
    <v-container>
      <v-row>
        <v-col cols="12">
          <v-card>
            <v-card-title>
              <v-icon left>mdi-docker</v-icon>
              Container Management
            </v-card-title>
            <v-card-text>
              <ErrorAlert :error="error" @clear="error = null" />

              <LoadingSpinner v-if="loading" />

              <div v-else>
                <div class="d-flex align-center justify-space-between mb-4">
                  <v-chip color="primary">
                    {{ containers?.length || 0 }} Containers
                  </v-chip>
                  
                  <div class="d-flex gap-2">
                    <v-btn
                      color="primary"
                      variant="outlined"
                      size="small"
                      @click="loadContainers"
                      :loading="loading"
                    >
                      <v-icon left>mdi-refresh</v-icon>
                      Refresh
                    </v-btn>
                    
                    <v-btn
                      color="success"
                      variant="outlined"
                      size="small"
                      @click="loadContainers"
                    >
                      <v-icon left>mdi-docker</v-icon>
                      View All
                    </v-btn>
                  </div>
                </div>

                <!-- Filter and Search -->
                <v-row class="mb-4">
                  <v-col cols="12" md="6">
                    <v-text-field
                      v-model="searchQuery"
                      label="Search containers..."
                      prepend-inner-icon="mdi-magnify"
                      variant="outlined"
                      density="compact"
                      clearable
                      @input="filterContainers"
                    />
                  </v-col>
                  <v-col cols="12" md="3">
                    <v-select
                      v-model="statusFilter"
                      label="Filter by status"
                      :items="statusOptions"
                      variant="outlined"
                      density="compact"
                      clearable
                      @update:model-value="filterContainers"
                    />
                  </v-col>
                  <v-col cols="12" md="3">
                    <v-select
                      v-model="sortBy"
                      label="Sort by"
                      :items="sortOptions"
                      variant="outlined"
                      density="compact"
                      @update:model-value="sortContainers"
                    />
                  </v-col>
                </v-row>

                <!-- Container Stats -->
                <v-row class="mb-4">
                  <v-col cols="12" sm="6" md="3">
                    <v-card color="green-lighten-5" variant="outlined">
                      <v-card-text class="text-center">
                        <v-icon color="green" size="large">mdi-play-circle</v-icon>
                        <div class="text-h6">{{ runningCount }}</div>
                        <div class="text-caption">Running</div>
                      </v-card-text>
                    </v-card>
                  </v-col>
                  <v-col cols="12" sm="6" md="3">
                    <v-card color="red-lighten-5" variant="outlined">
                      <v-card-text class="text-center">
                        <v-icon color="red" size="large">mdi-stop-circle</v-icon>
                        <div class="text-h6">{{ stoppedCount }}</div>
                        <div class="text-caption">Stopped</div>
                      </v-card-text>
                    </v-card>
                  </v-col>
                  <v-col cols="12" sm="6" md="3">
                    <v-card color="blue-lighten-5" variant="outlined">
                      <v-card-text class="text-center">
                        <v-icon color="blue" size="large">mdi-plus-circle</v-icon>
                        <div class="text-h6">{{ createdCount }}</div>
                        <div class="text-caption">Created</div>
                      </v-card-text>
                    </v-card>
                  </v-col>
                  <v-col cols="12" sm="6" md="3">
                    <v-card color="orange-lighten-5" variant="outlined">
                      <v-card-text class="text-center">
                        <v-icon color="orange" size="large">mdi-pause-circle</v-icon>
                        <div class="text-h6">{{ pausedCount }}</div>
                        <div class="text-caption">Paused</div>
                      </v-card-text>
                    </v-card>
                  </v-col>
                </v-row>

                <!-- Container List -->
                <div v-if="filteredContainers && filteredContainers.length > 0">
                  <ContainerCard
                    v-for="container in filteredContainers"
                    :key="container.id"
                    :container="container"
                  />
                </div>

                <v-empty-state
                  v-else-if="containers && containers.length === 0"
                  title="No containers found"
                  text="No Docker containers are currently available"
                >
                  <template v-slot:image>
                    <v-icon size="100" color="grey-lighten-1">mdi-docker</v-icon>
                  </template>
                </v-empty-state>

                <v-empty-state
                  v-else
                  title="No matching containers"
                  text="Try adjusting your search or filter criteria"
                >
                  <template v-slot:image>
                    <v-icon size="100" color="grey-lighten-1">mdi-magnify</v-icon>
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
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { apiService } from '../services/api'
import { useAuthStore } from '../stores/auth'
import AppLayout from '../components/AppLayout.vue'
import ErrorAlert from '../components/ErrorAlert.vue'
import LoadingSpinner from '../components/LoadingSpinner.vue'
import ContainerCard from '../components/ContainerCard.vue'
import type { Container } from '../types/api'

const router = useRouter()
const authStore = useAuthStore()

const containers = ref<Container[]>([])
const filteredContainers = ref<Container[]>([])
const loading = ref(true)
const error = ref<string | null>(null)
const searchQuery = ref('')
const statusFilter = ref('')
const sortBy = ref('name')

const statusOptions = [
  { title: 'Running', value: 'running' },
  { title: 'Stopped', value: 'exited' },
  { title: 'Created', value: 'created' },
  { title: 'Paused', value: 'paused' }
]

const sortOptions = [
  { title: 'Name', value: 'name' },
  { title: 'Status', value: 'state' },
  { title: 'Created', value: 'created' },
  { title: 'Image', value: 'image' }
]

const runningCount = computed(() => 
  containers.value.filter(c => c.state === 'running').length
)

const stoppedCount = computed(() => 
  containers.value.filter(c => c.state === 'exited').length
)

const createdCount = computed(() => 
  containers.value.filter(c => c.state === 'created').length
)

const pausedCount = computed(() => 
  containers.value.filter(c => c.state === 'paused').length
)

const loadContainers = async () => {
  try {
    loading.value = true
    error.value = null
    const response = await apiService.getContainers()
    containers.value = response.containers || []
    filteredContainers.value = [...containers.value]
    filterContainers()
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Failed to load containers'
    containers.value = []
    filteredContainers.value = []
  } finally {
    loading.value = false
  }
}

const filterContainers = () => {
  let filtered = [...containers.value]

  // Search filter
  if (searchQuery.value) {
    const query = searchQuery.value.toLowerCase()
    filtered = filtered.filter(container => 
      container.name.toLowerCase().includes(query) ||
      container.image.toLowerCase().includes(query) ||
      container.status.toLowerCase().includes(query) ||
      container.command.toLowerCase().includes(query)
    )
  }

  // Status filter
  if (statusFilter.value) {
    filtered = filtered.filter(container => container.state === statusFilter.value)
  }

  // Sort
  filtered.sort((a, b) => {
    switch (sortBy.value) {
      case 'name':
        return a.name.localeCompare(b.name)
      case 'state':
        return a.state.localeCompare(b.state)
      case 'created':
        return new Date(b.created).getTime() - new Date(a.created).getTime()
      case 'image':
        return a.image.localeCompare(b.image)
      default:
        return 0
    }
  })

  filteredContainers.value = filtered
}

const sortContainers = () => {
  filterContainers()
}

const handleLogout = () => {
  authStore.logout()
  router.push('/login')
}

onMounted(() => {
  loadContainers()
})
</script>

<style scoped>
.gap-2 {
  gap: 8px;
}
</style>
