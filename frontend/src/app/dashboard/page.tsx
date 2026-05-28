'use client';

import { useState, useEffect } from 'react';
import { useRouter } from 'next/navigation';
import { validateSession, logout, sessionManager } from '@/lib/api';
import { User } from '@/lib/types';

export default function DashboardPage() {
  const [user, setUser] = useState<User | null>(null);
  const [loading, setLoading] = useState(true);
  const router = useRouter();

  useEffect(() => {
    const checkSession = async () => {
      const sessionId = sessionManager.getSessionId();
      
      if (!sessionId) {
        router.push('/');
        return;
      }

      try {
        const response = await validateSession(sessionId);
        
        if (response.valid && response.session) {
          setUser({
            user_id: response.session.user_id,
            username: response.session.username,
            email: response.session.email,
            roles: response.session.roles,
          });
        } else {
          sessionManager.clearSession();
          router.push('/');
        }
      } catch (error) {
        sessionManager.clearSession();
        router.push('/');
      } finally {
        setLoading(false);
      }
    };

    checkSession();
  }, [router]);

  const handleLogout = async () => {
    const sessionId = sessionManager.getSessionId();
    if (sessionId) {
      await logout(sessionId);
    }
    sessionManager.clearSession();
    router.push('/');
  };

  if (loading) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-gray-100">
        <div className="text-xl">Loading...</div>
      </div>
    );
  }

  if (!user) {
    return null;
  }

  return (
    <div className="min-h-screen bg-gray-100">
      <nav className="bg-white shadow">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex justify-between h-16">
            <div className="flex items-center">
              <h1 className="text-xl font-bold">Dashboard</h1>
            </div>
            <div className="flex items-center space-x-4">
              <span className="text-gray-700">Welcome, {user.username}</span>
              <button
                onClick={handleLogout}
                className="bg-red-500 hover:bg-red-700 text-white px-4 py-2 rounded"
              >
                Logout
              </button>
            </div>
          </div>
        </div>
      </nav>

      <main className="max-w-7xl mx-auto py-6 sm:px-6 lg:px-8">
        <div className="bg-white overflow-hidden shadow-sm sm:rounded-lg">
          <div className="p-6">
            <h2 className="text-lg font-semibold mb-4">User Information</h2>
            <div className="space-y-2">
              <p><strong>User ID:</strong> {user.user_id}</p>
              <p><strong>Username:</strong> {user.username}</p>
              <p><strong>Email:</strong> {user.email}</p>
              <p><strong>Roles:</strong> {user.roles.join(', ')}</p>
            </div>
          </div>
        </div>
      </main>
    </div>
  );
}
