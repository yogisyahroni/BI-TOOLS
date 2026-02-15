'use client';

import React, { createContext, useContext, useState, useCallback, useEffect } from 'react';
import { useSession } from 'next-auth/react';
import { fetchWithAuth } from '@/lib/utils';

export interface Database {
  id: string;
  name: string;
  type: 'postgresql' | 'mysql' | 'mongodb' | 'snowflake' | 'bigquery';
  host: string;
  port: number;
  database: string;
  username: string;
  status: 'connected' | 'disconnected' | 'error';
  lastSync: string;
  createdAt: string;
}

interface DatabaseContextType {
  databases: Database[];
  selectedDatabase: Database | null;
  setSelectedDatabase: (db: Database | null) => void;
  addDatabase: (db: Database) => void;
  updateDatabase: (id: string, db: Partial<Database>) => void;
  deleteDatabase: (id: string) => void;
  testConnection: (db: Database) => Promise<boolean>;
  isLoading: boolean;
}

const DatabaseContext = createContext<DatabaseContextType | undefined>(undefined);

export function DatabaseProvider({ children }: { children: React.ReactNode }) {
  const { status } = useSession();
  const isAuthenticated = status === 'authenticated';

  const [databases, setDatabases] = useState<Database[]>([]);
  const [selectedDatabase, setSelectedDatabase] = useState<Database | null>(null);
  const [isLoading, setIsLoading] = useState(true);

  // Fetch databases from API - only when authenticated
  useEffect(() => {
    // Don't fetch if not authenticated or still loading
    if (status !== 'authenticated') {
      setIsLoading(false);
      return;
    }

    let isMounted = true;

    const fetchDatabases = async () => {
      // Don't set loading true on background refreshes to avoid UI flickering
      // only set on initial load if we have no data
      if (databases.length === 0) setIsLoading(true);

      try {
        const response = await fetchWithAuth('/api/go/connections');
        if (!isMounted) return;

        if (response.ok) {
          const result = await response.json();
          if (result.success && Array.isArray(result.data)) {
            const mappedDbs: Database[] = result.data.map((conn: any) => ({
              id: conn.id,
              name: conn.name,
              type: conn.type,
              host: conn.host || 'localhost',
              port: conn.port || 5432,
              database: conn.database || '',
              username: conn.username || '',
              status: conn.isActive ? 'connected' : 'disconnected',
              lastSync: conn.updatedAt || new Date().toISOString(),
              createdAt: conn.createdAt || new Date().toISOString(),
            }));
            setDatabases(mappedDbs);
            if (mappedDbs.length > 0 && !selectedDatabase) {
              setSelectedDatabase(mappedDbs[0]);
            }
          }
        } else if (response.status === 401) {
          // Stop polling if unauthorized
          console.warn('Database polling stopped due to 401 Unauthorized');
          return;
        }
      } catch (error) {
        console.error('Failed to fetch databases:', error);
      } finally {
        if (isMounted) setIsLoading(false);
      }
    };

    fetchDatabases();

    // Re-fetch every 30 seconds, but verify auth first
    const interval = setInterval(() => {
      if (status === 'authenticated') {
        fetchDatabases();
      }
    }, 30000);

    return () => {
      isMounted = false;
      clearInterval(interval);
    };
  }, [status, databases.length, selectedDatabase]); // Dependencies updated

  // Removed localStorage caching - API is source of truth

  const addDatabase = useCallback((db: Database) => {
    setDatabases((prev) => [...prev, db]);
  }, []);

  const updateDatabase = useCallback((id: string, updates: Partial<Database>) => {
    setDatabases((prev) =>
      prev.map((db) => (db.id === id ? { ...db, ...updates } : db))
    );
  }, []);

  const deleteDatabase = useCallback((id: string) => {
    setDatabases((prev) => prev.filter((db) => db.id !== id));
    if (selectedDatabase?.id === id) {
      setSelectedDatabase(null);
    }
  }, [selectedDatabase]);

  const testConnection = useCallback(async (db: Database): Promise<boolean> => {
    // Validate database object before making API call
    if (!db || !db.id) {
      return false;
    }

    setIsLoading(true);
    try {
      // Fixed: Use correct endpoint with ID parameter
      const response = await fetchWithAuth(`/api/go/connections/${db.id}/test`, {
        method: 'POST',
      });

      const isConnected = response.ok;
      updateDatabase(db.id, {
        status: isConnected ? 'connected' : 'error',
        lastSync: new Date().toISOString(),
      });

      return isConnected;
    } catch (error) {
      updateDatabase(db.id, { status: 'error' });
      return false;
    } finally {
      setIsLoading(false);
    }
  }, [updateDatabase]);

  return (
    <DatabaseContext.Provider
      value={{
        databases,
        selectedDatabase,
        setSelectedDatabase,
        addDatabase,
        updateDatabase,
        deleteDatabase,
        testConnection,
        isLoading,
      }}
    >
      {children}
    </DatabaseContext.Provider>
  );
}

export function useDatabase() {
  const context = useContext(DatabaseContext);
  if (!context) {
    throw new Error('useDatabase must be used within DatabaseProvider');
  }
  return context;
}
