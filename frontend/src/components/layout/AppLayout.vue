<template>
  <v-app>
    <v-navigation-drawer
      v-model="drawer"
      permanent
      width="280"
      class="navigation-drawer"
    >
      <v-list-item
        prepend-icon="mdi-cog"
        title="UPM Admin"
        subtitle="Proxy Manager"
        nav
        class="px-2"
      ></v-list-item>

      <v-divider></v-divider>

      <v-list density="compact" nav class="px-2">
        <v-list-item
          prepend-icon="mdi-view-dashboard"
          title="Dashboard"
          value="dashboard"
          :to="{ name: 'Dashboard' }"
          class="nav-item"
        ></v-list-item>

        <v-list-item
          prepend-icon="mdi-dns"
          title="DNS Management"
          value="dns"
          :to="{ name: 'DNS' }"
          class="nav-item"
        ></v-list-item>

        <v-list-item
          prepend-icon="mdi-docker"
          title="Containers"
          value="containers"
          :to="{ name: 'Containers' }"
          class="nav-item"
        ></v-list-item>

        <v-list-item
          prepend-icon="mdi-server-network"
          title="Proxies"
          value="proxies"
          :to="{ name: 'Proxies' }"
          class="nav-item"
        ></v-list-item>

        <v-list-item
          prepend-icon="mdi-certificate"
          title="Certificates"
          value="certificates"
          :to="{ name: 'Certificates' }"
          class="nav-item"
        ></v-list-item>

        <v-list-item
          prepend-icon="mdi-cog"
          title="Settings"
          value="settings"
          :to="{ name: 'Settings' }"
          class="nav-item"
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
import { ref } from 'vue';
import { useRouter } from 'vue-router';
import { useAuthStore } from '../../stores/auth';
import ThemeToggle from './ThemeToggle.vue';

const router = useRouter();
const authStore = useAuthStore();

const drawer = ref(true);

const handleLogout = () => {
  authStore.logout();
  router.push('/login');
};

defineEmits<{
  refresh: [];
}>();
</script>

<style scoped>
.navigation-drawer {
  transition: transform 0.3s ease;
}

.nav-item {
  min-height: 48px;
  margin: 4px 0;
  border-radius: 8px;
}

.nav-item :deep(.v-list-item__prepend) {
  margin-inline-end: 12px;
}

.nav-item :deep(.v-list-item__content) {
  padding: 8px 0;
}

/* Hover effects */
.nav-item:hover {
  background-color: rgba(var(--v-theme-primary), 0.08);
}

.nav-item.v-list-item--active {
  background-color: rgba(var(--v-theme-primary), 0.12);
  color: rgb(var(--v-theme-primary));
}

.nav-item.v-list-item--active :deep(.v-icon) {
  color: rgb(var(--v-theme-primary));
}
</style>
