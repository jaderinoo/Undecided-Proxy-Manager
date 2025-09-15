<template>
  <v-card class="mb-4" variant="outlined">
    <v-card-title class="d-flex align-center">
      <v-icon left :color="iconColor">{{ icon }}</v-icon>
      {{ title }}
      <v-spacer></v-spacer>
      <v-btn
        v-if="actionButton"
        :color="actionButton.color || 'primary'"
        variant="text"
        size="small"
        :loading="actionButton.loading"
        @click="$emit('action')"
      >
        <v-icon v-if="actionButton.icon" left>{{ actionButton.icon }}</v-icon>
        {{ actionButton.text }}
      </v-btn>
    </v-card-title>
    <v-card-text>
      <slot>
        <div v-if="content" class="d-flex align-center">
          <span class="text-body-1 mr-2">{{ contentLabel }}:</span>
          <v-chip v-if="chipContent" :color="chipColor" variant="outlined" :class="chipClass">
            {{ chipContent }}
          </v-chip>
          <span v-else class="text-body-2">{{ content }}</span>
        </div>
      </slot>
    </v-card-text>
  </v-card>
</template>

<script setup lang="ts">
interface ActionButton {
  text: string;
  icon?: string;
  color?: string;
  loading?: boolean;
}

interface Props {
  title: string;
  icon: string;
  iconColor?: string;
  content?: string;
  contentLabel?: string;
  chipContent?: string;
  chipColor?: string;
  chipClass?: string;
  actionButton?: ActionButton;
}

defineProps<Props>();

defineEmits<{
  action: [];
}>();
</script>
