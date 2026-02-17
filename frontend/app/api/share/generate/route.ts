export const dynamic = 'force-dynamic';

import { type NextRequest, NextResponse } from 'next/server';
import { db } from '@/lib/db';
import { getServerSession } from 'next-auth';
import { authOptions } from '@/lib/auth/auth-options';
import { nanoid } from 'nanoid';
import bcrypt from 'bcryptjs';

export async function POST(req: NextRequest) {
    try {
        const session = await getServerSession(authOptions);
        if (!session?.user?.id) return NextResponse.json({ error: 'Unauthorized' }, { status: 401 });

        const body = await req.json();
        const { resourceType, resourceId, expiresAt, password } = body;

        if (!['DASHBOARD', 'QUERY'].includes(resourceType)) {
            return NextResponse.json({ error: 'Invalid resource type' }, { status: 400 });
        }

        // Validate resource ownership
        if (resourceType === 'DASHBOARD') {
            const dashboard = await db.dashboard.findUnique({ where: { id: resourceId } });
            if (!dashboard || dashboard.userId !== session.user.id) return NextResponse.json({ error: 'Not found' }, { status: 404 });
        } else {
            const query = await db.savedQuery.findUnique({ where: { id: resourceId } });
            if (!query || query.userId !== session.user.id) return NextResponse.json({ error: 'Not found' }, { status: 404 });
        }

        const shareToken = nanoid(10); // Short, URL-safe

        let hashedPassword = null;
        if (password) {
            hashedPassword = await bcrypt.hash(password, 10);
        }

        if (resourceType === 'DASHBOARD') {
            await db.shareLink.create({
                data: {
                    resourceType: 'DASHBOARD',
                    dashboardId: resourceId,
                    token: shareToken,
                    expiresAt: expiresAt ? new Date(expiresAt) : null,
                    password: hashedPassword,
                    // settings: settings || {}, // basic ShareLink model doesn't have settings yet, typical for MVP
                    userId: session.user.id
                }
            });
        } else {
            // For now only DASHBOARD supported as per previous logic, or extend ShareLink for QUERY
            await db.shareLink.create({
                data: {
                    resourceType: 'QUERY',
                    queryId: resourceId,
                    token: shareToken,
                    expiresAt: expiresAt ? new Date(expiresAt) : null,
                    password: hashedPassword,
                    userId: session.user.id
                }
            });
        }

        const fullLink = `${process.env.NEXT_PUBLIC_APP_URL}/share/${shareToken}`;

        return NextResponse.json({
            success: true,
            token: shareToken,
            link: fullLink
        });

    } catch (_error) {
        return NextResponse.json({ error: 'Failed to generate share link' }, { status: 500 });
    }
}
