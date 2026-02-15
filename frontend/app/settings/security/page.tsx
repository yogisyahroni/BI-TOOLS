'use client';

export const dynamic = 'force-dynamic';

import { useState } from 'react';
import { Shield, Lock, Save, Loader2, KeyRound, AlertTriangle } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from '@/components/ui/card';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Separator } from '@/components/ui/separator';
import { Switch } from '@/components/ui/switch';
import { toast } from 'sonner';
import { authApi } from '@/lib/api/auth';
import { z } from 'zod';
import { passwordSchema } from '@/lib/validations/auth';

const changePasswordSchema = z.object({
    currentPassword: z.string().min(1, 'Current password is required'),
    newPassword: passwordSchema,
    confirmPassword: z.string().min(1, 'Please confirm your new password'),
}).refine((data) => data.newPassword === data.confirmPassword, {
    message: "Passwords don't match",
    path: ["confirmPassword"],
});

export default function SecuritySettingsPage() {
    const [isLoading, setIsLoading] = useState(false);
    const [formData, setFormData] = useState({
        currentPassword: '',
        newPassword: '',
        confirmPassword: '',
    });
    const [errors, setErrors] = useState<Record<string, string>>({});

    const handleCopy = (e: React.ChangeEvent<HTMLInputElement>) => {
        const { name, value } = e.target;
        setFormData(prev => ({ ...prev, [name]: value }));
        // Clear error when user types
        if (errors[name]) {
            setErrors(prev => {
                const newErrors = { ...prev };
                delete newErrors[name];
                return newErrors;
            });
        }
    };

    const handleChangePassword = async (e: React.FormEvent) => {
        e.preventDefault();
        setIsLoading(true);
        setErrors({});

        try {
            // Validate
            const validatedData = changePasswordSchema.parse(formData);

            // Call API
            await authApi.changePassword(validatedData.currentPassword, validatedData.newPassword);

            toast.success('Password updated successfully');
            setFormData({
                currentPassword: '',
                newPassword: '',
                confirmPassword: '',
            });
        } catch (err: any) {
            if (err instanceof z.ZodError) {
                const fieldErrors: Record<string, string> = {};
                err.errors.forEach(e => {
                    if (e.path[0]) fieldErrors[e.path[0] as string] = e.message;
                });
                setErrors(fieldErrors);
            } else {
                toast.error(err.message || 'Failed to update password');
            }
        } finally {
            setIsLoading(false);
        }
    };

    return (
        <div className="container max-w-4xl py-8">
            <div className="mb-8">
                <h1 className="text-3xl font-bold flex items-center gap-3">
                    <Shield className="h-8 w-8 text-primary" />
                    Security Settings
                </h1>
                <p className="text-muted-foreground mt-2">
                    Manage your account security and authentication preferences
                </p>
            </div>

            <div className="grid gap-8">
                {/* Password Change Section */}
                <Card>
                    <CardHeader>
                        <div className="flex items-center gap-2">
                            <KeyRound className="h-5 w-5 text-primary" />
                            <CardTitle>Change Password</CardTitle>
                        </div>
                        <CardDescription>
                            Update your password to keep your account secure.
                        </CardDescription>
                    </CardHeader>
                    <CardContent>
                        <form id="password-form" onSubmit={handleChangePassword} className="space-y-4 max-w-lg">
                            <div className="space-y-2">
                                <Label htmlFor="currentPassword">Current Password</Label>
                                <Input
                                    id="currentPassword"
                                    name="currentPassword"
                                    type="password"
                                    value={formData.currentPassword}
                                    onChange={handleCopy}
                                    disabled={isLoading}
                                />
                                {errors.currentPassword && <p className="text-sm text-destructive">{errors.currentPassword}</p>}
                            </div>
                            <div className="space-y-2">
                                <Label htmlFor="newPassword">New Password</Label>
                                <Input
                                    id="newPassword"
                                    name="newPassword"
                                    type="password"
                                    value={formData.newPassword}
                                    onChange={handleCopy}
                                    disabled={isLoading}
                                />
                                {errors.newPassword && <p className="text-sm text-destructive">{errors.newPassword}</p>}
                            </div>
                            <div className="space-y-2">
                                <Label htmlFor="confirmPassword">Confirm New Password</Label>
                                <Input
                                    id="confirmPassword"
                                    name="confirmPassword"
                                    type="password"
                                    value={formData.confirmPassword}
                                    onChange={handleCopy}
                                    disabled={isLoading}
                                />
                                {errors.confirmPassword && <p className="text-sm text-destructive">{errors.confirmPassword}</p>}
                            </div>
                        </form>
                    </CardContent>
                    <CardFooter className="border-t bg-muted/5 px-6 py-4">
                        <Button type="submit" form="password-form" disabled={isLoading}>
                            {isLoading ? <Loader2 className="h-4 w-4 animate-spin mr-2" /> : <Save className="h-4 w-4 mr-2" />}
                            Update Password
                        </Button>
                    </CardFooter>
                </Card>

                {/* Two-Factor Authentication (Placeholder) */}
                <Card className="opacity-80">
                    <CardHeader>
                        <div className="flex items-center gap-2">
                            <Lock className="h-5 w-5 text-primary" />
                            <CardTitle>Two-Factor Authentication</CardTitle>
                        </div>
                        <CardDescription>
                            Add an extra layer of security to your account.
                        </CardDescription>
                    </CardHeader>
                    <CardContent>
                        <div className="flex items-center justify-between">
                            <div className="space-y-1">
                                <p className="font-medium">Authenticator App</p>
                                <p className="text-sm text-muted-foreground">
                                    Use Google Authenticator or similar apps.
                                </p>
                            </div>
                            <Switch disabled />
                        </div>
                    </CardContent>
                    <CardFooter className="border-t bg-muted/5 px-6 py-3">
                        <p className="text-xs text-muted-foreground flex items-center gap-1">
                            <AlertTriangle className="h-3 w-3" />
                            Use of hardware keys is coming soon.
                        </p>
                    </CardFooter>
                </Card>

                {/* Sessions (Placeholder) */}
                <Card className="opacity-80">
                    <CardHeader>
                        <CardTitle>Active Sessions</CardTitle>
                        <CardDescription>
                            Manage devices logged into your account.
                        </CardDescription>
                    </CardHeader>
                    <CardContent>
                        <div className="space-y-4">
                            <div className="flex items-center justify-between p-3 border rounded-lg bg-background">
                                <div className="flex items-center gap-3">
                                    <div className="bg-primary/10 p-2 rounded-full">
                                        <Shield className="h-4 w-4 text-primary" />
                                    </div>
                                    <div>
                                        <p className="font-medium text-sm">Windows Chrome</p>
                                        <p className="text-xs text-muted-foreground">Jakarta, Indonesia • Current Session</p>
                                    </div>
                                </div>
                                <div className="h-2 w-2 rounded-full bg-green-500"></div>
                            </div>
                        </div>
                    </CardContent>
                </Card>
            </div>
        </div>
    );
}
