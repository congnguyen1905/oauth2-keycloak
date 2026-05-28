const GATEWAY_URL = process.env.GATEWAY_URL || 'http://localhost:8888';

import { LoginRequest, LoginResponse, SessionResponse } from './types';

export async function login(credentials: LoginRequest): Promise<LoginResponse> {
  const response = await fetch(`${GATEWAY_URL}/api/auth/login`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify(credentials),
  });

  if (!response.ok) {
    const error = await response.json();
    throw new Error(error.message || 'Login failed');
  }

  return response.json();
}

export async function validateSession(sessionId: string): Promise<SessionResponse> {
  const response = await fetch(`${GATEWAY_URL}/api/auth/validate`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({ session_id: sessionId }),
  });

  return response.json();
}

export async function logout(sessionId: string): Promise<void> {
  await fetch(`${GATEWAY_URL}/api/auth/logout`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({ session_id: sessionId }),
  });
}

// Session management with localStorage (only stores sessionId)
export const sessionManager = {
  getSessionId(): string | null {
    if (typeof window === 'undefined') return null;
    return localStorage.getItem('sessionId');
  },

  setSessionId(sessionId: string): void {
    if (typeof window === 'undefined') return;
    localStorage.setItem('sessionId', sessionId);
  },

  clearSession(): void {
    if (typeof window === 'undefined') return;
    localStorage.removeItem('sessionId');
  },
};
