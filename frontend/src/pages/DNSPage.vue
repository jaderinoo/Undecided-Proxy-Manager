<template>
  <AppLayout>
    <div class="dns-page">
      <div class="page-header">
        <h1 class="page-title">DNS Management</h1>
        <p class="page-description">Manage dynamic DNS configurations and records</p>
      </div>

      <!-- Public IP Display -->
      <div class="public-ip-section">
        <div class="card">
          <div class="card-header">
            <h3>Current Public IP</h3>
            <button @click="refreshPublicIP" :disabled="loadingPublicIP" class="btn btn-secondary">
              <i class="icon-refresh" :class="{ spinning: loadingPublicIP }"></i>
              Refresh
            </button>
          </div>
          <div class="card-content">
            <div class="ip-display">
              <span class="ip-label">Public IP:</span>
              <span class="ip-value">{{ publicIP || 'Loading...' }}</span>
            </div>
          </div>
        </div>
      </div>

      <!-- DNS Configurations -->
      <div class="dns-configs-section">
        <div class="section-header">
          <h2>DNS Configurations</h2>
          <button @click="showCreateConfigModal = true" class="btn btn-primary">
            <i class="icon-plus"></i>
            Add Configuration
          </button>
        </div>

        <div v-if="loadingConfigs" class="loading-container">
          <LoadingSpinner />
        </div>

        <div v-else-if="!dnsConfigs || dnsConfigs.length === 0" class="empty-state">
          <i class="icon-dns"></i>
          <h3>No DNS Configurations</h3>
          <p>Create your first DNS configuration to start managing dynamic DNS records.</p>
          <button @click="showCreateConfigModal = true" class="btn btn-primary">
            <i class="icon-plus"></i>
            Add Configuration
          </button>
        </div>

        <div v-else class="configs-grid">
          <div v-for="config in dnsConfigs || []" :key="config.id" class="config-card">
            <div class="config-header">
              <div class="config-info">
                <h3>{{ config.domain }}</h3>
                <span class="provider-badge">{{ config.provider }}</span>
              </div>
              <div class="config-actions">
                <button @click="editConfig(config)" class="btn btn-sm btn-secondary">
                  <i class="icon-edit"></i>
                </button>
                <button @click="deleteConfig(config.id)" class="btn btn-sm btn-danger">
                  <i class="icon-trash"></i>
                </button>
              </div>
            </div>
            
            <div class="config-details">
              <div class="detail-row">
                <span class="label">Status:</span>
                <span class="value" :class="{ active: config.is_active, inactive: !config.is_active }">
                  {{ config.is_active ? 'Active' : 'Inactive' }}
                </span>
              </div>
              <div class="detail-row">
                <span class="label">Last Update:</span>
                <span class="value">{{ formatDate(config.last_update) || 'Never' }}</span>
              </div>
              <div class="detail-row">
                <span class="label">Last IP:</span>
                <span class="value">{{ config.last_ip || 'Unknown' }}</span>
              </div>
            </div>

            <div class="config-records">
              <div class="records-header">
                <h4>DNS Records</h4>
                <button @click="openCreateRecordModal(config.id)" class="btn btn-sm btn-primary">
                  <i class="icon-plus"></i>
                  Add Record
                </button>
              </div>
              
              <div v-if="configRecords[config.id]?.length === 0" class="no-records">
                <p>No DNS records configured</p>
              </div>
              
              <div v-else class="records-list">
                <div v-for="record in configRecords[config.id] || []" :key="record.id" class="record-item">
                  <div class="record-info">
                    <span class="host">{{ record.host === '@' ? config.domain : `${record.host}.${config.domain}` }}</span>
                    <span class="ip">{{ record.current_ip || 'Not set' }}</span>
                  </div>
                  <div class="record-actions">
                    <button @click="updateRecordNow(record.id)" :disabled="loadingUpdates[record.id]" class="btn btn-sm btn-success">
                      <i class="icon-refresh" :class="{ spinning: loadingUpdates[record.id] }"></i>
                      Update
                    </button>
                    <button @click="editRecord(record)" class="btn btn-sm btn-secondary">
                      <i class="icon-edit"></i>
                    </button>
                    <button @click="deleteRecord(record.id)" class="btn btn-sm btn-danger">
                      <i class="icon-trash"></i>
                    </button>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- Create/Edit Config Modal -->
      <div v-if="showCreateConfigModal || showEditConfigModal" class="modal-overlay" @click="closeConfigModal">
        <div class="modal" @click.stop>
          <div class="modal-header">
            <h3>{{ showCreateConfigModal ? 'Create DNS Configuration' : 'Edit DNS Configuration' }}</h3>
            <button @click="closeConfigModal" class="btn btn-close">
              <i class="icon-close"></i>
            </button>
          </div>
          <div class="modal-content">
            <form @submit.prevent="saveConfig">
              <div class="form-group">
                <label for="provider">Provider</label>
                <select id="provider" v-model="configForm.provider" required>
                  <option value="namecheap">Namecheap</option>
                </select>
              </div>
              
              <div class="form-group">
                <label for="domain">Domain</label>
                <input 
                  id="domain" 
                  type="text" 
                  v-model="configForm.domain" 
                  placeholder="example.com"
                  required
                />
              </div>
              
              <div class="form-group">
                <label for="username">Username</label>
                <input 
                  id="username" 
                  type="text" 
                  v-model="configForm.username" 
                  placeholder="yourdomain.com"
                  required
                />
              </div>
              
              <div class="form-group">
                <label for="password">Password</label>
                <input 
                  id="password" 
                  type="password" 
                  v-model="configForm.password" 
                  placeholder="Dynamic DNS password"
                  required
                />
              </div>
              
              <div class="form-group">
                <label class="checkbox-label">
                  <input type="checkbox" v-model="configForm.is_active" />
                  <span class="checkmark"></span>
                  Active
                </label>
              </div>
            </form>
          </div>
          <div class="modal-footer">
            <button @click="closeConfigModal" class="btn btn-secondary">Cancel</button>
            <button @click="saveConfig" :disabled="savingConfig" class="btn btn-primary">
              <i class="icon-save" :class="{ spinning: savingConfig }"></i>
              {{ showCreateConfigModal ? 'Create' : 'Update' }}
            </button>
          </div>
        </div>
      </div>

      <!-- Create/Edit Record Modal -->
      <div v-if="showCreateRecordModal || showEditRecordModal" class="modal-overlay" @click="closeRecordModal">
        <div class="modal" @click.stop>
          <div class="modal-header">
            <h3>{{ showCreateRecordModal ? 'Create DNS Record' : 'Edit DNS Record' }}</h3>
            <button @click="closeRecordModal" class="btn btn-close">
              <i class="icon-close"></i>
            </button>
          </div>
          <div class="modal-content">
            <form @submit.prevent="saveRecord">
              <div class="form-group">
                <label for="host">Host</label>
                <input 
                  id="host" 
                  type="text" 
                  v-model="recordForm.host" 
                  placeholder="@ for root domain, www for subdomain"
                  required
                />
                <small class="form-help">Use "@" for root domain or enter subdomain name</small>
              </div>
              
              <div class="form-group">
                <label class="checkbox-label">
                  <input type="checkbox" v-model="recordForm.is_active" />
                  <span class="checkmark"></span>
                  Active
                </label>
              </div>
            </form>
          </div>
          <div class="modal-footer">
            <button @click="closeRecordModal" class="btn btn-secondary">Cancel</button>
            <button @click="saveRecord" :disabled="savingRecord" class="btn btn-primary">
              <i class="icon-save" :class="{ spinning: savingRecord }"></i>
              {{ showCreateRecordModal ? 'Create' : 'Update' }}
            </button>
          </div>
        </div>
      </div>

      <!-- Error Alert -->
      <ErrorAlert v-if="error" :message="error" @close="error = null" />
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import AppLayout from '../components/AppLayout.vue'
import LoadingSpinner from '../components/LoadingSpinner.vue'
import ErrorAlert from '../components/ErrorAlert.vue'
import apiService from '../services/api'
import type { 
  DNSConfig, 
  DNSConfigCreateRequest, 
  DNSConfigUpdateRequest,
  DNSRecord,
  DNSRecordCreateRequest,
  DNSRecordUpdateRequest,
  DNSUpdateResponse
} from '../types/api'

