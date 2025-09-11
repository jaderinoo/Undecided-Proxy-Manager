<template>
  <AppLayout>
    <v-container>
      <v-row>
        <v-col cols="12">
          <v-card>
            <v-card-title>
              <v-icon left>mdi-dns</v-icon>
              DNS Management
            </v-card-title>
            <v-card-text>
              <ErrorAlert v-if="error" :error="error" @clear="error = null" />

              <PageHeader :count="dnsConfigs?.length || 0" item-name="DNS Configurations" :show-refresh="false">
                <template #actions>
                  <v-btn color="success" variant="outlined" size="small" @click="showCreateConfigModal = true">
                    <v-icon left>mdi-plus</v-icon>
                    Add Configuration
                  </v-btn>
                </template>
              </PageHeader>

              <!-- Public IP Display -->
              <v-card class="mb-4" variant="outlined">
                <v-card-title>
                  <v-icon left>mdi-earth</v-icon>
                  Current Public IP
                  <v-spacer></v-spacer>
                  <v-btn color="primary" variant="outlined" size="small" @click="refreshPublicIP"
                    :loading="loadingPublicIP">
                    <v-icon left>mdi-refresh</v-icon>
                    Refresh
                  </v-btn>
                </v-card-title>
                <v-card-text>
                  <div class="d-flex align-center">
                    <span class="text-body-1 mr-2">Public IP:</span>
                    <v-chip color="primary" variant="outlined" class="font-mono">
                      {{ publicIP || 'Loading...' }}
                    </v-chip>
                  </div>
                </v-card-text>
              </v-card>

              <!-- Nginx IP Restrictions -->
              <v-card class="mb-4" variant="outlined">
                <v-card-title>
                  <v-icon left>mdi-shield-account</v-icon>
                  Nginx Access Control
                  <v-spacer></v-spacer>
                  <v-btn color="primary" variant="outlined" size="small" @click="showNginxIPModal = true">
                    <v-icon left>mdi-cog</v-icon>
                    Configure
                  </v-btn>
                </v-card-title>
                <v-card-text>
                  <div class="d-flex align-center">
                    <span class="text-body-1 mr-2">Allowed IP Ranges:</span>
                    <div v-if="nginxAllowedRanges.length === 0" class="text-grey">
                      No restrictions (all IPs allowed)
                    </div>
                    <div v-else>
                      <v-chip v-for="range in nginxAllowedRanges" :key="range" size="small" color="primary"
                        variant="outlined" class="mr-1">
                        {{ range }}
                      </v-chip>
                    </div>
                  </div>
                </v-card-text>
              </v-card>

              <!-- DNS Stats -->
              <StatsCards :stats="dnsStats" />

              <!-- Active Dynamic DNS Jobs -->
              <v-card class="mb-4" variant="outlined">
                <v-card-title>
                  <v-icon left>mdi-timer</v-icon>
                  Active Dynamic DNS Jobs
                  <v-spacer></v-spacer>
                  <v-btn color="primary" variant="outlined" size="small" @click="refreshAllData" :loading="loadingJobs">
                    <v-icon left>mdi-refresh</v-icon>
                    Refresh
                  </v-btn>
                </v-card-title>
                <v-card-text>
                  <div v-if="loadingJobs" class="text-center py-4">
                    <v-progress-circular indeterminate color="primary" size="32" class="mb-2"></v-progress-circular>
                    <p class="text-body-2 text-grey-darken-2">Loading scheduled jobs...</p>
                  </div>

                  <div v-else-if="!scheduledJobs || Object.keys(scheduledJobs).length === 0" class="text-center py-4">
                    <v-icon size="48" color="grey-lighten-1" class="mb-2">mdi-timer-off</v-icon>
                    <p class="text-body-1 text-grey-darken-2 mb-2">No Active Jobs</p>
                    <p class="text-body-2 text-grey-darken-1">Create DNS records with refresh rates to see scheduled
                      jobs here.</p>
                  </div>

                  <div v-else>
                    <v-list density="compact">
                      <v-list-item v-for="(job, recordId) in scheduledJobsWithNames" :key="recordId" class="mb-2">
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
                            @click="job.isPaused ? resumeScheduledJob(recordId) : pauseScheduledJob(recordId)"
                            :loading="stoppingJobs[recordId]" />
                        </template>
                      </v-list-item>
                    </v-list>
                  </div>
                </v-card-text>
              </v-card>

              <!-- DNS Configurations -->
              <div v-if="loadingConfigs" class="text-center py-8">
                <v-progress-circular indeterminate color="primary" size="64" class="mb-4"></v-progress-circular>
                <p class="text-body-1 text-grey-darken-2">
                  Loading DNS configurations...
                </p>
              </div>

              <div v-else-if="!dnsConfigs || dnsConfigs.length === 0" class="text-center py-8">
                <v-icon size="64" color="grey-lighten-1" class="mb-4">mdi-dns</v-icon>
                <h3 class="text-h5 font-weight-medium text-grey-darken-2 mb-2">
                  No DNS Configurations
                </h3>
                <p class="text-body-1 text-grey-darken-1 mb-4">
                  Create your first DNS configuration to start managing dynamic
                  DNS records.
                </p>
                <v-btn color="primary" prepend-icon="mdi-plus" @click="showCreateConfigModal = true">
                  Add Configuration
                </v-btn>
              </div>

              <div v-else class="dns-configs-grid">
                <v-row>
                  <v-col v-for="config in dnsConfigs || []" :key="config.id" cols="12" md="6" lg="4">
                    <v-card variant="outlined" class="mb-4">
                      <v-card-title>
                        <div class="d-flex align-center justify-space-between w-100">
                          <div>
                            <div class="text-h6">{{ config.domain }}</div>
                            <v-chip size="small" color="primary" variant="outlined">
                              {{ config.provider }}
                            </v-chip>
                          </div>
                          <div class="d-flex" style="gap: 4px">
                            <v-btn icon="mdi-pencil" size="small" variant="text" @click="editConfig(config)" />
                            <v-btn icon="mdi-delete" size="small" variant="text" color="error"
                              @click="deleteConfig(config.id)" />
                          </div>
                        </div>
                      </v-card-title>

                      <v-card-text>
                        <div class="mb-3">
                          <div class="d-flex justify-space-between align-center mb-1">
                            <span class="text-caption text-grey-darken-2">Status:</span>
                            <v-chip :color="config.is_active ? 'green' : 'orange'" size="small" variant="outlined">
                              {{ config.is_active ? 'Active' : 'Inactive' }}
                            </v-chip>
                          </div>
                          <div class="d-flex justify-space-between align-center mb-1">
                            <span class="text-caption text-grey-darken-2">Last Update:</span>
                            <span class="text-body-2">{{
                              formatDate(config.last_update) || 'Never'
                            }}</span>
                          </div>
                          <div class="d-flex justify-space-between align-center">
                            <span class="text-caption text-grey-darken-2">Last IP:</span>
                            <span class="text-body-2 font-mono">{{
                              config.last_ip || 'Unknown'
                            }}</span>
                          </div>
                        </div>

                        <v-divider class="my-3"></v-divider>

                        <div class="d-flex justify-space-between align-center mb-2">
                          <span class="text-subtitle-2">DNS Records</span>
                          <v-btn size="small" color="primary" variant="outlined" prepend-icon="mdi-plus"
                            @click="openCreateRecordModal(config.id)">
                            Add Record
                          </v-btn>
                        </div>

                        <div v-if="configRecords[config.id]?.length === 0" class="text-center py-4">
                          <v-icon color="grey-lighten-1">mdi-dns</v-icon>
                          <p class="text-caption text-grey-darken-1 mt-2">
                            No DNS records configured
                          </p>
                        </div>

                        <div v-else>
                          <v-list density="compact">
                            <v-list-item v-for="record in configRecords[config.id] || []" :key="record.id" class="px-0">
                              <template v-slot:prepend>
                                <v-icon size="small">mdi-dns</v-icon>
                              </template>

                              <v-list-item-title class="text-body-2">
                                {{
                                  record.host === '@'
                                    ? config.domain
                                    : `${record.host}.${config.domain}`
                                }}
                              </v-list-item-title>

                              <v-list-item-subtitle class="font-mono text-caption">
                                {{ record.current_ip || 'Not set' }}
                              </v-list-item-subtitle>

                              <v-list-item-subtitle v-if="record.allowed_ip_ranges"
                                class="text-caption text-grey-darken-1 mt-1">
                                Allowed: {{ record.allowed_ip_ranges }}
                              </v-list-item-subtitle>

                              <v-list-item-subtitle v-if="record.dynamic_dns_refresh_rate"
                                class="text-caption text-blue-darken-1 mt-1">
                                <v-icon size="x-small" class="mr-1">mdi-timer</v-icon>
                                Auto-refresh: {{ record.dynamic_dns_refresh_rate }} min
                              </v-list-item-subtitle>

                              <v-list-item-subtitle v-if="record.include_backend"
                                class="text-caption text-green-darken-1 mt-1">
                                <v-icon size="x-small" class="mr-1">mdi-application-cog</v-icon>
                                Backend API Access: {{ record.backend_url || 'Default' }}
                              </v-list-item-subtitle>

                              <template v-slot:append>
                                <div class="d-flex" style="gap: 4px">
                                  <v-btn icon="mdi-refresh" size="x-small" variant="text" color="success"
                                    :loading="loadingUpdates[record.id]" @click="updateRecordNow(record.id)" />
                                  <v-btn icon="mdi-file-document-edit" size="x-small" variant="text" color="blue"
                                    :loading="loadingRegen[record.id]" @click="regenerateConfig(record)"
                                    v-tooltip="'Regenerate Nginx Config'" />
                                  <v-btn icon="mdi-pencil" size="x-small" variant="text" @click="editRecord(record)" />
                                  <v-btn icon="mdi-delete" size="x-small" variant="text" color="error"
                                    @click="deleteRecord(record.id)" />
                                </div>
                              </template>
                            </v-list-item>
                          </v-list>
                        </div>
                      </v-card-text>
                    </v-card>
                  </v-col>
                </v-row>
              </div>
            </v-card-text>
          </v-card>
        </v-col>
      </v-row>
    </v-container>

    <!-- Create/Edit Config Modal -->
    <v-dialog v-model="showConfigDialog" max-width="600px" persistent>
      <v-card>
        <v-card-title class="d-flex align-center">
          <v-icon left>mdi-dns</v-icon>
          {{
            showCreateConfigModal
              ? 'Create DNS Configuration'
              : 'Edit DNS Configuration'
          }}
          <v-spacer></v-spacer>
          <v-btn icon @click="closeConfigModal">
            <v-icon>mdi-close</v-icon>
          </v-btn>
        </v-card-title>

        <v-card-text>
          <v-form @submit.prevent="saveConfig">
            <v-row>
              <v-col cols="12">
                <v-select v-model="configForm.provider" label="DNS Provider" :items="[
                  { title: 'Namecheap Dynamic DNS', value: 'namecheap' },
                  { title: 'No Dynamic DNS (Static)', value: 'static' }
                ]" required></v-select>
              </v-col>

              <v-col cols="12">
                <v-text-field v-model="configForm.domain" label="Domain" placeholder="example.com"
                  required></v-text-field>
              </v-col>

              <!-- Dynamic DNS credentials (only for namecheap) -->
              <template v-if="configForm.provider === 'namecheap'">
                <v-col cols="12">
                  <v-text-field v-model="configForm.username" label="Username" placeholder="yourdomain.com"
                    required></v-text-field>
                </v-col>

                <v-col cols="12">
                  <!-- Show password field for create or when changing password -->
                  <div v-if="showCreateConfigModal || changePassword">
                    <div v-if="changePassword" class="d-flex align-center mb-2">
                      <v-icon color="primary" class="mr-2">mdi-key-change</v-icon>
                      <span class="text-body-2 text-primary">Changing Password</span>
                      <v-spacer></v-spacer>
                      <v-btn variant="text" size="small" color="grey"
                        @click="changePassword = false; configForm.password = ''">
                        Cancel
                      </v-btn>
                    </div>
                    <v-text-field v-model="configForm.password" label="Dynamic DNS Password" type="password"
                      placeholder="Dynamic DNS password" :required="showCreateConfigModal"></v-text-field>
                  </div>

                  <!-- Show password change option for edit -->
                  <div v-else-if="showEditConfigModal">
                    <v-alert type="info" variant="outlined" class="mb-3" density="compact">
                      <template #prepend>
                        <v-icon>mdi-information</v-icon>
                      </template>
                      Password is encrypted and hidden for security.
                      <v-btn variant="text" size="small" color="primary" @click="changePassword = true">
                        Change Password
                      </v-btn>
                    </v-alert>
                  </div>
                </v-col>
              </template>

              <!-- Static DNS notice -->
              <v-col v-else-if="configForm.provider === 'static'" cols="12">
                <v-alert type="info" variant="outlined" density="compact">
                  <template #prepend>
                    <v-icon>mdi-information</v-icon>
                  </template>
                  Static DNS mode - no dynamic updates. DNS records will be managed manually.
                </v-alert>
              </v-col>

              <v-col cols="12">
                <v-checkbox v-model="configForm.is_active" label="Active" color="primary"></v-checkbox>
              </v-col>
            </v-row>
          </v-form>
        </v-card-text>

        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn @click="closeConfigModal" color="grey"> Cancel </v-btn>
          <v-btn @click="saveConfig" :loading="savingConfig" color="primary">
            {{ showCreateConfigModal ? 'Create' : 'Update' }}
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>

    <!-- Create/Edit Record Modal -->
    <v-dialog v-model="showRecordDialog" max-width="500px" persistent>
      <v-card>
        <v-card-title class="d-flex align-center">
          <v-icon left>mdi-dns</v-icon>
          {{ showCreateRecordModal ? 'Create DNS Record' : 'Edit DNS Record' }}
          <v-spacer></v-spacer>
          <v-btn icon @click="closeRecordModal">
            <v-icon>mdi-close</v-icon>
          </v-btn>
        </v-card-title>

        <v-card-text>
          <v-form @submit.prevent="saveRecord">
            <v-row>
              <v-col cols="12">
                <v-text-field v-model="recordForm.host" label="Host" placeholder="@ for root domain, www for subdomain"
                  hint="Use '@' for root domain or enter subdomain name" persistent-hint required></v-text-field>
              </v-col>

              <v-col cols="12">
                <v-textarea v-model="recordForm.allowed_ip_ranges" label="Allowed IP Ranges"
                  placeholder="192.168.1.0/24, 10.0.0.1, 203.0.113.0/24"
                  hint="Comma-separated list of IP addresses or CIDR ranges. Leave empty to allow all IPs."
                  persistent-hint rows="3"
                  :error-messages="validateIPRanges(recordForm.allowed_ip_ranges || '') ? [validateIPRanges(recordForm.allowed_ip_ranges || '')!] : []"
                  :error="!!validateIPRanges(recordForm.allowed_ip_ranges || '')"></v-textarea>

                <!-- Quick Fill Buttons -->
                <div class="mt-2">
                  <v-chip-group>
                    <v-chip size="small" color="primary" variant="outlined"
                      @click="quickFillIPRanges('192.168.50.2/24')">
                      <v-icon start size="small">mdi-plus</v-icon>
                      192.168.50.2/24
                    </v-chip>
                    <v-chip size="small" color="primary" variant="outlined" @click="quickFillIPRanges('10.6.0.1/24')">
                      <v-icon start size="small">mdi-plus</v-icon>
                      10.6.0.1/24
                    </v-chip>
                    <v-chip size="small" color="grey" variant="outlined" @click="clearIPRanges">
                      <v-icon start size="small">mdi-close</v-icon>
                      Clear
                    </v-chip>
                  </v-chip-group>
                </div>
              </v-col>

              <v-col cols="12">
                <v-text-field v-model.number="recordForm.dynamic_dns_refresh_rate"
                  label="Dynamic DNS Refresh Rate (minutes)" type="number" placeholder="e.g., 5, 10, 30, 60"
                  hint="Set refresh rate in minutes for automatic DNS updates. Leave empty to disable auto-refresh."
                  persistent-hint min="1" max="1440"
                  :error-messages="validateRefreshRate(recordForm.dynamic_dns_refresh_rate) ? [validateRefreshRate(recordForm.dynamic_dns_refresh_rate)!] : []"
                  :error="!!validateRefreshRate(recordForm.dynamic_dns_refresh_rate)">
                </v-text-field>
              </v-col>

              <v-col cols="12">
                <v-checkbox v-model="recordForm.include_backend" label="Include Backend" color="primary"
                  hint="Enable this to include backend API routes for this domain" persistent-hint></v-checkbox>
              </v-col>

              <v-col v-if="recordForm.include_backend" cols="12">
                <v-text-field v-model="recordForm.backend_url" label="Backend URL" color="primary"
                  hint="URL of the backend service (e.g., http://backend:6080, http://api.example.com:3000)"
                  persistent-hint placeholder="http://backend:6080"></v-text-field>
              </v-col>

              <v-col cols="12">
                <v-checkbox v-model="recordForm.is_active" label="Active" color="primary"></v-checkbox>
              </v-col>
            </v-row>
          </v-form>
        </v-card-text>

        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn @click="closeRecordModal" color="grey"> Cancel </v-btn>
          <v-btn @click="saveRecord" :loading="savingRecord" color="primary">
            {{ showCreateRecordModal ? 'Create' : 'Update' }}
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>

    <!-- Error Alert -->
    <ErrorAlert v-if="error" :error="error" @clear="error = null" />

    <!-- Nginx IP Restrictions Modal -->
    <v-dialog v-model="showNginxIPModal" max-width="600px" persistent>
      <v-card>
        <v-card-title class="d-flex align-center">
          <v-icon left>mdi-shield-account</v-icon>
          Configure Nginx IP Restrictions
          <v-spacer></v-spacer>
          <v-btn icon @click="closeNginxIPModal">
            <v-icon>mdi-close</v-icon>
          </v-btn>
        </v-card-title>

        <v-card-text>
          <v-form @submit.prevent="saveNginxIPRestrictions">
            <v-row>
              <v-col cols="12">
                <v-textarea v-model="nginxIPForm.allowedRanges" label="Allowed IP Ranges"
                  placeholder="192.168.50.0/24, 10.6.0.1/32"
                  hint="Comma-separated list of IP addresses or CIDR ranges. Leave empty to allow all IPs."
                  persistent-hint rows="4"
                  :error-messages="validateIPRanges(nginxIPForm.allowedRanges || '') ? [validateIPRanges(nginxIPForm.allowedRanges || '')!] : []"
                  :error="!!validateIPRanges(nginxIPForm.allowedRanges || '')"></v-textarea>

                <!-- Quick Fill Buttons -->
                <div class="mt-2">
                  <v-chip-group>
                    <v-chip size="small" color="primary" variant="outlined"
                      @click="quickFillNginxIPRanges('192.168.50.2/24')">
                      <v-icon start size="small">mdi-plus</v-icon>
                      192.168.50.2/24
                    </v-chip>
                    <v-chip size="small" color="primary" variant="outlined"
                      @click="quickFillNginxIPRanges('10.6.0.1/24')">
                      <v-icon start size="small">mdi-plus</v-icon>
                      10.6.0.1/24
                    </v-chip>
                    <v-chip size="small" color="grey" variant="outlined" @click="clearNginxIPRanges">
                      <v-icon start size="small">mdi-close</v-icon>
                      Clear All
                    </v-chip>
                  </v-chip-group>
                </div>
              </v-col>
            </v-row>
          </v-form>
        </v-card-text>

        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn @click="closeNginxIPModal" color="grey"> Cancel </v-btn>
          <v-btn @click="saveNginxIPRestrictions" :loading="savingNginxIP" color="primary">
            Save & Apply
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>

    <!-- Delete Confirmation Dialogs -->
    <ConfirmationDialog v-model:show="showDeleteConfigDialog" title="Delete DNS Configuration"
      message="Are you sure you want to delete this DNS configuration? This will also delete all associated records."
      icon="mdi-delete-alert" icon-color="error" confirm-text="Delete" confirm-color="error"
      @confirm="confirmDeleteConfig" />

    <ConfirmationDialog v-model:show="showDeleteRecordDialog" title="Delete DNS Record"
      message="Are you sure you want to delete this DNS record?" icon="mdi-delete-alert" icon-color="error"
      confirm-text="Delete" confirm-color="error" @confirm="confirmDeleteRecord" />
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from 'vue';
import AppLayout from '../components/AppLayout.vue';
import ConfirmationDialog from '../components/ConfirmationDialog.vue';
import ErrorAlert from '../components/ErrorAlert.vue';
import PageHeader from '../components/PageHeader.vue';
import StatsCards from '../components/StatsCards.vue';
import apiService from '../services/api';
import type {
  DNSConfig,
  DNSConfigCreateRequest,
  DNSConfigUpdateRequest,
  DNSRecord,
  DNSRecordCreateRequest,
  DNSRecordUpdateRequest,
  JobInfo
} from '../types/api';

