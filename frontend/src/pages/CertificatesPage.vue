<template>
  <AppLayout @refresh="loadCertificates">
    <v-container>
      <v-row>
        <v-col cols="12">
          <v-card>
            <v-card-title>
              <v-icon left>mdi-certificate</v-icon>
              SSL Certificates
            </v-card-title>
            <v-card-text>
              <ErrorAlert :error="error" @clear="error = null" />

              <LoadingSpinner v-if="isLoading" />

              <div v-else>
                <PageHeader
                  :count="certificates?.length || 0"
                  item-name="Certificates"
                  :loading="isLoading"
                  @refresh="loadCertificates"
                >
                  <template #actions>
                    <v-btn
                      color="success"
                      variant="outlined"
                      size="small"
                      @click="showCreateDialog = true"
                    >
                      <v-icon left>mdi-plus</v-icon>
                      Add Certificate
                    </v-btn>
                  </template>
                </PageHeader>

                <!-- Filter and Search -->
                <FilterBar
                  v-model:search-query="searchQuery"
                  v-model:status-filter="statusFilter"
                  v-model:sort-by="sortBy"
                  search-label="Search certificates..."
                  :status-options="statusFilterItems"
                  :sort-options="sortOptions"
                  @search="filterCertificates"
                />

                <!-- Certificate Stats -->
                <StatsCards :stats="certificateStats" />

                <!-- Certificate List -->
                <div v-if="filteredCertificates && filteredCertificates.length > 0">
                  <CertificateCard
                    v-for="certificate in filteredCertificates"
                    :key="certificate.id"
                    :certificate="certificate"
                    @deleted="handleCertificateDeleted"
                    @renewed="handleCertificateRenewed"
                  />
                </div>

                <v-empty-state
                  v-else-if="certificates && certificates.length === 0"
                  title="No certificates found"
                  text="No SSL certificates are currently available"
                />

                <v-empty-state
                  v-else
                  title="No matching certificates"
                  text="Try adjusting your search or filter criteria"
                />
              </div>
            </v-card-text>
          </v-card>
        </v-col>
      </v-row>
    </v-container>

    <!-- Create Certificate Dialog -->
    <v-dialog v-model="showCreateDialog" max-width="600">
      <v-card>
        <v-card-title class="d-flex align-center">
          <v-icon left color="primary">mdi-certificate-plus</v-icon>
          Add New Certificate
        </v-card-title>
        
        <v-card-text>
          <v-form @submit.prevent="createCertificate" ref="form">
            <v-row>
              <v-col cols="12">
                <v-text-field
                  v-model="newCertificate.domain"
                  label="Domain"
                  placeholder="example.com or *.example.com"
                  variant="outlined"
                  :rules="[v => !!v || 'Domain is required']"
                  required
                ></v-text-field>
              </v-col>
              
              <v-col cols="12">
                <v-text-field
                  v-model="newCertificate.cert_path"
                  label="Certificate Path"
                  placeholder="/etc/nginx/ssl/example.com.crt"
                  variant="outlined"
                  :rules="[v => !!v || 'Certificate path is required']"
                  required
                ></v-text-field>
              </v-col>
              
              <v-col cols="12">
                <v-text-field
                  v-model="newCertificate.key_path"
                  label="Private Key Path"
                  placeholder="/etc/nginx/ssl/example.com.key"
                  variant="outlined"
                  :rules="[v => !!v || 'Private key path is required']"
                  required
                ></v-text-field>
              </v-col>
              
              <v-col cols="12">
                <v-text-field
                  v-model="newCertificate.expires_at"
                  label="Expiration Date"
                  type="datetime-local"
                  variant="outlined"
                  :rules="[v => !!v || 'Expiration date is required']"
                  required
                ></v-text-field>
              </v-col>
            </v-row>
          </v-form>
        </v-card-text>
        
        <v-card-actions class="pa-4">
          <v-spacer></v-spacer>
          <v-btn
            color="grey"
            variant="text"
            @click="closeCreateDialog"
          >
            Cancel
          </v-btn>
          <v-btn
            color="primary"
            @click="createCertificate"
            :loading="isCreating"
          >
            Create Certificate
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
  </AppLayout>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import type { Certificate, CertificateCreateRequest } from '../types/api'
import apiService from '../services/api'
import AppLayout from '../components/AppLayout.vue'
import ErrorAlert from '../components/ErrorAlert.vue'
import LoadingSpinner from '../components/LoadingSpinner.vue'
import CertificateCard from '../components/CertificateCard.vue'
import PageHeader from '../components/PageHeader.vue'
import StatsCards from '../components/StatsCards.vue'
import FilterBar from '../components/FilterBar.vue'

const certificates = ref<Certificate[]>([])
const filteredCertificates = ref<Certificate[]>([])
const isLoading = ref(false)
const error = ref<string | null>(null)
const showCreateDialog = ref(false)
const isCreating = ref(false)
const searchQuery = ref('')
const statusFilter = ref('')
const sortBy = ref('domain')
const form = ref()

const newCertificate = ref<CertificateCreateRequest>({
  domain: '',
  cert_path: '',
  key_path: '',
  expires_at: ''
})

