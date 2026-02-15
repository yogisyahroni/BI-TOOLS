
import { SlideDeck } from '@/types/presentation';
import { api } from '@/lib/api';

export const presentationApi = {
    generate: async (dashboardId: string, prompt: string, providerId?: string) => {
        const response = await fetch('/api/ai/presentation', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({ dashboardId, prompt, providerId }),
        });

        if (!response.ok) {
            throw new Error('Failed to generate presentation');
        }

        return response.json();
    },
};