// Reactive data
const dnsConfigs = ref<DNSConfig[]>([]);
const configRecords = ref<Record<number, DNSRecord[]>>({});
const publicIP = ref<string>('');
const loadingConfigs = ref(false);
const loadingPublicIP = ref(false);
const loadingUpdates = ref<Record<number, boolean>>({});
const loadingRegen = ref<Record<number, boolean>>({});
const savingConfig = ref(false);
const savingRecord = ref(false);
const savingNginxIP = ref(false);
const nginxAllowedRanges = ref<string[]>([]);
const error = ref<string | null>(null);

// Scheduled jobs state
const scheduledJobs = ref<Record<number, JobInfo>>({});
const loadingJobs = ref(false);
const stoppingJobs = ref<Record<number, boolean>>({});
const countdownTimer = ref<number | null>(null);

// Computed property for scheduled jobs with display names and countdown
const scheduledJobsWithNames = computed(() => {
  const jobs: Record<number, { interval: number; displayName: string; nextUpdate: string; countdown: string; isPaused: boolean }> = {};
  for (const [recordId, jobInfo] of Object.entries(scheduledJobs.value)) {
    const id = parseInt(recordId);
    const nextUpdate = jobInfo.next_update;
    jobs[id] = {
      interval: jobInfo.interval,
      displayName: getRecordDisplayName(id),
      nextUpdate: nextUpdate,
      countdown: getCountdown(nextUpdate),
      isPaused: jobInfo.is_paused
    };
  }
  return jobs;
});

