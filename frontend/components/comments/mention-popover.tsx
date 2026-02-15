'use client';

import { useState, useEffect, useCallback } from 'react';
import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar';
import { Command, CommandEmpty, CommandGroup, CommandItem, CommandList } from '@/components/ui/command';
import { Popover, PopoverContent, PopoverTrigger } from '@/components/ui/popover';
import { Clock, User, Users } from 'lucide-react';
import type { CommentUser } from '@/types/comments';

interface MentionPopoverProps {
    children: React.ReactNode;
    query: string;
    isOpen: boolean;
    onOpenChange: (open: boolean) => void;
    onSelect: (user: CommentUser) => void;
    currentUserId: string;
    excludeUserIds?: string[];
}

export function MentionPopover({
    children,
    query,
    isOpen,
    onOpenChange,
    onSelect,
    currentUserId,
    excludeUserIds = [],
}: MentionPopoverProps) {
    const [users, setUsers] = useState<CommentUser[]>([]);
    const [recentUsers, setRecentUsers] = useState<CommentUser[]>([]);
    const [isLoading, setIsLoading] = useState(false);

    // Fetch users matching the query
    const searchUsers = useCallback(async (searchQuery: string) => {
        if (!searchQuery || searchQuery.length < 1) {
            setUsers([]);
            return;
        }

        setIsLoading(true);
        try {
            const params = new URLSearchParams({
                q: searchQuery,
                limit: '10',
            });
            
            const response = await fetch(`/api/go/comments/mentions/search?${params}`);
            if (response.ok) {
                const data = await response.json();
                // Filter out current user and excluded users
                const filtered = (data as CommentUser[]).filter(
                    (u) => u.id !== currentUserId && !excludeUserIds.includes(u.id)
                );
                setUsers(filtered);
            }
        } catch (error) {
            console.error('Failed to search users:', error);
            setUsers([]);
        } finally {
            setIsLoading(false);
        }
    }, [currentUserId, excludeUserIds]);

    // Fetch recent mentions
    const fetchRecentMentions = useCallback(async () => {
        try {
            const params = new URLSearchParams({
                limit: '5',
            });
            
            const response = await fetch(`/api/go/comments/mentions/recent?${params}`);
            if (response.ok) {
                const data = await response.json();
                setRecentUsers(data as CommentUser[]);
            }
        } catch (error) {
            console.error('Failed to fetch recent mentions:', error);
        }
    }, []);

    // Search when query changes
    useEffect(() => {
        const timeoutId = setTimeout(() => {
            searchUsers(query);
        }, 150);

        return () => clearTimeout(timeoutId);
    }, [query, searchUsers]);

    // Fetch recent mentions when popover opens
    useEffect(() => {
        if (isOpen) {
            fetchRecentMentions();
        }
    }, [isOpen, fetchRecentMentions]);

    const handleSelect = (user: CommentUser) => {
        onSelect(user);
        onOpenChange(false);
    };

    const getInitials = (name: string) => {
        return name
            .split(' ')
            .map((n) => n[0])
            .join('')
            .toUpperCase()
            .slice(0, 2);
    };

    return (
        <Popover open={isOpen} onOpenChange={onOpenChange}>
            <PopoverTrigger asChild>{children}</PopoverTrigger>
            <PopoverContent 
                className="w-72 p-0" 
                align="start" 
                side="top"
                sideOffset={5}
            >
                <Command>
                    <CommandList>
                        {query.length > 0 && (
                            <CommandGroup heading="Search Results">
                                {isLoading ? (
                                    <CommandItem disabled>
                                        <div className="flex items-center gap-2 py-1">
                                            <div className="w-4 h-4 border-2 border-primary border-t-transparent rounded-full animate-spin" />
                                            <span className="text-sm text-muted-foreground">Searching...</span>
                                        </div>
                                    </CommandItem>
                                ) : users.length === 0 ? (
                                    <CommandEmpty>
                                        <div className="flex flex-col items-center py-4 text-muted-foreground">
                                            <User className="w-8 h-8 mb-2 opacity-50" />
                                            <span className="text-sm">No users found</span>
                                        </div>
                                    </CommandEmpty>
                                ) : (
                                    users.map((user) => (
                                        <CommandItem
                                            key={user.id}
                                            onSelect={() => handleSelect(user)}
                                            className="flex items-center gap-2 py-2 cursor-pointer"
                                        >
                                            <Avatar className="w-6 h-6">
                                                <AvatarImage src={user.image || user.avatar} />
                                                <AvatarFallback className="text-xs">
                                                    {getInitials(user.name)}
                                                </AvatarFallback>
                                            </Avatar>
                                            <div className="flex flex-col min-w-0">
                                                <span className="text-sm font-medium truncate">
                                                    {user.name}
                                                </span>
                                                {user.username && (
                                                    <span className="text-xs text-muted-foreground">
                                                        @{user.username}
                                                    </span>
                                                )}
                                            </div>
                                        </CommandItem>
                                    ))
                                )}
                            </CommandGroup>
                        )}

                        {query.length === 0 && recentUsers.length > 0 && (
                            <CommandGroup heading="Recent Mentions">
                                <div className="flex items-center gap-1 px-2 py-1 text-xs text-muted-foreground">
                                    <Clock className="w-3 h-3" />
                                    <span>Recently mentioned</span>
                                </div>
                                {recentUsers.map((user) => (
                                    <CommandItem
                                        key={user.id}
                                        onSelect={() => handleSelect(user)}
                                        className="flex items-center gap-2 py-2 cursor-pointer"
                                    >
                                        <Avatar className="w-6 h-6">
                                            <AvatarImage src={user.image || user.avatar} />
                                            <AvatarFallback className="text-xs">
                                                {getInitials(user.name)}
                                            </AvatarFallback>
                                        </Avatar>
                                        <div className="flex flex-col min-w-0">
                                            <span className="text-sm font-medium truncate">
                                                {user.name}
                                            </span>
                                            {user.username && (
                                                <span className="text-xs text-muted-foreground">
                                                    @{user.username}
                                                </span>
                                            )}
                                        </div>
                                    </CommandItem>
                                ))}
                            </CommandGroup>
                        )}

                        {query.length === 0 && recentUsers.length === 0 && (
                            <CommandGroup>
                                <div className="flex flex-col items-center py-4 text-muted-foreground">
                                    <Users className="w-8 h-8 mb-2 opacity-50" />
                                    <span className="text-sm">Type @ to mention someone</span>
                                </div>
                            </CommandGroup>
                        )}
                    </CommandList>
                </Command>
            </PopoverContent>
        </Popover>
    );
}

