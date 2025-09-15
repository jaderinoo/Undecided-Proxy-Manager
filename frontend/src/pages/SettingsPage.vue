<template>
  <AppLayout>
    <v-container>
      <PageHeader
        :showChip="false"
        :count="0"
        item-name="Settings"
        subtitle="Manage application settings and configuration"
      />

      <!-- Success/Error Messages -->
      <v-alert
        v-if="successMessage"
        type="success"
        closable
        @click:close="successMessage = null"
        class="mb-4"
      >
        {{ successMessage }}
      </v-alert>

      <v-alert
        v-if="errorMessage"
        type="error"
        closable
        @click:close="errorMessage = null"
        class="mb-4"
      >
        {{ errorMessage }}
      </v-alert>

      <v-row>
        <!-- Core Settings (Read-only) -->
        <v-col cols="12" md="6">
          <v-card>
            <v-card-title class="d-flex align-center">
              <v-icon class="mr-2">mdi-cog</v-icon>
              Core Settings
              <v-spacer />
              <v-chip color="info" size="small">Read-only</v-chip>
            </v-card-title>
            <v-card-text>
              <p class="text-caption mb-4">
                These settings are managed through environment variables and
                cannot be changed via the UI.
              </p>

              <v-form>
                <v-text-field
                  v-model="coreSettings.database_path"
                  label="Database Path"
                  readonly
                  variant="outlined"
                  density="compact"
                  class="mb-2"
                />

                <v-text-field
                  v-model="coreSettings.environment"
                  label="Environment"
                  readonly
                  variant="outlined"
                  density="compact"
                  class="mb-2"
                />

                <v-text-field
                  v-model="coreSettings.backend_port"
                  label="Backend Port"
                  readonly
                  variant="outlined"
                  density="compact"
                  class="mb-2"
                />


                <v-text-field
                  v-model="coreSettings.letsencrypt_email"
                  label="Let's Encrypt Email"
                  readonly
                  variant="outlined"
                  density="compact"
                  class="mb-2"
                />

                <v-text-field
                  v-model="coreSettings.letsencrypt_webroot"
                  label="Let's Encrypt Webroot"
                  readonly
                  variant="outlined"
                  density="compact"
                  class="mb-2"
                />

                <v-text-field
                  v-model="coreSettings.letsencrypt_cert_path"
                  label="Let's Encrypt Cert Path"
                  readonly
                  variant="outlined"
                  density="compact"
                  class="mb-2"
                />


                <v-text-field
                  v-model="coreSettings.public_ip_service"
                  label="Public IP Service"
                  readonly
                  variant="outlined"
                  density="compact"
                />
              </v-form>
            </v-card-text>
          </v-card>
        </v-col>

        <!-- UI Settings (Editable) -->
        <v-col cols="12" md="6">
          <v-card>
            <v-card-title class="d-flex align-center">
              <v-icon class="mr-2">mdi-tune</v-icon>
              UI Settings
              <v-spacer />
              <v-chip color="success" size="small">Editable</v-chip>
            </v-card-title>
            <v-card-text>
              <p class="text-caption mb-4">
                These settings can be managed through the UI and are stored in
                the database.
              </p>

              <v-form ref="settingsForm" v-model="formValid">
                <v-text-field
                  v-model="uiSettings.display_name"
                  label="Display Name"
                  hint="Name shown in the UI"
                  variant="outlined"
                  density="compact"
                  class="mb-2"
                  :rules="[rules.required]"
                />

                <v-select
                  v-model="uiSettings.theme"
                  label="Theme"
                  :items="themeOptions"
                  variant="outlined"
                  density="compact"
                  class="mb-2"
                />

                <v-select
                  v-model="uiSettings.language"
                  label="Language"
                  :items="languageOptions"
                  variant="outlined"
                  density="compact"
                  class="mb-2"
                />
              </v-form>

              <v-card-actions class="px-0">
                <v-btn
                  color="primary"
                  :loading="saving"
                  :disabled="!formValid"
                  @click="saveSettings"
                >
                  <v-icon class="mr-2">mdi-content-save</v-icon>
                  Save Settings
                </v-btn>

                <v-btn
                  color="secondary"
                  variant="outlined"
                  :disabled="saving"
                  @click="resetSettings"
                >
                  <v-icon class="mr-2">mdi-refresh</v-icon>
                  Reset
                </v-btn>
              </v-card-actions>
            </v-card-text>
          </v-card>
        </v-col>
      </v-row>

      <!-- Environment Variables Info -->
      <v-row class="mt-4">
        <v-col cols="12">
          <v-card color="info" variant="outlined">
            <v-card-title class="d-flex align-center">
              <v-icon class="mr-2">mdi-information</v-icon>
              Environment Variables
              <v-spacer />
              <v-btn
                color="primary"
                variant="outlined"
                size="small"
                @click="copyEnvTemplate"
                :loading="copying"
              >
                <v-icon class="mr-2">mdi-content-copy</v-icon>
                Copy Template
              </v-btn>
            </v-card-title>
            <v-card-text>
              <p class="mb-3">
                To modify core settings, update the following environment
                variables in your <code>.env</code> file:
              </p>

              <v-expansion-panels variant="accordion" class="mb-3">
                <v-expansion-panel>
                  <v-expansion-panel-title>
                    <v-icon class="mr-2">mdi-file-code</v-icon>
                    View .env Template
                  </v-expansion-panel-title>
                  <v-expansion-panel-text>
                    <div class="env-template">
                      <div data-comment="true">
                        # UPM (Undecided Proxy Manager) Environment
                        Configuration
                      </div>
                      <div data-comment="true">
                        # This file is automatically loaded by Docker Compose
                      </div>
                      <div></div>
                      <div data-comment="true"># Database Configuration</div>
                      <div>DB_PATH=/data/upm-dev.db</div>
                      <div></div>
                      <div data-comment="true"># Server Configuration</div>
                      <div>GO_ENV=development</div>
                      <div>BACKEND_PORT=6081</div>
                      <div></div>
                      <div data-comment="true"># Production Port Configuration</div>
                      <div>PROD_NGINX_HTTP_PORT=80</div>
                      <div>PROD_NGINX_HTTPS_PORT=443</div>
                      <div>PROD_BACKEND_PORT=6080</div>
                      <div>PROD_FRONTEND_PORT=6070</div>
                      <div></div>
                      <div data-comment="true">
                        # Let's Encrypt Configuration (Core Settings - Read-only
                        in UI)
                      </div>
                      <div>LETSENCRYPT_EMAIL=jaderinoo@gmail.com</div>
                      <div>LETSENCRYPT_WEBROOT=/var/www/html</div>
                      <div>LETSENCRYPT_CERT_PATH=/etc/letsencrypt</div>
                      <div></div>
                      <div data-comment="true">
                        # DNS Configuration (UI-manageable settings)
                      </div>
                      <div>PUBLIC_IP_SERVICE=https://api.ipify.org</div>
                      <div></div>
                      <div data-comment="true"># Encryption Configuration</div>
                      <div data-comment="true"># Generate a secure 32-byte key for production: openssl rand -base64 32</div>
                      <div>ENCRYPTION_KEY=upm-default-encryption-key-32byt</div>
                      <div></div>
                      <div data-comment="true"># Nginx Configuration</div>
                      <div>NGINX_CONFIG_PATH=/etc/nginx/sites-available</div>
                      <div>NGINX_RELOAD_CMD=docker exec undecided-proxy-manager-nginx-1 nginx -s reload</div>
                      <div>NGINX_CONTAINER_NAME=undecided-proxy-manager-nginx-1</div>
                      <div></div>
                      <div data-comment="true"># Development Nginx Configuration (for docker-compose.dev.yml)</div>
                      <div>DEV_NGINX_RELOAD_CMD=docker exec undecided-proxy-manager-dev-nginx-1 nginx -s reload</div>
                      <div>DEV_NGINX_CONTAINER_NAME=undecided-proxy-manager-dev-nginx-1</div>
                    </div>
                  </v-expansion-panel-text>
                </v-expansion-panel>
              </v-expansion-panels>
            </v-card-text>
          </v-card>
        </v-col>
      </v-row>
    </v-container>
  </AppLayout>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue';