// Modal states
const showCreateConfigModal = ref(false);
const showEditConfigModal = ref(false);
const showCreateRecordModal = ref(false);
const showEditRecordModal = ref(false);
const showDeleteConfigDialog = ref(false);
const showDeleteRecordDialog = ref(false);
const showNginxIPModal = ref(false);
const deletingConfigId = ref<number | null>(null);
const deletingRecordId = ref<number | null>(null);

// Form data
const configForm = ref<DNSConfigCreateRequest & { is_active: boolean }>({
  provider: 'namecheap',
  domain: '',
  username: '',
  password: '',
  is_active: true,
});

// Password change state
const changePassword = ref(false);

const recordForm = ref<DNSRecordCreateRequest & { is_active: boolean }>({
  config_id: 0,
  host: '',
  allowed_ip_ranges: '',
  dynamic_dns_refresh_rate: undefined,
  include_backend: false,
  backend_url: '',
  is_active: true,
});

const nginxIPForm = ref({
  allowedRanges: '',
});

const editingConfig = ref<DNSConfig | null>(null);
const editingRecord = ref<DNSRecord | null>(null);

// Computed
const showConfigDialog = computed({
  get: () => showCreateConfigModal.value || showEditConfigModal.value,
  set: value => {
    if (!value) {
      showCreateConfigModal.value = false;
      showEditConfigModal.value = false;
    }
  },
});

