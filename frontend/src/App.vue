<template>
  <v-app>
    <v-app-bar color="primary" dark>
      <v-app-bar-title>
        <v-icon left>mdi-proxy</v-icon>
        UPM Dashboard
      </v-app-bar-title>
      <v-spacer></v-spacer>
      <v-btn icon @click="loadProxies">
        <v-icon>mdi-refresh</v-icon>
      </v-btn>
    </v-app-bar>

    <v-main>
      <v-container>
        <v-row>
          <v-col cols="12">
            <v-card>
              <v-card-title>
                <v-icon left>mdi-server-network</v-icon>
                Proxy Management
              </v-card-title>
              <v-card-text>
                <v-alert
                  v-if="error"
                  type="error"
                  closable
                  @click:close="error = null"
                >
                  {{ error }}
                </v-alert>

                <v-progress-linear
                  v-if="loading"
                  indeterminate
                  color="primary"
                ></v-progress-linear>

                <div v-else>
                  <v-chip color="primary" class="mb-4">
                    {{ proxies.length }} Proxies
                  </v-chip>

                  <v-list>
                    <v-list-item
                      v-for="proxy in proxies"
                      :key="proxy.id"
                      class="mb-2"
                    >
                      <template v-slot:prepend>
                        <v-avatar :color="getStatusColor(proxy.status)">
                          <v-icon color="white">mdi-server</v-icon>
                        </v-avatar>
                      </template>

                      <v-list-item-title>{{ proxy.name }}</v-list-item-title>
                      <v-list-item-subtitle>
                        {{ proxy.domain }} → {{ proxy.target_url }}
                      </v-list-item-subtitle>

                      <template v-slot:append>
                        <v-chip
                          :color="getStatusColor(proxy.status)"
                          size="small"
                        >
                          {{ proxy.status }}
                        </v-chip>
                        <v-chip
                          v-if="proxy.ssl_enabled"
                          color="success"
                          size="small"
                          class="ml-2"
                        >
                          <v-icon left size="small">mdi-lock</v-icon>
                          SSL
                        </v-chip>
                      </template>
                    </v-list-item>
                  </v-list>

                  <v-empty-state
                    v-if="proxies.length === 0"
                    title="No proxies found"
                    text="Create your first proxy to get started"
                  >
                    <template v-slot:image>
                      <v-icon size="100" color="grey-lighten-1">mdi-server-network</v-icon>
                    </template>
                  </v-empty-state>
                </div>
              </v-card-text>
            </v-card>
          </v-col>
        </v-row>
      </v-container>
    </v-main>
  </v-app>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { apiService } from './services/api'
import type { Proxy } from './types/api'

const proxies = ref<Proxy[]>([])
const loading = ref(true)
const error = ref<string | null>(null)

const loadProxies = async () => {
  try {
    loading.value = true
    error.value = null
    const response = await apiService.getProxies()
    proxies.value = response.data
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Failed to load proxies'
  } finally {
    loading.value = false
  }
}

const getStatusColor = (status: string): string => {
  switch (status) {
    case 'active':
      return 'success'
    case 'inactive':
      return 'warning'
    case 'error':
      return 'error'
    default:
      return 'grey'
  }
}

onMounted(() => {
  loadProxies()
})
</script>

<style>
/* Vuetify handles most styling, but we can add custom styles here if needed */
</style>
