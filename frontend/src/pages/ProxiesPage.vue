<template>
  <AppLayout @refresh="loadProxies">
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
                <PageHeader
                  :count="proxies?.length || 0"
                  item-name="Proxies"
                  :loading="loading"
                  @refresh="loadProxies"
                >
                  <template #actions>
                    <v-btn
                      color="success"
                      variant="outlined"
                      size="small"
                      @click="showCreateDialog = true"
                    >
                      <v-icon left>mdi-plus</v-icon>
                      Add Proxy
                    </v-btn>

                    <v-btn
                      color="primary"
                      variant="outlined"
                      size="small"
                      @click="showContainerDialog = true"
                    >
                      <v-icon left>mdi-docker</v-icon>
                      From Container
                    </v-btn>

                    <v-btn
                      color="orange"
                      variant="outlined"
                      size="small"
                      @click="reloadNginx"
                      :loading="reloadingNginx"
                    >
                      <v-icon left>mdi-reload</v-icon>
                      Reload Nginx
                    </v-btn>
                  </template>
                </PageHeader>

                <!-- Filter and Search -->
                <FilterBar
                  v-model:search-query="searchQuery"
                  v-model:status-filter="statusFilter"
                  v-model:sort-by="sortBy"
                  search-label="Search proxies..."
                  :status-options="statusOptions"
                  :sort-options="sortOptions"
                  @search="filterProxies"
                />

                <!-- Proxy Stats -->
                <StatsCards :stats="proxyStats" />

                <!-- Proxy List -->
                <div v-if="filteredProxies && filteredProxies.length > 0">
                  <ProxyCard
                    v-for="proxy in filteredProxies"
                    :key="proxy.id"
                    :proxy="proxy"
                    @edit="editProxy"
                    @delete="deleteProxy"
                  />
                </div>

                <v-empty-state
                  v-else-if="proxies && proxies.length === 0"
                  title="No proxies found"
                  text="No proxy configurations are currently available"
                />

                <v-empty-state
                  v-else
                  title="No matching proxies"
                  text="Try adjusting your search or filter criteria"
                />
              </div>
            </v-card-text>
          </v-card>
        </v-col>
      </v-row>
    </v-container>

    <!-- Create/Edit Dialog -->
    <v-dialog v-model="showCreateDialog" max-width="600px" persistent>
      <v-card>
        <v-card-title>
          <v-icon left>mdi-server-plus</v-icon>
          {{ editingProxy ? 'Edit Proxy' : 'Create New Proxy' }}
        </v-card-title>

        <v-card-text>
          <v-form ref="formRef" v-model="formValid">
            <v-row>
              <v-col cols="12">
                <v-text-field
                  v-model="form.name"
                  label="Proxy Name"
                  variant="outlined"
                  density="compact"
                  :rules="[v => !!v || 'Name is required']"
                  required
                />
              </v-col>

              <v-col cols="12">
                <v-text-field
                  v-model="form.domain"
                  label="Domain"
                  variant="outlined"
                  density="compact"
                  placeholder="example.com"
                  :rules="[v => !!v || 'Domain is required']"
                  required
                />
              </v-col>

              <v-col cols="12">
                <v-text-field
                  v-model="form.target_url"
                  label="Target URL"
                  variant="outlined"
                  density="compact"
                  placeholder="http://localhost:3000"
                  :rules="[v => !!v || 'Target URL is required']"
                  required
                />
              </v-col>

              <v-col cols="12">
                <v-switch
                  v-model="form.ssl_enabled"
                  label="Enable SSL"
                  color="primary"
                  hide-details
                />
              </v-col>
            </v-row>
          </v-form>
        </v-card-text>

        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn
            color="grey"
            variant="text"
            @click="cancelEdit"
            :disabled="saving"
          >
            Cancel
          </v-btn>
          <v-btn
            color="primary"
            @click="saveProxy"
            :loading="saving"
            :disabled="!formValid"
          >
            {{ editingProxy ? 'Update' : 'Create' }}
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>

    <!-- Delete Confirmation Dialog -->
    <ConfirmationDialog
      v-model:show="showDeleteDialog"
      title="Confirm Delete"
      :message="deleteMessage"
      icon="mdi-delete-alert"
      icon-color="error"
      confirm-text="Delete"
      confirm-color="error"
      :loading="deleting"
      @confirm="confirmDelete"
    />

    <!-- Container Selection Dialog -->
    <v-dialog v-model="showContainerDialog" max-width="800px">
      <v-card>
        <v-card-title>
          <v-icon left>mdi-docker</v-icon>
          Create Proxy from Container
        </v-card-title>

        <v-card-text>
          <ErrorAlert :error="containerError" @clear="containerError = null" />

          <LoadingSpinner v-if="loadingContainers" />

          <div v-else-if="containers && containers.length > 0">
            <v-text-field
              v-model="containerSearchQuery"
              label="Search containers..."
              prepend-inner-icon="mdi-magnify"
              variant="outlined"
              density="compact"
              clearable
              class="mb-4"
            />

            <v-list>
              <v-list-item
                v-for="container in filteredContainerList"
                :key="container.id"
                class="mb-2"
                :class="{
                  'bg-grey-lighten-4': selectedContainer?.id === container.id,
                }"
                @click="selectContainer(container)"
                :disabled="container.state !== 'running'"
              >
                <template v-slot:prepend>
                  <v-icon
                    :color="getContainerStatusColor(container.state)"
                    size="small"
                  >
                    {{ getContainerStatusIcon(container.state) }}
                  </v-icon>
                </template>

                <v-list-item-title class="text-body-1">
                  {{ container.name || 'Unnamed Container' }}
                </v-list-item-title>

                <v-list-item-subtitle class="text-caption">
                  {{ container.image }} • {{ container.state }}
                  <span v-if="container.ports && container.ports.length > 0">
                    • Ports:
                    {{ container.ports.map(p => p.public_port).join(', ') }}
                  </span>
                </v-list-item-subtitle>

                <template v-slot:append>
                  <v-chip
                    :color="getContainerStatusColor(container.state)"
                    size="x-small"
                    variant="outlined"
                    class="mr-2"
                  >
                    {{ container.state }}
                  </v-chip>

                  <v-btn
                    v-if="container.state === 'running'"
                    color="primary"
                    variant="outlined"
                    size="small"
                    @click.stop="createProxyFromContainer(container)"
                  >
                    <v-icon left size="small">mdi-plus</v-icon>
                    Create Proxy
                  </v-btn>

                  <v-tooltip v-else>
                    <template v-slot:activator="{ props }">
                      <v-btn
                        color="grey"
                        variant="outlined"
                        size="small"
                        disabled
                        v-bind="props"
                      >
                        <v-icon left size="small">mdi-plus</v-icon>
                        Create Proxy
                      </v-btn>
                    </template>
                    <span>Container must be running to create proxy</span>
                  </v-tooltip>
                </template>
              </v-list-item>
            </v-list>
          </div>

          <v-empty-state
            v-else
            title="No containers found"
            text="No Docker containers are currently available"
          />
        </v-card-text>

        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn
            color="grey"
            variant="text"
            @click="showContainerDialog = false"
          >
            Cancel
          </v-btn>
          <v-btn
            color="primary"
            variant="outlined"
            @click="loadContainers"
            :loading="loadingContainers"
          >
            <v-icon left>mdi-refresh</v-icon>
            Refresh
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
  </AppLayout>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue';
