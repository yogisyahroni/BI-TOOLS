import type { 
    Comment, 
    CreateCommentRequest, 
    UpdateCommentRequest, 
    ResolveCommentRequest,
    CommentsResponse,
    CreateAnnotationRequest,
    CommentUser,
    CommentFilter
} from '@/types/comments';

const API_BASE = '/api/go';

export const commentApi = {
    // GET /api/go/comments?entityType=pipeline&entityId=xxx&parentId=root&isResolved=false
    list: async (
        entityType: string, 
        entityId: string, 
        filter?: Partial<CommentFilter>
    ): Promise<CommentsResponse> => {
        const params = new URLSearchParams({
            entityType,
            entityId,
        });

        if (filter?.parentId !== undefined) {
            params.set('parentId', filter.parentId);
        }
        if (filter?.isResolved !== undefined) {
            params.set('isResolved', String(filter.isResolved));
        }
        if (filter?.sortBy) {
            params.set('sortBy', filter.sortBy);
        }
        if (filter?.sortOrder) {
            params.set('sortOrder', filter.sortOrder);
        }
        if (filter?.limit) {
            params.set('limit', String(filter.limit));
        }
        if (filter?.offset) {
            params.set('offset', String(filter.offset));
        }

        const res = await fetch(`${API_BASE}/comments?${params.toString()}`);
        if (!res.ok) {
            const error = await res.json();
            throw new Error(error.error || 'Failed to fetch comments');
        }
        return res.json();
    },

    // GET /api/go/comments/:id
    get: async (id: string): Promise<Comment> => {
        const res = await fetch(`${API_BASE}/comments/${id}`);
        if (!res.ok) {
            const error = await res.json();
            throw new Error(error.error || 'Failed to fetch comment');
        }
        return res.json();
    },

    // POST /api/go/comments
    create: async (data: CreateCommentRequest): Promise<Comment> => {
        const res = await fetch(`${API_BASE}/comments`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(data),
        });
        if (!res.ok) {
            const error = await res.json();
            throw new Error(error.error || 'Failed to create comment');
        }
        return res.json();
    },

    // PUT /api/go/comments/:id
    update: async (id: string, data: UpdateCommentRequest): Promise<Comment> => {
        const res = await fetch(`${API_BASE}/comments/${id}`, {
            method: 'PUT',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(data),
        });
        if (!res.ok) {
            const error = await res.json();
            throw new Error(error.error || 'Failed to update comment');
        }
        return res.json();
    },

    // DELETE /api/go/comments/:id
    delete: async (id: string): Promise<void> => {
        const res = await fetch(`${API_BASE}/comments/${id}`, {
            method: 'DELETE',
        });
        if (!res.ok) {
            const error = await res.json();
            throw new Error(error.error || 'Failed to delete comment');
        }
    },

    // GET /api/go/comments/:id/replies
    getReplies: async (parentId: string): Promise<Comment[]> => {
        const res = await fetch(`${API_BASE}/comments/${parentId}/replies`);
        if (!res.ok) {
            const error = await res.json();
            throw new Error(error.error || 'Failed to fetch replies');
        }
        return res.json();
    },

    // PATCH /api/go/comments/:id/resolve
    resolve: async (id: string): Promise<Comment> => {
        const res = await fetch(`${API_BASE}/comments/${id}/resolve`, {
            method: 'PATCH',
            headers: { 'Content-Type': 'application/json' },
        });
        if (!res.ok) {
            const error = await res.json();
            throw new Error(error.error || 'Failed to resolve comment');
        }
        return res.json();
    },

    // PATCH /api/go/comments/:id/unresolve
    unresolve: async (id: string): Promise<Comment> => {
        const res = await fetch(`${API_BASE}/comments/${id}/unresolve`, {
            method: 'PATCH',
            headers: { 'Content-Type': 'application/json' },
        });
        if (!res.ok) {
            const error = await res.json();
            throw new Error(error.error || 'Failed to unresolve comment');
        }
        return res.json();
    },

    // GET /api/go/comments/mentions/search?q=query
    searchMentions: async (query: string, limit: number = 10): Promise<CommentUser[]> => {
        const params = new URLSearchParams({
            q: query,
            limit: String(limit),
        });

        const res = await fetch(`${API_BASE}/comments/mentions/search?${params.toString()}`);
        if (!res.ok) {
            const error = await res.json();
            throw new Error(error.error || 'Failed to search mentions');
        }
        return res.json();
    },

    // GET /api/go/comments/mentions/recent
    getRecentMentions: async (limit: number = 5): Promise<CommentUser[]> => {
        const params = new URLSearchParams({
            limit: String(limit),
        });

        const res = await fetch(`${API_BASE}/comments/mentions/recent?${params.toString()}`);
        if (!res.ok) {
            const error = await res.json();
            throw new Error(error.error || 'Failed to fetch recent mentions');
        }
        return res.json();
    },

    // Annotation endpoints

    // GET /api/go/comments/annotations/chart/:chartId
    getAnnotationsByChart: async (chartId: string): Promise<Comment[]> => {
        const res = await fetch(`${API_BASE}/comments/annotations/chart/${chartId}`);
        if (!res.ok) {
            const error = await res.json();
            throw new Error(error.error || 'Failed to fetch annotations');
        }
        return res.json();
    },

    // POST /api/go/comments/annotations
    createAnnotation: async (data: CreateAnnotationRequest): Promise<Comment> => {
        const res = await fetch(`${API_BASE}/comments/annotations`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(data),
        });
        if (!res.ok) {
            const error = await res.json();
            throw new Error(error.error || 'Failed to create annotation');
        }
        return res.json();
    },

    // PUT /api/go/comments/annotations/:id
    updateAnnotation: async (id: string, data: CreateAnnotationRequest): Promise<Comment> => {
        const res = await fetch(`${API_BASE}/comments/annotations/${id}`, {
            method: 'PUT',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(data),
        });
        if (!res.ok) {
            const error = await res.json();
            throw new Error(error.error || 'Failed to update annotation');
        }
        return res.json();
    },

    // DELETE /api/go/comments/annotations/:id
    deleteAnnotation: async (id: string): Promise<void> => {
        const res = await fetch(`${API_BASE}/comments/annotations/${id}`, {
            method: 'DELETE',
        });
        if (!res.ok) {
            const error = await res.json();
            throw new Error(error.error || 'Failed to delete annotation');
        }
    },
};

// React Query compatible API wrapper
export const commentApiClient = {
    list: commentApi.list,
    get: commentApi.get,
    create: commentApi.create,
    update: commentApi.update,
    delete: commentApi.delete,
    getReplies: commentApi.getReplies,
    resolve: commentApi.resolve,
    unresolve: commentApi.unresolve,
    searchMentions: commentApi.searchMentions,
    getRecentMentions: commentApi.getRecentMentions,
    getAnnotationsByChart: commentApi.getAnnotationsByChart,
    createAnnotation: commentApi.createAnnotation,
    updateAnnotation: commentApi.updateAnnotation,
    deleteAnnotation: commentApi.deleteAnnotation,
};
