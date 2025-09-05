<template>
  <v-app>
    <v-navigation-drawer
      v-model="drawer"
      :rail="rail"
      permanent
      @click="rail = false"
    >
      <v-list-item
        prepend-icon="mdi-cog"
        title="UPM Admin"
        subtitle="Proxy Manager"
        nav
      >
        <template v-slot:append>
          <v-btn
            variant="text"
            icon="mdi-chevron-left"
            @click.stop="rail = !rail"
          ></v-btn>
        </template>
      </v-list-item>

      <v-divider></v-divider>

      <v-list density="compact" nav>
        <v-list-item
          prepend-icon="mdi-view-dashboard"
          title="Dashboard"
          value="dashboard"
          :to="{ name: 'Dashboard' }"
        ></v-list-item>

        <v-list-item
          prepend-icon="mdi-server-network"
          title="Proxies"
          value="proxies"
          :to="{ name: 'Proxies' }"
        ></v-list-item>

        <v-list-item
          prepend-icon="mdi-certificate"
          title="Certificates"
          value="certificates"
          :to="{ name: 'Certificates' }"
        ></v-list-item>

        <v-list-item
          prepend-icon="mdi-docker"
          title="Containers"
          value="containers"
          :to="{ name: 'Containers' }"
        ></v-list-item>

        <v-list-item
          prepend-icon="mdi-dns"
          title="DNS Management"
          value="dns"
          :to="{ name: 'DNS' }"
        ></v-list-item>
      </v-list>
    </v-navigation-drawer>

    <v-app-bar color="primary" dark>
      <v-app-bar-nav-icon @click="drawer = !drawer"></v-app-bar-nav-icon>
      
      <v-app-bar-title>
        <v-icon left>mdi-proxy</v-icon>
        UPM Dashboard
      </v-app-bar-title>
      
      <v-spacer></v-spacer>
      
      <!-- Actions -->
      <v-btn icon @click="$emit('refresh')" class="mr-2">
        <v-icon>mdi-refresh</v-icon>
      </v-btn>
      
      <!-- Theme Toggle -->
      <ThemeToggle class="mr-2" />
      
      <v-btn icon @click="handleLogout">
        <v-icon>mdi-logout</v-icon>
      </v-btn>
    </v-app-bar> 

    <v-main>
      <slot />
    </v-main>
  </v-app>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'
import ThemeToggle from './ThemeToggle.vue'

const router = useRouter()
const authStore = useAuthStore()

const drawer = ref(true)
const rail = ref(false)

const handleLogout = () => {
  authStore.logout()
  router.push('/login')
}

defineEmits<{
  refresh: []
}>()
</script>