import { apiService } from '../services/api';
import AppLayout from '../components/AppLayout.vue';
import ErrorAlert from '../components/ErrorAlert.vue';
import LoadingSpinner from '../components/LoadingSpinner.vue';
import ProxyCard from '../components/ProxyCard.vue';
import PageHeader from '../components/PageHeader.vue';
import StatsCards from '../components/StatsCards.vue';
import FilterBar from '../components/FilterBar.vue';
import ConfirmationDialog from '../components/ConfirmationDialog.vue';
import type {
  Proxy,
  ProxyCreateRequest,
  ProxyUpdateRequest,
  Container,
} from '../types/api';

const proxies = ref<Proxy[]>([]);
const filteredProxies = ref<Proxy[]>([]);
const loading = ref(true);
const error = ref<string | null>(null);
const searchQuery = ref('');
const statusFilter = ref('');
const sortBy = ref('name');

// Container-related state
const containers = ref<Container[]>([]);
const loadingContainers = ref(false);
const containerError = ref<string | null>(null);
const containerSearchQuery = ref('');
const selectedContainer = ref<Container | null>(null);

// Dialog states
const showCreateDialog = ref(false);
const showDeleteDialog = ref(false);
const showContainerDialog = ref(false);
const editingProxy = ref<Proxy | null>(null);
const deletingProxy = ref<Proxy | null>(null);
const saving = ref(false);
const deleting = ref(false);
const formValid = ref(false);
const reloadingNginx = ref(false);

