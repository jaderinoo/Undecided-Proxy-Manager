<template>
  <AppLayout @refresh="loadContainers">
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
                <PageHeader
                  :count="containers?.length || 0"
                  item-name="Containers"
                  :loading="loading"
                  @refresh="loadContainers"
                >
                  <template #actions>
                    <v-btn
                      color="success"
                      variant="text"
                      size="small"
                      @click="loadContainers"
                    >
                      <v-icon left>mdi-docker</v-icon>
                      View All
                    </v-btn>
                  </template>
                </PageHeader>

                <!-- Filter and Search -->
                <FilterBar
                  v-model:search-query="searchQuery"
                  v-model:status-filter="statusFilter"
                  v-model:sort-by="sortBy"
                  search-label="Search containers..."
                  :status-options="statusOptions"
                  :sort-options="sortOptions"
                  @search="filterContainers"
                />

                <!-- Container Stats -->
                <StatsCards :stats="containerStats" />

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
                />

                <v-empty-state
                  v-else
                  title="No matching containers"
                  text="Try adjusting your search or filter criteria"
                />
              </div>
            </v-card-text>
          </v-card>
        </v-col>
      </v-row>
    </v-container>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue';
import ContainerCard from '../components/container/ContainerCard.vue';
import AppLayout from '../components/layout/AppLayout.vue';
import ErrorAlert from '../components/ui/ErrorAlert.vue';
import FilterBar from '../components/ui/FilterBar.vue';
import LoadingSpinner from '../components/ui/LoadingSpinner.vue';
import PageHeader from '../components/ui/PageHeader.vue';
import StatsCards from '../components/ui/StatsCards.vue';
import { apiService } from '../services/api';
import type { Container, Proxy } from '../types/api';

const containers = ref<Container[]>([]);
const filteredContainers = ref<Container[]>([]);
const proxies = ref<Proxy[]>([]);
const loading = ref(true);
const error = ref<string | null>(null);
const searchQuery = ref('');
const statusFilter = ref('');
const sortBy = ref('name');

const statusOptions = [
  { title: 'Running', value: 'running' },
  { title: 'Stopped', value: 'exited' },
  { title: 'Created', value: 'created' },
  { title: 'Paused', value: 'paused' },
];

const sortOptions = [
  { title: 'Name', value: 'name' },
  { title: 'Status', value: 'state' },
  { title: 'Created', value: 'created' },
  { title: 'Image', value: 'image' },
];

const runningCount = computed(
  () => containers.value.filter(c => c.state === 'running').length
);

const stoppedCount = computed(
  () => containers.value.filter(c => c.state === 'exited').length
);

const createdCount = computed(
  () => containers.value.filter(c => c.state === 'created').length
);

const pausedCount = computed(
  () => containers.value.filter(c => c.state === 'paused').length
);

const containerStats = computed(() => [
  {
    key: 'running',
    value: runningCount.value,
    label: 'Running',
    icon: 'mdi-play-circle',
    color: 'green-lighten-5',
    iconColor: 'green',
  },
  {
    key: 'stopped',
    value: stoppedCount.value,
    label: 'Stopped',
    icon: 'mdi-stop-circle',
    color: 'red-lighten-5',
    iconColor: 'red',
  },
  {
    key: 'created',
    value: createdCount.value,
    label: 'Created',
    icon: 'mdi-plus-circle',
    color: 'blue-lighten-5',
    iconColor: 'blue',
  },
  {
    key: 'paused',
    value: pausedCount.value,
    label: 'Paused',
    icon: 'mdi-pause-circle',
    color: 'orange-lighten-5',
    iconColor: 'orange',
  },
]);

const loadContainers = async () => {
  try {
    loading.value = true;
    error.value = null;
    const response = await apiService.getContainers();
    containers.value = response.containers || [];
    filteredContainers.value = [...containers.value];
    filterContainers();
    updateContainerProxyRelationships();
  } catch (err) {
    error.value =
      err instanceof Error ? err.message : 'Failed to load containers';
    containers.value = [];
    filteredContainers.value = [];
  } finally {
    loading.value = false;
  }
};

const loadProxies = async () => {
  try {
    const response = await apiService.getProxies();
    proxies.value = response.data || [];
    updateContainerProxyRelationships();
  } catch (err) {
    console.error('Failed to load proxies:', err);
    proxies.value = [];
  }
};

// Relationship matching logic
const updateContainerProxyRelationships = () => {
  // Clear existing relationships
  containers.value.forEach(container => {
    container.connected_proxies = [];
  });
  proxies.value.forEach(proxy => {
    proxy.connected_containers = [];
  });

  // Match containers to proxies based on target URL
  proxies.value.forEach(proxy => {
    const targetUrl = new URL(proxy.target_url);
    const targetPort =
      parseInt(targetUrl.port) || (targetUrl.protocol === 'https:' ? 443 : 80);

    containers.value.forEach(container => {
      if (container.state === 'running' && container.ports) {
        // Check if any container port matches the proxy target port
        const matchingPort = container.ports.find(
          port =>
            port.public_port === targetPort ||
            (targetUrl.hostname === 'localhost' &&
              port.public_port === targetPort)
        );

        if (matchingPort) {
          // Add container to proxy's connected containers
          if (!proxy.connected_containers) {
            proxy.connected_containers = [];
          }
          proxy.connected_containers.push(container);

          // Add proxy to container's connected proxies
          if (!container.connected_proxies) {
            container.connected_proxies = [];
          }
          container.connected_proxies.push(proxy);
        }
      }
    });
  });
};

const filterContainers = () => {
  let filtered = [...containers.value];

  // Search filter
  if (searchQuery.value) {
    const query = searchQuery.value.toLowerCase();
    filtered = filtered.filter(
      container =>
        container.name.toLowerCase().includes(query) ||
        container.image.toLowerCase().includes(query) ||
        container.status.toLowerCase().includes(query) ||
        container.command.toLowerCase().includes(query)
    );
  }

  // Status filter
  if (statusFilter.value) {
    filtered = filtered.filter(
      container => container.state === statusFilter.value
    );
  }

  // Sort
  filtered.sort((a, b) => {
    // Always prioritize containers with connected proxies first
    const aHasProxies = a.connected_proxies && a.connected_proxies.length > 0;
    const bHasProxies = b.connected_proxies && b.connected_proxies.length > 0;

    if (aHasProxies && !bHasProxies) return -1;
    if (bHasProxies && !aHasProxies) return 1;

    // Then prioritize running containers
    if (a.state === 'running' && b.state !== 'running') return -1;
    if (b.state === 'running' && a.state !== 'running') return 1;

    switch (sortBy.value) {
      case 'name':
        return a.name.localeCompare(b.name);
      case 'state':
        return a.state.localeCompare(b.state);
      case 'created':
        return new Date(b.created).getTime() - new Date(a.created).getTime();
      case 'image':
        return a.image.localeCompare(b.image);
      default:
        return 0;
    }
  });

  filteredContainers.value = filtered;
};

// const sortContainers = () => {
//   filterContainers();
// };

onMounted(async () => {
  await Promise.all([loadContainers(), loadProxies()]);
});
</script>

<style scoped>
.gap-2 {
  gap: 8px;
}
</style>