// Reactive data
const dnsConfigs = ref<DNSConfig[]>([])
const configRecords = ref<Record<number, DNSRecord[]>>({})
const publicIP = ref<string>('')
const loadingConfigs = ref(false)
const loadingPublicIP = ref(false)
const loadingUpdates = ref<Record<number, boolean>>({})
const savingConfig = ref(false)
const savingRecord = ref(false)
const error = ref<string | null>(null)

// Modal states
const showCreateConfigModal = ref(false)
const showEditConfigModal = ref(false)
const showCreateRecordModal = ref(false)
const showEditRecordModal = ref(false)

// Form data
const configForm = ref<DNSConfigCreateRequest & { is_active: boolean }>({
  provider: 'namecheap',
  domain: '',
  username: '',
  password: '',
  is_active: true
})

const recordForm = ref<DNSRecordCreateRequest & { is_active: boolean }>({
  config_id: 0,
  host: '',
  is_active: true
})

const editingConfig = ref<DNSConfig | null>(null)
const editingRecord = ref<DNSRecord | null>(null)

// Computed
const selectedConfigId = computed(() => {
  if (showCreateRecordModal.value) {
    return recordForm.value.config_id
  }
  return editingRecord.value?.config_id || 0
})

// Methods
const loadDNSConfigs = async () => {
  try {
    loadingConfigs.value = true
    const response = await apiService.getDNSConfigs()
    dnsConfigs.value = response.configs || []
    
    // Load records for each config
    for (const config of dnsConfigs.value) {
      await loadDNSRecords(config.id)
    }
  } catch (err) {
    error.value = `Failed to load DNS configurations: ${err}`
    dnsConfigs.value = []
  } finally {
    loadingConfigs.value = false
  }
}

