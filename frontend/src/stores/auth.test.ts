import { createPinia, setActivePinia } from 'pinia';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { makeToken } from '../test/makeToken';
import { useAuthStore } from './auth';

vi.mock('../services/api', () => ({
  apiService: {
    login: vi.fn(),
    setAuthToken: vi.fn(),
    clearAuthToken: vi.fn(),
    getSettings: vi.fn(),
  },
}));

import { apiService } from '../services/api';

const futureToken = () =>
  makeToken({ is_admin: true, exp: Math.floor(Date.now() / 1000) + 3600 });
const expiredToken = () =>
  makeToken({ is_admin: true, exp: Math.floor(Date.now() / 1000) - 3600 });

describe('useAuthStore', () => {
  beforeEach(() => {
    localStorage.clear();
    setActivePinia(createPinia());
    vi.clearAllMocks();
  });

  describe('login', () => {
    it('stores the token and user, and persists the token to localStorage', async () => {
      const token = futureToken();
      const user = { id: 1, username: 'admin', email: 'admin@upm.local', is_active: true, created_at: '', updated_at: '' };
      (apiService.login as ReturnType<typeof vi.fn>).mockResolvedValue({ data: { token, user } });

      const store = useAuthStore();
      await store.login({ password: 'secret' });

      expect(store.token).toBe(token);
      expect(store.user).toEqual(user);
      expect(localStorage.getItem('upm_token')).toBe(token);
      expect(apiService.setAuthToken).toHaveBeenCalledWith(token);
    });

    it('logs out and rethrows when the backend rejects the login', async () => {
      (apiService.login as ReturnType<typeof vi.fn>).mockRejectedValue(new Error('bad password'));

      const store = useAuthStore();
      await expect(store.login({ password: 'wrong' })).rejects.toThrow('bad password');

      expect(store.token).toBeNull();
      expect(store.user).toBeNull();
      expect(localStorage.getItem('upm_token')).toBeNull();
    });
  });

  describe('logout', () => {
    it('clears state, localStorage, and the api client auth header', () => {
      localStorage.setItem('upm_token', futureToken());
      const store = useAuthStore();

      store.logout();

      expect(store.token).toBeNull();
      expect(store.user).toBeNull();
      expect(localStorage.getItem('upm_token')).toBeNull();
      expect(apiService.clearAuthToken).toHaveBeenCalled();
    });
  });

  describe('validateSession', () => {
    it('returns false without calling the backend when there is no token', async () => {
      const store = useAuthStore();
      const valid = await store.validateSession();

      expect(valid).toBe(false);
      expect(apiService.getSettings).not.toHaveBeenCalled();
    });

    it('logs out and returns false, without calling the backend, when the token is locally expired', async () => {
      localStorage.setItem('upm_token', expiredToken());
      const store = useAuthStore();

      const valid = await store.validateSession();

      expect(valid).toBe(false);
      expect(store.token).toBeNull();
      expect(apiService.getSettings).not.toHaveBeenCalled();
    });

    it('returns true when the token is unexpired and the backend confirms it', async () => {
      localStorage.setItem('upm_token', futureToken());
      (apiService.getSettings as ReturnType<typeof vi.fn>).mockResolvedValue({});
      const store = useAuthStore();

      const valid = await store.validateSession();

      expect(valid).toBe(true);
      expect(store.token).not.toBeNull();
    });

    it('logs out and returns false when the backend rejects an unexpired token', async () => {
      localStorage.setItem('upm_token', futureToken());
      (apiService.getSettings as ReturnType<typeof vi.fn>).mockRejectedValue(new Error('401'));
      const store = useAuthStore();

      const valid = await store.validateSession();

      expect(valid).toBe(false);
      expect(store.token).toBeNull();
    });

    it('re-checks with the backend every time (no permanent cache)', async () => {
      localStorage.setItem('upm_token', futureToken());
      (apiService.getSettings as ReturnType<typeof vi.fn>).mockResolvedValue({});
      const store = useAuthStore();

      await store.validateSession();
      await store.validateSession();
      await store.validateSession();

      expect(apiService.getSettings).toHaveBeenCalledTimes(3);
    });
  });

  describe('logoutIfExpired', () => {
    it('does nothing and returns false when there is no token', () => {
      const store = useAuthStore();
      expect(store.logoutIfExpired()).toBe(false);
      expect(apiService.clearAuthToken).not.toHaveBeenCalled();
    });

    it('does nothing and returns false when the token is not expired', () => {
      localStorage.setItem('upm_token', futureToken());
      const store = useAuthStore();

      expect(store.logoutIfExpired()).toBe(false);
      expect(store.token).not.toBeNull();
    });

    it('logs out and returns true when the token is expired', () => {
      localStorage.setItem('upm_token', expiredToken());
      const store = useAuthStore();

      expect(store.logoutIfExpired()).toBe(true);
      expect(store.token).toBeNull();
      expect(localStorage.getItem('upm_token')).toBeNull();
    });
  });
});
