<template>
  <AppLayout @refresh="loadProxies">
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
                class="mr-2"
              >
                <v-icon left>mdi-refresh</v-icon>
                Refresh
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
                    @edit="openEditProxyDialog"
                    @delete="openDeleteProxyDialog"
                  />
                </v-list>

                <v-empty-state
                  v-if="proxies && proxies.length === 0"
                  title="No proxies found"
                  text="Create your first proxy to get started"
                >
                  <template v-slot:image>
                    <v-icon size="100" color="grey-lighten-1"
                      >mdi-server-network</v-icon
                    >
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
              <ErrorAlert
                :error="containerError"
                @clear="containerError = null"
              />

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
                          class="mr-2"
                        >
                          {{ container.state }}
                        </v-chip>
                        <v-btn
                          icon
                          size="small"
                          variant="text"
                          color="primary"
                          @click="openCreateProxyForContainer(container)"
                          :disabled="container.state !== 'running'"
                          v-tooltip="
                            container.state === 'running'
                              ? 'Create proxy for this container'
                              : 'Container must be running to create proxy'
                          "
                        >
                          <v-icon size="small">mdi-plus-circle</v-icon>
                        </v-btn>
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
                      {{
                        showAllContainers
                          ? 'mdi-chevron-up'
                          : 'mdi-chevron-down'
                      }}
                    </v-icon>
                    {{
                      showAllContainers
                        ? 'Show Less'
                        : `Show All ${containers.length} Containers`
                    }}
                  </v-btn>
                </div>

                <v-empty-state
                  v-else
                  title="No containers found"
                  text="No Docker containers are currently available"
                >
                  <template v-slot:image>
                    <v-icon size="100" color="grey-lighten-1"
                      >mdi-docker</v-icon
                    >
                  </template>
                </v-empty-state>
              </div>
            </v-card-text>
          </v-card>
        </v-col>
      </v-row>
    </v-container>

    <!-- Create Proxy Dialog -->
    <ProxyFormDialog
      v-model:show="createProxyDialog"
      :initial-data="containerFormData"
      @save="handleCreateProxy"
      @cancel="closeCreateProxyDialog"
    />

    <!-- Edit Proxy Dialog -->
    <ProxyFormDialog
      v-model:show="editProxyDialog"
      :editing-proxy="editingProxy"
      @save="handleEditProxy"
      @cancel="closeEditProxyDialog"
    />

    <!-- Delete Confirmation Dialog -->
    <v-dialog v-model="deleteProxyDialog" max-width="400px">
      <v-card>
        <v-card-title>
          <v-icon left color="error">mdi-delete</v-icon>
          Delete Proxy
        </v-card-title>
        <v-card-text>
          Are you sure you want to delete the proxy "{{ proxyToDelete?.name }}"?
          This action cannot be undone.
        </v-card-text>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="grey" variant="text" @click="closeDeleteProxyDialog">
            Cancel
          </v-btn>
          <v-btn color="error" @click="deleteProxy" :loading="deletingProxy">
            Delete
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue';
import AppLayout from '../components/AppLayout.vue';
import ErrorAlert from '../components/ErrorAlert.vue';
import LoadingSpinner from '../components/LoadingSpinner.vue';
import ProxyCard from '../components/ProxyCard.vue';
import ProxyFormDialog from '../components/ProxyFormDialog.vue';
import { apiService } from '../services/api';
import type {
  Container,
  Proxy,
  ProxyCreateRequest,
  ProxyUpdateRequest,
} from '../types/api';

const proxies = ref<Proxy[]>([]);
const containers = ref<Container[]>([]);
const loading = ref(true);
const loadingContainers = ref(false);
const error = ref<string | null>(null);
const containerError = ref<string | null>(null);
const showAllContainers = ref(false);

// Dialog state
const createProxyDialog = ref(false);
const editProxyDialog = ref(false);
const creatingProxy = ref(false);
const updatingProxy = ref(false);
const editingProxy = ref<Proxy | null>(null);
const containerFormData = ref<Partial<ProxyCreateRequest> | undefined>(undefined);

// Delete proxy dialog state
const deleteProxyDialog = ref(false);
const deletingProxy = ref(false);
const proxyToDelete = ref<Proxy | null>(null);

// Nginx reload state
const reloadingNginx = ref(false);

// Computed properties
const sslCount = computed(
  () => proxies.value.filter(p => p.ssl_enabled).length
);

