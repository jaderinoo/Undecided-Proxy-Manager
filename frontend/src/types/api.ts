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
  ws_enabled?: boolean;
  ssl_path?: string;
  status: 'active' | 'inactive' | 'error';
  created_at: string;
  updated_at: string;
  // Computed fields for UI
  connected_containers?: Container[];
  certificate?: Certificate;
}

export interface ProxyCreateRequest {
  name: string;
  domain: string;
  target_url: string;
  ssl_enabled: boolean;
  ws_enabled?: boolean;
}

export interface ProxyUpdateRequest {
  id?: number;
  name?: string;
  domain?: string;
  target_url?: string;
  ssl_enabled?: boolean;
  ws_enabled?: boolean;
}

export interface ProxyResponse {
  data: Proxy;
  ssl_status?: 'certificate_generated' | 'certificate_failed';
  ssl_message?: string;
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
  issuer?: string;
  created_at: string;
  updated_at: string;
}

export interface CertificateCreateRequest {
  domain: string;
  cert_path: string;
  key_path: string;
  expires_at: string;
}

export interface CertificateUpdateRequest {
  domain?: string;
  cert_path?: string;
  key_path?: string;
  expires_at?: string;
  is_valid?: boolean;
}

// Container Types
export interface Container {
  id: string;
  name: string;
  image: string;
  image_id: string;
  status: string;
  state: string;
  created: string;
  started_at?: string;
  finished_at?: string;
  ports: PortMapping[];
  labels: Record<string, string>;
  command: string;
  size_rw: number;
  size_root_fs: number;
  network_mode: string;
  mounts: Mount[];
  // Computed fields for UI
  connected_proxies?: Proxy[];
}

export interface PortMapping {
  ip: string;
  private_port: number;
  public_port: number;
  type: string;
}

export interface Mount {
  type: string;
  source: string;
  destination: string;
  mode: string;
  rw: boolean;
  propagation: string;
}

export interface ContainerListResponse {
  containers: Container[];
  count: number;
}

// DNS Types
export interface DNSConfig {
  id: number;
  provider: 'namecheap' | 'static';
  domain: string;
  username?: string; // Optional for static DNS
  password?: string; // Optional since it's hidden in API responses
  is_active: boolean;
  last_update?: string;
  last_ip?: string;
  created_at: string;
  updated_at: string;
}

export interface DNSConfigCreateRequest {
  provider: string;
  domain: string;
  username?: string;
  password?: string;
  is_active?: boolean;
}

export interface DNSConfigUpdateRequest {
  provider?: string;
  domain?: string;
  username?: string;
  password?: string;
  is_active?: boolean;
}

export interface DNSRecord {
  id: number;
  config_id: number;
  host: string;
  current_ip?: string;
  allowed_ip_ranges?: string;
  dynamic_dns_refresh_rate?: number;
  include_backend: boolean;
  backend_url?: string;
  last_update?: string;
  is_active: boolean;
  created_at: string;
  updated_at: string;
}

export interface DNSRecordCreateRequest {
  config_id: number;
  host: string;
  allowed_ip_ranges?: string;
  dynamic_dns_refresh_rate?: number;
  include_backend: boolean;
  backend_url?: string;
}

export interface DNSRecordUpdateRequest {
  host?: string;
  allowed_ip_ranges?: string;
  dynamic_dns_refresh_rate?: number;
  include_backend?: boolean;
  backend_url?: string;
  is_active?: boolean;
}

export interface DNSUpdateResponse {
  success: boolean;
  message: string;
  new_ip?: string;
  updated_at?: string;
}

export interface DNSStatus {
  config_id: number;
  domain: string;
  provider: string;
  is_active: boolean;
  last_update?: string;
  last_ip?: string;
  record_count: number;
  records?: DNSRecord[];
}

export interface JobInfo {
  interval: number; // Duration in nanoseconds
  last_started: string; // ISO timestamp
  next_update: string; // ISO timestamp
  is_paused: boolean;
  paused_at?: string; // ISO timestamp
}

// Settings Types
export interface Settings {
  core_settings: CoreSettings;
  ui_settings: UISettings;
}

export interface CoreSettings {
  database_path: string;
  environment: string;
  backend_port: string;
  admin_password: string; // Masked for security
  jwt_secret: string; // Masked for security
  letsencrypt_email: string;
  letsencrypt_webroot: string;
  letsencrypt_cert_path: string;
  dns_check_interval: string;
  public_ip_service: string;
}

export interface UISettings {
  display_name: string;
  theme: string;
  language: string;
}

export interface SettingsUpdateRequest {
  display_name?: string;
  theme?: string;
  language?: string;
}
