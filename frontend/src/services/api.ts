import type {
  ApiResponse,
  AuthResponse,
  Certificate,
  CertificateCreateRequest,
  CertificateUpdateRequest,
  Container,
  ContainerListResponse,
  DNSConfig,
  DNSConfigCreateRequest,
  DNSConfigUpdateRequest,
  DNSRecord,
  DNSRecordCreateRequest,
  DNSRecordUpdateRequest,
  DNSStatus,
  DNSUpdateResponse,
  JobInfo,
  Proxy,
  ProxyCreateRequest,
  ProxyResponse,
  ProxyUpdateRequest,
  Settings,
  SettingsUpdateRequest,
  User,
  UserCreateRequest,
  UserLoginRequest,
} from '../types/api';

// Determine API base URL based on environment
const getApiBaseUrl = () => {
  // Development mode: Frontend on localhost:6071, Backend on localhost:6081
  if (window.location.hostname === 'localhost' && window.location.port === '6071') {
    return 'http://localhost:6081';
  }

  // Production mode: Use the same hostname and protocol as the frontend
  // This ensures API calls go through nginx proxy instead of direct backend access
  return `${window.location.protocol}//${window.location.host}`;
};

const API_BASE_URL = getApiBaseUrl();

class ApiService {
  private baseUrl: string;
  private authToken: string | null = null;

  constructor(baseUrl: string = API_BASE_URL) {
    this.baseUrl = baseUrl;
  }

  setAuthToken(token: string) {
    this.authToken = token;
  }

  clearAuthToken() {
    this.authToken = null;
  }

  private async request<T>(
    endpoint: string,
    options: RequestInit = {}
  ): Promise<T> {
    const url = `${this.baseUrl}${endpoint}`;

    const headers: Record<string, string> = {
      'Content-Type': 'application/json',
      ...(options.headers as Record<string, string>),
    };

    // Add authorization header if we have a token
    if (this.authToken) {
      headers['Authorization'] = `Bearer ${this.authToken}`;
    }

    const response = await fetch(url, {
      headers,
      ...options,
    });

    if (!response.ok) {
      throw new Error(`API Error: ${response.status} ${response.statusText}`);
    }

    // Handle 204 No Content responses (like DELETE operations)
    if (response.status === 204) {
      return {} as T;
    }

    return response.json();
  }

  // Health Check
  async getHealth(): Promise<{ status: string; message: string }> {
    return this.request('/health');
  }

  // Proxy endpoints
  async getProxies(): Promise<ApiResponse<Proxy[]>> {
    return this.request('/api/v1/proxies');
  }

  async getProxy(id: number): Promise<ApiResponse<Proxy>> {
    return this.request(`/api/v1/proxies/${id}`);
  }

  async createProxy(proxy: ProxyCreateRequest): Promise<ProxyResponse> {
    return this.request('/api/v1/proxies', {
      method: 'POST',
      body: JSON.stringify(proxy),
    });
  }

  async updateProxy(
    id: number,
    proxy: ProxyUpdateRequest
  ): Promise<ProxyResponse> {
    return this.request(`/api/v1/proxies/${id}`, {
      method: 'PUT',
      body: JSON.stringify(proxy),
    });
  }

  async deleteProxy(id: number): Promise<{ message: string }> {
    return this.request(`/api/v1/proxies/${id}`, {
      method: 'DELETE',
    });
  }

  async getProxyCertificate(id: number): Promise<ApiResponse<Certificate>> {
    return this.request(`/api/v1/proxies/${id}/certificate`);
  }

  // User endpoints
  async getUsers(): Promise<ApiResponse<User[]>> {
    return this.request('/api/v1/users');
  }

  async getUser(id: number): Promise<ApiResponse<User>> {
    return this.request(`/api/v1/users/${id}`);
  }

  async createUser(user: UserCreateRequest): Promise<ApiResponse<User>> {
    return this.request('/api/v1/users', {
      method: 'POST',
      body: JSON.stringify(user),
    });
  }

  async updateUser(
    id: number,
    user: Partial<User>
  ): Promise<ApiResponse<User>> {
    return this.request(`/api/v1/users/${id}`, {
      method: 'PUT',
      body: JSON.stringify(user),
    });
  }

  async deleteUser(id: number): Promise<{ message: string }> {
    return this.request(`/api/v1/users/${id}`, {
      method: 'DELETE',
    });
  }

  // Auth endpoints
  async login(
    credentials: UserLoginRequest
  ): Promise<ApiResponse<AuthResponse>> {
    return this.request('/api/v1/auth/login', {
      method: 'POST',
      body: JSON.stringify(credentials),
    });
  }

  async register(
    userData: UserCreateRequest
  ): Promise<ApiResponse<AuthResponse>> {
    return this.request('/api/v1/auth/register', {
      method: 'POST',
      body: JSON.stringify(userData),
    });
  }

  // Container endpoints
  async getContainers(): Promise<ContainerListResponse> {
    return this.request('/api/v1/containers');
  }

  async getContainer(id: string): Promise<ApiResponse<Container>> {
    return this.request(`/api/v1/containers/${id}`);
  }

  async getContainerStats(id: string): Promise<any> {
    return this.request(`/api/v1/containers/${id}/stats`);
  }

  // Nginx management endpoints
  async reloadNginx(): Promise<{ message: string }> {
    return this.request('/api/v1/nginx/reload', {
      method: 'POST',
    });
  }

  async testNginxConfig(): Promise<{ message: string }> {
    return this.request('/api/v1/nginx/test', {
      method: 'POST',
    });
  }

