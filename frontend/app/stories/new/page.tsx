'use client';

import { useState, useEffect } from 'react';
import { useRouter } from 'next/navigation';
import { useDatabase } from '@/contexts/database-context';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { ArrowRight, Loader2 } from 'lucide-react';
import { toast } from 'sonner';

export default function NewStoryPage() {
    const router = useRouter();
    const { selectedDatabase } = useDatabase();
    const [dashboards, setDashboards] = useState<any[]>([]);
    const [selectedDashboardId, setSelectedDashboardId] = useState<string>('');
    const [isLoading, setIsLoading] = useState(false);
    const [isCreating, setIsCreating] = useState(false);

    useEffect(() => {
        if (selectedDatabase?.id) {
            fetchDashboards(selectedDatabase.id);
        }
    }, [selectedDatabase?.id]);

    const fetchDashboards = async (dbId: string) => {
        setIsLoading(true);
        try {
            const res = await fetch(`/api/dashboards?database_id=${dbId}`);
            if (!res.ok) throw new Error('Failed to fetch dashboards');
            const data = await res.json();
            setDashboards(data);
        } catch (error) {
            toast.error('Failed to load dashboards');
            console.error(error);
        } finally {
            setIsLoading(false);
        }
    };

    const handleCreate = () => {
        if (!selectedDashboardId) return;
        setIsCreating(true);
        // Simulate creation delay or just redirect
        setTimeout(() => {
            router.push(`/stories/draft?dashboardId=${selectedDashboardId}`);
        }, 500);
    };

    return (
        <div className="container mx-auto max-w-2xl py-20">
            <Card>
                <CardHeader>
                    <CardTitle>Create New Story</CardTitle>
                    <CardDescription>
                        Select a dashboard to generate your AI-powered presentation from.
                    </CardDescription>
                </CardHeader>
                <CardContent className="space-y-6">
                    <div className="space-y-2">
                        <label className="text-sm font-medium">Select Dashboard</label>
                        <Select
                            value={selectedDashboardId}
                            onValueChange={setSelectedDashboardId}
                            disabled={isLoading}
                        >
                            <SelectTrigger>
                                <SelectValue placeholder="Select a dashboard..." />
                            </SelectTrigger>
                            <SelectContent>
                                {dashboards.map((db) => (
                                    <SelectItem key={db.id} value={db.id}>
                                        {db.name}
                                    </SelectItem>
                                ))}
                                {dashboards.length === 0 && !isLoading && (
                                    <div className="p-2 text-sm text-muted-foreground text-center">
                                        No dashboards found in this database.
                                    </div>
                                )}
                            </SelectContent>
                        </Select>
                    </div>

                    <div className="flex justify-end gap-2">
                        <Button variant="outline" onClick={() => router.back()}>
                            Cancel
                        </Button>
                        <Button
                            onClick={handleCreate}
                            disabled={!selectedDashboardId || isCreating}
                        >
                            {isCreating ? (
                                <>
                                    <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                                    Creating...
                                </>
                            ) : (
                                <>
                                    Create Story
                                    <ArrowRight className="ml-2 h-4 w-4" />
                                </>
                            )}
                        </Button>
                    </div>
                </CardContent>
            </Card>
        </div>
    );
}
