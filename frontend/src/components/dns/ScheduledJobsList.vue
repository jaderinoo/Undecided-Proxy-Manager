<template>
  <v-card class="mb-4" variant="outlined">
    <v-card-title class="d-flex align-center">
      <v-icon left>mdi-timer</v-icon>
      Active Dynamic DNS Jobs
      <v-spacer></v-spacer>
      <v-btn color="primary" variant="text" size="small" @click="$emit('refresh')" :loading="loading">
        <v-icon left>mdi-refresh</v-icon>
        Refresh
      </v-btn>
    </v-card-title>
    <v-card-text>
      <div v-if="loading" class="text-center py-4">
        <v-progress-circular indeterminate color="primary" size="32" class="mb-2"></v-progress-circular>
        <p class="text-body-2 text-grey-darken-2">Loading scheduled jobs...</p>
      </div>

      <div v-else-if="!jobs || Object.keys(jobs).length === 0" class="text-center py-4">
        <v-icon size="48" color="grey-lighten-1" class="mb-2">mdi-timer-off</v-icon>
        <p class="text-body-1 text-grey-darken-2 mb-2">No Active Jobs</p>
        <p class="text-body-2 text-grey-darken-1">Create DNS records with refresh rates to see scheduled
          jobs here.</p>
      </div>

      <div v-else>
        <v-list density="compact">
          <v-list-item v-for="(job, recordId) in jobs" :key="recordId" class="mb-2">
            <template v-slot:prepend>
              <v-icon color="green" size="small">mdi-timer</v-icon>
            </template>

            <v-list-item-title class="text-body-2">
              {{ job.displayName }}
            </v-list-item-title>

            <v-list-item-subtitle class="text-caption">
              <v-chip size="x-small" :color="job.isPaused ? 'orange' : 'blue'" variant="outlined"
                class="mr-2">
                {{ job.isPaused ? 'Paused' : job.countdown }}
              </v-chip>
              <span class="text-grey-darken-1">
                {{ job.isPaused ? 'Paused' : 'Next update' }}
              </span>
            </v-list-item-subtitle>

            <template v-slot:append>
              <v-btn :icon="job.isPaused ? 'mdi-play' : 'mdi-pause'" size="x-small" variant="text"
                :color="job.isPaused ? 'success' : 'warning'"
                @click="job.isPaused ? $emit('resume', parseInt(recordId)) : $emit('pause', parseInt(recordId))"
                :loading="stoppingJobs[recordId]" />
            </template>
          </v-list-item>
        </v-list>
      </div>
    </v-card-text>
  </v-card>
</template>

<script setup lang="ts">
interface Job {
  interval: number;
  displayName: string;
  nextUpdate: string;
  countdown: string;
  isPaused: boolean;
}

interface Props {
  jobs: Record<string, Job>;
  loading: boolean;
  stoppingJobs: Record<string, boolean>;
}

defineProps<Props>();

defineEmits<{
  refresh: [];
  pause: [recordId: number];
  resume: [recordId: number];
}>();
</script>
