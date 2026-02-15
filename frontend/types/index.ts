export interface Connection {
    id: string;
    name: string;
    type: string; // postgres, mysql, etc.
    database: string;
    host?: string;
    port?: number;
    userId: string;
    createdAt: string;
    updatedAt: string;
}

export interface Collection {
    id: string;
    name: string;
    description?: string;
    projectId?: string;
    createdAt: string;
    updatedAt: string;
}
