import { fetchWithAuth } from '@/lib/utils';
import { Connection } from '@/types/index';
import { SchemaTable } from '@/types/visual-query';


export type ConnectionSchema = SchemaTable[];

export const connectionsApi = {
    getSchema: async (connectionId: string): Promise<ConnectionSchema> => {
        const response = await fetchWithAuth(`/api/go/connections/${connectionId}/schema`);
        if (!response.ok) {
            throw new Error('Failed to fetch schema');
        }
        const json = await response.json();
        return json.data;
    },


    list: async (): Promise<Connection[]> => {
        const response = await fetchWithAuth('/api/go/connections');
        if (!response.ok) throw new Error('Failed to fetch connections');
        const json = await response.json();
        return json.data;
    }
};
