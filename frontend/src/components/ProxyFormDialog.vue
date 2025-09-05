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
              <v-switch
                v-model="form.ssl_enabled"
                label="Enable SSL"
                color="primary"
                hide-details
              />
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
                  <strong>Let's Encrypt Certificate:</strong> When SSL is enabled, a free SSL certificate will be automatically generated using Let's Encrypt for the specified domain. This may take a few moments.
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
import { ref, watch } from 'vue';
import type { Proxy, ProxyCreateRequest, ProxyUpdateRequest } from '../types/api';

interface Props {
  show: boolean;
  editingProxy?: Proxy | null;
  initialData?: Partial<ProxyCreateRequest>;
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