// Hook for managing mention state
export function useMentions() {
    const [isOpen, setIsOpen] = useState(false);
    const [query, setQuery] = useState('');
    const [mentionStartIndex, setMentionStartIndex] = useState<number | null>(null);

    const handleInputChange = (
        value: string,
        cursorPosition: number
    ): { shouldShowMentions: boolean; mentionQuery: string } => {
        // Find the last @ before cursor
        const textBeforeCursor = value.slice(0, cursorPosition);
        const lastAtIndex = textBeforeCursor.lastIndexOf('@');

        if (lastAtIndex === -1) {
            setIsOpen(false);
            setQuery('');
            return { shouldShowMentions: false, mentionQuery: '' };
        }

        // Check if there's a space between @ and cursor (which would close the mention)
        const textAfterAt = textBeforeCursor.slice(lastAtIndex + 1);
        if (textAfterAt.includes(' ')) {
            setIsOpen(false);
            setQuery('');
            return { shouldShowMentions: false, mentionQuery: '' };
        }

        // Show mentions
        setMentionStartIndex(lastAtIndex);
        setQuery(textAfterAt);
        setIsOpen(true);
        
        return { shouldShowMentions: true, mentionQuery: textAfterAt };
    };

    const insertMention = (
        currentValue: string,
        cursorPosition: number,
        username: string
    ): { newValue: string; newCursorPosition: number } => {
        if (mentionStartIndex === null) {
            return { newValue: currentValue, newCursorPosition: cursorPosition };
        }

        const beforeMention = currentValue.slice(0, mentionStartIndex);
        const afterMention = currentValue.slice(cursorPosition);
        const newValue = `${beforeMention}@${username} ${afterMention}`;
        const newCursorPosition = mentionStartIndex + username.length + 2; // +2 for @ and space

        setIsOpen(false);
        setQuery('');
        setMentionStartIndex(null);

        return { newValue, newCursorPosition };
    };

    const closeMentions = () => {
        setIsOpen(false);
        setQuery('');
        setMentionStartIndex(null);
    };

    return {
        isOpen,
        query,
        mentionStartIndex,
        setIsOpen,
        handleInputChange,
        insertMention,
        closeMentions,
    };
}
