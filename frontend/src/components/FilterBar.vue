<template>
  <v-row class="mb-4">
    <v-col cols="12" md="6">
      <v-text-field
        :model-value="searchQuery"
        :label="searchLabel"
        prepend-inner-icon="mdi-magnify"
        variant="outlined"
        density="compact"
        clearable
        @update:model-value="$emit('update:searchQuery', $event)"
        @input="$emit('search')"
      />
    </v-col>
    <v-col cols="12" md="3">
      <v-select
        :model-value="statusFilter"
        :label="statusLabel"
        :items="statusOptions"
        variant="outlined"
        density="compact"
        clearable
        @update:model-value="$emit('update:statusFilter', $event)"
      />
    </v-col>
    <v-col cols="12" md="3">
      <v-select
        :model-value="sortBy"
        :label="sortLabel"
        :items="sortOptions"
        variant="outlined"
        density="compact"
        @update:model-value="$emit('update:sortBy', $event)"
      />
    </v-col>
  </v-row>
</template>

<script setup lang="ts">
interface FilterOption {
  title: string
  value: string
}

interface Props {
  searchQuery: string
  statusFilter: string
  sortBy: string
  searchLabel?: string
  statusLabel?: string
  sortLabel?: string
  statusOptions: FilterOption[]
  sortOptions: FilterOption[]
}

withDefaults(defineProps<Props>(), {
  searchLabel: 'Search...',
  statusLabel: 'Filter by status',
  sortLabel: 'Sort by'
})

defineEmits<{
  'update:searchQuery': [value: string]
  'update:statusFilter': [value: string]
  'update:sortBy': [value: string]
  search: []
}>()
</script>
