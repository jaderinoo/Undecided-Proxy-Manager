<template>
  <v-card class="mb-2" variant="outlined">
    <v-card-title class="card-header">
      <div class="proxy-info">
        <v-icon :color="getStatusIconColor(proxy.status)" size="small">
          {{ getStatusIcon(proxy.status) }}
        </v-icon>
        <span class="responsive-title">{{ proxy.name }}</span>
        <v-chip :color="getStatusChipColor(proxy.status)" size="x-small">
          {{ proxy.status }}
        </v-chip>
        <v-chip v-if="proxy.ssl_enabled" color="success" size="x-small">
          <v-icon left size="x-small">mdi-lock</v-icon>
          SSL
        </v-chip>
      </div>
      <div class="card-actions">
        <v-btn
          v-if="proxy.ssl_enabled"
          size="small"
          variant="text"
          color="info"
          @click="showCertificateInfo = true"
        >
          <v-icon>mdi-information</v-icon>
          <span class="d-none d-sm-inline">Info</span>
        </v-btn>
        <v-btn
          size="small"
          variant="text"
          color="primary"
          @click="$emit('regenerate', proxy)"
          :loading="regenerating"
        >
          <v-tooltip activator="parent" location="top">
            Regenerate nginx proxy configuration
          </v-tooltip>
          <v-icon>mdi-cog-refresh</v-icon>
          <span class="d-none d-sm-inline">Regenerate</span>
        </v-btn>
        <v-btn
          size="small"
          variant="text"
          color="grey-darken-1"
          @click="$emit('edit', proxy)"
        >
          <v-icon>mdi-pencil</v-icon>
          <span class="d-none d-sm-inline">Edit</span>
        </v-btn>
        <v-btn
          size="small"
          variant="text"
          color="error"
          @click="$emit('delete', proxy)"
        >
          <v-icon>mdi-delete</v-icon>
          <span class="d-none d-sm-inline">Delete</span>
        </v-btn>
      </div>
    </v-card-title>

    <v-card-text class="card-content">
      <div class="details-grid">
        <div class="detail-item">
          <div class="detail-label">Domain</div>
          <div class="detail-value">{{ proxy.domain }}</div>
        </div>

        <div class="detail-item">
          <div class="detail-label">Target URL</div>
          <div class="detail-value">{{ proxy.target_url }}</div>
        </div>

        <div class="detail-item">
          <div class="detail-label">Created</div>
          <div class="detail-value">{{ formatDate(proxy.created_at) }}</div>
        </div>

        <div class="detail-item">
          <div class="detail-label">Updated</div>
          <div class="detail-value">{{ formatDate(proxy.updated_at) }}</div>
        </div>
      </div>

      <!-- Connected Containers -->
      <div v-if="
        proxy.connected_containers && proxy.connected_containers.length > 0
      " class="mt-2">
        <div class="text-caption text-grey-darken-2 mb-1">
          Connected Containers
        </div>
        <div class="d-flex flex-wrap gap-1">
          <v-chip v-for="container in proxy.connected_containers" :key="container.id" size="x-small" color="primary"
            variant="outlined">
            <v-icon left size="x-small">mdi-docker</v-icon>
            {{ container.name || 'Unnamed' }}
          </v-chip>
        </div>
      </div>
    </v-card-text>

    <!-- Certificate Information Dialog -->
    <v-dialog v-model="showCertificateInfo" max-width="600px">
      <v-card>
        <v-card-title>
          <v-icon left>mdi-certificate</v-icon>
          SSL Certificate Information
        </v-card-title>

        <v-card-text>
          <div v-if="loadingCertificate" class="text-center py-4">
            <v-progress-circular indeterminate color="primary"></v-progress-circular>
            <div class="mt-2">Loading certificate information...</div>
          </div>

          <div v-else-if="certificateError" class="text-center py-4">
            <v-icon color="error" size="large">mdi-alert-circle</v-icon>
            <div class="mt-2 text-error">{{ certificateError }}</div>
          </div>

          <div v-else-if="certificate">
            <v-row>
              <v-col cols="12">
                <div class="text-caption text-grey-darken-2 mb-1">Domain</div>
                <div class="text-body-2 text-grey-darken-3 mb-3">
                  {{ certificate.domain }}
                </div>

                <div class="text-caption text-grey-darken-2 mb-1">Issuer</div>
                <div class="text-body-2 text-grey-darken-3 mb-3">
                  {{ certificate.issuer || 'Let\'s Encrypt' }}
                </div>

                <div class="text-caption text-grey-darken-2 mb-1">Expires</div>
                <div class="text-body-2 text-grey-darken-3 mb-3">
                  {{ formatDate(certificate.expires_at) }}
                  <v-chip :color="certificate.is_valid ? 'success' : 'error'" size="x-small" class="ml-2">
                    {{ certificate.is_valid ? 'Valid' : 'Expired' }}
                  </v-chip>
                </div>

                <div class="text-caption text-grey-darken-2 mb-1">Certificate Path</div>
                <div class="text-body-2 text-grey-darken-3 mb-3 font-mono text-caption">
                  {{ certificate.cert_path }}
                </div>

                <div class="text-caption text-grey-darken-2 mb-1">Private Key Path</div>
                <div class="text-body-2 text-grey-darken-3 mb-3 font-mono text-caption">
                  {{ certificate.key_path }}
                </div>

                <div class="text-caption text-grey-darken-2 mb-1">Created</div>
                <div class="text-body-2 text-grey-darken-3">
                  {{ formatDate(certificate.created_at) }}
                </div>
              </v-col>
            </v-row>
          </div>
        </v-card-text>

        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="grey" variant="text" @click="showCertificateInfo = false">
            Close
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
  </v-card>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue';
