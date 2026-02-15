import api from '@/lib/api';
import { Dashboard } from '@/lib/types';

export const nlApi = {
    parseFilter: async (text: string) => {
        const response = await api.post<any>('/nl/filter', { text });
        return response.data;
    },

    generateDashboard: async (text: string) => {
        const response = await api.post<Dashboard>('/nl/dashboard', { text });
        return response.data;
    },

    generateStory: async (data: any) => {
        const response = await api.post<{ story: string }>('/nl/story', { data });
        return response.data;
    }
};
