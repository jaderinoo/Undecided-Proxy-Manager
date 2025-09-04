<template>
  <v-list-item class="mb-2">
    <template v-slot:prepend>
      <v-avatar :color="getStatusColor(proxy.status)">
        <v-icon color="white">mdi-server</v-icon>
      </v-avatar>
    </template>

    <v-list-item-title>{{ proxy.name }}</v-list-item-title>
    <v-list-item-subtitle>
      {{ proxy.domain }} → {{ proxy.target_url }}
    </v-list-item-subtitle>

    <template v-slot:append>
      <v-chip
        :color="getStatusColor(proxy.status)"
        size="small"
      >
        {{ proxy.status }}
      </v-chip>
      <v-chip
        v-if="proxy.ssl_enabled"
        color="success"
        size="small"
        class="ml-2"
      >
        <v-icon left size="small">mdi-lock</v-icon>
        SSL
      </v-chip>
    </template>
  </v-list-item>
</template>

<script setup lang="ts">
import type { Proxy } from '../types/api'

defineProps<{
  proxy: Proxy
}>()

const getStatusColor = (status: string): string => {
  switch (status) {
    case 'active':
      return 'success'
    case 'inactive':
      return 'warning'
    case 'error':
      return 'error'
    default:
      return 'grey'
  }
}
</script>
