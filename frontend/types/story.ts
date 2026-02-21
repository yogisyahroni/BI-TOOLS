export interface DataBinding {
  dashboard_id?: string;
  card_id?: string;
}

export interface Slide {
  title: string;
  content: string;
  image?: string;
  layout: "title" | "bullet_points" | "image_text" | "chart";
  notes?: string;
  data_binding?: DataBinding;
  query_result?: any;
  visualization_config?: any;
  query_error?: string;
}

export interface Story {
  id: string;
  title: string;
  description?: string;
  dashboard_id?: string;
  is_public?: boolean;
  share_token?: string;
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

export interface CreateManualStoryRequest {
  title: string;
  description?: string;
  slides?: Slide[];
}

export interface UpdateStoryRequest {
  title?: string;
  description?: string;
  content?: {
    slides: Slide[];
  };
}
