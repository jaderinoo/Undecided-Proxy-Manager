<template>
  <div class="action-buttons">
    <v-btn
      v-for="button in buttons"
      :key="button.key"
      :color="button.color"
      :variant="button.variant"
      :size="button.size"
      :loading="button.loading"
      :disabled="button.disabled"
      @click="button.onClick"
      v-tooltip="button.tooltip"
      :class="{ 'v-btn--icon': !button.text }"
    >
      <v-icon v-if="button.icon">{{ button.icon }}</v-icon>
      <span v-if="button.text">{{ button.text }}</span>
    </v-btn>
  </div>
</template>

<script setup lang="ts">
interface ActionButton {
  key: string;
  color?: string;
  variant?: 'outlined' | 'text' | 'flat' | 'elevated' | 'tonal' | 'plain';
  size?: string;
  icon?: string;
  text?: string;
  loading?: boolean;
  disabled?: boolean;
  tooltip?: string;
  onClick: () => void;
}

interface Props {
  buttons: ActionButton[];
}

defineProps<Props>();
</script>

<style scoped>
.action-buttons {
  display: flex;
  align-items: center;
  gap: var(--space-1);
  flex-wrap: wrap;
  justify-content: flex-end;
  min-width: 0;
}

/* Bauhaus button styling */
.action-buttons .v-btn {
  flex-shrink: 0;
  flex-grow: 0;
  white-space: nowrap;
  min-width: auto;
  max-width: none;
  font-weight: 500;
  border-radius: var(--radius-md);
  transition: all 0.2s ease;
  border: 2px solid transparent;
}


/* Tablet and smaller screens */
@media (max-width: 768px) {
  .action-buttons {
    gap: var(--space-1);
    justify-content: center;
    flex-wrap: wrap;
  }

  .action-buttons .v-btn {
    min-width: auto;
    padding: var(--space-1) var(--space-2);
    font-size: 0.875rem;
  }
}

/* Mobile devices - stack vertically with natural width */
@media (max-width: 600px) {
  .action-buttons {
    flex-direction: column;
    align-items: center;
    gap: var(--space-1);
    width: 100%;
    justify-content: center;
  }

  .action-buttons .v-btn {
    width: auto;
    min-width: 120px;
    justify-content: center;
    min-height: 40px;
    padding: var(--space-1) var(--space-2);
    flex-grow: 0;
    flex-shrink: 0;
  }
}

/* Very small screens */
@media (max-width: 400px) {
  .action-buttons .v-btn {
    font-size: 0.75rem;
    padding: var(--space-1) var(--space-2);
    min-height: 36px;
    min-width: 100px;
    width: auto;
  }

  .action-buttons .v-btn--icon {
    width: 40px;
    height: 40px;
    min-width: 40px;
  }
}

/* Touch optimization */
@media (max-width: 768px) {
  .action-buttons .v-btn {
    min-height: 36px;
    touch-action: manipulation;
  }
}

/* Text handling */
.action-buttons .v-btn span {
  display: inline-block;
  max-width: 100%;
  overflow: hidden;
  text-overflow: ellipsis;
}

/* Icon button sizing */
.action-buttons .v-btn--icon {
  min-width: 40px;
  min-height: 40px;
}

@media (max-width: 768px) {
  .action-buttons .v-btn--icon {
    min-width: 36px;
    min-height: 36px;
  }
}
</style>
