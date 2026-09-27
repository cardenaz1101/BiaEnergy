import axios, { AxiosError } from 'axios';
import { session } from '../auth/session';

export const http = axios.create({
  baseURL: import.meta.env.VITE_API_URL || '/api',
  headers: { 'Content-Type': 'application/json' },
});

type UnauthorizedListener = () => void;
const unauthorizedListeners = new Set<UnauthorizedListener>();

export function onUnauthorized(listener: UnauthorizedListener) {
  unauthorizedListeners.add(listener);
  return () => {
    unauthorizedListeners.delete(listener);
  };
}

http.interceptors.request.use((config) => {
  const token = session.getToken();
  if (token) config.headers.Authorization = `Bearer ${token}`;
  return config;
});

http.interceptors.response.use(
  (response) => response,
  (error: AxiosError) => {
    const isLogin = error.config?.url?.endsWith('/auth/login');
    if (error.response?.status === 401 && !isLogin) {
      session.clear();
      unauthorizedListeners.forEach((listener) => listener());
    }
    return Promise.reject(error);
  },
);

export function errorMessage(error: unknown): string {
  if (error instanceof AxiosError) {
    const apiMessage = (error.response?.data as { error?: string } | undefined)?.error;
    if (apiMessage) return apiMessage;
    if (!error.response) return 'No se pudo conectar con la API. ¿Está corriendo el backend?';
  }
  return error instanceof Error ? error.message : 'Error inesperado';
}
