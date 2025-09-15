<template>
  <v-card>
    <v-card-title>
      <TitleWithActions
        title="Proxy Overview"
        icon="mdi-server-network"
        :buttons="actionButtons"
      />
    </v-card-title>
    <v-card-text>
      <ErrorAlert :error="error" @clear="$emit('clear-error')" />

      <LoadingSpinner v-if="loading" />

      <div v-else>
        <StatsCards :stats="proxyStats" />

        <v-divider class="my-4"></v-divider>

        <div class="d-flex justify-space-between align-center mb-3">
          <h3 class="text-h6">Recent Proxies</h3>
          <v-btn
            color="primary"
            variant="text"
            size="small"
            @click="$emit('view-all')"
          >
            View All
            <v-icon right>mdi-arrow-right</v-icon>
          </v-btn>
        </div>

        <div v-if="recentProxies.length > 0">
          <v-list density="compact">
            <v-list-item
              v-for="proxy in recentProxies"
              :key="proxy.id"
              class="px-0"
            >
              <template v-slot:prepend>
                <v-icon
                  :color="getProxyStatusColor(proxy.status)"
                  size="small"
                >
                  {{ getProxyStatusIcon(proxy.status) }}
                </v-icon>
              </template>

              <v-list-item-title>{{ proxy.name }}</v-list-item-title>
              <v-list-item-subtitle>{{ proxy.domain }}</v-list-item-subtitle>

              <template v-slot:append>
                <v-chip
                  class="mr-4"
                  :color="getProxyStatusColor(proxy.status)"
                  size="x-small"
                  variant="outlined"
                >
                  {{ proxy.status }}
                </v-chip>
                <v-chip
                  v-if="proxy.ssl_enabled"
                  color="success"
                  size="x-small"
                  class="ml-1"
                >
                  SSL
                </v-chip>
              </template>
            </v-list-item>
          </v-list>
        </div>

        <v-empty-state
          v-else
          title="No proxies found"
          text="Create your first proxy to get started"
        >
          <template v-slot:default>
            <v-btn
              color="primary"
              @click="$emit('add-proxy')"
            >
              <v-icon left>mdi-plus</v-icon>
              Add Proxy
            </v-btn>
          </template>
        </v-empty-state>
      </div>
    </v-card-text>
  </v-card>
</template>

<script setup lang="ts">
import { computed } from 'vue';
import type { Proxy } from '../../types/api';
import ErrorAlert from '../ui/ErrorAlert.vue';
import LoadingSpinner from '../ui/LoadingSpinner.vue';
import StatsCards from '../ui/StatsCards.vue';
import TitleWithActions from '../ui/TitleWithActions.vue';

interface Props {
  proxies: Proxy[];
  loading: boolean;
  error: string | null;
}

const props = defineProps<Props>();

const emit = defineEmits<{
  refresh: [];
  'clear-error': [];
  'view-all': [];
  'add-proxy': [];
}>();

const actionButtons = computed(() => [
  {
    key: 'refresh',
    color: 'primary',
    variant: 'text' as const,
    size: 'small',
    icon: 'mdi-refresh',
    loading: props.loading,
    tooltip: 'Refresh proxies',
    onClick: () => emit('refresh'),
  },
]);

// Computed properties
const activeProxies = computed(
  () => props.proxies.filter(p => p.status === 'active').length
);

const inactiveProxies = computed(
  () => props.proxies.filter(p => p.status === 'inactive').length
);

const sslCount = computed(
  () => props.proxies.filter(p => p.ssl_enabled).length
);

const errorProxies = computed(
  () => props.proxies.filter(p => p.status === 'error').length
);

const recentProxies = computed(() => {
  return props.proxies
    .sort((a, b) => new Date(b.created_at).getTime() - new Date(a.created_at).getTime())
    .slice(0, 5); // Show only the 5 most recent proxies
});

const proxyStats = computed(() => [
  {
    key: 'active',
    value: activeProxies.value,
    label: 'Active Proxies',
    icon: 'mdi-check-circle',
    color: 'green-lighten-5',
    iconColor: 'green',
  },
  {
    key: 'inactive',
    value: inactiveProxies.value,
    label: 'Inactive Proxies',
    icon: 'mdi-pause-circle',
    color: 'orange-lighten-5',
    iconColor: 'orange',
  },
  {
    key: 'ssl',
    value: sslCount.value,
    label: 'SSL Enabled',
    icon: 'mdi-lock',
    color: 'blue-lighten-5',
    iconColor: 'blue',
  },
  {
    key: 'error',
    value: errorProxies.value,
    label: 'Error Proxies',
    icon: 'mdi-alert-circle',
    color: 'red-lighten-5',
    iconColor: 'red',
  },
]);

// Proxy status helper methods
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

const getProxyStatusIcon = (status: string): string => {
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
</script>
