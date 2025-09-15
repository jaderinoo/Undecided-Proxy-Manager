<template>
  <v-card class="mb-2" :color="getStatusColor()" variant="outlined">
    <v-card-title class="d-flex align-center py-2">
      <v-icon :color="getStatusIconColor()" class="mr-2" size="small">
        {{ getStatusIcon() }}
      </v-icon>
      <span class="text-subtitle-1 font-weight-medium text-grey-darken-3">{{
        certificate.domain
      }}</span>
      <v-spacer></v-spacer>
      <v-chip :color="getValidityChipColor()" size="x-small" class="mr-2">
        {{ certificate.is_valid ? 'Valid' : 'Invalid' }}
      </v-chip>
      <v-chip :color="getExpiryChipColor()" size="x-small" class="mr-2">
        {{ getExpiryText() }}
      </v-chip>
      <v-btn
        size="small"
        variant="text"
        color="grey-darken-1"
        @click="toggleExpanded"
        class="action-btn"
      >
        <v-icon left size="small">{{
          expanded ? 'mdi-chevron-up' : 'mdi-chevron-down'
        }}</v-icon>
        <span class="d-none d-sm-inline">{{ expanded ? 'Collapse' : 'Expand' }}</span>
      </v-btn>
    </v-card-title>

    <v-card-text class="py-2">
      <v-row>
        <v-col cols="12" md="6">
          <div class="text-caption text-grey-darken-2 mb-1">
            Certificate Path
          </div>
          <div class="text-body-2 text-grey-darken-3 mb-1 font-mono">
            {{ certificate.cert_path }}
          </div>

          <div class="text-caption text-grey-darken-2 mb-1">Key Path</div>
          <div class="text-body-2 text-grey-darken-3 mb-1 font-mono">
            {{ certificate.key_path }}
          </div>
        </v-col>

        <v-col cols="12" md="6">
          <div class="text-caption text-grey-darken-2 mb-1">Expires At</div>
          <div class="text-body-2 text-grey-darken-3 mb-1">
            {{ formatDate(certificate.expires_at) }}
          </div>

          <div class="text-caption text-grey-darken-2 mb-1">Created</div>
          <div class="text-body-2 text-grey-darken-3 mb-1">
            {{ formatDate(certificate.created_at) }}
          </div>
        </v-col>
      </v-row>

      <!-- Associated Proxies -->
      <div v-if="proxies.length > 0" class="mt-2">
        <div class="text-caption text-grey-darken-2 mb-1">
          Associated Proxies
        </div>
        <div class="d-flex flex-wrap gap-1">
          <v-chip
            v-for="proxy in proxies"
            :key="proxy.id"
            size="x-small"
            :color="getProxyStatusColor(proxy.status)"
            variant="outlined"
          >
            <v-icon left size="x-small">mdi-server-network</v-icon>
            {{ proxy.name }}
          </v-chip>
        </div>
      </div>

      <!-- Expanded Details -->
      <v-expand-transition>
        <div v-if="expanded" class="mt-2">
          <v-divider class="mb-2"></v-divider>

          <v-row>
            <v-col cols="12" md="6">
              <div class="text-caption text-grey-darken-2 mb-1">
                Last Updated
              </div>
              <div class="text-body-2 text-grey-darken-3 mb-2">
                {{ formatDate(certificate.updated_at) }}
              </div>

              <div class="text-caption text-grey-darken-2 mb-1">
                Days Until Expiry
              </div>
              <div class="text-body-2 text-grey-darken-3 mb-2">
                {{ getDaysUntilExpiry() }}
              </div>
            </v-col>

            <v-col cols="12" md="6">
              <div class="text-caption text-grey-darken-2 mb-1">Actions</div>
              <div class="d-flex flex-wrap gap-1">
                <v-btn
                  size="x-small"
                  color="primary"
                  variant="outlined"
                  @click="renewCertificate"
                  :loading="isRenewing"
                >
                  <v-icon left size="x-small">mdi-refresh</v-icon>
                  Renew
                </v-btn>
                <v-btn
                  size="x-small"
                  color="info"
                  variant="outlined"
                  @click="viewProxies"
                >
                  <v-icon left size="x-small">mdi-server-network</v-icon>
                  View Proxies
                </v-btn>
                <v-btn
                  size="x-small"
                  color="error"
                  variant="outlined"
                  @click="deleteCertificate"
                  :loading="isDeleting"
                >
                  <v-icon left size="x-small">mdi-delete</v-icon>
                  Delete
                </v-btn>
              </div>
            </v-col>
          </v-row>

          <!-- Proxies List -->
          <div v-if="showProxies && proxies.length > 0" class="mt-2">
            <div class="text-caption text-grey-darken-2 mb-1">
              Proxy Details
            </div>
            <v-list density="compact">
              <v-list-item
                v-for="proxy in proxies"
                :key="proxy.id"
                class="px-0 py-1"
              >
                <template v-slot:prepend>
                  <v-icon size="x-small" color="grey-darken-2"
                    >mdi-server-network</v-icon
                  >
                </template>
                <v-list-item-title class="text-body-2 text-grey-darken-3">
                  {{ proxy.name }}
                </v-list-item-title>
                <v-list-item-subtitle class="text-caption text-grey-darken-2">
                  {{ proxy.domain }} → {{ proxy.target_url }}
                </v-list-item-subtitle>
                <template v-slot:append>
                  <v-chip
                    :color="getProxyStatusColor(proxy.status)"
                    size="x-small"
                    variant="outlined"
                  >
                    {{ proxy.status }}
                  </v-chip>
                </template>
              </v-list-item>
            </v-list>
          </div>
        </div>
      </v-expand-transition>
    </v-card-text>
  </v-card>
