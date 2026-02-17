'use client';

import { useState } from 'react';
import { Label } from '@/components/ui/label';
import { Input } from '@/components/ui/input';
import { Checkbox } from '@/components/ui/checkbox';
import { Button } from '@/components/ui/button';
import {
    Accordion,
    AccordionContent,
    AccordionItem,
    AccordionTrigger,
} from '@/components/ui/accordion';
import { Mail, Webhook, Bell, MessageSquare, _Plus, Trash2 } from 'lucide-react';
import type { AlertChannelInput, AlertNotificationChannel } from '@/types/alerts';
import { ALERT_CHANNELS } from '@/types/alerts';

interface NotificationConfigProps {
    channels: AlertChannelInput[];
    onChange: (channels: AlertChannelInput[]) => void;
}

export function NotificationConfig({ channels, onChange }: NotificationConfigProps) {
    const [_newChannelType, _setNewChannelType] = useState<AlertNotificationChannel>('email');

    const addChannel = (type: AlertNotificationChannel) => {
        const existingIndex = channels.findIndex((c) => c.channelType === type);
        if (existingIndex >= 0) {
            // Enable existing channel
            const updated = [...channels];
            updated[existingIndex] = { ...updated[existingIndex], isEnabled: true };
            onChange(updated);
        } else {
            // Add new channel with default config
            onChange([
                ...channels,
                {
                    channelType: type,
                    isEnabled: true,
                    config: getDefaultConfig(type),
                },
            ]);
        }
    };

    const removeChannel = (index: number) => {
        const updated = [...channels];
        updated.splice(index, 1);
        onChange(updated);
    };

    const updateChannel = (index: number, updates: Partial<AlertChannelInput>) => {
        const updated = [...channels];
        updated[index] = { ...updated[index], ...updates };
        onChange(updated);
    };

    const getDefaultConfig = (type: AlertNotificationChannel) => {
        switch (type) {
            case 'email':
                return { recipients: [] };
            case 'webhook':
                return { url: '', headers: {} };
            case 'slack':
            case 'teams':
                return { url: '' };
            case 'in_app':
                return {};
            default:
                return {};
        }
    };

    const getChannelIcon = (type: AlertNotificationChannel) => {
        switch (type) {
            case 'email':
                return <Mail className="h-4 w-4" />;
            case 'webhook':
                return <Webhook className="h-4 w-4" />;
            case 'in_app':
                return <Bell className="h-4 w-4" />;
            case 'slack':
                return <MessageSquare className="h-4 w-4" />;
            case 'teams':
                return <MessageSquare className="h-4 w-4" />; // Using MessageSquare for Teams as well, or could import Users if preferred
        }
    };

    return (
        <div className="space-y-4">
            <p className="text-sm text-gray-500">
                Configure where notifications should be sent when this alert triggers.
            </p>

            {/* Add Channel Buttons */}
            <div className="flex flex-wrap gap-2">
                {ALERT_CHANNELS.map((channel) => {
                    const isActive = channels.some(
                        (c) => c.channelType === channel.value && c.isEnabled
                    );
                    return (
                        <Button
                            key={channel.value}
                            variant={isActive ? 'default' : 'outline'}
                            size="sm"
                            onClick={() => addChannel(channel.value)}
                        >
                            {getChannelIcon(channel.value)}
                            <span className="ml-2">{channel.label}</span>
                        </Button>
                    );
                })}
            </div>

            {/* Configured Channels */}
            <Accordion type="multiple" className="w-full">
                {channels.map((channel, index) => (
                    <AccordionItem key={`${channel.channelType}-${index}`} value={`item-${index}`}>
                        <AccordionTrigger className="hover:no-underline">
                            <div className="flex items-center gap-2">
                                {getChannelIcon(channel.channelType)}
                                <span className="capitalize">{channel.channelType === 'teams' ? 'Microsoft Teams' : channel.channelType}</span>
                                <Checkbox
                                    checked={channel.isEnabled}
                                    onCheckedChange={(checked) =>
                                        updateChannel(index, { isEnabled: checked as boolean })
                                    }
                                    onClick={(e) => e.stopPropagation()}
                                />
                            </div>
                        </AccordionTrigger>
                        <AccordionContent>
                            <div className="space-y-3 pt-2">
                                {channel.channelType === 'email' && (
                                    <div className="space-y-2">
                                        <Label>Email Recipients</Label>
                                        <Input
                                            placeholder="Enter email addresses, separated by commas"
                                            value={(channel.config?.recipients as string[])?.join(', ') || ''}
                                            onChange={(e) =>
                                                updateChannel(index, {
                                                    config: {
                                                        ...channel.config,
                                                        recipients: e.target.value
                                                            .split(',')
                                                            .map((s) => s.trim())
                                                            .filter(Boolean),
                                                    },
                                                })
                                            }
                                        />
                                    </div>
                                )}

                                {(channel.channelType === 'webhook' || channel.channelType === 'slack' || channel.channelType === 'teams') && (
                                    <div className="space-y-2">
                                        <Label>Webhook URL</Label>
                                        <Input
                                            type="url"
                                            placeholder="https://..."
                                            value={(channel.config?.url as string) || ''}
                                            onChange={(e) =>
                                                updateChannel(index, {
                                                    config: { ...channel.config, url: e.target.value },
                                                })
                                            }
                                        />
                                    </div>
                                )}

                                {channel.channelType === 'webhook' && (
                                    <div className="space-y-2">
                                        <Label>Custom Headers (JSON)</Label>
                                        <Input
                                            placeholder='{"Authorization": "Bearer token"}'
                                            value={JSON.stringify(channel.config?.headers || {})}
                                            onChange={(e) => {
                                                try {
                                                    const headers = JSON.parse(e.target.value);
                                                    updateChannel(index, {
                                                        config: { ...channel.config, headers },
                                                    });
                                                } catch {
                                                    // Invalid JSON, ignore
                                                }
                                            }}
                                        />
                                    </div>
                                )}

                                {channel.channelType === 'in_app' && (
                                    <p className="text-sm text-gray-500">
                                        In-app notifications will appear in your notification center.
                                    </p>
                                )}

                                <Button
                                    variant="ghost"
                                    size="sm"
                                    className="text-red-600"
                                    onClick={() => removeChannel(index)}
                                >
                                    <Trash2 className="h-4 w-4 mr-2" />
                                    Remove Channel
                                </Button>
                            </div>
                        </AccordionContent>
                    </AccordionItem>
                ))}
            </Accordion>

            {channels.length === 0 && (
                <div className="text-center py-8 text-gray-500">
                    <Bell className="h-8 w-8 mx-auto mb-2 text-gray-300" />
                    <p>No notification channels configured</p>
                    <p className="text-sm">Add at least one channel to receive alerts</p>
                </div>
            )}
        </div>
    );
}
