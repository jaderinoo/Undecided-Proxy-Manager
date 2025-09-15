<template>
  <v-dialog
    :model-value="show"
    max-width="400px"
    @update:model-value="$emit('update:show', $event)"
  >
    <v-card>
      <v-card-title>
        <v-icon left :color="iconColor">{{ icon }}</v-icon>
        {{ title }}
      </v-card-title>

      <v-card-text>
        {{ message }}
      </v-card-text>

      <v-card-actions>
        <v-spacer></v-spacer>
        <v-btn
          color="grey"
          variant="text"
          @click="$emit('update:show', false)"
          :disabled="loading"
        >
          Cancel
        </v-btn>
        <v-btn
          :color="confirmColor"
          @click="$emit('confirm')"
          :loading="loading"
        >
          {{ confirmText }}
        </v-btn>
      </v-card-actions>
    </v-card>
  </v-dialog>
</template>

<script setup lang="ts">
interface Props {
  show: boolean;
  title: string;
  message: string;
  icon?: string;
  iconColor?: string;
  confirmText?: string;
  confirmColor?: string;
  loading?: boolean;
}

withDefaults(defineProps<Props>(), {
  icon: 'mdi-help-circle',
  iconColor: 'primary',
  confirmText: 'Confirm',
  confirmColor: 'primary',
  loading: false,
});

defineEmits<{
  'update:show': [value: boolean];
  confirm: [];
}>();
</script>
