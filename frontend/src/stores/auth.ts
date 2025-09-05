import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { apiService } from '../services/api'
import type { User, UserLoginRequest, AuthResponse } from '../types/api'

export const useAuthStore = defineStore('auth', () => {
  // State
  const user = ref<User | null>(null)
  const token = ref<string | null>(localStorage.getItem('upm_token'))
  const loading = ref(false)

  // Getters
  const isAuthenticated = computed(() => !!token.value)
  // Single admin auth - always admin if authenticated
  const isAdmin = computed(() => isAuthenticated.value)

  // Actions
  const login = async (credentials: UserLoginRequest) => {
    try {
      loading.value = true
      const response = await apiService.login(credentials)
      
      // Store the token and user data
      token.value = response.data.token
      user.value = response.data.user
      
      // Persist token to localStorage
      localStorage.setItem('upm_token', response.data.token)
      
      // Set default authorization header for future requests
      apiService.setAuthToken(response.data.token)
      
      return response.data
    } catch (error) {
      // Clear any existing auth data on login failure
      logout()
      throw error
    } finally {
      loading.value = false
    }
  }

  const logout = () => {
    user.value = null
    token.value = null
    localStorage.removeItem('upm_token')
    apiService.clearAuthToken()
  }

  const initializeAuth = () => {
    // Check if we have a stored token
    if (token.value) {
      // Set the auth token for API requests
      apiService.setAuthToken(token.value)
      
      // Create a basic user object (we could validate the token here)
      user.value = {
        id: 1,
        username: 'admin',
        email: 'admin@upm.local',
        is_active: true,
        created_at: new Date().toISOString(),
        updated_at: new Date().toISOString()
      }
    }
  }

  return {
    // State
    user,
    token,
    loading,
    
    // Getters
    isAuthenticated,
    isAdmin,
    
    // Actions
    login,
    logout,
    initializeAuth
  }
})