  async getAdminIPRestrictions(): Promise<{ allowed_ranges: string[] }> {
    return this.request('/api/v1/nginx/admin-ip-restrictions');
  }

  async updateAdminIPRestrictions(allowedRanges: string[]): Promise<{ message: string }> {
    return this.request('/api/v1/nginx/admin-ip-restrictions', {
      method: 'PUT',
      body: JSON.stringify({ allowed_ranges: allowedRanges }),
    });
  }

  async regenerateProxyConfig(domain: string): Promise<{ message: string }> {
    return this.request(`/api/v1/nginx/regenerate-config?domain=${encodeURIComponent(domain)}`, {
      method: 'POST',
    });
  }

  // DNS management endpoints
  async getDNSConfigs(): Promise<{ configs: DNSConfig[] }> {
    return this.request('/api/v1/dns/configs');
  }

  async getDNSConfig(id: number): Promise<{ config: DNSConfig }> {
    return this.request(`/api/v1/dns/configs/${id}`);
  }

  async createDNSConfig(
    data: DNSConfigCreateRequest
  ): Promise<{ config: DNSConfig }> {
    return this.request('/api/v1/dns/configs', {
      method: 'POST',
      body: JSON.stringify(data),
    });
  }

  async updateDNSConfig(
    id: number,
    data: DNSConfigUpdateRequest
  ): Promise<{ config: DNSConfig }> {
    return this.request(`/api/v1/dns/configs/${id}`, {
      method: 'PUT',
      body: JSON.stringify(data),
    });
  }

  async deleteDNSConfig(id: number): Promise<{ message: string }> {
    return this.request(`/api/v1/dns/configs/${id}`, {
      method: 'DELETE',
    });
  }

  async getDNSRecords(configId: number): Promise<{ records: DNSRecord[] }> {
    return this.request(`/api/v1/dns/records?config_id=${configId}`);
  }

  async getDNSRecord(id: number): Promise<{ record: DNSRecord }> {
    return this.request(`/api/v1/dns/records/${id}`);
  }

  async createDNSRecord(
    data: DNSRecordCreateRequest
  ): Promise<{ record: DNSRecord }> {
    return this.request('/api/v1/dns/records', {
      method: 'POST',
      body: JSON.stringify(data),
    });
  }

  async updateDNSRecord(
    id: number,
    data: DNSRecordUpdateRequest
  ): Promise<{ record: DNSRecord }> {
    return this.request(`/api/v1/dns/records/${id}`, {
      method: 'PUT',
      body: JSON.stringify(data),
    });
  }

  async deleteDNSRecord(id: number): Promise<{ message: string }> {
    return this.request(`/api/v1/dns/records/${id}`, {
      method: 'DELETE',
    });
  }

  async updateDNSRecordNow(
    id: number
  ): Promise<{ response: DNSUpdateResponse }> {
    return this.request(`/api/v1/dns/records/${id}/update`, {
      method: 'POST',
    });
  }

  async updateAllDNSRecords(): Promise<{ responses: DNSUpdateResponse[] }> {
    return this.request('/api/v1/dns/update-all', {
      method: 'POST',
    });
  }

  async getDNSStatus(): Promise<{ statuses: DNSStatus[] }> {
    return this.request('/api/v1/dns/status');
  }

  async getScheduledJobs(): Promise<{ active_jobs: Record<number, JobInfo> }> {
    return this.request('/api/v1/dns/scheduled-jobs');
  }

  async pauseScheduledJob(recordId: number): Promise<{ message: string }> {
    return this.request(`/api/v1/dns/scheduled-jobs/${recordId}/pause`, {
      method: 'POST'
    });
  }

  async resumeScheduledJob(recordId: number): Promise<{ message: string }> {
    return this.request(`/api/v1/dns/scheduled-jobs/${recordId}/resume`, {
      method: 'POST'
    });
  }

  async getPublicIP(): Promise<{ ip: string }> {
    return this.request('/api/v1/dns/public-ip');
  }

  // Certificate endpoints
  async getCertificates(): Promise<ApiResponse<Certificate[]>> {
    return this.request('/api/v1/certificates');
  }

  async getCertificate(id: number): Promise<ApiResponse<Certificate>> {
    return this.request(`/api/v1/certificates/${id}`);
  }

  async createCertificate(
    certificate: CertificateCreateRequest
  ): Promise<ApiResponse<Certificate>> {
    return this.request('/api/v1/certificates', {
      method: 'POST',
      body: JSON.stringify(certificate),
    });
  }

  async updateCertificate(
    id: number,
    certificate: CertificateUpdateRequest
  ): Promise<ApiResponse<Certificate>> {
    return this.request(`/api/v1/certificates/${id}`, {
      method: 'PUT',
      body: JSON.stringify(certificate),
    });
  }

  async deleteCertificate(id: number): Promise<{ message: string }> {
    return this.request(`/api/v1/certificates/${id}`, {
      method: 'DELETE',
    });
  }

  async getCertificateProxies(id: number): Promise<ApiResponse<Proxy[]>> {
    return this.request(`/api/v1/certificates/${id}/proxies`);
  }

  async renewCertificate(id: number): Promise<ApiResponse<Certificate>> {
    return this.request(`/api/v1/certificates/${id}/renew`, {
      method: 'POST',
    });
  }

  // Settings endpoints
  async getSettings(): Promise<Settings> {
    return this.request('/api/v1/settings');
  }

  async updateSettings(settings: SettingsUpdateRequest): Promise<Settings> {
    return this.request('/api/v1/settings', {
      method: 'PUT',
      body: JSON.stringify(settings),
    });
  }
}

export const apiService = new ApiService();
export default apiService;
