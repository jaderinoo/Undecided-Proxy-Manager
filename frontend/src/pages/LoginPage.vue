<template>
  <v-app>
    <v-main>
      <v-container fluid fill-height>
        <v-row align="center" justify="center">
          <v-col cols="12" sm="8" md="6" lg="4">
            <v-card elevation="8" class="pa-6">
              <v-card-title class="text-center mb-6">
                <v-icon size="48" color="primary" class="mb-2">mdi-proxy</v-icon>
                <h2 class="text-h4 font-weight-bold">UPM Login</h2>
                <p class="text-subtitle-1 text-grey-darken-1 mt-2">
                  Undecided Proxy Manager
                </p>
              </v-card-title>

              <v-card-text>
                <ErrorAlert :error="error" @clear="error = null" />

                <v-form @submit.prevent="handleLogin" ref="form">
                  <v-text-field
                    v-model="credentials.password"
                    label="Password"
                    prepend-inner-icon="mdi-lock"
                    variant="outlined"
                    type="password"
                    :rules="[rules.required]"
                    :disabled="loading"
                    class="mb-6"
                  />

                  <v-btn
                    type="submit"
                    color="primary"
                    size="large"
                    block
                    :loading="loading"
                    :disabled="!isFormValid"
                    class="mb-4"
                  >
                    <v-icon left>mdi-login</v-icon>
                    Login
                  </v-btn>
                </v-form>

                <v-divider class="my-4"></v-divider>

                <v-alert
                  type="info"
                  variant="tonal"
                  class="text-body-2"
                >
                  <v-icon left>mdi-information</v-icon>
                  <strong>Pi-hole Style Authentication</strong><br>
                  UPM uses single admin authentication. Set the <code>ADMIN_PASSWORD</code> environment variable to configure access.
                </v-alert>
              </v-card-text>
            </v-card>
          </v-col>
        </v-row>
      </v-container>
    </v-main>
  </v-app>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'
import ErrorAlert from '../components/ErrorAlert.vue'
import type { UserLoginRequest } from '../types/api'

const router = useRouter()
const authStore = useAuthStore()

const credentials = ref<UserLoginRequest>({
  password: ''
})

const loading = ref(false)
const error = ref<string | null>(null)

const rules = {
  required: (value: string) => !!value || 'This field is required'
}

const isFormValid = computed(() => {
  return !!credentials.value.password
})

const handleLogin = async () => {
  if (!isFormValid.value) return

  try {
    loading.value = true
    error.value = null
    
    await authStore.login(credentials.value)
    router.push('/')
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Login failed'
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.v-card {
  border-radius: 16px;
}

code {
  background-color: rgba(0, 0, 0, 0.1);
  padding: 2px 6px;
  border-radius: 4px;
  font-family: 'Courier New', monospace;
}
</style>
