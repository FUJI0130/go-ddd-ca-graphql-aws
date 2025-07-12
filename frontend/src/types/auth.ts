// src/types/auth.ts

export interface AuthUser {
  id: string;
  username: string;
  role: string;
  createdAt: string;
  updatedAt: string;
  lastLoginAt?: string;
}

export interface AuthState {
  isAuthenticated: boolean;
  isLoading: boolean;
  user: AuthUser | null;
  error: string | null;
}

export interface LoginCredentials {
  username: string;
  password: string;
}

export interface AuthContextType extends AuthState {
  login: (credentials: LoginCredentials) => Promise<void>;
  logout: () => Promise<void>;
  resetAuthError: () => void;
  checkAuthStatus: () => Promise<void>;
}

// ユーザーロール定義
export enum UserRole {
  ADMIN = 'admin',
  MANAGER = 'manager', 
  TESTER = 'tester'
}

// 認証エラー型
export interface AuthError {
  message: string;
  code?: string;
}