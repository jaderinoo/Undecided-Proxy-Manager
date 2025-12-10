<template>
  <v-dialog :model-value="show" @update:model-value="$emit('update:show', $event)" max-width="600px" persistent>
    <v-card>
      <v-card-title>
        <v-icon left>mdi-server-plus</v-icon>
        {{ editingProxy ? 'Edit Proxy' : 'Create New Proxy' }}
      </v-card-title>

      <v-card-text>
        <v-form ref="formRef" v-model="formValid">
          <v-row>
            <v-col cols="12">
              <v-text-field
                v-model="form.name"
                label="Proxy Name"
                variant="outlined"
                density="compact"
                :rules="[v => !!v || 'Name is required']"
                required
              />
            </v-col>

            <v-col cols="12">
              <v-text-field
                v-model="form.domain"
                label="Domain"
                variant="outlined"
                density="compact"
                placeholder="example.com"
                :rules="[v => !!v || 'Domain is required']"
                required
              />
            </v-col>

            <v-col cols="12">
              <v-text-field
                v-model="form.target_url"
                label="Target URL"
                variant="outlined"
                density="compact"
                placeholder="http://localhost:3000"
                :rules="[v => !!v || 'Target URL is required']"
                required
              />
            </v-col>

            <v-col cols="12">
              <div class="d-flex align-center mb-2">
                <v-switch
                  v-model="form.ssl_enabled"
                  :label="sslSwitchLabel"
                  color="primary"
                  hide-details
                  class="mr-2"
                />
                <v-chip
                  v-if="existingCertificate"
                  color="success"
                  size="small"
                  variant="flat"
                  prepend-icon="mdi-certificate"
                >
                  Certificate Exists
                </v-chip>
              </div>
              <v-alert
                v-if="form.ssl_enabled"
                type="info"
                variant="tonal"
                density="compact"
                class="mt-2"
              >
                <template v-slot:prepend>
                  <v-icon>mdi-information</v-icon>
                </template>
                <div class="text-caption">
                  <span v-if="existingCertificate">
                    An existing certificate will be used for this domain.
                  </span>
                  <span v-else>
                    A free Let's Encrypt certificate will be generated for this domain.
                  </span>
                </div>
              </v-alert>
            </v-col>
          </v-row>
        </v-form>
      </v-card-text>

      <v-card-actions>
        <v-spacer></v-spacer>
        <v-btn
          color="grey"
          variant="text"
          @click="cancel"
          :disabled="saving"
        >
          Cancel
        </v-btn>
        <v-btn
          color="primary"
          @click="save"
          :loading="saving"
          :disabled="!formValid"
        >
          {{ editingProxy ? 'Update' : 'Create' }}
        </v-btn>
      </v-card-actions>
    </v-card>
  </v-dialog>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue';
import type { Certificate, Proxy, ProxyCreateRequest, ProxyUpdateRequest } from '../../types/api';

interface Props {
  show: boolean;
  editingProxy?: Proxy | null;
  initialData?: Partial<ProxyCreateRequest>;
  certificates?: Certificate[];
}

const props = defineProps<Props>();

const emit = defineEmits<{
  'update:show': [value: boolean];
  'save': [data: ProxyCreateRequest | ProxyUpdateRequest, isEdit: boolean];
  'cancel': [];
}>();

const formRef = ref();
const formValid = ref(false);
const saving = ref(false);

const form = ref<ProxyCreateRequest & { id?: number }>({
  name: '',
  domain: '',
  target_url: '',
  ssl_enabled: false,
});

// Check if a certificate exists for the current domain
const existingCertificate = computed(() => {
  if (!form.value.domain || !props.certificates) {
    return null;
  }
  return props.certificates.find(cert => cert.domain === form.value.domain);
});

// Update SSL switch label based on whether cert exists
const sslSwitchLabel = computed(() => {
  if (existingCertificate.value) {
    return 'Enable SSL (use existing certificate)';
  }
  return 'Enable SSL (creates Let\'s Encrypt cert)';
});

// Watch for changes in props to update form
watch(() => props.show, (newValue) => {
  if (newValue) {
    resetForm();
    // If we have initial data and no editing proxy, populate the form
    if (props.initialData && !props.editingProxy) {
      populateFormFromInitialData();
    }
  }
});

watch(() => props.editingProxy, (newValue) => {
  if (newValue) {
    form.value = {
      id: newValue.id,
      name: newValue.name,
      domain: newValue.domain,
      target_url: newValue.target_url,
      ssl_enabled: newValue.ssl_enabled,
    };
  }
});

watch(() => props.initialData, (newValue) => {
  if (newValue && !props.editingProxy && props.show) {
    populateFormFromInitialData();
  }
});

const populateFormFromInitialData = () => {
  if (props.initialData) {
    form.value = {
      name: props.initialData.name || '',
      domain: props.initialData.domain || '',
      target_url: props.initialData.target_url || '',
      ssl_enabled: props.initialData.ssl_enabled || false,
    };
  }
};

const resetForm = () => {
  form.value = {
    name: '',
    domain: '',
    target_url: '',
    ssl_enabled: false,
  };
  formValid.value = false;
  formRef.value?.reset();
};

const cancel = () => {
  emit('cancel');
  emit('update:show', false);
};

const save = () => {
  if (!formValid.value) return;

  const isEdit = !!props.editingProxy;
  const data = isEdit
    ? { ...form.value } as ProxyUpdateRequest
    : { ...form.value } as ProxyCreateRequest;

  emit('save', data, isEdit);
  };
</script>
