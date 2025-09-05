<template>
  <v-btn icon @click="toggleTheme">
    <v-icon>{{
      theme.global.name.value === 'dark'
        ? 'mdi-weather-sunny'
        : 'mdi-weather-night'
    }}</v-icon>
  </v-btn>
</template>

<script setup>
import { useTheme } from 'vuetify';
const theme = useTheme();
const toggleTheme = () => {
  const newTheme = theme.global.name.value === 'light' ? 'dark' : 'light';
  theme.global.name.value = newTheme;
  document.documentElement.dataset.theme = newTheme;
  localStorage.setItem('theme', newTheme);
};
</script>
