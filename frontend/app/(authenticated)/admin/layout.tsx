'use client';

import Link from 'next/link';
import { usePathname } from 'next/navigation';
import { cn } from '@/lib/utils';
import { Building2, Users, Activity, LayoutDashboard } from 'lucide-react';

const navigation = [
    {
        name: 'Overview',
        href: '/admin',
        icon: LayoutDashboard,
    },
    {
        name: 'Organizations',
        href: '/admin/organizations',
        icon: Building2,
    },
    {
        name: 'Users',
        href: '/admin/users',
        icon: Users,
    },
    {
        name: 'System Health',
        href: '/admin/system',
        icon: Activity,
    },
];

export default function AdminLayout({
    children,
}: {
    children: React.ReactNode;
}) {
    const pathname = usePathname();

    return (
        <div className="min-h-screen bg-background">
            {/* Navigation */}
            <div className="border-b">
                <div className="container mx-auto">
                    <nav className="flex items-center gap-6 px-6 py-4">
                        <div className="font-semibold text-lg">Admin Panel</div>
                        <div className="flex items-center gap-1 flex-1">
                            {navigation.map((item) => {
                                const Icon = item.icon;
                                const isActive = pathname === item.href;

                                return (
                                    <Link
                                        key={item.href}
                                        href={item.href}
                                        className={cn(
                                            'flex items-center gap-2 px-3 py-2 rounded-md text-sm font-medium transition-colors',
                                            isActive
                                                ? 'bg-primary text-primary-foreground'
                                                : 'text-muted-foreground hover:text-foreground hover:bg-muted'
                                        )}
                                    >
                                        <Icon className="h-4 w-4" />
                                        {item.name}
                                    </Link>
                                );
                            })}
                        </div>
                    </nav>
                </div>
            </div>

            {/* Content */}
            <main>{children}</main>
        </div>
    );
}