const runningContainers = computed(
  () => containers.value.filter(c => c.state === 'running').length
);

const stoppedContainers = computed(
  () => containers.value.filter(c => c.state === 'exited').length
);

const createdContainers = computed(
  () => containers.value.filter(c => c.state === 'created').length
);

const displayedContainers = computed(() => {
  if (showAllContainers.value) {
    return containers.value;
  }
  return containers.value.slice(0, 5);
});

const hasMoreContainers = computed(() => containers.value.length > 5);

const loadProxies = async () => {
  try {
    loading.value = true;
    error.value = null;
    const response = await apiService.getProxies();
    proxies.value = response.data || [];
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Failed to load proxies';
    proxies.value = []; // Ensure proxies is always an array
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
  } catch (err) {
    containerError.value =
      err instanceof Error ? err.message : 'Failed to load containers';
    containers.value = [];
  } finally {
    loadingContainers.value = false;
  }
};

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

const toggleContainerDisplay = () => {
  showAllContainers.value = !showAllContainers.value;
};

// Create proxy dialog methods
const openCreateProxyDialog = () => {
  containerFormData.value = undefined;
  createProxyDialog.value = true;
};

const closeCreateProxyDialog = () => {
  createProxyDialog.value = false;
  containerFormData.value = undefined;
  error.value = null;
};

const openCreateProxyForContainer = (container: Container) => {
  // Pre-fill the form with container information
  containerFormData.value = {
    name: `${container.name || 'container'}-proxy`,
    domain: `${container.name || 'container'}.example.com`,
    target_url: getContainerTargetUrl(container),
    ssl_enabled: false,
  };
  createProxyDialog.value = true;
};

const handleCreateProxy = async (data: ProxyCreateRequest | ProxyUpdateRequest, _isEdit: boolean) => {
  try {
    creatingProxy.value = true;
    error.value = null;

    const response = await apiService.createProxy(data as ProxyCreateRequest);

    // Add the new proxy to the list
    if (response.data) {
      proxies.value.unshift(response.data);
    }

    // Close dialog and reset form
    closeCreateProxyDialog();
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Failed to create proxy';
  } finally {
    creatingProxy.value = false;
  }
};

// Edit proxy dialog methods
const openEditProxyDialog = (proxy: Proxy) => {
  editingProxy.value = proxy;
  editProxyDialog.value = true;
};

const closeEditProxyDialog = () => {
  editProxyDialog.value = false;
  editingProxy.value = null;
  error.value = null;
};

const handleEditProxy = async (data: ProxyCreateRequest | ProxyUpdateRequest, _isEdit: boolean) => {
  if (!editingProxy.value) return;

  try {
    updatingProxy.value = true;
    error.value = null;

    const response = await apiService.updateProxy(editingProxy.value.id, data as ProxyUpdateRequest);

    // Update the proxy in the list
    if (response.data) {
      const proxyIndex = proxies.value.findIndex(p => p.id === editingProxy.value!.id);
      if (proxyIndex !== -1) {
        proxies.value[proxyIndex] = response.data;
      }
    }

    // Close dialog
    closeEditProxyDialog();
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Failed to update proxy';
  } finally {
    updatingProxy.value = false;
  }
};

// Delete proxy dialog methods
const openDeleteProxyDialog = (proxy: Proxy) => {
  proxyToDelete.value = proxy;
  deleteProxyDialog.value = true;
};

const closeDeleteProxyDialog = () => {
  deleteProxyDialog.value = false;
  proxyToDelete.value = null;
  error.value = null;
};

const deleteProxy = async () => {
  if (!proxyToDelete.value) return;

  try {
    deletingProxy.value = true;
    error.value = null;

    // Store the proxy ID before deletion
    const proxyId = proxyToDelete.value.id;

    await apiService.deleteProxy(proxyId);

    // Remove the proxy from the list
    const proxyIndex = proxies.value.findIndex(p => p.id === proxyId);
    if (proxyIndex !== -1) {
      proxies.value.splice(proxyIndex, 1);
    }

    // Close dialog
    closeDeleteProxyDialog();
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Failed to delete proxy';
  } finally {
    deletingProxy.value = false;
  }
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
  // Load all data when component mounts
  await Promise.all([loadProxies(), loadContainers()]);
});
</script>

<style>
/* Vuetify handles most styling, but we can add custom styles here if needed */
</style>