const showRecordDialog = computed({
  get: () => showCreateRecordModal.value || showEditRecordModal.value,
  set: value => {
    if (!value) {
      showCreateRecordModal.value = false;
      showEditRecordModal.value = false;
    }
  },
});

const dnsStats = computed(() => [
  {
    key: 'configs',
    value: dnsConfigs.value?.length || 0,
    label: 'Configurations',
    icon: 'mdi-dns',
    color: 'blue-lighten-5',
    iconColor: 'blue',
  },
  {
    key: 'active',
    value: dnsConfigs.value?.filter(c => c.is_active).length || 0,
    label: 'Active',
    icon: 'mdi-check-circle',
    color: 'green-lighten-5',
    iconColor: 'green',
  },
  {
    key: 'records',
    value: Object.values(configRecords.value).flat().length,
    label: 'Total Records',
    icon: 'mdi-list',
    color: 'orange-lighten-5',
    iconColor: 'orange',
  },
  {
    key: 'public-ip',
    value: publicIP.value ? 'Set' : 'Unknown',
    label: 'Public IP',
    icon: 'mdi-earth',
    color: 'purple-lighten-5',
    iconColor: 'purple',
  },
]);

// Methods
const loadDNSConfigs = async () => {
  try {
    loadingConfigs.value = true;
    const response = await apiService.getDNSConfigs();
    dnsConfigs.value = response.configs || [];

    // Load records for each config
    for (const config of dnsConfigs.value) {
      await loadDNSRecords(config.id);
    }

    // Load scheduled jobs after records are loaded
    await loadScheduledJobs();
  } catch (err) {
    error.value = `Failed to load DNS configurations: ${err}`;
    dnsConfigs.value = [];
  } finally {
    loadingConfigs.value = false;
  }
};

