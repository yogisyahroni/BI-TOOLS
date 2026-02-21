import { create } from "zustand";
import { persist } from "zustand/middleware";
import { fetchWithAuth } from "@/lib/utils";

export interface Database {
  id: string;
  name: string;
  type: "postgresql" | "mysql" | "mongodb" | "snowflake" | "bigquery";
  host: string;
  port: number;
  database: string;
  username: string;
  status: "connected" | "disconnected" | "error";
  lastSync: string;
  createdAt: string;
}

interface DatabaseState {
  databases: Database[];
  selectedDatabase: Database | null;
  isLoading: boolean;
  setSelectedDatabase: (db: Database | null) => void;
  addDatabase: (db: Database) => void;
  updateDatabase: (id: string, db: Partial<Database>) => void;
  deleteDatabase: (id: string) => void;
  fetchDatabases: () => Promise<void>;
  testConnection: (db: Database) => Promise<boolean>;
  initialize: (isAuthenticated: boolean) => void;
}

export const useDatabaseStore = create<DatabaseState>()(
  persist(
    (set, get) => ({
      databases: [],
      selectedDatabase: null,
      isLoading: true,

      setSelectedDatabase: (db) => set({ selectedDatabase: db }),

      addDatabase: (db) => set((state) => ({ databases: [...state.databases, db] })),

      updateDatabase: (id, updates) =>
        set((state) => ({
          databases: state.databases.map((db) => (db.id === id ? { ...db, ...updates } : db)),
        })),

      deleteDatabase: (id) =>
        set((state) => {
          const newDatabases = state.databases.filter((db) => db.id !== id);
          return {
            databases: newDatabases,
            selectedDatabase: state.selectedDatabase?.id === id ? null : state.selectedDatabase,
          };
        }),

      fetchDatabases: async () => {
        const { databases } = get();
        // Optionally avoid UI flicker if already loaded
        if (databases.length === 0) set({ isLoading: true });

        try {
          const response = await fetchWithAuth("/api/go/connections");
          if (response.ok) {
            const result = await response.json();
            if (result.success && Array.isArray(result.data)) {
              // eslint-disable-next-line @typescript-eslint/no-explicit-any
              const mappedDbs: Database[] = result.data.map((conn: any) => ({
                id: conn.id,
                name: conn.name,
                type: conn.type,
                host: conn.host || "localhost",
                port: conn.port || 5432,
                database: conn.database || "",
                username: conn.username || "",
                status: conn.isActive ? "connected" : "disconnected",
                lastSync: conn.updatedAt || new Date().toISOString(),
                createdAt: conn.createdAt || new Date().toISOString(),
              }));
              set((state) => ({
                databases: mappedDbs,
                // Keep the previously selected database if it still exists in the new list, else pick the first
                selectedDatabase:
                  mappedDbs.find((d) => d.id === state.selectedDatabase?.id) ||
                  (mappedDbs.length > 0 ? mappedDbs[0] : null),
              }));
            }
          } else if (response.status === 401) {
            console.warn("Database polling stopped due to 401 Unauthorized");
          }
        } catch (error) {
          console.error("Failed to fetch databases:", error);
        } finally {
          set({ isLoading: false });
        }
      },

      testConnection: async (db: Database): Promise<boolean> => {
        if (!db || !db.id) return false;
        set({ isLoading: true });
        try {
          const response = await fetchWithAuth(`/api/go/connections/${db.id}/test`, {
            method: "POST",
          });
          const isConnected = response.ok;
          get().updateDatabase(db.id, {
            status: isConnected ? "connected" : "error",
            lastSync: new Date().toISOString(),
          });
          return isConnected;
        } catch (error) {
          get().updateDatabase(db.id, { status: "error" });
          return false;
        } finally {
          set({ isLoading: false });
        }
      },

      initialize: (isAuthenticated: boolean) => {
        if (!isAuthenticated) {
          set({ isLoading: false, databases: [], selectedDatabase: null });
        }
      },
    }),
    {
      name: "databaseStore",
      partialize: (state) => ({ selectedDatabase: state.selectedDatabase }),
    },
  ),
);
