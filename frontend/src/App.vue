<template>
  <v-app>
    <router-view />
  </v-app>
</template>

<script setup lang="ts">
import { onMounted } from 'vue'
import { useAuthStore } from './stores/auth'

const authStore = useAuthStore()

onMounted(() => {
  // Initialize authentication
  authStore.initializeAuth()
})
</script>

<style>
/* Vuetify handles most styling, but we can add custom styles here if needed */
</style>