const loadDNSRecords = async (configId: number) => {
  try {
    const response = await apiService.getDNSRecords(configId);
    configRecords.value[configId] = response.records || [];
  } catch (err) {
    console.error(`Failed to load DNS records for config ${configId}:`, err);
    configRecords.value[configId] = [];
  }
};

const refreshPublicIP = async () => {
  try {
    loadingPublicIP.value = true;
    const response = await apiService.getPublicIP();
    publicIP.value = response.ip;
  } catch (err) {
    error.value = `Failed to get public IP: ${err}`;
  } finally {
    loadingPublicIP.value = false;
  }
};

const editConfig = (config: DNSConfig) => {
  editingConfig.value = config;
  configForm.value = {
    provider: config.provider,
    domain: config.domain,
    username: config.username,
    password: '', // Don't populate password for security
    is_active: config.is_active,
  };
  changePassword.value = false; // Reset password change state
  showEditConfigModal.value = true;
};

const deleteConfig = async (id: number) => {
  deletingConfigId.value = id;
  showDeleteConfigDialog.value = true;
};

const confirmDeleteConfig = async () => {
  if (!deletingConfigId.value) return;

  try {
    await apiService.deleteDNSConfig(deletingConfigId.value);
    await loadDNSConfigs();
    showDeleteConfigDialog.value = false;
    deletingConfigId.value = null;
  } catch (err) {
    error.value = `Failed to delete DNS configuration: ${err}`;
  }
};

