export interface User {
  user_id: string;
  username: string;
  email: string;
  roles: string[];
}

export interface SessionResponse {
  valid: boolean;
  session?: User & {
    session_id: string;
    expires_at: string;
  };
  error?: string;
}

export interface LoginRequest {
  username: string;
  password: string;
}

export interface LoginResponse {
  session_id: string;
  user: {
    username: string;
    email: string;
    roles: string[];
  };
}