</template>

<script setup lang="ts">
import { ref } from 'vue';
import apiService from '../../services/api';
import type { Certificate, Proxy } from '../../types/api';

interface Props {
  certificate: Certificate;
}

const props = defineProps<Props>();

const emit = defineEmits<{
  deleted: [id: number];
  renewed: [certificate: Certificate];
}>();

const expanded = ref(false);
const showProxies = ref(false);
const proxies = ref<Proxy[]>([]);
const isRenewing = ref(false);
const isDeleting = ref(false);

const toggleExpanded = () => {
  expanded.value = !expanded.value;
};

const viewProxies = async () => {
  if (showProxies.value) {
    showProxies.value = false;
    return;
  }

  try {
    const response = await apiService.getCertificateProxies(
      props.certificate.id
    );
    proxies.value = response.data;
    showProxies.value = true;
  } catch (error) {
    console.error('Failed to fetch certificate proxies:', error);
  }
};

const renewCertificate = async () => {
  if (isRenewing.value) return;

  isRenewing.value = true;
  try {
    const response = await apiService.renewCertificate(props.certificate.id);
    emit('renewed', response.data);
  } catch (error) {
    console.error('Failed to renew certificate:', error);
  } finally {
    isRenewing.value = false;
  }
};

const deleteCertificate = async () => {
  if (isDeleting.value) return;

  if (
    !confirm(
      `Are you sure you want to delete the certificate for ${props.certificate.domain}?`
    )
  ) {
    return;
  }

  isDeleting.value = true;
  try {
    await apiService.deleteCertificate(props.certificate.id);
    emit('deleted', props.certificate.id);
  } catch (error) {
    console.error('Failed to delete certificate:', error);
  } finally {
    isDeleting.value = false;
  }
};

const getStatusColor = () => {
  if (!props.certificate.is_valid) return 'red-lighten-5';

  const now = new Date();
  const expiry = new Date(props.certificate.expires_at);
  const daysUntilExpiry = Math.ceil(
    (expiry.getTime() - now.getTime()) / (1000 * 60 * 60 * 24)
  );

  if (daysUntilExpiry < 0) return 'red-lighten-5';
  if (daysUntilExpiry < 7) return 'orange-lighten-5';
  if (daysUntilExpiry < 30) return 'yellow-lighten-5';
  return 'green-lighten-5';
};

const getStatusIconColor = () => {
  if (!props.certificate.is_valid) return 'red';

  const now = new Date();
  const expiry = new Date(props.certificate.expires_at);
  const daysUntilExpiry = Math.ceil(
    (expiry.getTime() - now.getTime()) / (1000 * 60 * 60 * 24)
  );

  if (daysUntilExpiry < 0) return 'red';
  if (daysUntilExpiry < 7) return 'orange';
  if (daysUntilExpiry < 30) return 'yellow';
  return 'green';
};

const getStatusIcon = () => {
  if (!props.certificate.is_valid) return 'mdi-certificate-remove';

  const now = new Date();
  const expiry = new Date(props.certificate.expires_at);
  const daysUntilExpiry = Math.ceil(
    (expiry.getTime() - now.getTime()) / (1000 * 60 * 60 * 24)
  );

  if (daysUntilExpiry < 0) return 'mdi-certificate-remove';
  if (daysUntilExpiry < 7) return 'mdi-certificate-alert';
  if (daysUntilExpiry < 30) return 'mdi-certificate-clock';
  return 'mdi-certificate';
};

const getValidityChipColor = () => {
  return props.certificate.is_valid ? 'green' : 'red';
};

const getExpiryChipColor = () => {
  const now = new Date();
  const expiry = new Date(props.certificate.expires_at);
  const daysUntilExpiry = Math.ceil(
    (expiry.getTime() - now.getTime()) / (1000 * 60 * 60 * 24)
  );

  if (daysUntilExpiry < 0) return 'red';
  if (daysUntilExpiry < 7) return 'orange';
  if (daysUntilExpiry < 30) return 'yellow';
  return 'green';
};

const getExpiryText = () => {
  const now = new Date();
  const expiry = new Date(props.certificate.expires_at);
  const daysUntilExpiry = Math.ceil(
    (expiry.getTime() - now.getTime()) / (1000 * 60 * 60 * 24)
  );

  if (daysUntilExpiry < 0) return 'Expired';
  if (daysUntilExpiry === 0) return 'Today';
  if (daysUntilExpiry === 1) return 'Tomorrow';
  if (daysUntilExpiry < 7) return `${daysUntilExpiry}d`;
  if (daysUntilExpiry < 30) return `${daysUntilExpiry}d`;
  return `${daysUntilExpiry}d`;
};

const getDaysUntilExpiry = () => {
  const now = new Date();
  const expiry = new Date(props.certificate.expires_at);
  const daysUntilExpiry = Math.ceil(
    (expiry.getTime() - now.getTime()) / (1000 * 60 * 60 * 24)
  );

  if (daysUntilExpiry < 0) return 'Expired';
  if (daysUntilExpiry === 0) return 'Expires today';
  if (daysUntilExpiry === 1) return 'Expires tomorrow';
  return `${daysUntilExpiry} days`;
};

const getProxyStatusColor = (status: string): string => {
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
.font-mono {
  font-family: 'Courier New', monospace;
}

.action-btn {
  min-width: auto;
  padding: 4px 8px;
  font-size: 0.75rem;
  text-transform: none;
  letter-spacing: normal;
}

/* On mobile, make buttons more touchable */
@media (max-width: 768px) {
  .action-btn {
    min-height: 36px;
    touch-action: manipulation;
    padding: 8px 12px;
  }
}
</style>
