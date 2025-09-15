<template>
  <v-card :color="color" variant="outlined">
    <v-card-text class="text-center">
      <v-icon :color="iconColor" size="large">{{ icon }}</v-icon>
      <div class="text-h6">{{ value }}</div>
      <div class="text-caption">{{ label }}</div>
    </v-card-text>
  </v-card>
</template>

<script setup lang="ts">
interface Props {
  value: string | number;
  label: string;
  icon: string;
  color?: string;
  iconColor?: string;
}

withDefaults(defineProps<Props>(), {
  color: 'primary',
  iconColor: 'primary'
});
</script>