import AppLayout from '../components/layout/AppLayout.vue';
import PageHeader from '../components/ui/PageHeader.vue';
import apiService from '../services/api';
import type {
    CoreSettings,
    SettingsUpdateRequest,
    UISettings,
} from '../types/api';

// Reactive data
const coreSettings = ref<CoreSettings>({
  database_path: '',
  environment: '',
  backend_port: '',
  letsencrypt_email: '',
  letsencrypt_webroot: '',
  letsencrypt_cert_path: '',
  public_ip_service: '',
});

const uiSettings = ref<UISettings>({
  display_name: '',
  theme: '',
  language: '',
});

const originalUISettings = ref<UISettings>({
  display_name: '',
  theme: '',
  language: '',
});

const loading = ref(false);
const saving = ref(false);
const copying = ref(false);
const formValid = ref(false);
const successMessage = ref<string | null>(null);
const errorMessage = ref<string | null>(null);

// Options
const themeOptions = [
  { title: 'Auto', value: 'auto' },
  { title: 'Light', value: 'light' },
  { title: 'Dark', value: 'dark' },
];

const languageOptions = [
  { title: 'English', value: 'en' },
  { title: 'Spanish', value: 'es' },
  { title: 'French', value: 'fr' },
  { title: 'German', value: 'de' },
];

