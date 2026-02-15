export interface Slide {
    title: string;
    layout: 'bullet_points' | 'chart_focus' | 'title_and_body' | 'title_only';
    bullet_points?: string[];
    speaker_notes?: string;
    chart_id?: string;
}

export interface SlideDeck {
    title: string;
    description: string;
    slides: Slide[];
}

export interface GeneratePresentationRequest {
    dashboardId: string;
    prompt: string;
}
