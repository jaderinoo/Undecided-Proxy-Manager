<template>
  <v-alert
    v-if="error"
    type="error"
    closable
    @click:close="$emit('clear')"
    class="mb-4"
  >
    {{ error }}
  </v-alert>
</template>

<script setup lang="ts">
defineProps<{
  error: string | null;
}>();

defineEmits<{
  clear: [];
}>();
</script>