const loadDNSRecords = async (configId: number) => {
  try {
    const response = await apiService.getDNSRecords(configId)
    configRecords.value[configId] = response.records || []
  } catch (err) {
    console.error(`Failed to load DNS records for config ${configId}:`, err)
    configRecords.value[configId] = []
  }
}

const refreshPublicIP = async () => {
  try {
    loadingPublicIP.value = true
    const response = await apiService.getPublicIP()
    publicIP.value = response.ip
  } catch (err) {
    error.value = `Failed to get public IP: ${err}`
  } finally {
    loadingPublicIP.value = false
  }
}

const editConfig = (config: DNSConfig) => {
  editingConfig.value = config
  configForm.value = {
    provider: config.provider,
    domain: config.domain,
    username: config.username,
    password: config.password,
    is_active: config.is_active
  }
  showEditConfigModal.value = true
}

const deleteConfig = async (id: number) => {
  if (!confirm('Are you sure you want to delete this DNS configuration? This will also delete all associated records.')) {
    return
  }
  
  try {
    await apiService.deleteDNSConfig(id)
    await loadDNSConfigs()
  } catch (err) {
    error.value = `Failed to delete DNS configuration: ${err}`
  }
}

const saveConfig = async () => {
  try {
    savingConfig.value = true
    
    if (showCreateConfigModal.value) {
      await apiService.createDNSConfig(configForm.value)
    } else if (editingConfig.value) {
      const updateData: DNSConfigUpdateRequest = {
        provider: configForm.value.provider,
        domain: configForm.value.domain,
        username: configForm.value.username,
        password: configForm.value.password,
        is_active: configForm.value.is_active
      }
      await apiService.updateDNSConfig(editingConfig.value.id, updateData)
    }
    
    await loadDNSConfigs()
    closeConfigModal()
  } catch (err) {
    error.value = `Failed to save DNS configuration: ${err}`
  } finally {
    savingConfig.value = false
  }
}

const openCreateRecordModal = (configId: number) => {
  recordForm.value = {
    config_id: configId,
    host: '',
    is_active: true
  }
  showCreateRecordModal.value = true
}

const editRecord = (record: DNSRecord) => {
  editingRecord.value = record
  recordForm.value = {
    config_id: record.config_id,
    host: record.host,
    is_active: record.is_active
  }
  showEditRecordModal.value = true
}

const deleteRecord = async (id: number) => {
  if (!confirm('Are you sure you want to delete this DNS record?')) {
    return
  }
  
  try {
    await apiService.deleteDNSRecord(id)
    await loadDNSConfigs()
  } catch (err) {
    error.value = `Failed to delete DNS record: ${err}`
  }
}

const saveRecord = async () => {
  try {
    savingRecord.value = true
    
    if (showCreateRecordModal.value) {
      await apiService.createDNSRecord(recordForm.value)
    } else if (editingRecord.value) {
      const updateData: DNSRecordUpdateRequest = {
        host: recordForm.value.host,
        is_active: recordForm.value.is_active
      }
      await apiService.updateDNSRecord(editingRecord.value.id, updateData)
    }
    
    await loadDNSConfigs()
    closeRecordModal()
  } catch (err) {
    error.value = `Failed to save DNS record: ${err}`
  } finally {
    savingRecord.value = false
  }
}

