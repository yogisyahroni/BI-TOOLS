export interface Slide {
    title: string;
    content: string;
    image?: string;
    layout: 'title' | 'bullet_points' | 'image_text' | 'chart';
    notes?: string;
}

export interface Story {
    id: string;
    title: string;
    description?: string;
    dashboard_id?: string;
    content: {
        slides: Slide[];
    };
    created_at: string;
    updated_at: string;
}

export interface CreateStoryRequest {
    dashboard_id: string;
    prompt: string;
}

export interface UpdateStoryRequest {
    title?: string;
    description?: string;
    content?: {
        slides: Slide[];
    };
}
