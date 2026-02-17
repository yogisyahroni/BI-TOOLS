'use client';

export const dynamic = 'force-dynamic';

import { useState, useEffect } from 'react';
import { fetchWithAuth } from '@/lib/utils';
import { WorkspaceMembers } from '@/components/workspace/workspace-members';
import {
    Card,
    CardContent,
    CardDescription,
    CardHeader,
    CardTitle,
} from '@/components/ui/card';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Button } from '@/components/ui/button';
import {
    Select,
    SelectContent,
    SelectItem,
    SelectTrigger,
    SelectValue,
} from '@/components/ui/select';
import { Save } from 'lucide-react';
import { toast } from 'sonner';
import { useWorkspace } from '@/hooks/use-workspace';

export default function WorkspaceSettings() {
    const { workspaceId } = useWorkspace();
    const [name, setName] = useState('');
    const [plan, setPlan] = useState('FREE');
    const [loading, setLoading] = useState(false);

    useEffect(() => {
        if (workspaceId) {
            // Mock fetch or initial state since API might not exist yet for GET /workspaces/:id
            // For now we just set defaults or leave empty to be filled
            setName('Default Workspace');
        }
    }, [workspaceId]);

    const handleSave = async () => {
        if (!workspaceId) return;
        setLoading(true);
        try {
            await fetchWithAuth(`/api/workspaces/${workspaceId}`, {
                method: 'PATCH',
                body: JSON.stringify({ name, plan }),
            });
            toast.success('Workspace updated');
        } catch (error) {
            toast.error('Failed to update workspace');
        } finally {
            setLoading(false);
        }
    };

    if (!workspaceId) return null;

    return (
        <div className="space-y-6">
            <Card>
                <CardHeader>
                    <CardTitle>Workspace Settings</CardTitle>
                    <CardDescription>
                        Manage your workspace preferences and subscription plan.
                    </CardDescription>
                </CardHeader>
                <CardContent className="space-y-4">
                    <div className="space-y-2">
                        <Label>Workspace Name</Label>
                        <Input
                            value={name}
                            onChange={(e) => setName(e.target.value)}
                            placeholder="My Workspace"
                        />
                    </div>
                    <div className="space-y-2">
                        <Label>Subscription Plan</Label>
                        <Select value={plan} onValueChange={setPlan}>
                            <SelectTrigger>
                                <SelectValue />
                            </SelectTrigger>
                            <SelectContent>
                                <SelectItem value="FREE">
                                    <div>
                                        <div className="font-medium">Free</div>
                                        <div className="text-xs text-muted-foreground">
                                            Basic features
                                        </div>
                                    </div>
                                </SelectItem>
                                <SelectItem value="PRO">
                                    <div>
                                        <div className="font-medium">Pro</div>
                                        <div className="text-xs text-muted-foreground">
                                            Advanced analytics
                                        </div>
                                    </div>
                                </SelectItem>
                                <SelectItem value="ENTERPRISE">
                                    <div>
                                        <div className="font-medium">Enterprise</div>
                                        <div className="text-xs text-muted-foreground">
                                            Unlimited everything
                                        </div>
                                    </div>
                                </SelectItem>
                            </SelectContent>
                        </Select>
                    </div>

                    <Button onClick={handleSave} disabled={loading}>
                        <Save className="mr-2 h-4 w-4" />
                        {loading ? 'Saving...' : 'Save Changes'}
                    </Button>
                </CardContent>
            </Card>

            <WorkspaceMembers
                workspaceId={workspaceId}
                isOwner={true}
                isAdmin={true}
            />
        </div>
    );
}
