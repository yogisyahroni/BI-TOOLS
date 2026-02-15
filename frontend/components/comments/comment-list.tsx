'use client';

import { useState, useMemo } from 'react';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { ScrollArea } from '@/components/ui/scroll-area';
import { Tabs, TabsList, TabsTrigger } from '@/components/ui/tabs';
import {
    Select,
    SelectContent,
    SelectItem,
    SelectTrigger,
    SelectValue,
} from '@/components/ui/select';
import { 
    MessageSquare, 
    CheckCircle2, 
    Circle,
    ArrowUpDown,
    MessageCircle,
    Filter
} from 'lucide-react';
import { CommentThread } from '../collaboration/comment-thread';
import type { 
    Comment, 
    CommentEntityType, 
    CommentSortOption, 
    CommentFilterOption,
    CreateCommentRequest 
} from '@/types/comments';

interface CommentListProps {
    entityType: CommentEntityType;
    entityId: string;
    comments: Comment[];
    currentUserId: string;
    isLoading?: boolean;
    onCreateComment: (data: CreateCommentRequest) => Promise<void>;
    onResolveComment: (commentId: string, isResolved: boolean) => Promise<void>;
    onDeleteComment: (commentId: string) => Promise<void>;
    onEditComment: (commentId: string, content: string) => Promise<void>;
    onReply?: (parentId: string) => Promise<void>;
    className?: string;
}