const updateRecordNow = async (recordId: number) => {
  try {
    loadingUpdates.value[recordId] = true
    const response = await apiService.updateDNSRecordNow(recordId)
    
    if (response.response.success) {
      await loadDNSConfigs()
    } else {
      error.value = `DNS update failed: ${response.response.message}`
    }
  } catch (err) {
    error.value = `Failed to update DNS record: ${err}`
  } finally {
    loadingUpdates.value[recordId] = false
  }
}

const closeConfigModal = () => {
  showCreateConfigModal.value = false
  showEditConfigModal.value = false
  editingConfig.value = null
  configForm.value = {
    provider: 'namecheap',
    domain: '',
    username: '',
    password: '',
    is_active: true
  }
}

const closeRecordModal = () => {
  showCreateRecordModal.value = false
  showEditRecordModal.value = false
  editingRecord.value = null
  recordForm.value = {
    config_id: 0,
    host: '',
    is_active: true
  }
}

const formatDate = (dateString?: string) => {
  if (!dateString) return null
  return new Date(dateString).toLocaleString()
}

// Lifecycle
onMounted(() => {
  loadDNSConfigs()
  refreshPublicIP()
})
</script>

<style scoped>
.dns-page {
  padding: 2rem;
  max-width: 1200px;
  margin: 0 auto;
}

.page-header {
  margin-bottom: 2rem;
}

.page-title {
  font-size: 2rem;
  font-weight: 600;
  color: var(--text-primary);
  margin: 0 0 0.5rem 0;
}

.page-description {
  color: var(--text-secondary);
  margin: 0;
}

.public-ip-section {
  margin-bottom: 2rem;
}

.card {
  background: var(--bg-secondary);
  border-radius: 8px;
  border: 1px solid var(--border-color);
  overflow: hidden;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 1rem 1.5rem;
  border-bottom: 1px solid var(--border-color);
  background: var(--bg-tertiary);
}

.card-header h3 {
  margin: 0;
  font-size: 1.1rem;
  font-weight: 600;
  color: var(--text-primary);
}

.card-content {
  padding: 1.5rem;
}

.ip-display {
  display: flex;
  align-items: center;
  gap: 1rem;
}

.ip-label {
  font-weight: 500;
  color: var(--text-secondary);
}

.ip-value {
  font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
  font-size: 1.1rem;
  font-weight: 600;
  color: var(--text-primary);
  background: var(--bg-primary);
  padding: 0.5rem 1rem;
  border-radius: 4px;
  border: 1px solid var(--border-color);
}

.section-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 1.5rem;
}

.section-header h2 {
  margin: 0;
  font-size: 1.5rem;
  font-weight: 600;
  color: var(--text-primary);
}

.configs-grid {
  display: grid;
  gap: 1.5rem;
  grid-template-columns: repeat(auto-fill, minmax(400px, 1fr));
}

.config-card {
  background: var(--bg-secondary);
  border-radius: 8px;
  border: 1px solid var(--border-color);
  overflow: hidden;
}

.config-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 1rem 1.5rem;
  border-bottom: 1px solid var(--border-color);
  background: var(--bg-tertiary);
}

.config-info h3 {
  margin: 0 0 0.5rem 0;
  font-size: 1.1rem;
  font-weight: 600;
  color: var(--text-primary);
}

.provider-badge {
  display: inline-block;
  padding: 0.25rem 0.5rem;
  background: var(--accent-color);
  color: white;
  border-radius: 4px;
  font-size: 0.8rem;
  font-weight: 500;
  text-transform: uppercase;
}

.config-actions {
  display: flex;
  gap: 0.5rem;
}

.config-details {
  padding: 1rem 1.5rem;
  border-bottom: 1px solid var(--border-color);
}

.detail-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 0.5rem;
}

.detail-row:last-child {
  margin-bottom: 0;
}

.detail-row .label {
  font-weight: 500;
  color: var(--text-secondary);
}

.detail-row .value {
  color: var(--text-primary);
}

.detail-row .value.active {
  color: var(--success-color);
}

.detail-row .value.inactive {
  color: var(--warning-color);
}

