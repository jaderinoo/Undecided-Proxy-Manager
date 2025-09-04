import type { 
  ApiResponse, 
  Proxy, 
  ProxyCreateRequest, 
  ProxyUpdateRequest,
  User,
  UserCreateRequest,
  UserLoginRequest,
  AuthResponse
} from '../types/api'

const API_BASE_URL = import.meta.env.VITE_API_URL || 'http://localhost:6081'

class ApiService {
  private baseUrl: string

  constructor(baseUrl: string = API_BASE_URL) {
    this.baseUrl = baseUrl
  }

  private async request<T>(endpoint: string, options: RequestInit = {}): Promise<T> {
    const url = `${this.baseUrl}${endpoint}`
    const response = await fetch(url, {
      headers: {
        'Content-Type': 'application/json',
        ...options.headers,
      },
      ...options,
    })

    if (!response.ok) {
      throw new Error(`API Error: ${response.status} ${response.statusText}`)
    }

    return response.json()
  }

  // Health Check
  async getHealth(): Promise<{ status: string; message: string }> {
    return this.request('/health')
  }

  // Proxy endpoints
  async getProxies(): Promise<ApiResponse<Proxy[]>> {
    return this.request('/api/v1/proxies')
  }

  async getProxy(id: number): Promise<ApiResponse<Proxy>> {
    return this.request(`/api/v1/proxies/${id}`)
  }

  async createProxy(proxy: ProxyCreateRequest): Promise<ApiResponse<Proxy>> {
    return this.request('/api/v1/proxies', {
      method: 'POST',
      body: JSON.stringify(proxy),
    })
  }

  async updateProxy(id: number, proxy: ProxyUpdateRequest): Promise<ApiResponse<Proxy>> {
    return this.request(`/api/v1/proxies/${id}`, {
      method: 'PUT',
      body: JSON.stringify(proxy),
    })
  }

  async deleteProxy(id: number): Promise<{ message: string }> {
    return this.request(`/api/v1/proxies/${id}`, {
      method: 'DELETE',
    })
  }

  // User endpoints
  async getUsers(): Promise<ApiResponse<User[]>> {
    return this.request('/api/v1/users')
  }

  async getUser(id: number): Promise<ApiResponse<User>> {
    return this.request(`/api/v1/users/${id}`)
  }

  async createUser(user: UserCreateRequest): Promise<ApiResponse<User>> {
    return this.request('/api/v1/users', {
      method: 'POST',
      body: JSON.stringify(user),
    })
  }

  async updateUser(id: number, user: Partial<User>): Promise<ApiResponse<User>> {
    return this.request(`/api/v1/users/${id}`, {
      method: 'PUT',
      body: JSON.stringify(user),
    })
  }

  async deleteUser(id: number): Promise<{ message: string }> {
    return this.request(`/api/v1/users/${id}`, {
      method: 'DELETE',
    })
  }

  // Auth endpoints
  async login(credentials: UserLoginRequest): Promise<ApiResponse<AuthResponse>> {
    return this.request('/api/v1/auth/login', {
      method: 'POST',
      body: JSON.stringify(credentials),
    })
  }

  async register(userData: UserCreateRequest): Promise<ApiResponse<AuthResponse>> {
    return this.request('/api/v1/auth/register', {
      method: 'POST',
      body: JSON.stringify(userData),
    })
  }
}

export const apiService = new ApiService()
export default apiService
