<template>
  <div class="title-with-actions">
    <!-- Desktop layout: title and buttons side by side -->
      <div class="title-section">
        <v-icon v-if="icon" left>{{ icon }}</v-icon>
        <span class="responsive-title">{{ title }}</span>
      </div>

    <!-- Action buttons container -->
    <div class="actions-container" :class="{ 'mobile-layout': isMobile }">
      <ActionButtons :buttons="buttons" />
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted, onUnmounted, ref } from 'vue';
import ActionButtons from './ActionButtons.vue';

interface ActionButton {
  key: string;
  color?: string;
  variant?: 'flat' | 'text' | 'elevated' | 'tonal' | 'outlined' | 'plain';
  size?: string;
  icon?: string;
  text?: string;
  loading?: boolean;
  disabled?: boolean;
  tooltip?: string;
  onClick: () => void;
}

interface Props {
  title: string;
  icon?: string;
  buttons: ActionButton[];
}

defineProps<Props>();

const isMobile = ref(false);

const checkMobile = () => {
  isMobile.value = window.innerWidth <= 768;
};

onMounted(() => {
  checkMobile();
  window.addEventListener('resize', checkMobile);
});

onUnmounted(() => {
  window.removeEventListener('resize', checkMobile);
});
</script>

<style scoped>
.title-with-actions {
  display: flex;
  align-items: center;
  justify-content: space-between;
  width: 100%;
  min-height: 48px;
  gap: var(--space-2);
  min-width: 0;
  overflow: hidden;
}

.title-section {
  display: flex;
  align-items: center;
  flex-shrink: 0;
  min-width: 0;
  max-width: 60%;
  gap: var(--space-1);
}

.actions-container {
  display: flex;
  align-items: center;
  flex-shrink: 0;
  flex-grow: 0;
  min-width: 0;
  max-width: 40%;
}

/* Mobile layout - stack vertically when necessary */
@media (max-width: 600px) {
  .title-with-actions {
    flex-direction: column;
    align-items: stretch;
    gap: var(--space-2);
  }

  .title-section {
    justify-content: center;
    text-align: center;
    max-width: 100%;
  }

  .actions-container.mobile-layout {
    width: 100%;
    justify-content: center;
    max-width: 100%;
  }
}

/* Tablet layout - maintain horizontal with adjusted spacing */
@media (max-width: 768px) and (min-width: 601px) {
  .title-section {
    max-width: 50%;
  }

  .actions-container {
    max-width: 50%;
  }
}

/* Very small screens */
@media (max-width: 480px) {
  .title-with-actions {
    gap: var(--space-1);
  }

  .title-text {
    font-size: 1.125rem;
  }
}
</style>
