'use client';

import { useState, useEffect } from 'react';
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter } from '@/components/ui/dialog';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { Switch } from '@/components/ui/switch';
import { DashboardCard, VisualizationConfig } from '@/lib/types';
import { useSavedQueries } from '@/hooks/use-saved-queries';
import { Card, CardContent } from '@/components/ui/card';
import { Separator } from '@/components/ui/separator';

interface EditCardDialogProps {
    open: boolean;
    onOpenChange: (open: boolean) => void;
    card: DashboardCard;
    onSave: (cardId: string, updates: Partial<DashboardCard>) => void;
}

export function EditCardDialog({ open, onOpenChange, card, onSave }: EditCardDialogProps) {
    const [title, setTitle] = useState(card.title);
    const [description, setDescription] = useState(card.description || '');

    // Interaction Config
    const [clickAction, setClickAction] = useState<any>(card.visualizationConfig?.clickAction || 'filter');
    const [drillType, setDrillType] = useState<string>(card.visualizationConfig?.drillConfig?.type || 'dashboard');
    const [targetDashboardId, setTargetDashboardId] = useState(card.visualizationConfig?.drillConfig?.dashboardId || '');
    const [targetUrl, setTargetUrl] = useState(card.visualizationConfig?.drillConfig?.url || '');
    const [openInNewTab, setOpenInNewTab] = useState(card.visualizationConfig?.drillConfig?.openInNewTab || false);

    // Tooltip State
    const [tooltipTemplate, setTooltipTemplate] = useState(card.visualizationConfig?.tooltipTemplate || '');

    // Fetch dashboards for dropdown
    const [dashboards, setDashboards] = useState<{ id: string, name: string }[]>([]);
    useEffect(() => {
        if (open && clickAction === 'drill') {
            fetch('/api/go/dashboards').then(res => res.json()).then(data => {
                if (data.success) setDashboards(data.data);
            });
        }
    }, [open, clickAction]);

    const handleSave = () => {
        const updates: Partial<DashboardCard> = {
            title,
            description,
        };

        if (card.type === 'visualization') {
            updates.visualizationConfig = {
                ...card.visualizationConfig,
                ...card.query?.visualizationConfig, // Base config
                clickAction,
                drillConfig: clickAction === 'drill' ? {
                    type: drillType as 'dashboard' | 'url',
                    dashboardId: drillType === 'dashboard' ? targetDashboardId : undefined,
                    url: drillType === 'url' ? targetUrl : undefined,
                    openInNewTab
                } : undefined,
                tooltipTemplate // Save tooltip template
            } as VisualizationConfig;
        }

        onSave(card.id, updates);
        onOpenChange(false);
    };

    return (
        <Dialog open={open} onOpenChange={onOpenChange}>
            <DialogContent className="sm:max-w-[600px] max-h-[80vh] overflow-y-auto">
                <DialogHeader>
                    <DialogTitle>Edit Card: {card.title}</DialogTitle>
                </DialogHeader>

                <Tabs defaultValue="general" className="w-full">
                    <TabsList className="grid w-full grid-cols-2">
                        <TabsTrigger value="interaction">Interaction</TabsTrigger>
                        <TabsTrigger value="tooltip">Tooltip</TabsTrigger>
                    </TabsList>

                    <TabsContent value="general" className="space-y-4 py-4">
                        <div className="space-y-2">
                            <Label>Title</Label>
                            <Input value={title} onChange={(e) => setTitle(e.target.value)} />
                        </div>
                        <div className="space-y-2">
                            <Label>Description</Label>
                            <Input value={description} onChange={(e) => setDescription(e.target.value)} />
                        </div>
                    </TabsContent>

                    <TabsContent value="interactions" className="space-y-4 py-4">
                        <div className="space-y-2">
                            <Label>On Click Action</Label>
                            <Select value={clickAction} onValueChange={setClickAction}>
                                <SelectTrigger>
                                    <SelectValue />
                                </SelectTrigger>
                                <SelectContent>
                                    <SelectItem value="none">None</SelectItem>
                                    <SelectItem value="filter">Filter Dashboard (Cross-Filter)</SelectItem>
                                    <SelectItem value="drill">Drill Through</SelectItem>
                                </SelectContent>
                            </Select>
                        </div>

                        {clickAction === 'drill' && (
                            <Card>
                                <CardContent className="pt-6 space-y-4">
                                    <div className="space-y-2">
                                        <Label>Target Type</Label>
                                        <Select value={drillType} onValueChange={setDrillType}>
                                            <SelectTrigger><SelectValue /></SelectTrigger>
                                            <SelectContent>
                                                <SelectItem value="dashboard">Another Dashboard</SelectItem>
                                                <SelectItem value="url">External URL</SelectItem>
                                            </SelectContent>
                                        </Select>
                                    </div>

                                    {drillType === 'dashboard' && (
                                        <div className="space-y-2">
                                            <Label>Target Dashboard</Label>
                                            <Select value={targetDashboardId} onValueChange={setTargetDashboardId}>
                                                <SelectTrigger><SelectValue placeholder="Select dashboard" /></SelectTrigger>
                                                <SelectContent>
                                                    {dashboards.map(d => (
                                                        <SelectItem key={d.id} value={d.id}>{d.name}</SelectItem>
                                                    ))}
                                                </SelectContent>
                                            </Select>
                                        </div>
                                    )}

                                    {drillType === 'url' && (
                                        <div className="space-y-2">
                                            <Label>URL Template</Label>
                                            <Input
                                                value={targetUrl}
                                                onChange={(e) => setTargetUrl(e.target.value)}
                                                placeholder="https://example.com/details?id={{value}}"
                                            />
                                            <p className="text-xs text-muted-foreground">Use {'{{value}}'} or {'{{series}}'} as placeholders.</p>
                                        </div>
                                    )}

                                    <div className="flex items-center space-x-2">
                                        <Switch id="new-tab" checked={openInNewTab} onCheckedChange={setOpenInNewTab} />
                                        <Label htmlFor="new-tab">Open in new tab</Label>
                                    </div>
                                </CardContent>
                            </Card>
                        )}
                    </TabsContent>
                </Tabs>

                <DialogFooter>
                    <Button variant="outline" onClick={() => onOpenChange(false)}>Cancel</Button>
                    <Button onClick={handleSave}>Save Changes</Button>
                </DialogFooter>
            </DialogContent>
        </Dialog>
    );
}