// Form data
const form = ref<ProxyCreateRequest & { id?: number }>({
  name: '',
  domain: '',
  target_url: '',
  ssl_enabled: false,
});

const formRef = ref();

const statusOptions = [
  { title: 'Active', value: 'active' },
  { title: 'Inactive', value: 'inactive' },
  { title: 'Error', value: 'error' },
];

const sortOptions = [
  { title: 'Name', value: 'name' },
  { title: 'Status', value: 'status' },
  { title: 'Domain', value: 'domain' },
  { title: 'Created', value: 'created_at' },
];

const activeCount = computed(
  () => proxies.value.filter(p => p.status === 'active').length
);

const inactiveCount = computed(
  () => proxies.value.filter(p => p.status === 'inactive').length
);

const errorCount = computed(
  () => proxies.value.filter(p => p.status === 'error').length
);

const sslCount = computed(
  () => proxies.value.filter(p => p.ssl_enabled).length
);

const proxyStats = computed(() => [
  {
    key: 'active',
    value: activeCount.value,
    label: 'Active',
    icon: 'mdi-check-circle',
    color: 'green-lighten-5',
    iconColor: 'green',
  },
  {
    key: 'inactive',
    value: inactiveCount.value,
    label: 'Inactive',
    icon: 'mdi-pause-circle',
    color: 'orange-lighten-5',
    iconColor: 'orange',
  },
  {
    key: 'error',
    value: errorCount.value,
    label: 'Error',
    icon: 'mdi-alert-circle',
    color: 'red-lighten-5',
    iconColor: 'red',
  },
  {
    key: 'ssl',
    value: sslCount.value,
    label: 'SSL Enabled',
    icon: 'mdi-lock',
    color: 'blue-lighten-5',
    iconColor: 'blue',
  },
]);

const deleteMessage = computed(
  () =>
    `Are you sure you want to delete the proxy "${deletingProxy.value?.name || ''}"? This action cannot be undone.`
);

// Container filtering
const filteredContainerList = computed(() => {
  let filtered = [...containers.value];

  if (containerSearchQuery.value) {
    const query = containerSearchQuery.value.toLowerCase();
    filtered = filtered.filter(
      container =>
        container.name.toLowerCase().includes(query) ||
        container.image.toLowerCase().includes(query) ||
        container.status.toLowerCase().includes(query)
    );
  }

  return filtered;
});

const loadProxies = async () => {
  try {
    loading.value = true;
    error.value = null;
    const response = await apiService.getProxies();
    proxies.value = response.data || [];
    filteredProxies.value = [...proxies.value];
    filterProxies();
    updateProxyContainerRelationships();
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Failed to load proxies';
    proxies.value = [];
    filteredProxies.value = [];
  } finally {
    loading.value = false;
  }
};

const loadContainers = async () => {
  try {
    loadingContainers.value = true;
    containerError.value = null;
    const response = await apiService.getContainers();
    containers.value = response.containers || [];
    updateProxyContainerRelationships();
  } catch (err) {
    containerError.value =
      err instanceof Error ? err.message : 'Failed to load containers';
    containers.value = [];
  } finally {
    loadingContainers.value = false;
  }
};

const filterProxies = () => {
  let filtered = [...proxies.value];

  // Search filter
  if (searchQuery.value) {
    const query = searchQuery.value.toLowerCase();
    filtered = filtered.filter(
      proxy =>
        proxy.name.toLowerCase().includes(query) ||
        proxy.domain.toLowerCase().includes(query) ||
        proxy.target_url.toLowerCase().includes(query)
    );
  }

  // Status filter
  if (statusFilter.value) {
    filtered = filtered.filter(proxy => proxy.status === statusFilter.value);
  }

  // Sort
  filtered.sort((a, b) => {
    switch (sortBy.value) {
      case 'name':
        return a.name.localeCompare(b.name);
      case 'status':
        return a.status.localeCompare(b.status);
      case 'domain':
        return a.domain.localeCompare(b.domain);
      case 'created_at':
        return (
          new Date(b.created_at).getTime() - new Date(a.created_at).getTime()
        );
      default:
        return 0;
    }
  });

  filteredProxies.value = filtered;
};

const sortProxies = () => {
  filterProxies();
};

