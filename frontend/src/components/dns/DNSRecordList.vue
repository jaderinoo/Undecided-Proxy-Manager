<template>
  <div v-if="records.length === 0" class="text-center py-4">
    <v-icon color="grey-lighten-1">mdi-dns</v-icon>
    <p class="text-caption text-grey-darken-1 mt-2">
      No DNS records configured
    </p>
  </div>

  <div v-else>
    <v-list density="compact">
      <v-list-item v-for="record in records" :key="record.id" class="px-0">
        <template v-slot:prepend>
          <v-icon size="small">mdi-dns</v-icon>
        </template>

        <v-list-item-title class="text-body-2">
          {{ getRecordDisplayName(record) }}
        </v-list-item-title>

        <v-list-item-subtitle class="font-mono text-caption">
          {{ record.current_ip || 'Not set' }}
        </v-list-item-subtitle>

        <v-list-item-subtitle v-if="record.allowed_ip_ranges"
          class="text-caption text-grey-darken-1 mt-1">
          Allowed: {{ record.allowed_ip_ranges }}
        </v-list-item-subtitle>

        <v-list-item-subtitle v-if="record.dynamic_dns_refresh_rate"
          class="text-caption text-blue-darken-1 mt-1">
          <v-icon size="x-small" class="mr-1">mdi-timer</v-icon>
          Auto-refresh: {{ record.dynamic_dns_refresh_rate }} min
        </v-list-item-subtitle>

        <v-list-item-subtitle v-if="record.include_backend"
          class="text-caption text-green-darken-1 mt-1">
          <v-icon size="x-small" class="mr-1">mdi-application-cog</v-icon>
          Backend API Access: {{ record.backend_url || 'Default' }}
        </v-list-item-subtitle>

        <template v-slot:append>
          <div class="d-flex" style="gap: 4px">
            <v-btn icon="mdi-refresh" size="x-small" variant="text" color="success"
              :loading="loadingUpdates[record.id]" @click="$emit('update', record.id)" />
            <v-btn icon="mdi-file-document-edit" size="x-small" variant="text" color="blue"
              :loading="loadingRegen[record.id]" @click="$emit('regenerate', record)"
              v-tooltip="'Regenerate Nginx Config'" />
            <v-btn icon="mdi-pencil" size="x-small" variant="text" @click="$emit('edit', record)" />
            <v-btn icon="mdi-delete" size="x-small" variant="text" color="error"
              @click="$emit('delete', record.id)" />
          </div>
        </template>
      </v-list-item>
    </v-list>
  </div>
</template>

<script setup lang="ts">
import type { DNSRecord } from '../../types/api';

interface Props {
  records: DNSRecord[];
  configDomain?: string;
  loadingUpdates: Record<number, boolean>;
  loadingRegen: Record<number, boolean>;
}

const props = defineProps<Props>();

defineEmits<{
  update: [recordId: number];
  regenerate: [record: DNSRecord];
  edit: [record: DNSRecord];
  delete: [recordId: number];
}>();

const getRecordDisplayName = (record: DNSRecord): string => {
  if (props.configDomain) {
    return record.host === '@' ? props.configDomain : `${record.host}.${props.configDomain}`;
  }
  return record.host;
};
</script>

<style scoped>
.font-mono {
  font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
}
</style>
