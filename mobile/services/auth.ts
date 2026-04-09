import { AxiosError } from 'axios';
import { api, handleApiError } from './api';
import { useAuthStore } from '../stores/authStore';
import { User, ApiResponse } from '../types';

interface LoginCredentials {
  email: string;
  password: string;
}

interface RegisterData {
  email: string;
  password: string;
  display_name: string;
}

interface AuthTokens {
  access_token: string;
  refresh_token: string;
  expires_in: number;
}

interface AuthResponse {
  user: User;
  tokens: AuthTokens;
}

/**
 * Login with email and password
 */
export async function login(credentials: LoginCredentials): Promise<AuthResponse> {
  try {
    const response = await api.post<ApiResponse<AuthResponse>>('/auth/login', credentials);
    const { user, tokens } = response.data.data;

    useAuthStore.getState().login(user, tokens.access_token, tokens.refresh_token);

    return response.data.data;
  } catch (error) {
    const apiError = handleApiError(error as AxiosError);
    throw new Error(apiError.message);
  }
}

/**
 * Register a new user
 */
export async function register(data: RegisterData): Promise<AuthResponse> {
  try {
    const response = await api.post<ApiResponse<AuthResponse>>('/auth/register', data);
    const { user, tokens } = response.data.data;

    useAuthStore.getState().login(user, tokens.access_token, tokens.refresh_token);

    return response.data.data;
  } catch (error) {
    const apiError = handleApiError(error as AxiosError);
    throw new Error(apiError.message);
  }
}

/**
 * Refresh access token using refresh token
 */
export async function refreshToken(refreshTokenValue: string): Promise<AuthTokens> {
  try {
    const response = await api.post<ApiResponse<AuthTokens>>('/auth/refresh', {
      refresh_token: refreshTokenValue,
    });

    const { access_token, refresh_token } = response.data.data;
    useAuthStore.getState().setTokens(access_token, refresh_token);

    return response.data.data;
  } catch (error) {
    const apiError = handleApiError(error as AxiosError);
    throw new Error(apiError.message);
  }
}

/**
 * Logout user and invalidate tokens
 */
export async function logout(): Promise<void> {
  const { refreshToken: currentRefreshToken, clearAuth } = useAuthStore.getState();

  try {
    if (currentRefreshToken) {
      await api.post('/auth/logout', {
        refresh_token: currentRefreshToken,
      });
    }
  } catch (error) {
    // Ignore errors during logout, still clear auth
  } finally {
    clearAuth();
  }
}

/**
 * Get current authenticated user
 */
export async function getCurrentUser(): Promise<User> {
  try {
    const response = await api.get<ApiResponse<User>>('/auth/me');
    useAuthStore.getState().setUser(response.data.data);
    return response.data.data;
  } catch (error) {
    const apiError = handleApiError(error as AxiosError);
    throw new Error(apiError.message);
  }
}
