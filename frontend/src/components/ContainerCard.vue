<template>
  <v-card class="mb-3" :color="getStatusColor()" variant="outlined">
    <v-card-title class="d-flex align-center">
      <v-icon :color="getStatusIconColor()" class="mr-2">
        {{ getStatusIcon() }}
      </v-icon>
      <span class="text-h6">{{ container.name || 'Unnamed Container' }}</span>
      <v-spacer></v-spacer>
      <v-chip 
        :color="getStatusChipColor()" 
        size="small"
        class="mr-2"
      >
        {{ container.state }}
      </v-chip>
      <v-btn
        icon
        size="small"
        @click="toggleExpanded"
      >
        <v-icon>{{ expanded ? 'mdi-chevron-up' : 'mdi-chevron-down' }}</v-icon>
      </v-btn>
    </v-card-title>

    <v-card-text>
      <v-row>
        <v-col cols="12" md="6">
          <div class="text-subtitle2 text-grey-darken-1 mb-1">Image</div>
          <div class="text-body-2 mb-2">{{ container.image }}</div>
          
          <div class="text-subtitle2 text-grey-darken-1 mb-1">Status</div>
          <div class="text-body-2 mb-2">{{ container.status }}</div>
        </v-col>
        
        <v-col cols="12" md="6">
          <div class="text-subtitle2 text-grey-darken-1 mb-1">Created</div>
          <div class="text-body-2 mb-2">{{ formatDate(container.created) }}</div>
          
          <div class="text-subtitle2 text-grey-darken-1 mb-1">Network</div>
          <div class="text-body-2 mb-2">{{ container.network_mode }}</div>
        </v-col>
      </v-row>

      <!-- Ports -->
      <div v-if="container.ports && container.ports.length > 0" class="mt-3">
        <div class="text-subtitle2 text-grey-darken-1 mb-2">Ports</div>
        <div class="d-flex flex-wrap gap-2">
          <v-chip
            v-for="port in container.ports"
            :key="`${port.private_port}-${port.public_port}`"
            size="small"
            color="primary"
            variant="outlined"
          >
            {{ port.public_port }}:{{ port.private_port }}/{{ port.type }}
          </v-chip>
        </div>
      </div>

      <!-- Expanded Details -->
      <v-expand-transition>
        <div v-if="expanded" class="mt-4">
          <v-divider class="mb-4"></v-divider>
          
          <v-row>
            <v-col cols="12" md="6">
              <div class="text-subtitle2 text-grey-darken-1 mb-2">Container ID</div>
              <div class="text-body-2 mb-3 font-mono">{{ container.id.substring(0, 12) }}...</div>
              
              <div class="text-subtitle2 text-grey-darken-1 mb-2">Command</div>
              <div class="text-body-2 mb-3 font-mono">{{ container.command || 'N/A' }}</div>
              
              <div class="text-subtitle2 text-grey-darken-1 mb-2">Size</div>
              <div class="text-body-2 mb-3">
                RW: {{ formatBytes(container.size_rw) }}<br>
                RootFS: {{ formatBytes(container.size_root_fs) }}
              </div>
            </v-col>
            
            <v-col cols="12" md="6">
              <div v-if="container.started_at" class="mb-3">
                <div class="text-subtitle2 text-grey-darken-1 mb-1">Started</div>
                <div class="text-body-2">{{ formatDate(container.started_at) }}</div>
              </div>
              
              <div v-if="container.finished_at" class="mb-3">
                <div class="text-subtitle2 text-grey-darken-1 mb-1">Finished</div>
                <div class="text-body-2">{{ formatDate(container.finished_at) }}</div>
              </div>
            </v-col>
          </v-row>

          <!-- Mounts -->
          <div v-if="container.mounts && container.mounts.length > 0" class="mt-3">
            <div class="text-subtitle2 text-grey-darken-1 mb-2">Mounts</div>
            <v-list density="compact">
              <v-list-item
                v-for="mount in container.mounts"
                :key="mount.destination"
                class="px-0"
              >
                <template v-slot:prepend>
                  <v-icon size="small">mdi-folder</v-icon>
                </template>
                <v-list-item-title class="text-body-2">
                  {{ mount.source }} → {{ mount.destination }}
                </v-list-item-title>
                <v-list-item-subtitle class="text-caption">
                  {{ mount.type }} ({{ mount.mode }})
                </v-list-item-subtitle>
              </v-list-item>
            </v-list>
          </div>

          <!-- Labels -->
          <div v-if="container.labels && Object.keys(container.labels).length > 0" class="mt-3">
            <div class="text-subtitle2 text-grey-darken-1 mb-2">Labels</div>
            <div class="d-flex flex-wrap gap-1">
              <v-chip
                v-for="(value, key) in container.labels"
                :key="key"
                size="x-small"
                color="grey"
                variant="outlined"
              >
                {{ key }}={{ value }}
              </v-chip>
            </div>
          </div>
        </div>
      </v-expand-transition>
    </v-card-text>
  </v-card>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import type { Container } from '../types/api'

interface Props {
  container: Container
}

const props = defineProps<Props>()
const expanded = ref(false)

const toggleExpanded = () => {
  expanded.value = !expanded.value
}

const getStatusColor = () => {
  switch (props.container.state) {
    case 'running':
      return 'green-lighten-5'
    case 'exited':
      return 'red-lighten-5'
    case 'created':
      return 'blue-lighten-5'
    case 'paused':
      return 'orange-lighten-5'
    default:
      return 'grey-lighten-5'
  }
}

const getStatusIconColor = () => {
  switch (props.container.state) {
    case 'running':
      return 'green'
    case 'exited':
      return 'red'
    case 'created':
      return 'blue'
    case 'paused':
      return 'orange'
    default:
      return 'grey'
  }
}

const getStatusIcon = () => {
  switch (props.container.state) {
    case 'running':
      return 'mdi-play-circle'
    case 'exited':
      return 'mdi-stop-circle'
    case 'created':
      return 'mdi-plus-circle'
    case 'paused':
      return 'mdi-pause-circle'
    default:
      return 'mdi-help-circle'
  }
}

const getStatusChipColor = () => {
  switch (props.container.state) {
    case 'running':
      return 'green'
    case 'exited':
      return 'red'
    case 'created':
      return 'blue'
    case 'paused':
      return 'orange'
    default:
      return 'grey'
  }
}

const formatDate = (dateString: string) => {
  return new Date(dateString).toLocaleString()
}

const formatBytes = (bytes: number) => {
  if (bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
}
</script>

<style scoped>
.font-mono {
  font-family: 'Courier New', monospace;
}
</style>
