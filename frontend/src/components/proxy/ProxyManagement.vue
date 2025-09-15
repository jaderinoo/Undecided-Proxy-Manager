<template>
  <v-card>
    <v-card-title>
      <TitleWithActions
        title="Proxy Management"
        icon="mdi-server-network"
        :buttons="actionButtons"
      />
    </v-card-title>
    <v-card-text>
      <ErrorAlert :error="error" @clear="$emit('clear-error')" />

      <LoadingSpinner v-if="loading" />

      <div v-else>
        <v-list>
          <ProxyCard
            v-for="proxy in proxies"
            :key="proxy.id"
            :proxy="proxy"
            @edit="$emit('edit-proxy', $event)"
            @delete="$emit('delete-proxy', $event)"
          />
        </v-list>

        <v-empty-state
          v-if="proxies && proxies.length === 0"
          title="No proxies found"
          text="Create your first proxy to get started"
        >
          <template v-slot:default>
            <v-icon size="100" color="grey-lighten-1"
              >mdi-server-network</v-icon
            >
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
import TitleWithActions from '../ui/TitleWithActions.vue';
import ProxyCard from './ProxyCard.vue';

interface Props {
  proxies: Proxy[];
  loading: boolean;
  reloadingNginx: boolean;
  error: string | null;
}

const props = defineProps<Props>();

const emit = defineEmits<{
  'create-proxy': [];
  'refresh': [];
  'reload-nginx': [];
  'edit-proxy': [proxy: Proxy];
  'delete-proxy': [proxy: Proxy];
  'clear-error': [];
}>();

const actionButtons = computed(() => [
  {
    key: 'create',
    color: 'success',
    variant: 'text',
    size: 'default',
    icon: 'mdi-plus',
    text: 'Add Proxy',
    onClick: () => emit('create-proxy'),
  },
  {
    key: 'refresh',
    color: 'primary',
    variant: 'text',
    size: 'small',
    icon: 'mdi-refresh',
    loading: props.loading,
    tooltip: 'Refresh proxies',
    onClick: () => emit('refresh'),
  },
  {
    key: 'reload-nginx',
    color: 'orange',
    variant: 'text',
    size: 'small',
    icon: 'mdi-reload',
    loading: props.reloadingNginx,
    tooltip: 'Reload Nginx',
    onClick: () => emit('reload-nginx'),
  },
]);
</script>
