import api from '@/lib/api';

export interface TermColumnMapping {
    id: string;
    term_id: string;
    data_source_id?: string;
    table_name: string;
    column_name: string;
    metric_id?: string;
    created_at?: string;
}

export interface BusinessTerm {
    id: string;
    workspace_id: string;
    name: string;
    definition: string;
    synonyms: string[];
    owner_id: string;
    status: 'draft' | 'approved' | 'deprecated';
    tags: string[];
    created_at: string;
    updated_at: string;
    related_columns?: TermColumnMapping[];
}

export const glossaryApi = {
    listTerms: async () => {
        const response = await api.get<BusinessTerm[]>('/glossary/terms');
        return response.data;
    },

    createTerm: async (term: Partial<BusinessTerm>) => {
        const response = await api.post<BusinessTerm>('/glossary/terms', term);
        return response.data;
    },

    updateTerm: async (id: string, term: Partial<BusinessTerm>) => {
        const response = await api.put<BusinessTerm>(`/glossary/terms/${id}`, term);
        return response.data;
    },

    deleteTerm: async (id: string) => {
        await api.delete(`/glossary/terms/${id}`);
    },

    getTerm: async (id: string) => {
        const response = await api.get<BusinessTerm>(`/glossary/terms/${id}`);
        return response.data;
    }
};
