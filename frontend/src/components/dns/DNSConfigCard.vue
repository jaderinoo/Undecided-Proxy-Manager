<template>
  <v-card variant="outlined" class="mb-4">
    <v-card-title>
      <div class="d-flex align-center justify-space-between w-100">
        <div>
          <div class="text-h6">{{ config.domain }}</div>
          <v-chip size="small" color="primary" variant="outlined">
            {{ config.provider }}
          </v-chip>
        </div>
        <div class="d-flex" style="gap: 4px">
          <v-btn icon="mdi-pencil" size="small" variant="text" @click="$emit('edit', config)" />
          <v-btn icon="mdi-delete" size="small" variant="text" color="error"
            @click="$emit('delete', config.id)" />
        </div>
      </div>
    </v-card-title>

    <v-card-text>
      <div class="mb-3">
        <div class="d-flex justify-space-between align-center mb-1">
          <span class="text-caption text-grey-darken-2">Status:</span>
          <v-chip :color="config.is_active ? 'green' : 'orange'" size="small" variant="outlined">
            {{ config.is_active ? 'Active' : 'Inactive' }}
          </v-chip>
        </div>
        <div class="d-flex justify-space-between align-center mb-1">
          <span class="text-caption text-grey-darken-2">Last Update:</span>
          <span class="text-body-2">{{
            formatDate(config.last_update) || 'Never'
          }}</span>
        </div>
        <div class="d-flex justify-space-between align-center">
          <span class="text-caption text-grey-darken-2">Last IP:</span>
          <span class="text-body-2 font-mono">{{
            config.last_ip || 'Unknown'
          }}</span>
        </div>
      </div>

      <v-divider class="my-3"></v-divider>

      <div class="d-flex justify-space-between align-center mb-2">
        <span class="text-subtitle-2">DNS Records</span>
        <v-btn size="small" color="primary" variant="text" prepend-icon="mdi-plus"
          @click="$emit('add-record', config.id)">
          Add Record
        </v-btn>
      </div>

      <DNSRecordList
        :records="records"
        :loading-updates="loadingUpdates"
        :loading-regen="loadingRegen"
        @update="(recordId) => $emit('update-record', recordId)"
        @regenerate="(record) => $emit('regenerate', record)"
        @edit="(record) => $emit('edit-record', record)"
        @delete="(recordId) => $emit('delete-record', recordId)"
      />
    </v-card-text>
  </v-card>
</template>

<script setup lang="ts">
import type { DNSConfig, DNSRecord } from '../../types/api';
import DNSRecordList from './DNSRecordList.vue';

interface Props {
  config: DNSConfig;
  records: DNSRecord[];
  loadingUpdates: Record<number, boolean>;
  loadingRegen: Record<number, boolean>;
}

defineProps<Props>();

defineEmits<{
  edit: [config: DNSConfig];
  delete: [configId: number];
  'add-record': [configId: number];
  'update-record': [recordId: number];
  regenerate: [record: DNSRecord];
  'edit-record': [record: DNSRecord];
  'delete-record': [recordId: number];
}>();

const formatDate = (dateString?: string) => {
  if (!dateString) return null;
  return new Date(dateString).toLocaleString();
};
</script>

<style scoped>
.font-mono {
  font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
}
</style>
