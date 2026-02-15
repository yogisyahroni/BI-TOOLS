'use client';

import { useState } from 'react';
import { Button } from '@/components/ui/button';
import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar';
import { Badge } from '@/components/ui/badge';
import { ScrollArea } from '@/components/ui/scroll-area';
import { 
    MessageSquare, 
    CheckCircle2, 
    CornerDownRight, 
    MoreHorizontal,
    Trash2,
    Edit2,
    RotateCcw,
    Bell
} from 'lucide-react';
import {
    DropdownMenu,
    DropdownMenuContent,
    DropdownMenuItem,
    DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import { formatDistanceToNow } from 'date-fns';
import { CommentInput } from './comment-input';
import type { Comment, CommentEntityType, CreateCommentRequest, CommentUser } from '@/types/comments';

// Helper to parse and highlight mentions
function HighlightMentions({ content }: { content: string }) {
    // Split content by mention pattern
    const parts = content.split(/(@\w+)/g);
    
    return (
        <>
            {parts.map((part, index) => {
                if (part.startsWith('@')) {
                    return (
                        <span 
                            key={index} 
                            className="text-primary font-medium bg-primary/10 px-1 rounded"
                        >
                            {part}
                        </span>
                    );
                }
                return <span key={index}>{part}</span>;
            })}
        </>
    );
}

interface CommentItemProps {
    comment: Comment;
    currentUserId: string;
    onReply: (parentId: string) => void;
    onResolve: (commentId: string, isResolved: boolean) => void;
    onDelete: (commentId: string) => void;
    onEdit: (commentId: string, content: string) => void;
    depth?: number;
}

function CommentItem({ 
    comment, 
    currentUserId, 
    onReply, 
    onResolve, 
    onDelete, 
    onEdit,
    depth = 0 
}: CommentItemProps) {
    const [isReplying, setIsReplying] = useState(false);
    const [isEditing, setIsEditing] = useState(false);
    const [editContent, setEditContent] = useState(comment.content);
    const [showReplies, setShowReplies] = useState(depth < 1);
    const isOwner = comment.userId === currentUserId;
    const canResolve = isOwner || depth === 0; // Simplified - real logic would check entity ownership

    const getInitials = (name: string) => {
        return name
            .split(' ')
            .map((n) => n[0])
            .join('')
            .toUpperCase()
            .slice(0, 2);
    };

    const handleSaveEdit = () => {
        if (editContent.trim() && editContent !== comment.content) {
            onEdit(comment.id, editContent.trim());
        }
        setIsEditing(false);
    };

    const user = comment.user || {
        id: comment.userId,
        name: 'Unknown User',
        email: '',
    };

    return (
        <div className={`${depth > 0 ? 'ml-8 mt-3' : ''}`}>
            <div className="flex gap-3">
                <Avatar className="w-8 h-8 flex-shrink-0">
                    <AvatarImage src={user.image} />
                    <AvatarFallback className="text-xs bg-primary/10">
                        {getInitials(user.name)}
                    </AvatarFallback>
                </Avatar>

                <div className="flex-1 min-w-0">
                    <div className="flex items-center justify-between gap-2">
                        <div className="flex items-center gap-2 min-w-0">
                            <span className="text-sm font-semibold truncate">{user.name}</span>
                            <span className="text-xs text-muted-foreground">
                                {formatDistanceToNow(new Date(comment.createdAt))} ago
                            </span>
                            {comment.isResolved && (
                                <Badge variant="secondary" className="text-xs bg-green-100 text-green-700">
                                    <CheckCircle2 className="w-3 h-3 mr-1" />
                                    Resolved
                                </Badge>
                            )}
                        </div>

                        <DropdownMenu>
                            <DropdownMenuTrigger asChild>
                                <Button variant="ghost" size="sm" className="h-6 w-6 p-0 opacity-0 group-hover:opacity-100 transition-opacity">
                                    <MoreHorizontal className="h-4 w-4" />
                                </Button>
                            </DropdownMenuTrigger>
                            <DropdownMenuContent align="end">
                                {isOwner && (
                                    <DropdownMenuItem onClick={() => setIsEditing(true)}>
                                        <Edit2 className="w-4 h-4 mr-2" />
                                        Edit
                                    </DropdownMenuItem>
                                )}
                                {canResolve && (
                                    <DropdownMenuItem onClick={() => onResolve(comment.id, !comment.isResolved)}>
                                        {comment.isResolved ? (
                                            <>
                                                <RotateCcw className="w-4 h-4 mr-2" />
                                                Unresolve
                                            </>
                                        ) : (
                                            <>
                                                <CheckCircle2 className="w-4 h-4 mr-2" />
                                                Resolve
                                            </>
                                        )}
                                    </DropdownMenuItem>
                                )}
                                {(isOwner || canResolve) && (
                                    <DropdownMenuItem 
                                        onClick={() => onDelete(comment.id)}
                                        className="text-destructive"
                                    >
                                        <Trash2 className="w-4 h-4 mr-2" />
                                        Delete
                                    </DropdownMenuItem>
                                )}
                            </DropdownMenuContent>
                        </DropdownMenu>
                    </div>

                    {isEditing ? (
                        <div className="mt-2 space-y-2">
                            <textarea
                                value={editContent}
                                onChange={(e) => setEditContent(e.target.value)}
                                className="w-full min-h-[60px] p-2 text-sm border rounded-md resize-none focus:outline-none focus:ring-2 focus:ring-primary"
                                autoFocus
                            />
                            <div className="flex gap-2">
                                <Button size="sm" onClick={handleSaveEdit}>Save</Button>
                                <Button size="sm" variant="ghost" onClick={() => {
                                    setIsEditing(false);
                                    setEditContent(comment.content);
                                }}>
                                    Cancel
                                </Button>
                            </div>
                        </div>
                    ) : (
                        <div className={`text-sm mt-1 p-2 rounded-md ${comment.isResolved ? 'bg-muted/50' : 'bg-muted'}`}>
                            <div className="whitespace-pre-wrap">
                                <HighlightMentions content={comment.content} />
                            </div>
                        </div>
                    )}

                    {/* Annotation indicator */}
                    {comment.annotation && (
                        <div className="mt-2 flex items-center gap-2 text-xs text-muted-foreground">
                            <div 
                                className="w-3 h-3 rounded-full"
                                style={{ backgroundColor: comment.annotation.color }}
                            />
                            <span>Annotation on chart</span>
                            <span className="text-xs">
                                ({comment.annotation.type === 'point' ? 'Point' : 
                                  comment.annotation.type === 'range' ? 'Range' : 'Text'})
                            </span>
                        </div>
                    )}

                    {/* Action buttons */}
                    <div className="flex items-center gap-4 mt-2 text-xs text-muted-foreground">
                        <button
                            className="hover:text-foreground transition-colors"
                            onClick={() => setIsReplying(!isReplying)}
                        >
                            Reply
                        </button>
                        
                        {comment.replies && comment.replies.length > 0 && (
                            <button
                                className="hover:text-foreground transition-colors flex items-center gap-1"
                                onClick={() => setShowReplies(!showReplies)}
                            >
                                <CornerDownRight className="w-3 h-3" />
                                {showReplies ? 'Hide' : 'Show'} {comment.replies.length} {comment.replies.length === 1 ? 'reply' : 'replies'}
                            </button>
                        )}

                        {comment.mentions && comment.mentions.length > 0 && (
                            <div className="flex items-center gap-1">
                                <Bell className="w-3 h-3" />
                                {comment.mentions.length} {comment.mentions.length === 1 ? 'mention' : 'mentions'}
                            </div>
                        )}
                    </div>

                    {/* Reply input */}
                    {isReplying && (
                        <div className="mt-3">
                            <CommentInput
                                onSubmit={async (data) => {
                                    await onReply(comment.id);
                                    setIsReplying(false);
                                }}
                                entityType={comment.entityType}
                                entityId={comment.entityId}
                                parentId={comment.id}
                                placeholder={`Reply to ${user.name}...`}
                                currentUserId={currentUserId}
                                onCancel={() => setIsReplying(false)}
                                submitLabel="Reply"
                            />
                        </div>
                    )}

                    {/* Replies */}
                    {showReplies && comment.replies && comment.replies.length > 0 && (
                        <div className="mt-3">
                            {comment.replies.map((reply) => (
                                <CommentItem
                                    key={reply.id}
                                    comment={reply}
                                    currentUserId={currentUserId}
                                    onReply={onReply}
                                    onResolve={onResolve}
                                    onDelete={onDelete}
                                    onEdit={onEdit}
                                    depth={depth + 1}
                                />
                            ))}
                        </div>
                    )}
                </div>
            </div>
        </div>
    );
}

interface CommentThreadProps {
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

export function CommentThread({
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
}: CommentThreadProps) {
    const unresolvedCount = comments.filter(c => !c.isResolved && !c.parentId).length;
    const totalCount = comments.filter(c => !c.parentId).length;

    if (isLoading) {
        return (
            <div className={`flex flex-col h-[400px] ${className}`}>
                <div className="p-3 border-b bg-muted/50 flex items-center justify-between">
                    <h3 className="font-semibold text-sm flex items-center">
                        <MessageSquare className="w-4 h-4 mr-2" /> Discussion
                    </h3>
                </div>
                <div className="flex-1 flex items-center justify-center">
                    <div className="flex items-center gap-2 text-muted-foreground">
                        <div className="w-4 h-4 border-2 border-primary border-t-transparent rounded-full animate-spin" />
                        <span className="text-sm">Loading discussions...</span>
                    </div>
                </div>
            </div>
        );
    }

    return (
        <div className={`flex flex-col h-[500px] ${className}`}>
            {/* Header */}
            <div className="p-3 border-b bg-muted/50 flex items-center justify-between">
                <h3 className="font-semibold text-sm flex items-center">
                    <MessageSquare className="w-4 h-4 mr-2" /> Discussion
                </h3>
                <div className="flex items-center gap-2">
                    {unresolvedCount > 0 && (
                        <Badge variant="secondary" className="text-xs">
                            {unresolvedCount} unresolved
                        </Badge>
                    )}
                    <span className="text-xs text-muted-foreground">
                        {totalCount} threads
                    </span>
                </div>
            </div>

            {/* Comments list */}
            <ScrollArea className="flex-1 p-4">
                {comments.length === 0 ? (
                    <div className="text-center text-muted-foreground py-8 text-sm">
                        <MessageSquare className="w-12 h-12 mx-auto mb-3 opacity-30" />
                        <p>No comments yet. Start a discussion!</p>
                    </div>
                ) : (
                    <div className="space-y-6">
                        {comments
                            .filter(comment => !comment.parentId) // Only top-level comments
                            .map(comment => (
                                <CommentItem
                                    key={comment.id}
                                    comment={comment}
                                    currentUserId={currentUserId}
                                    onReply={async (parentId) => {
                                        if (onReply) {
                                            await onReply(parentId);
                                        }
                                    }}
                                    onResolve={onResolveComment}
                                    onDelete={onDeleteComment}
                                    onEdit={onEditComment}
                                />
                            ))}
                    </div>
                )}
            </ScrollArea>

            {/* Input area */}
            <div className="p-3 border-t bg-muted/30">
                <CommentInput
                    onSubmit={onCreateComment}
                    entityType={entityType}
                    entityId={entityId}
                    placeholder="Add to the discussion..."
                    currentUserId={currentUserId}
                    submitLabel="Post"
                />
            </div>
        </div>
    );
}
