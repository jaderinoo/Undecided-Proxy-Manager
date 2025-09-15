<template>
  <v-card>
    <v-card-title>
      <TitleWithActions
        title="Container Overview"
        icon="mdi-docker"
        :buttons="actionButtons"
      />
    </v-card-title>
    <v-card-text>
      <ErrorAlert
        :error="error"
        @clear="$emit('clear-error')"
      />

      <LoadingSpinner v-if="loading" />

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
                  @click="$emit('create-proxy-for-container', container)"
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
            @click="$emit('toggle-display')"
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
          <template v-slot:default>
            <v-icon size="100" color="grey-lighten-1"
              >mdi-docker</v-icon
            >
          </template>
        </v-empty-state>
      </div>
    </v-card-text>
  </v-card>
</template>

<script setup lang="ts">
import { computed } from 'vue';
import type { Container } from '../../types/api';
import ErrorAlert from '../ui/ErrorAlert.vue';
import LoadingSpinner from '../ui/LoadingSpinner.vue';
import TitleWithActions from '../ui/TitleWithActions.vue';

interface Props {
  containers: Container[];
  loading: boolean;
  error: string | null;
  showAllContainers: boolean;
  maxDisplayedContainers?: number;
}

const props = withDefaults(defineProps<Props>(), {
  maxDisplayedContainers: 5,
});

const emit = defineEmits<{
  'refresh': [];
  'clear-error': [];
  'create-proxy-for-container': [container: Container];
  'toggle-display': [];
}>();

const actionButtons = computed(() => [
  {
    key: 'refresh',
    color: 'primary',
    variant: 'text',
    size: 'small',
    icon: 'mdi-refresh',
    loading: props.loading,
    tooltip: 'Refresh containers',
    onClick: () => emit('refresh'),
  },
]);

const displayedContainers = computed(() => {
  if (props.showAllContainers) {
    return props.containers;
  }
  return props.containers.slice(0, props.maxDisplayedContainers);
});

const hasMoreContainers = computed(() => {
  return props.containers.length > props.maxDisplayedContainers;
});

const runningContainers = computed(() => {
  return props.containers.filter(c => c.state === 'running').length;
});

const stoppedContainers = computed(() => {
  return props.containers.filter(c => c.state === 'exited' || c.state === 'stopped').length;
});

const createdContainers = computed(() => {
  return props.containers.filter(c => c.state === 'created').length;
});

const getContainerStatusColor = (state: string) => {
  switch (state) {
    case 'running':
      return 'green';
    case 'exited':
    case 'stopped':
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
    case 'stopped':
      return 'mdi-stop-circle';
    case 'created':
      return 'mdi-plus-circle';
    case 'paused':
      return 'mdi-pause-circle';
    default:
      return 'mdi-help-circle';
  }
};
</script>
