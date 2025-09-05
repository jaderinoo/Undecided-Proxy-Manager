// API Response Types
export interface ApiResponse<T> {
  data: T;
  count?: number;
}

// Proxy Types
export interface Proxy {
  id: number;
  name: string;
  domain: string;
  target_url: string;
  ssl_enabled: boolean;
  ssl_path?: string;
  status: 'active' | 'inactive' | 'error';
  created_at: string;
  updated_at: string;
}

export interface ProxyCreateRequest {
  name: string;
  domain: string;
  target_url: string;
  ssl_enabled: boolean;
}

export interface ProxyUpdateRequest {
  name?: string;
  domain?: string;
  target_url?: string;
  ssl_enabled?: boolean;
}

// User Types
export interface User {
  id: number;
  username: string;
  email: string;
  is_active: boolean;
  created_at: string;
  updated_at: string;
}

export interface UserCreateRequest {
  username: string;
  email: string;
  password: string;
}

export interface UserLoginRequest {
  password: string;
}

export interface AuthResponse {
  token: string;
  user: User;
}

// Certificate Types
export interface Certificate {
  id: number;
  domain: string;
  cert_path: string;
  key_path: string;
  expires_at: string;
  is_valid: boolean;
  created_at: string;
  updated_at: string;
}