const saveConfig = async () => {
  try {
    savingConfig.value = true;

    if (showCreateConfigModal.value) {
      // Prepare data based on provider type
      const createData: any = {
        provider: configForm.value.provider,
        domain: configForm.value.domain,
        is_active: configForm.value.is_active,
      };

      // Only include credentials for dynamic DNS providers
      if (configForm.value.provider === 'namecheap') {
        createData.username = configForm.value.username;
        createData.password = configForm.value.password;
      }

      await apiService.createDNSConfig(createData);
    } else if (editingConfig.value) {
      const updateData: DNSConfigUpdateRequest = {
        provider: configForm.value.provider,
        domain: configForm.value.domain,
        is_active: configForm.value.is_active,
      };

      // Only include credentials for dynamic DNS providers
      if (configForm.value.provider === 'namecheap') {
        updateData.username = configForm.value.username;

        // Only include password if it's being changed
        if (changePassword.value && configForm.value.password) {
          updateData.password = configForm.value.password;
        }
      }

      await apiService.updateDNSConfig(editingConfig.value.id, updateData);
    }

    await loadDNSConfigs();
    closeConfigModal();
  } catch (err) {
    error.value = `Failed to save DNS configuration: ${err}`;
  } finally {
    savingConfig.value = false;
  }
};

const openCreateRecordModal = (configId: number) => {
  recordForm.value = {
    config_id: configId,
    host: '',
    allowed_ip_ranges: '',
    dynamic_dns_refresh_rate: undefined,
    include_backend: false,
    backend_url: '',
    is_active: true,
  };
  showCreateRecordModal.value = true;
};

const editRecord = (record: DNSRecord) => {
  editingRecord.value = record;
  recordForm.value = {
    config_id: record.config_id,
    host: record.host,
    allowed_ip_ranges: record.allowed_ip_ranges || '',
    dynamic_dns_refresh_rate: record.dynamic_dns_refresh_rate,
    include_backend: record.include_backend,
    backend_url: record.backend_url || '',
    is_active: record.is_active,
  };
  showEditRecordModal.value = true;
};

const deleteRecord = async (id: number) => {
  deletingRecordId.value = id;
  showDeleteRecordDialog.value = true;
};

const confirmDeleteRecord = async () => {
  if (!deletingRecordId.value) return;

  try {
    await apiService.deleteDNSRecord(deletingRecordId.value);
    await loadDNSConfigs();
    await loadScheduledJobs();
    showDeleteRecordDialog.value = false;
    deletingRecordId.value = null;
  } catch (err) {
    error.value = `Failed to delete DNS record: ${err}`;
  }
};

const saveRecord = async () => {
  try {
    savingRecord.value = true;

    // Validate IP ranges
    const ipValidationError = validateIPRanges(recordForm.value.allowed_ip_ranges || '');
    if (ipValidationError) {
      error.value = `IP range validation error: ${ipValidationError}`;
      return;
    }

    // Validate refresh rate
    const refreshRateValidationError = validateRefreshRate(recordForm.value.dynamic_dns_refresh_rate);
    if (refreshRateValidationError) {
      error.value = `Refresh rate validation error: ${refreshRateValidationError}`;
      return;
    }

    if (showCreateRecordModal.value) {
      await apiService.createDNSRecord(recordForm.value);
    } else if (editingRecord.value) {
      const updateData: DNSRecordUpdateRequest = {
        host: recordForm.value.host,
        allowed_ip_ranges: recordForm.value.allowed_ip_ranges,
        dynamic_dns_refresh_rate: recordForm.value.dynamic_dns_refresh_rate,
        include_backend: recordForm.value.include_backend,
        backend_url: recordForm.value.backend_url,
        is_active: recordForm.value.is_active,
      };
      await apiService.updateDNSRecord(editingRecord.value.id, updateData);
    }

    await loadDNSConfigs();
    await loadScheduledJobs();
    closeRecordModal();
  } catch (err) {
    error.value = `Failed to save DNS record: ${err}`;
  } finally {
    savingRecord.value = false;
  }
};

