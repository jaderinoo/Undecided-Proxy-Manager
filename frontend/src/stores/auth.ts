import { defineStore } from 'pinia';
import { computed, ref } from 'vue';
import { apiService } from '../services/api';
import type { User, UserLoginRequest } from '../types/api';

export const useAuthStore = defineStore('auth', () => {
  // State
  const user = ref<User | null>(null);
  const token = ref<string | null>(localStorage.getItem('upm_token'));
  const loading = ref(false);
  const sessionValidated = ref(false);

  // Getters
  const isAuthenticated = computed(() => !!token.value);
  // Single admin auth - always admin if authenticated
  const isAdmin = computed(() => isAuthenticated.value);

  // Actions
  const login = async (credentials: UserLoginRequest) => {
    try {
      loading.value = true;
      const response = await apiService.login(credentials);

      // Store the token and user data
      token.value = response.data.token;
      user.value = response.data.user;

      // Persist token to localStorage
      localStorage.setItem('upm_token', response.data.token);

      // Set default authorization header for future requests
      apiService.setAuthToken(response.data.token);
      sessionValidated.value = true;

      return response.data;
    } catch (error) {
      // Clear any existing auth data on login failure
      logout();
      throw error;
    } finally {
      loading.value = false;
    }
  };

  const logout = () => {
    user.value = null;
    token.value = null;
    sessionValidated.value = false;
    localStorage.removeItem('upm_token');
    apiService.clearAuthToken();
  };

  const initializeAuth = () => {
    if (token.value) {
      apiService.setAuthToken(token.value);
      user.value = {
        id: 1,
        username: 'admin',
        email: 'admin@upm.local',
        is_active: true,
        created_at: new Date().toISOString(),
        updated_at: new Date().toISOString(),
      };
    }
  };

  const validateSession = async (): Promise<boolean> => {
    if (!token.value) {
      return false;
    }

    if (sessionValidated.value) {
      return true;
    }

    try {
      await apiService.getSettings();
      sessionValidated.value = true;
      return true;
    } catch {
      if (token.value) {
        logout();
      }
      return false;
    }
  };

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
    initializeAuth,
    validateSession,
  };
});