const statusFilterItems = [
  { title: 'All Status', value: '' },
  { title: 'Valid', value: 'valid' },
  { title: 'Invalid', value: 'invalid' },
  { title: 'Expiring Soon', value: 'expiring' }
]

const sortOptions = [
  { title: 'Domain', value: 'domain' },
  { title: 'Status', value: 'status' },
  { title: 'Expires', value: 'expires_at' },
  { title: 'Created', value: 'created_at' }
]

const validCertificates = computed(() => 
  certificates.value.filter(cert => cert.is_valid).length
)

const invalidCertificates = computed(() => 
  certificates.value.filter(cert => !cert.is_valid).length
)

const expiringSoon = computed(() => 
  certificates.value.filter(cert => {
    const now = new Date()
    const expiry = new Date(cert.expires_at)
    const daysUntilExpiry = Math.ceil((expiry.getTime() - now.getTime()) / (1000 * 60 * 60 * 24))
    return daysUntilExpiry <= 30 && daysUntilExpiry > 0
  }).length
)

const sslEnabledCount = computed(() => 
  certificates.value.filter(cert => cert.is_valid).length
)

const certificateStats = computed(() => [
  {
    key: 'valid',
    value: validCertificates.value,
    label: 'Valid',
    icon: 'mdi-check-circle',
    color: 'green-lighten-5',
    iconColor: 'green'
  },
  {
    key: 'invalid',
    value: invalidCertificates.value,
    label: 'Invalid',
    icon: 'mdi-alert-circle',
    color: 'red-lighten-5',
    iconColor: 'red'
  },
  {
    key: 'expiring',
    value: expiringSoon.value,
    label: 'Expiring Soon',
    icon: 'mdi-clock-alert',
    color: 'orange-lighten-5',
    iconColor: 'orange'
  },
  {
    key: 'ssl',
    value: sslEnabledCount.value,
    label: 'SSL Enabled',
    icon: 'mdi-lock',
    color: 'blue-lighten-5',
    iconColor: 'blue'
  }
])

const loadCertificates = async () => {
  try {
    isLoading.value = true
    error.value = null
    const response = await apiService.getCertificates()
    certificates.value = response.data || []
    filteredCertificates.value = [...certificates.value]
    filterCertificates()
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Failed to load certificates'
    certificates.value = []
    filteredCertificates.value = []
  } finally {
    isLoading.value = false
  }
}

const filterCertificates = () => {
  let filtered = [...certificates.value]

  // Search filter
  if (searchQuery.value) {
    const query = searchQuery.value.toLowerCase()
    filtered = filtered.filter(cert => 
      cert.domain.toLowerCase().includes(query) ||
      cert.cert_path.toLowerCase().includes(query) ||
      cert.key_path.toLowerCase().includes(query)
    )
  }

  // Status filter
  if (statusFilter.value) {
    const now = new Date()
    filtered = filtered.filter(cert => {
      const expiry = new Date(cert.expires_at)
      const daysUntilExpiry = Math.ceil((expiry.getTime() - now.getTime()) / (1000 * 60 * 60 * 24))

      switch (statusFilter.value) {
        case 'valid':
          return cert.is_valid && daysUntilExpiry > 0
        case 'invalid':
          return !cert.is_valid || daysUntilExpiry <= 0
        case 'expiring':
          return daysUntilExpiry <= 30 && daysUntilExpiry > 0
        default:
          return true
      }
    })
  }

  // Sort
  filtered.sort((a, b) => {
    switch (sortBy.value) {
      case 'domain':
        return a.domain.localeCompare(b.domain)
      case 'status':
        return a.is_valid === b.is_valid ? 0 : a.is_valid ? -1 : 1
      case 'expires_at':
        return new Date(a.expires_at).getTime() - new Date(b.expires_at).getTime()
      case 'created_at':
        return new Date(b.created_at).getTime() - new Date(a.created_at).getTime()
      default:
        return 0
    }
  })

  filteredCertificates.value = filtered
}


const createCertificate = async () => {
  if (isCreating.value) return

  const { valid } = await form.value.validate()
  if (!valid) return

  try {
    isCreating.value = true
    error.value = null
    const response = await apiService.createCertificate(newCertificate.value)
    certificates.value.unshift(response.data)
    filterCertificates()
    closeCreateDialog()
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Failed to create certificate'
  } finally {
    isCreating.value = false
  }
}

const closeCreateDialog = () => {
  showCreateDialog.value = false
  newCertificate.value = {
    domain: '',
    cert_path: '',
    key_path: '',
    expires_at: ''
  }
  form.value?.reset()
}

const handleCertificateDeleted = (id: number) => {
  certificates.value = certificates.value.filter(cert => cert.id !== id)
  filterCertificates()
}

const handleCertificateRenewed = (certificate: Certificate) => {
  const index = certificates.value.findIndex(cert => cert.id === certificate.id)
  if (index !== -1) {
    certificates.value[index] = certificate
    filterCertificates()
  }
}


onMounted(() => {
  loadCertificates()
})
</script>

<style scoped>
.gap-2 {
  gap: 8px;
}
</style>