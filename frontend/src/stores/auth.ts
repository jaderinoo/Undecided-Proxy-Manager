import { defineStore } from 'pinia';
import { computed, ref } from 'vue';
import { apiService } from '../services/api';
import type { User, UserLoginRequest } from '../types/api';
import { isTokenExpired } from '../utils/jwt';

export const useAuthStore = defineStore('auth', () => {
  // State
  const user = ref<User | null>(null);
  const token = ref<string | null>(localStorage.getItem('upm_token'));
  const loading = ref(false);

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

    if (isTokenExpired(token.value)) {
      logout();
      return false;
    }

    try {
      await apiService.getSettings();
      return true;
    } catch {
      if (token.value) {
        logout();
      }
      return false;
    }
  };

  // Fast, local-only check (no network round-trip) for whether the current
  // token has passed its expiry. Used to detect a stale session on a tab
  // that's been left open without any navigation to trigger validateSession.
  // Returns true if the session was expired (and has now been logged out).
  const logoutIfExpired = (): boolean => {
    if (token.value && isTokenExpired(token.value)) {
      logout();
      return true;
    }
    return false;
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
    logoutIfExpired,
  };
});