// Validation rules
const rules = {
  required: (value: string) => !!value || 'This field is required',
  url: (value: string) => {
    try {
      new URL(value);
      return true;
    } catch {
      return 'Please enter a valid URL';
    }
  },
  interval: (value: string) => {
    const intervalRegex = /^\d+[smhd]$/;
    return (
      intervalRegex.test(value) ||
      'Please enter a valid interval (e.g., 5m, 1h, 30s)'
    );
  },
};

// Methods
const loadSettings = async () => {
  try {
    loading.value = true;
    errorMessage.value = null;
    const settings = await apiService.getSettings();

    coreSettings.value = settings.core_settings;
    uiSettings.value = settings.ui_settings;
    originalUISettings.value = { ...settings.ui_settings };
  } catch (error) {
    console.error('Failed to load settings:', error);
    errorMessage.value = 'Failed to load settings';
  } finally {
    loading.value = false;
  }
};

const saveSettings = async () => {
  try {
    saving.value = true;
    errorMessage.value = null;
    successMessage.value = null;

    const updateData: SettingsUpdateRequest = {
      display_name: uiSettings.value.display_name,
      theme: uiSettings.value.theme,
      language: uiSettings.value.language,
    };

    const updatedSettings = await apiService.updateSettings(updateData);

    // Update local state with response
    coreSettings.value = updatedSettings.core_settings;
    uiSettings.value = updatedSettings.ui_settings;
    originalUISettings.value = { ...updatedSettings.ui_settings };

    successMessage.value = 'Settings saved successfully';
  } catch (error) {
    console.error('Failed to save settings:', error);
    errorMessage.value = 'Failed to save settings';
  } finally {
    saving.value = false;
  }
};

const resetSettings = () => {
  uiSettings.value = { ...originalUISettings.value };
  errorMessage.value = null;
  successMessage.value = null;
};

const copyEnvTemplate = async () => {
  try {
    copying.value = true;
    const envTemplate = `# UPM (Undecided Proxy Manager) Environment Configuration
# This file is automatically loaded by Docker Compose

# Database Configuration
DB_PATH=/data/upm-dev.db

# Server Configuration
GO_ENV=development
BACKEND_PORT=6081

# Production Port Configuration
PROD_NGINX_HTTP_PORT=80
PROD_NGINX_HTTPS_PORT=443
PROD_BACKEND_PORT=6080
PROD_FRONTEND_PORT=6070


# Let's Encrypt Configuration (Core Settings - Read-only in UI)
LETSENCRYPT_EMAIL=jaderinoo@gmail.com
LETSENCRYPT_WEBROOT=/var/www/html
LETSENCRYPT_CERT_PATH=/etc/letsencrypt

# DNS Configuration (UI-manageable settings)
PUBLIC_IP_SERVICE=https://api.ipify.org

# Encryption Configuration
# Generate a secure 32-byte key for production: openssl rand -base64 32
ENCRYPTION_KEY=upm-default-encryption-key-32byt

# Nginx Configuration
NGINX_CONFIG_PATH=/etc/nginx/sites-available
NGINX_RELOAD_CMD=docker exec undecided-proxy-manager-nginx-1 nginx -s reload
NGINX_CONTAINER_NAME=undecided-proxy-manager-nginx-1

# Development Nginx Configuration (for docker-compose.dev.yml)
DEV_NGINX_RELOAD_CMD=docker exec undecided-proxy-manager-dev-nginx-1 nginx -s reload
DEV_NGINX_CONTAINER_NAME=undecided-proxy-manager-dev-nginx-1`;

    await navigator.clipboard.writeText(envTemplate);
    successMessage.value = 'Environment template copied to clipboard!';
    setTimeout(() => {
      successMessage.value = null;
    }, 3000);
  } catch (error) {
    console.error('Failed to copy to clipboard:', error);
    errorMessage.value = 'Failed to copy to clipboard';
  } finally {
    copying.value = false;
  }
};

// Lifecycle
onMounted(() => {
  loadSettings();
});
</script>

<style scoped>
.v-code {
  background-color: rgba(0, 0, 0, 0.05);
  padding: 16px;
  border-radius: 6px;
  font-family: 'Courier New', monospace;
  white-space: pre-line;
  max-height: 500px;
  overflow-y: auto;
  border: 1px solid rgba(0, 0, 0, 0.1);
}

.v-code div {
  margin-bottom: 3px;
  line-height: 1.5;
  padding-left: 8px;
  text-indent: -8px;
}

.v-code div:last-child {
  margin-bottom: 0;
}

.env-template {
  font-size: 0.9rem;
  line-height: 1.4;
  background-color: #f8f9fa;
  border: 1px solid #e9ecef;
}

.env-template div {
  padding-left: 12px;
  text-indent: -12px;
  margin-bottom: 4px;
}

/* Add some visual separation for comment lines */
.env-template div[data-comment='true'] {
  color: #6c757d;
  font-style: italic;
}

/* Style empty lines */
.env-template div:empty {
  height: 8px;
  margin-bottom: 8px;
}
</style>
