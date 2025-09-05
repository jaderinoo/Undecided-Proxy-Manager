<template>
  <div class="d-flex align-center justify-space-between mb-4">
    <v-chip v-if="showChip" :color="chipColor" :variant="chipVariant">
      {{ count }} {{ itemName }}
    </v-chip>
    
    <div class="d-flex gap-2">
      <v-btn
        v-if="showRefresh"
        color="primary"
        variant="outlined"
        size="small"
        @click="$emit('refresh')"
        :loading="loading"
      >
        <v-icon left>mdi-refresh</v-icon>
        Refresh
      </v-btn>
      
      <slot name="actions" />
    </div>
  </div>
</template>

<script setup lang="ts">
interface Props {
  count: number
  itemName: string
  chipColor?: string
  chipVariant?: 'flat' | 'outlined' | 'text' | 'tonal'
  showRefresh?: boolean
  loading?: boolean
  showChip?: boolean
}

withDefaults(defineProps<Props>(), {
  showChip: true,
  chipColor: 'primary',
  chipVariant: 'flat',
  showRefresh: true,
  loading: false
})

defineEmits<{
  refresh: []
}>()
</script>

<style scoped>
.gap-2 {
  gap: 8px;
}
</style>
