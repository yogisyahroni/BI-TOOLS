import { create } from "zustand";
import { type User } from "@/lib/types";

interface UserState {
  user: User | null;
  isLoading: boolean;
  isAuthenticated: boolean;

  // Actions
  setUser: (user: User | null) => void;
  setIsLoading: (isLoading: boolean) => void;
  login: (email: string, password: string) => Promise<void>;
  logout: () => Promise<void>;
  signup: (email: string, name: string, password: string) => Promise<void>;
}

export const useUserStore = create<UserState>((set) => ({
  user: null,
  isLoading: false, // Default false since NextAuth handles session
  isAuthenticated: false,

  setUser: (user) => set({ user, isAuthenticated: user !== null }),
  setIsLoading: (isLoading) => set({ isLoading }),

  login: async (email, password) => {
    // Dev mode: auth handled by NextAuth
  },

  logout: async () => {
    // Dev mode: auth handled by NextAuth
    set({ user: null, isAuthenticated: false });
  },

  signup: async (email, name, password) => {
    // Dev mode: auth handled by NextAuth
  },
}));
