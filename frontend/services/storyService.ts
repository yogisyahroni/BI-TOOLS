import { Story, CreateStoryRequest, UpdateStoryRequest } from '@/types/story';

const API_BASE_URL = '/api';

export const storyService = {
    async getStories(): Promise<Story[]> {
        const response = await fetch(`${API_BASE_URL}/stories`);
        if (!response.ok) {
            throw new Error('Failed to fetch stories');
        }
        return response.json();
    },

    async getStory(id: string): Promise<Story> {
        const response = await fetch(`${API_BASE_URL}/stories/${id}`);
        if (!response.ok) {
            throw new Error('Failed to fetch story');
        }
        return response.json();
    },

    async createStory(data: CreateStoryRequest): Promise<Story> {
        const response = await fetch(`${API_BASE_URL}/stories`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify(data),
        });

        if (!response.ok) {
            const errorData = await response.json();
            throw new Error(errorData.error || 'Failed to create story');
        }

        return response.json();
    },

    async updateStory(id: string, data: UpdateStoryRequest): Promise<Story> {
        const response = await fetch(`${API_BASE_URL}/stories/${id}`, {
            method: 'PUT',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify(data),
        });

        if (!response.ok) {
            throw new Error('Failed to update story');
        }
        return response.json();
    },

    async deleteStory(id: string): Promise<void> {
        const response = await fetch(`${API_BASE_URL}/stories/${id}`, {
            method: 'DELETE',
        });

        if (!response.ok) {
            throw new Error('Failed to delete story');
        }
    },

    async exportPPTX(id: string, title: string): Promise<void> {
        const response = await fetch(`${API_BASE_URL}/stories/${id}/export`, {
            method: 'POST',
        });

        if (!response.ok) {
            throw new Error('Failed to export PPTX');
        }

        const blob = await response.blob();
        const url = window.URL.createObjectURL(blob);
        const a = document.createElement('a');
        a.href = url;
        a.download = `${title.replace(/[^a-z0-9]/gi, '_').toLowerCase()}.pptx`;
        document.body.appendChild(a);
        a.click();
        window.URL.revokeObjectURL(url);
        document.body.removeChild(a);
    },
};