export function CommentList({
    entityType,
    entityId,
    comments,
    currentUserId,
    isLoading = false,
    onCreateComment,
    onResolveComment,
    onDeleteComment,
    onEditComment,
    onReply,
    className = '',
}: CommentListProps) {
    const [sortBy, setSortBy] = useState<CommentSortOption>('newest');
    const [filterBy, setFilterBy] = useState<CommentFilterOption>('all');

    // Calculate stats
    const stats = useMemo(() => {
        const total = comments.filter(c => !c.parentId).length;
        const resolved = comments.filter(c => !c.parentId && c.isResolved).length;
        const unresolved = total - resolved;
        const totalReplies = comments.filter(c => c.parentId).length;

        return { total, resolved, unresolved, totalReplies };
    }, [comments]);

    // Filter and sort comments
    const filteredAndSortedComments = useMemo(() => {
        let filtered = [...comments];

        // Filter by status
        if (filterBy === 'resolved') {
            filtered = filtered.filter(c => c.isResolved);
        } else if (filterBy === 'unresolved') {
            filtered = filtered.filter(c => !c.isResolved);
        }

        // Sort
        filtered.sort((a, b) => {
            let comparison = 0;
            
            switch (sortBy) {
                case 'newest':
                    comparison = new Date(b.createdAt).getTime() - new Date(a.createdAt).getTime();
                    break;
                case 'oldest':
                    comparison = new Date(a.createdAt).getTime() - new Date(b.createdAt).getTime();
                    break;
                case 'popular':
                    // Sort by reply count as a proxy for popularity
                    const aReplies = a.replies?.length || 0;
                    const bReplies = b.replies?.length || 0;
                    comparison = bReplies - aReplies;
                    break;
            }
            
            return comparison;
        });

        return filtered;
    }, [comments, sortBy, filterBy]);

    if (isLoading) {
        return (
            <div className={`flex flex-col h-full ${className}`}>
                <div className="flex items-center justify-center h-64">
                    <div className="flex items-center gap-2 text-muted-foreground">
                        <div className="w-5 h-5 border-2 border-primary border-t-transparent rounded-full animate-spin" />
                        <span>Loading comments...</span>
                    </div>
                </div>
            </div>
        );
    }

    return (
        <div className={`flex flex-col h-full ${className}`}>
            {/* Header Stats */}
            <div className="p-4 border-b bg-muted/30">
                <div className="flex items-center justify-between mb-4">
                    <h2 className="text-lg font-semibold flex items-center gap-2">
                        <MessageSquare className="w-5 h-5" />
                        Comments & Discussion
                    </h2>
                </div>

                {/* Stats Cards */}
                <div className="grid grid-cols-3 gap-3">
                    <div className="p-3 bg-background rounded-lg border">
                        <div className="flex items-center gap-2 text-muted-foreground mb-1">
                            <MessageCircle className="w-4 h-4" />
                            <span className="text-xs">Total Threads</span>
                        </div>
                        <span className="text-2xl font-bold">{stats.total}</span>
                    </div>
                    
                    <div className="p-3 bg-background rounded-lg border">
                        <div className="flex items-center gap-2 text-yellow-600 mb-1">
                            <Circle className="w-4 h-4" />
                            <span className="text-xs">Unresolved</span>
                        </div>
                        <span className="text-2xl font-bold">{stats.unresolved}</span>
                    </div>
                    
                    <div className="p-3 bg-background rounded-lg border">
                        <div className="flex items-center gap-2 text-green-600 mb-1">
                            <CheckCircle2 className="w-4 h-4" />
                            <span className="text-xs">Resolved</span>
                        </div>
                        <span className="text-2xl font-bold">{stats.resolved}</span>
                    </div>
                </div>

                {/* Filters and Sort */}
                <div className="flex items-center justify-between mt-4 gap-3">
                    <Tabs 
                        value={filterBy} 
                        onValueChange={(v) => setFilterBy(v as CommentFilterOption)}
                        className="flex-1"
                    >
                        <TabsList className="grid w-full grid-cols-3">
                            <TabsTrigger value="all" className="text-xs">
                                <Filter className="w-3 h-3 mr-1" />
                                All
                            </TabsTrigger>
                            <TabsTrigger value="unresolved" className="text-xs">
                                <Circle className="w-3 h-3 mr-1" />
                                Open
                                {stats.unresolved > 0 && (
                                    <Badge variant="secondary" className="ml-1 text-[10px] px-1">
                                        {stats.unresolved}
                                    </Badge>
                                )}
                            </TabsTrigger>
                            <TabsTrigger value="resolved" className="text-xs">
                                <CheckCircle2 className="w-3 h-3 mr-1" />
                                Resolved
                            </TabsTrigger>
                        </TabsList>
                    </Tabs>

                    <Select value={sortBy} onValueChange={(v) => setSortBy(v as CommentSortOption)}>
                        <SelectTrigger className="w-[130px]">
                            <ArrowUpDown className="w-3 h-3 mr-2" />
                            <SelectValue />
                        </SelectTrigger>
                        <SelectContent>
                            <SelectItem value="newest">Newest</SelectItem>
                            <SelectItem value="oldest">Oldest</SelectItem>
                            <SelectItem value="popular">Popular</SelectItem>
                        </SelectContent>
                    </Select>
                </div>
            </div>

            {/* Comments List */}
            <ScrollArea className="flex-1">
                {filteredAndSortedComments.length === 0 ? (
                    <div className="flex flex-col items-center justify-center h-64 text-muted-foreground">
                        <MessageSquare className="w-12 h-12 mb-3 opacity-30" />
                        <p className="text-sm">
                            {filterBy === 'all' 
                                ? 'No comments yet. Start the discussion!' 
                                : filterBy === 'resolved'
                                ? 'No resolved comments'
                                : 'No unresolved comments - great job!'
                            }
                        </p>
                        {filterBy !== 'all' && (
                            <Button 
                                variant="link" 
                                size="sm" 
                                onClick={() => setFilterBy('all')}
                                className="mt-2"
                            >
                                Show all comments
                            </Button>
                        )}
                    </div>
                ) : (
                    <div className="p-4">
                        <CommentThread
                            entityType={entityType}
                            entityId={entityId}
                            comments={filteredAndSortedComments}
                            currentUserId={currentUserId}
                            isLoading={false}
                            onCreateComment={onCreateComment}
                            onResolveComment={onResolveComment}
                            onDeleteComment={onDeleteComment}
                            onEditComment={onEditComment}
                            onReply={onReply}
                        />
                    </div>
                )}
            </ScrollArea>
        </div>
    );
}
