<template>
  <v-row class="mb-4">
    <v-col
      v-for="stat in stats"
      :key="stat.key"
      cols="12"
      sm="6"
      :md="12 / stats.length"
    >
      <v-card :color="stat.color" variant="outlined">
        <v-card-text class="text-center">
          <v-icon :color="stat.iconColor" size="large">{{ stat.icon }}</v-icon>
          <div class="text-h6 text-grey-darken-3">{{ stat.value }}</div>
          <div class="text-caption text-grey-darken-2">{{ stat.label }}</div>
        </v-card-text>
      </v-card>
    </v-col>
  </v-row>
</template>

<script setup lang="ts">
interface StatItem {
  key: string
  value: number | string
  label: string
  icon: string
  color: string
  iconColor: string
}

interface Props {
  stats: StatItem[]
}

defineProps<Props>()
</script>
