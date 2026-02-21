import {
  BaseConnector,
  ConnectionConfig,
  type SchemaInfo,
  type QueryResult,
} from "./base-connector";
import * as _fs from "fs";

/**
 * Parquet Connector Implementation
 * Apache Parquet columnar storage format
 *
 * Uses parquetjs library for reading
 */
export class ParquetConnector extends BaseConnector {
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  private data: any[] = [];
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  private schema: any = null;
  private tableName: string = "parquet_data";

  async testConnection(): Promise<{ success: boolean; error?: string }> {
    try {
      const filePath = this.config.filePath;

      if (!filePath) {
        return {
          success: false,
          error: "File path is required (URL not supported for Parquet)",
        };
      }

      // Import parquetjs library (lazy load)
      // @ts-expect-error
      const parquetjs = await import("parquetjs");

      // Open parquet file
      const reader = await parquetjs.ParquetReader.openFile(filePath);

      // Get schema
      this.schema = reader.getSchema();

      // Read all rows
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      const cursor = reader.getCursor();
      let record = null;
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      const rows: any[] = [];

      while ((record = await cursor.next())) {
        rows.push(record);
      }

      await reader.close();

      this.data = rows;

      if (this.data.length === 0) {
        return {
          success: false,
          error: "Parquet file is empty",
        };
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
      }

      return { success: true };
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
    } catch (error: any) {
      return {
        success: false,
        error: `Parquet file read failed: ${error.message}`,
      };
    }
  }

  async fetchSchema(): Promise<SchemaInfo> {
    if (!this.schema || this.data.length === 0) {
      throw new Error("Not connected. Call testConnection() first.");
    }

    const schemaInfo: SchemaInfo = { tables: [] };
    // eslint-disable-next-line @typescript-eslint/no-explicit-any

    // Convert Parquet schema to our format
    const fields = this.schema.fields || this.schema.fieldList || [];

    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    const columns = fields.map((field: any) => {
      // Map Parquet types to SQL types
      const parquetType = field.primitiveType || field.type || "BYTE_ARRAY";
      let sqlType = "TEXT";

      switch (parquetType) {
        case "INT32":
        case "INT64":
          sqlType = "INTEGER";
          break;
        case "FLOAT":
        case "DOUBLE":
          sqlType = "REAL";
          break;
        case "BOOLEAN":
          sqlType = "BOOLEAN";
          break;
        case "INT96": // Timestamp
          sqlType = "TIMESTAMP";
          break;
        default:
          sqlType = "TEXT";
      }

      return {
        name: field.name,
        type: sqlType,
        nullable: field.optional !== false,
        isPrimary: false,
        isForeign: false,
        description: field.description || undefined,
      };
    });

    schemaInfo.tables.push({
      name: this.tableName,
      schema: "parquet",
      rowCount: this.data.length,
      columns,
    });

    return schemaInfo;
  }

  async executeQuery(sql: string): Promise<QueryResult> {
    const startTime = Date.now();

    if (this.data.length === 0) {
      throw new Error("Not connected. Call testConnection() first.");
    }

    // Use alasql for in-memory SQL execution
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    const alasql = await import("alasql");

    // Register data as alasql table
    alasql.default.tables[this.tableName] = { data: this.data };

    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    const result = alasql.default(sql) as any[];
    const executionTime = Date.now() - startTime;

    const columns = result.length > 0 ? Object.keys(result[0]) : [];

    return {
      columns,
      rows: result,
      rowCount: result.length,
      executionTime,
    };
  }
  // eslint-disable-next-line @typescript-eslint/no-explicit-any

  async disconnect(): Promise<void> {
    this.data = [];
    this.schema = null;
  }

  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  async *extractData(): AsyncGenerator<any[]> {
    if (this.data.length === 0) {
      // Attempt to load if not loaded (optional, depending on flow)
      const testResult = await this.testConnection();
      if (!testResult.success) {
        throw new Error(testResult.error || "Failed to load parquet data");
      }
    }

    // Yield all data as a single batch for now (Parquet is file-based)
    if (this.data.length > 0) {
      yield this.data;
    }
  }

  validateConfig(): { valid: boolean; errors: string[] } {
    const errors: string[] = [];

    if (!this.config.filePath) {
      errors.push("File path is required for Parquet connector (URL not supported)");
    }

    const filePath = this.config.filePath || "";
    if (!filePath.toLowerCase().endsWith(".parquet")) {
      errors.push("File must have .parquet extension");
    }

    return {
      valid: errors.length === 0,
      errors: [...super.validateConfig().errors, ...errors],
    };
  }
}