import { apiService } from '../../services/api';
import type { Certificate, Proxy } from '../../types/api';

const props = defineProps<{
  proxy: Proxy;
  regenerating?: boolean;
}>();

defineEmits<{
  edit: [proxy: Proxy];
  delete: [proxy: Proxy];
  regenerate: [proxy: Proxy];
}>();

const showCertificateInfo = ref(false);
const loadingCertificate = ref(false);
const certificateError = ref<string | null>(null);
const certificate = ref<Certificate | null>(null);

const loadCertificate = async (proxyId: number) => {
  try {
    loadingCertificate.value = true;
    certificateError.value = null;
    const response = await apiService.getProxyCertificate(proxyId);
    certificate.value = response.data;
  } catch (err) {
    certificateError.value = err instanceof Error ? err.message : 'Failed to load certificate information';
    certificate.value = null;
  } finally {
    loadingCertificate.value = false;
  }
};

// Watch for dialog opening to load certificate data
watch(showCertificateInfo, (newValue) => {
  if (newValue && props.proxy.ssl_enabled) {
    loadCertificate(props.proxy.id);
  }
});

const getStatusColor = (status: string): string => {
  switch (status) {
    case 'active':
      return 'green-lighten-5';
    case 'inactive':
      return 'orange-lighten-5';
    case 'error':
      return 'red-lighten-5';
    default:
      return 'grey-lighten-5';
  }
};

const getStatusIconColor = (status: string): string => {
  switch (status) {
    case 'active':
      return 'green';
    case 'inactive':
      return 'orange';
    case 'error':
      return 'red';
    default:
      return 'grey';
  }
};

const getStatusIcon = (status: string): string => {
  switch (status) {
    case 'active':
      return 'mdi-check-circle';
    case 'inactive':
      return 'mdi-pause-circle';
    case 'error':
      return 'mdi-alert-circle';
    default:
      return 'mdi-help-circle';
  }
};

const getStatusChipColor = (status: string): string => {
  switch (status) {
    case 'active':
      return 'green';
    case 'inactive':
      return 'orange';
    case 'error':
      return 'red';
    default:
      return 'grey';
  }
};

const formatDate = (dateString: string) => {
  return new Date(dateString).toLocaleString();
};
</script>

<style scoped>
.proxy-info {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: var(--space-1);
  min-width: 0;
  flex: 1;
}

/* Mobile layout improvements */
@media (max-width: 600px) {
  .card-header {
    flex-direction: column;
    align-items: center;
    gap: var(--space-2);
  }

  .proxy-info {
    justify-content: center;
    text-align: center;
  }

  .card-actions {
    flex-direction: column;
    align-items: center;
    gap: var(--space-1);
    width: 100%;
  }

  .card-actions .v-btn {
    width: auto;
    min-width: 120px;
    justify-content: center;
  }
}

/* Very small screens */
@media (max-width: 400px) {
  .card-actions .v-btn span {
    display: none;
  }

  .card-actions .v-btn {
    min-width: 40px;
    width: auto;
  }
}

/* Touch optimization */
@media (max-width: 768px) {
  .card-actions .v-btn {
    min-height: 36px;
    touch-action: manipulation;
  }
}
</style>
