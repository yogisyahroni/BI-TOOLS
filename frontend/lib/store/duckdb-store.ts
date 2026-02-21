import { create } from "zustand";
import * as duckdb from "@duckdb/duckdb-wasm";

interface DuckDBState {
  db: duckdb.AsyncDuckDB | null;
  connection: duckdb.AsyncDuckDBConnection | null;
  isInitialized: boolean;
  initError: string | null;

  // Actions
  initialize: () => Promise<void>;
  ingestArrowBuffer: (tableName: string, buffer: Uint8Array) => Promise<void>;
  query: (sql: string) => Promise<any[]>;
}

export const useDuckDBStore = create<DuckDBState>((set, get) => ({
  db: null,
  connection: null,
  isInitialized: false,
  initError: null,

  initialize: async () => {
    if (get().isInitialized || get().db) return;

    try {
      // 1. Instantiate the web worker bundles for DuckDB
      const JSDELIVR_BUNDLES = duckdb.getJsDelivrBundles();

      // Configure Wasm URL
      const bundle = await duckdb.selectBundle(JSDELIVR_BUNDLES);

      const worker_url = URL.createObjectURL(
        new Blob([`importScripts("${bundle.mainWorker!}");`], { type: "text/javascript" }),
      );

      const worker = new Worker(worker_url);
      const logger = new duckdb.ConsoleLogger();
      const db = new duckdb.AsyncDuckDB(logger, worker);

      await db.instantiate(bundle.mainModule, bundle.pthreadWorker);

      const conn = await db.connect();

      set({ db, connection: conn, isInitialized: true, initError: null });
      console.log("🦆 DuckDB-Wasm initialized successfully");
    } catch (error: any) {
      console.error("Failed to initialize DuckDB-Wasm:", error);
      set({ initError: error.message, isInitialized: false });
    }
  },

  ingestArrowBuffer: async (tableName: string, buffer: Uint8Array) => {
    const { db, connection, isInitialized } = get();
    if (!isInitialized || !db || !connection) {
      throw new Error("DuckDB is not initialized yet");
    }

    try {
      console.log(`🦆 Ingesting Arrow IPC stream into local table: ${tableName}`);
      // 1. Mount the raw buffer into DuckDB's virtual filesystem
      await db.registerFileBuffer(`tmp_${tableName}.arrow`, buffer);

      // 2. Load the Arrow file into a physical table
      await connection.query(`
        CREATE TABLE IF NOT EXISTS "${tableName}" AS SELECT * FROM scan_arrow_ipc('tmp_${tableName}.arrow');
        -- If table exists, we might want to insert logic or just drop/replace
        -- For robust ad-hoc execution, recreate it:
        DROP TABLE IF EXISTS "${tableName}";
        CREATE TABLE "${tableName}" AS SELECT * FROM scan_arrow_ipc('tmp_${tableName}.arrow');
      `);

      console.log(`🦆 Successfully loaded ${tableName} from Arrow IPC stream.`);
    } catch (error: any) {
      console.error("Failed to ingest Arrow buffer:", error);
      throw error;
    }
  },

  query: async (sql: string) => {
    const { connection, isInitialized } = get();
    if (!isInitialized || !connection) {
      throw new Error("DuckDB is not initialized yet");
    }

    const result = await connection.query(sql);
    // Arrow Table to JavaScript Array of Objects
    return result.toArray().map((row) => row.toJSON());
  },
}));
