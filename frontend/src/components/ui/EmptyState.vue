<template>
  <div class="text-center py-8">
    <v-icon :size="iconSize" :color="iconColor" class="mb-4">{{ icon }}</v-icon>
    <h3 class="text-h5 font-weight-medium text-grey-darken-2 mb-2">
      {{ title }}
    </h3>
    <p class="text-body-1 text-grey-darken-1 mb-4">
      {{ description }}
    </p>
    <v-btn
      v-if="actionButton"
      :color="actionButton.color || 'primary'"
      :variant="actionButton.variant || 'text'"
      :prepend-icon="actionButton.icon"
      @click="$emit('action')"
    >
      {{ actionButton.text }}
    </v-btn>
  </div>
</template>

<script setup lang="ts">
interface ActionButton {
  text: string;
  icon?: string;
  color?: string;
  variant?: 'flat' | 'text' | 'elevated' | 'tonal' | 'outlined' | 'plain';
}

interface Props {
  icon: string;
  title: string;
  description: string;
  iconSize?: number | string;
  iconColor?: string;
  actionButton?: ActionButton;
}

withDefaults(defineProps<Props>(), {
  iconSize: 64,
  iconColor: 'grey-lighten-1'
});

defineEmits<{
  action: [];
}>();
</script>