const updateRecordNow = async (recordId: number) => {
  try {
    loadingUpdates.value[recordId] = true;
    const response = await apiService.updateDNSRecordNow(recordId);

    if (response.response.success) {
      await loadDNSConfigs();
    } else {
      error.value = `DNS update failed: ${response.response.message}`;
    }
  } catch (err) {
    error.value = `Failed to update DNS record: ${err}`;
  } finally {
    loadingUpdates.value[recordId] = false;
  }
};

const regenerateConfig = async (record: DNSRecord) => {
  let domain = '';
  try {
    loadingRegen.value[record.id] = true;

    // Get the domain name for this record
    const config = dnsConfigs.value.find(c => c.id === record.config_id);
    if (!config) {
      error.value = 'DNS configuration not found for this record';
      return;
    }

    domain = record.host === '@' ? config.domain : `${record.host}.${config.domain}`;

    // Call the regenerate config API
    await apiService.regenerateProxyConfig(domain);

    // Show success message
    error.value = null; // Clear any existing errors
    // You could add a success notification here if you have a notification system

  } catch (err) {
    error.value = `Failed to regenerate nginx config for ${domain}: ${err}`;
  } finally {
    loadingRegen.value[record.id] = false;
  }
};


const closeConfigModal = () => {
  showCreateConfigModal.value = false;
  showEditConfigModal.value = false;
  editingConfig.value = null;
  changePassword.value = false;
  configForm.value = {
    provider: 'namecheap',
    domain: '',
    username: '',
    password: '',
    is_active: true,
  };
};

const closeRecordModal = () => {
  showCreateRecordModal.value = false;
  showEditRecordModal.value = false;
  editingRecord.value = null;
  recordForm.value = {
    config_id: 0,
    host: '',
    allowed_ip_ranges: '',
    dynamic_dns_refresh_rate: undefined,
    include_backend: false,
    backend_url: '',
    is_active: true,
  };
};

const formatDate = (dateString?: string) => {
  if (!dateString) return null;
  return new Date(dateString).toLocaleString();
};

// IP range validation
const validateIPRanges = (ipRanges: string): string | null => {
  if (!ipRanges.trim()) {
    return null; // Empty is valid
  }

  const ranges = ipRanges.split(',').map(r => r.trim()).filter(r => r);

  for (const range of ranges) {
    // Check if it's a CIDR notation
    if (range.includes('/')) {
      const parts = range.split('/');
      if (parts.length !== 2) {
        return `Invalid CIDR format: ${range}`;
      }

      const ip = parts[0];
      const cidr = parseInt(parts[1]);

      if (isNaN(cidr) || cidr < 0 || cidr > 32) {
        return `Invalid CIDR prefix: ${range}`;
      }

      if (!isValidIP(ip)) {
        return `Invalid IP address in CIDR: ${range}`;
      }
    } else {
      // Single IP address
      if (!isValidIP(range)) {
        return `Invalid IP address: ${range}`;
      }
    }
  }

  return null;
};

