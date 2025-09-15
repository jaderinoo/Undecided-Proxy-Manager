<template>
  <AppLayout @refresh="loadProxies">
    <v-container>
      <!-- Quick Stats -->
      <v-row class="mb-4">
        <v-col cols="12" sm="6" md="4">
          <QuickStatCard
            :value="proxies?.length || 0"
            label="Proxies"
            icon="mdi-server-network"
            color="primary"
            icon-color="primary"
          />
        </v-col>
        <v-col cols="12" sm="6" md="4">
          <QuickStatCard
            :value="containers?.length || 0"
            label="Containers"
            icon="mdi-docker"
            color="success"
            icon-color="success"
          />
        </v-col>
        <v-col cols="12" sm="6" md="4">
          <QuickStatCard
            :value="sslCount"
            label="SSL Enabled"
            icon="mdi-shield-check"
            color="warning"
            icon-color="warning"
          />
        </v-col>
      </v-row>

      <v-row>
        <v-col cols="12">
          <ProxyOverview
            :proxies="proxies"
            :loading="loading"
            :error="error"
            @refresh="loadProxies"
            @clear-error="error = null"
            @view-all="$router.push('/proxies')"
            @add-proxy="$router.push('/proxies')"
          />
        </v-col>

        <v-col cols="12">
          <ContainerOverview
            :containers="containers"
            :loading="loadingContainers"
            :error="containerError"
            :show-all-containers="showAllContainers"
            @refresh="loadContainers"
            @clear-error="containerError = null"
            @create-proxy-for-container="openCreateProxyForContainer"
            @toggle-display="toggleContainerDisplay"
          />
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
import ContainerOverview from '../components/container/ContainerOverview.vue';
import AppLayout from '../components/layout/AppLayout.vue';
import ProxyFormDialog from '../components/proxy/ProxyFormDialog.vue';
import ProxyOverview from '../components/proxy/ProxyOverview.vue';
import QuickStatCard from '../components/ui/QuickStatCard.vue';
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


const toggleContainerDisplay = () => {
  showAllContainers.value = !showAllContainers.value;
};


onMounted(async () => {
  // Load all data when component mounts
  await Promise.all([loadProxies(), loadContainers()]);
});
</script>

<style>
/* Vuetify handles most styling, but we can add custom styles here if needed */
</style>
