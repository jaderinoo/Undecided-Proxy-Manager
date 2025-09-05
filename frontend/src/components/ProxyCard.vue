<template>
  <v-card class="mb-2" :color="getStatusColor(proxy.status)" variant="outlined">
    <v-card-title class="d-flex align-center py-2">
      <v-icon :color="getStatusIconColor(proxy.status)" class="mr-2" size="small">
        {{ getStatusIcon(proxy.status) }}
      </v-icon>
      <span class="text-subtitle-1 font-weight-medium text-grey-darken-3">{{ proxy.name }}</span>
      <v-spacer></v-spacer>
      <v-chip 
        :color="getStatusChipColor(proxy.status)" 
        size="x-small"
        class="mr-2"
      >
        {{ proxy.status }}
      </v-chip>
      <v-chip
        v-if="proxy.ssl_enabled"
        color="success"
        size="x-small"
        class="mr-2"
      >
        <v-icon left size="x-small">mdi-lock</v-icon>
        SSL
      </v-chip>
      <v-btn
        icon
        size="x-small"
        @click="$emit('edit', proxy)"
      >
        <v-icon size="small" color="grey-darken-2">mdi-pencil</v-icon>
      </v-btn>
      <v-btn
        icon
        size="x-small"
        @click="$emit('delete', proxy)"
        class="ml-1"
      >
        <v-icon size="small" color="grey-darken-2">mdi-delete</v-icon>
      </v-btn>
    </v-card-title>

    <v-card-text class="py-2">
      <v-row>
        <v-col cols="12" md="6">
          <div class="text-caption text-grey-darken-2 mb-1">Domain</div>
          <div class="text-body-2 text-grey-darken-3 mb-1">{{ proxy.domain }}</div>
          
          <div class="text-caption text-grey-darken-2 mb-1">Target URL</div>
          <div class="text-body-2 text-grey-darken-3 mb-1">{{ proxy.target_url }}</div>
        </v-col>
        
        <v-col cols="12" md="6">
          <div class="text-caption text-grey-darken-2 mb-1">Created</div>
          <div class="text-body-2 text-grey-darken-3 mb-1">{{ formatDate(proxy.created_at) }}</div>
          
          <div class="text-caption text-grey-darken-2 mb-1">Updated</div>
          <div class="text-body-2 text-grey-darken-3 mb-1">{{ formatDate(proxy.updated_at) }}</div>
        </v-col>
      </v-row>

      <!-- Connected Containers -->
      <div v-if="proxy.connected_containers && proxy.connected_containers.length > 0" class="mt-2">
        <div class="text-caption text-grey-darken-2 mb-1">Connected Containers</div>
        <div class="d-flex flex-wrap gap-1">
          <v-chip
            v-for="container in proxy.connected_containers"
            :key="container.id"
            size="x-small"
            color="primary"
            variant="outlined"
          >
            <v-icon left size="x-small">mdi-docker</v-icon>
            {{ container.name || 'Unnamed' }}
          </v-chip>
        </div>
      </div>
    </v-card-text>
  </v-card>
</template>

<script setup lang="ts">
import type { Proxy } from '../types/api'

defineProps<{
  proxy: Proxy
}>()

defineEmits<{
  edit: [proxy: Proxy]
  delete: [proxy: Proxy]
}>()

const getStatusColor = (status: string): string => {
  switch (status) {
    case 'active':
      return 'green-lighten-5'
    case 'inactive':
      return 'orange-lighten-5'
    case 'error':
      return 'red-lighten-5'
    default:
      return 'grey-lighten-5'
  }
}

const getStatusIconColor = (status: string): string => {
  switch (status) {
    case 'active':
      return 'green'
    case 'inactive':
      return 'orange'
    case 'error':
      return 'red'
    default:
      return 'grey'
  }
}

const getStatusIcon = (status: string): string => {
  switch (status) {
    case 'active':
      return 'mdi-check-circle'
    case 'inactive':
      return 'mdi-pause-circle'
    case 'error':
      return 'mdi-alert-circle'
    default:
      return 'mdi-help-circle'
  }
}

const getStatusChipColor = (status: string): string => {
  switch (status) {
    case 'active':
      return 'green'
    case 'inactive':
      return 'orange'
    case 'error':
      return 'red'
    default:
      return 'grey'
  }
}

const formatDate = (dateString: string) => {
  return new Date(dateString).toLocaleString()
}
</script>