const isValidIP = (ip: string): boolean => {
  const ipv4Regex = /^(?:(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)$/;
  const ipv6Regex = /^(?:[0-9a-fA-F]{1,4}:){7}[0-9a-fA-F]{1,4}$/;
  return ipv4Regex.test(ip) || ipv6Regex.test(ip);
};

// Refresh rate validation
const validateRefreshRate = (refreshRate: number | null | undefined): string | null => {
  if (refreshRate === null || refreshRate === undefined) {
    return null; // Empty is valid (no auto-refresh)
  }

  if (isNaN(refreshRate) || refreshRate < 1) {
    return 'Refresh rate must be at least 1 minute';
  }

  if (refreshRate > 1440) {
    return 'Refresh rate cannot exceed 1440 minutes (24 hours)';
  }

  return null;
};

// Scheduled jobs methods
const loadScheduledJobs = async () => {
  try {
    loadingJobs.value = true;
    const response = await apiService.getScheduledJobs();
    scheduledJobs.value = response.active_jobs || {};
    startCountdownTimer();
  } catch (err: any) {
    error.value = err.response?.data?.error || 'Failed to load scheduled jobs';
  } finally {
    loadingJobs.value = false;
  }
};

const startCountdownTimer = () => {
  // Clear existing timer
  if (countdownTimer.value) {
    clearInterval(countdownTimer.value);
  }

  // Start new timer that updates every second
  countdownTimer.value = setInterval(() => {
    // Check if any jobs are due and refresh if needed
    let needsRefresh = false;
    for (const [, jobInfo] of Object.entries(scheduledJobs.value)) {
      const now = new Date().getTime();
      const next = new Date(jobInfo.next_update).getTime();
      if (next <= now) {
        needsRefresh = true;
        break;
      }
    }

    if (needsRefresh) {
      // Refresh scheduled jobs to get updated times
      loadScheduledJobs();
    } else {
      // Force reactivity update by creating a new object
      scheduledJobs.value = { ...scheduledJobs.value };
    }
  }, 1000);
};

const stopCountdownTimer = () => {
  if (countdownTimer.value) {
    clearInterval(countdownTimer.value);
    countdownTimer.value = null;
  }
};

const refreshAllData = async () => {
  await loadDNSConfigs();
};

const pauseScheduledJob = async (recordId: number) => {
  try {
    stoppingJobs.value[recordId] = true;
    await apiService.pauseScheduledJob(recordId);
    // Refresh the scheduled jobs to get updated state
    await loadScheduledJobs();
  } catch (err: any) {
    error.value = err.response?.data?.error || 'Failed to pause scheduled job';
  } finally {
    stoppingJobs.value[recordId] = false;
  }
};

const resumeScheduledJob = async (recordId: number) => {
  try {
    stoppingJobs.value[recordId] = true;
    await apiService.resumeScheduledJob(recordId);
    // Refresh the scheduled jobs to get updated state
    await loadScheduledJobs();
  } catch (err: any) {
    error.value = err.response?.data?.error || 'Failed to resume scheduled job';
  } finally {
    stoppingJobs.value[recordId] = false;
  }
};

const getRecordDisplayName = (recordId: number): string => {
  // Find the record in all configs
  for (const config of dnsConfigs.value) {
    const records = configRecords.value[config.id] || [];
    const record = records.find(r => r.id === recordId);
    if (record) {
      const hostname = record.host === '@' ? config.domain : `${record.host}.${config.domain}`;
      const status = record.is_active ? '' : ' (inactive)';
      return `${hostname}${status}`;
    }
  }
  // If record not found, try to refresh and return fallback
  console.log(`Record ${recordId} not found in loaded records. Available records:`, Object.keys(configRecords.value));
  return `Record ID: ${recordId}`;
};

const getCountdown = (nextUpdate: string): string => {
  if (!nextUpdate) {
    return 'Unknown';
  }

  const now = new Date().getTime();
  const next = new Date(nextUpdate).getTime();

  // Check if the date is valid
  if (isNaN(next)) {
    console.error('Invalid nextUpdate date:', nextUpdate);
    return 'Invalid date';
  }

  const diff = next - now;

  // If due or overdue, show a very short countdown or refresh
  if (diff <= 0) {
    return 'Due now';
  }

  const minutes = Math.floor(diff / (1000 * 60));
  const seconds = Math.floor((diff % (1000 * 60)) / 1000);

  if (minutes < 1) {
    return `${seconds}s`;
  } else if (minutes < 60) {
    return `${minutes}m ${seconds}s`;
  } else {
    const hours = Math.floor(minutes / 60);
    const remainingMinutes = minutes % 60;
    return `${hours}h ${remainingMinutes}m`;
  }
};

// Quick fill functions
const quickFillIPRanges = (range: string) => {
  const currentRanges = recordForm.value.allowed_ip_ranges || '';

  if (currentRanges.trim()) {
    // If there are existing ranges, add the new one with a comma
    recordForm.value.allowed_ip_ranges = `${currentRanges}, ${range}`;
  } else {
    // If empty, just set the range
    recordForm.value.allowed_ip_ranges = range;
  }
};

const clearIPRanges = () => {
  recordForm.value.allowed_ip_ranges = '';
};

// Nginx IP management methods
const saveNginxIPRestrictions = async () => {
  try {
    savingNginxIP.value = true;

    // Validate IP ranges
    const ipValidationError = validateIPRanges(nginxIPForm.value.allowedRanges || '');
    if (ipValidationError) {
      error.value = `IP range validation error: ${ipValidationError}`;
      return;
    }

    // Parse ranges
    const ranges = nginxIPForm.value.allowedRanges
      .split(',')
      .map(r => r.trim())
      .filter(r => r);

    // Update nginx configuration
    await apiService.updateAdminIPRestrictions(ranges);

    // Update local state
    nginxAllowedRanges.value = ranges;

    closeNginxIPModal();
  } catch (err) {
    error.value = `Failed to update nginx IP restrictions: ${err}`;
  } finally {
    savingNginxIP.value = false;
  }
};

const closeNginxIPModal = () => {
  showNginxIPModal.value = false;
  nginxIPForm.value = {
    allowedRanges: nginxAllowedRanges.value.join(', '),
  };
};

const quickFillNginxIPRanges = (range: string) => {
  const currentRanges = nginxIPForm.value.allowedRanges || '';

  if (currentRanges.trim()) {
    nginxIPForm.value.allowedRanges = `${currentRanges}, ${range}`;
  } else {
    nginxIPForm.value.allowedRanges = range;
  }
};

const clearNginxIPRanges = () => {
  nginxIPForm.value.allowedRanges = '';
};

// Lifecycle
onMounted(() => {
  loadDNSConfigs();
  refreshPublicIP();
  loadNginxIPRestrictions();
});

onUnmounted(() => {
  stopCountdownTimer();
});

const loadNginxIPRestrictions = async () => {
  try {
    const response = await apiService.getAdminIPRestrictions();
    nginxAllowedRanges.value = response.allowed_ranges || [];
    nginxIPForm.value.allowedRanges = nginxAllowedRanges.value.join(', ');
  } catch (err) {
    console.error('Failed to load nginx IP restrictions:', err);
    // Fallback to default ranges
    nginxAllowedRanges.value = ['192.168.50.2/24', '10.6.0.1/24'];
    nginxIPForm.value.allowedRanges = nginxAllowedRanges.value.join(', ');
  }
};
</script>

<style scoped>
.font-mono {
  font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
}
</style>