.config-records {
  padding: 1rem 1.5rem;
}

.records-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 1rem;
}

.records-header h4 {
  margin: 0;
  font-size: 1rem;
  font-weight: 600;
  color: var(--text-primary);
}

.no-records {
  text-align: center;
  padding: 2rem;
  color: var(--text-secondary);
}

.records-list {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.record-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0.75rem;
  background: var(--bg-primary);
  border-radius: 6px;
  border: 1px solid var(--border-color);
}

.record-info {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}

.record-info .host {
  font-weight: 500;
  color: var(--text-primary);
}

.record-info .ip {
  font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
  font-size: 0.9rem;
  color: var(--text-secondary);
}

.record-actions {
  display: flex;
  gap: 0.5rem;
}

.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}

.modal {
  background: var(--bg-secondary);
  border-radius: 8px;
  border: 1px solid var(--border-color);
  width: 90%;
  max-width: 500px;
  max-height: 90vh;
  overflow: hidden;
  box-shadow: 0 10px 25px rgba(0, 0, 0, 0.2);
}

.modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 1.5rem;
  border-bottom: 1px solid var(--border-color);
  background: var(--bg-tertiary);
}

.modal-header h3 {
  margin: 0;
  font-size: 1.2rem;
  font-weight: 600;
  color: var(--text-primary);
}

.modal-content {
  padding: 1.5rem;
  max-height: 60vh;
  overflow-y: auto;
}

.modal-footer {
  display: flex;
  justify-content: flex-end;
  gap: 1rem;
  padding: 1.5rem;
  border-top: 1px solid var(--border-color);
  background: var(--bg-tertiary);
}

.form-group {
  margin-bottom: 1.5rem;
}

.form-group:last-child {
  margin-bottom: 0;
}

.form-group label {
  display: block;
  margin-bottom: 0.5rem;
  font-weight: 500;
  color: var(--text-primary);
}

.form-group input,
.form-group select {
  width: 100%;
  padding: 0.75rem;
  border: 1px solid var(--border-color);
  border-radius: 4px;
  background: var(--bg-primary);
  color: var(--text-primary);
  font-size: 1rem;
}

.form-group input:focus,
.form-group select:focus {
  outline: none;
  border-color: var(--accent-color);
  box-shadow: 0 0 0 2px rgba(var(--accent-color-rgb), 0.2);
}

.form-help {
  display: block;
  margin-top: 0.25rem;
  font-size: 0.875rem;
  color: var(--text-secondary);
}

.checkbox-label {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  cursor: pointer;
  font-weight: normal;
}

.checkbox-label input[type="checkbox"] {
  width: auto;
  margin: 0;
}

.btn {
  display: inline-flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.5rem 1rem;
  border: none;
  border-radius: 4px;
  font-size: 0.9rem;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s ease;
  text-decoration: none;
}

.btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.btn-primary {
  background: var(--accent-color);
  color: white;
}

.btn-primary:hover:not(:disabled) {
  background: var(--accent-color-dark);
}

.btn-secondary {
  background: var(--bg-tertiary);
  color: var(--text-primary);
  border: 1px solid var(--border-color);
}

.btn-secondary:hover:not(:disabled) {
  background: var(--bg-primary);
}

.btn-success {
  background: var(--success-color);
  color: white;
}

.btn-success:hover:not(:disabled) {
  background: var(--success-color-dark);
}

.btn-danger {
  background: var(--danger-color);
  color: white;
}

.btn-danger:hover:not(:disabled) {
  background: var(--danger-color-dark);
}

.btn-sm {
  padding: 0.375rem 0.75rem;
  font-size: 0.8rem;
}

.btn-close {
  background: none;
  color: var(--text-secondary);
  padding: 0.5rem;
}

.btn-close:hover {
  color: var(--text-primary);
}

.spinning {
  animation: spin 1s linear infinite;
}

@keyframes spin {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}

.empty-state {
  text-align: center;
  padding: 3rem;
  color: var(--text-secondary);
}

.empty-state i {
  font-size: 3rem;
  margin-bottom: 1rem;
  opacity: 0.5;
}

.empty-state h3 {
  margin: 0 0 0.5rem 0;
  font-size: 1.2rem;
  color: var(--text-primary);
}

.empty-state p {
  margin: 0 0 1.5rem 0;
}

.loading-container {
  display: flex;
  justify-content: center;
  padding: 2rem;
}
</style>
