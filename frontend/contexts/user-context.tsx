'use client';

import { createContext, useContext, useState, useEffect, type ReactNode } from 'react';
import { type User } from '@/lib/types';

interface UserContextType {
  user: User | null;
  isLoading: boolean;
  isAuthenticated: boolean;
  login: (email: string, password: string) => Promise<void>;
  logout: () => Promise<void>;
  signup: (email: string, name: string, password: string) => Promise<void>;
}

const UserContext = createContext<UserContextType | undefined>(undefined);

export function UserProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<User | null>(null);
  const [isLoading, setIsLoading] = useState(true);

  // Initialize loading state - immediately set to false since we use NextAuth session
  useEffect(() => {
    setIsLoading(false);
  }, []);

  const login = async (email: string, password: string) => {
    // Dev mode: auth handled by NextAuth
  };

  const logout = async () => {
    // Dev mode: auth handled by NextAuth
  };

  const signup = async (email: string, name: string, password: string) => {
    // Dev mode: auth handled by NextAuth
  };

  return (
    <UserContext.Provider
      value={{
        user,
        isLoading,
        isAuthenticated: user !== null,
        login,
        logout,
        signup,
      }}
    >
      {children}
    </UserContext.Provider>
  );
}

export function useUser() {
  const context = useContext(UserContext);
  if (context === undefined) {
    throw new Error('useUser must be used within a UserProvider');
  }
  return context;
}
