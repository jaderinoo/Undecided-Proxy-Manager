<template>
  <v-app>
    <v-app-bar color="primary" dark>
      <v-app-bar-title>
        <v-icon left>mdi-proxy</v-icon>
        UPM Dashboard
      </v-app-bar-title>
      <v-spacer></v-spacer>
      
      <!-- User Info -->
      <v-chip color="white" text-color="primary" class="mr-2">
        <v-icon left>mdi-account</v-icon>
        {{ authStore.user?.username }}
      </v-chip>
      
      <!-- Actions -->
      <v-btn icon @click="$emit('refresh')" class="mr-2">
        <v-icon>mdi-refresh</v-icon>
      </v-btn>
      
      <v-btn icon @click="$emit('logout')">
        <v-icon>mdi-logout</v-icon>
      </v-btn>
    </v-app-bar> 

    <v-main>
      <slot />
    </v-main>
  </v-app>
</template>

<script setup lang="ts">
import { useAuthStore } from '../stores/auth'

const authStore = useAuthStore()

defineEmits<{
  refresh: []
  logout: []
}>()
</script>
