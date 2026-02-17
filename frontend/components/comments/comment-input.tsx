'use client';

import { useState, useRef, useCallback } from 'react';
import { Button } from '@/components/ui/button';
import { Textarea } from '@/components/ui/textarea';
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from '@/components/ui/tooltip';
import { Send, Bold, Italic, Code, Quote, AtSign } from 'lucide-react';
import { MentionPopover, useMentions } from './mention-popover';
import type { CommentUser, CreateCommentRequest } from '@/types/comments';

interface CommentInputProps {
    onSubmit: (data: CreateCommentRequest) => Promise<void>;
    entityType: CreateCommentRequest['entityType'];
    entityId: string;
    parentId?: string;
    placeholder?: string;
    autoFocus?: boolean;
    currentUserId: string;
    onCancel?: () => void;
    submitLabel?: string;
    className?: string;
}

export function CommentInput({
    onSubmit,
    entityType,
    entityId,
    parentId,
    placeholder = 'Write a comment...',
    autoFocus = false,
    currentUserId,
    onCancel,
    submitLabel = 'Comment',
    className = '',
}: CommentInputProps) {
    const [content, setContent] = useState('');
    const [isSubmitting, setIsSubmitting] = useState(false);
    const textareaRef = useRef<HTMLTextAreaElement>(null);
    const [cursorPosition, setCursorPosition] = useState(0);

    const {
        isOpen,
        query,
        handleInputChange,
        insertMention,
        closeMentions,
        setIsOpen,
    } = useMentions();

    const handleChange = (e: React.ChangeEvent<HTMLTextAreaElement>) => {
        const value = e.target.value;
        const position = e.target.selectionStart;
        
        setContent(value);
        setCursorPosition(position);
        handleInputChange(value, position);
    };

    const handleKeyDown = (e: React.KeyboardEvent<HTMLTextAreaElement>) => {
        // Close mentions on Escape
        if (e.key === 'Escape' && isOpen) {
            e.preventDefault();
            closeMentions();
            return;
        }

        // Submit on Ctrl+Enter or Cmd+Enter
        if (e.key === 'Enter' && (e.ctrlKey || e.metaKey)) {
            e.preventDefault();
            handleSubmit();
            return;
        }

        // Cancel on Escape (if not in mentions)
        if (e.key === 'Escape' && onCancel && !isOpen) {
            e.preventDefault();
            onCancel();
            return;
        }
    };

    const handleSelectMention = useCallback(
        (user: CommentUser) => {
            const username = user.username || user.name.replace(/\s+/g, '').toLowerCase();
            const { newValue, newCursorPosition } = insertMention(
                content,
                cursorPosition,
                username
            );
            setContent(newValue);
            
            // Focus and set cursor position after state update
            setTimeout(() => {
                if (textareaRef.current) {
                    textareaRef.current.focus();
                    textareaRef.current.setSelectionRange(newCursorPosition, newCursorPosition);
                    setCursorPosition(newCursorPosition);
                }
            }, 0);
        },
        [content, cursorPosition, insertMention]
    );

    const handleSubmit = async () => {
        if (!content.trim() || isSubmitting) return;

        setIsSubmitting(true);
        try {
            await onSubmit({
                entityType,
                entityId,
                content: content.trim(),
                parentId,
            });
            setContent('');
            closeMentions();
        } catch (error) {
            console.error('Failed to submit comment:', error);
        } finally {
            setIsSubmitting(false);
        }
    };

    const insertFormatting = (before: string, after: string = '') => {
        const textarea = textareaRef.current;
        if (!textarea) return;

        const start = textarea.selectionStart;
        const end = textarea.selectionEnd;
        const selectedText = content.slice(start, end);
        const newContent = content.slice(0, start) + before + selectedText + after + content.slice(end);
        
        setContent(newContent);
        
        setTimeout(() => {
            textarea.focus();
            if (selectedText) {
                textarea.setSelectionRange(start + before.length, end + before.length);
            } else {
                textarea.setSelectionRange(start + before.length, start + before.length);
            }
        }, 0);
    };

    const formatButtons = [
        { icon: Bold, label: 'Bold', action: () => insertFormatting('**', '**') },
        { icon: Italic, label: 'Italic', action: () => insertFormatting('*', '*') },
        { icon: Code, label: 'Code', action: () => insertFormatting('`', '`') },
        { icon: Quote, label: 'Quote', action: () => insertFormatting('> ', '') },
    ];

    const _getInitials = (name: string) => {
        return name
            .split(' ')
            .map((n) => n[0])
            .join('')
            .toUpperCase()
            .slice(0, 2);
    };

    return (
        <div className={`space-y-2 ${className}`}>
            <div className="relative">
                <MentionPopover
                    query={query}
                    isOpen={isOpen}
                    onOpenChange={setIsOpen}
                    onSelect={handleSelectMention}
                    currentUserId={currentUserId}
                >
                    <div>
                        <Textarea
                            ref={textareaRef}
                            value={content}
                            onChange={handleChange}
                            onKeyDown={handleKeyDown}
                            onClick={(e) => setCursorPosition(e.currentTarget.selectionStart)}
                            onKeyUp={(e) => setCursorPosition(e.currentTarget.selectionStart)}
                            placeholder={placeholder}
                            autoFocus={autoFocus}
                            className="min-h-[80px] resize-none"
                            disabled={isSubmitting}
                        />
                    </div>
                </MentionPopover>
            </div>

            <div className="flex items-center justify-between">
                <TooltipProvider delayDuration={300}>
                    <div className="flex items-center gap-1">
                        {formatButtons.map(({ icon: Icon, label, action }) => (
                            <Tooltip key={label}>
                                <TooltipTrigger asChild>
                                    <Button
                                        type="button"
                                        variant="ghost"
                                        size="sm"
                                        className="h-8 w-8 p-0"
                                        onClick={action}
                                        disabled={isSubmitting}
                                    >
                                        <Icon className="h-4 w-4" />
                                    </Button>
                                </TooltipTrigger>
                                <TooltipContent>
                                    <p>{label}</p>
                                </TooltipContent>
                            </Tooltip>
                        ))}

                        <Tooltip>
                            <TooltipTrigger asChild>
                                <Button
                                    type="button"
                                    variant="ghost"
                                    size="sm"
                                    className="h-8 w-8 p-0"
                                    onClick={() => {
                                        const textarea = textareaRef.current;
                                        if (textarea) {
                                            const pos = textarea.selectionStart;
                                            const newContent = content.slice(0, pos) + '@' + content.slice(pos);
                                            setContent(newContent);
                                            setTimeout(() => {
                                                textarea.focus();
                                                textarea.setSelectionRange(pos + 1, pos + 1);
                                                handleInputChange(newContent, pos + 1);
                                            }, 0);
                                        }
                                    }}
                                    disabled={isSubmitting}
                                >
                                    <AtSign className="h-4 w-4" />
                                </Button>
                            </TooltipTrigger>
                            <TooltipContent>
                                <p>Mention (@)</p>
                            </TooltipContent>
                        </Tooltip>
                    </div>
                </TooltipProvider>

                <div className="flex items-center gap-2">
                    {onCancel && (
                        <Button
                            type="button"
                            variant="ghost"
                            size="sm"
                            onClick={onCancel}
                            disabled={isSubmitting}
                        >
                            Cancel
                        </Button>
                    )}
                    <Button
                        type="button"
                        size="sm"
                        onClick={handleSubmit}
                        disabled={!content.trim() || isSubmitting}
                    >
                        {isSubmitting ? (
                            <>
                                <div className="w-4 h-4 border-2 border-white border-t-transparent rounded-full animate-spin mr-2" />
                                Posting...
                            </>
                        ) : (
                            <>
                                <Send className="w-4 h-4 mr-2" />
                                {submitLabel}
                            </>
                        )}
                    </Button>
                </div>
            </div>

            <div className="text-xs text-muted-foreground">
                Press <kbd className="px-1 py-0.5 bg-muted rounded">Ctrl</kbd> +{' '}
                <kbd className="px-1 py-0.5 bg-muted rounded">Enter</kbd> to post
            </div>
        </div>
    );
}