// Relationship matching logic
const updateProxyContainerRelationships = () => {
  // Clear existing relationships
  proxies.value.forEach(proxy => {
    proxy.connected_containers = [];
  });
  containers.value.forEach(container => {
    container.connected_proxies = [];
  });

  // Match proxies to containers based on target URL
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

const editProxy = (proxy: Proxy) => {
  editingProxy.value = proxy;
  form.value = {
    id: proxy.id,
    name: proxy.name,
    domain: proxy.domain,
    target_url: proxy.target_url,
    ssl_enabled: proxy.ssl_enabled,
  };
  showCreateDialog.value = true;
};

const deleteProxy = (proxy: Proxy) => {
  deletingProxy.value = proxy;
  showDeleteDialog.value = true;
};

const cancelEdit = () => {
  showCreateDialog.value = false;
  editingProxy.value = null;
  form.value = {
    name: '',
    domain: '',
    target_url: '',
    ssl_enabled: false,
  };
  formRef.value?.reset();
};

const saveProxy = async () => {
  if (!formValid.value) return;

  try {
    saving.value = true;
    error.value = null;

    if (editingProxy.value) {
      // Update existing proxy
      const updateData: ProxyUpdateRequest = {
        name: form.value.name,
        domain: form.value.domain,
        target_url: form.value.target_url,
        ssl_enabled: form.value.ssl_enabled,
      };
      await apiService.updateProxy(editingProxy.value.id, updateData);
    } else {
      // Create new proxy
      const createData: ProxyCreateRequest = {
        name: form.value.name,
        domain: form.value.domain,
        target_url: form.value.target_url,
        ssl_enabled: form.value.ssl_enabled,
      };
      await apiService.createProxy(createData);
    }

    await loadProxies();
    cancelEdit();
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Failed to save proxy';
  } finally {
    saving.value = false;
  }
};

const confirmDelete = async () => {
  if (!deletingProxy.value) return;

  try {
    deleting.value = true;
    error.value = null;

    await apiService.deleteProxy(deletingProxy.value.id);
    await loadProxies();
    showDeleteDialog.value = false;
    deletingProxy.value = null;
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Failed to delete proxy';
  } finally {
    deleting.value = false;
  }
};

// Container utility methods
const getContainerStatusColor = (state: string) => {
  switch (state) {
    case 'running':
      return 'green';
    case 'exited':
      return 'red';
    case 'created':
      return 'blue';
    case 'paused':
      return 'orange';
    default:
      return 'grey';
  }
};

const getContainerStatusIcon = (state: string) => {
  switch (state) {
    case 'running':
      return 'mdi-play-circle';
    case 'exited':
      return 'mdi-stop-circle';
    case 'created':
      return 'mdi-plus-circle';
    case 'paused':
      return 'mdi-pause-circle';
    default:
      return 'mdi-help-circle';
  }
};

const getContainerTargetUrl = (container: Container) => {
  // Try to find a port mapping for common web ports
  const webPorts = [80, 3000, 5000, 8000, 8080, 9000];

  for (const port of webPorts) {
    const portMapping = container.ports?.find(p => p.private_port === port);
    if (portMapping) {
      return `http://localhost:${portMapping.public_port}`;
    }
  }

  // If no web port found, use the first available port
  if (container.ports && container.ports.length > 0) {
    const firstPort = container.ports[0];
    return `http://localhost:${firstPort.public_port}`;
  }

  // Fallback to localhost:3000
  return 'http://localhost:3000';
};

const selectContainer = (container: Container) => {
  selectedContainer.value = container;
};

const createProxyFromContainer = (container: Container) => {
  // Pre-fill the form with container information
  form.value = {
    name: `${container.name || 'container'}-proxy`,
    domain: `${container.name || 'container'}.example.com`,
    target_url: getContainerTargetUrl(container),
    ssl_enabled: false,
  };

  // Close container dialog and open create dialog
  showContainerDialog.value = false;
  showCreateDialog.value = true;
  formValid.value = false;
};

const reloadNginx = async () => {
  try {
    reloadingNginx.value = true;
    error.value = null;

    await apiService.reloadNginx();

    // Show success message (you could add a toast notification here)
    console.log('Nginx reloaded successfully');
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Failed to reload nginx';
  } finally {
    reloadingNginx.value = false;
  }
};

onMounted(async () => {
  await Promise.all([loadProxies(), loadContainers()]);
});
</script>

<style scoped>
.gap-2 {
  gap: 8px;
}
</style>
