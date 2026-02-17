export const dynamic = 'force-dynamic';

import { NextResponse } from 'next/server';
import { db } from '@/lib/db';
import { getServerSession } from 'next-auth';
import { authOptions } from '@/lib/auth/auth-options';

const DEV_USER_ID = 'user_123';

export async function GET() {
    try {
        const session = await getServerSession(authOptions);
        // Note: For dev environment we might allowing seeing invalid sessions if we are auto-seeding
        // But strictly speaking /me should return the session user.
        // The previous logic seemed to mix Auth checks with seeding.
        // We will stick to the previous logic of "Find or Create Dev User" if specific email matches.

        // However, usually /me depends on session. 
        // If the goal is "get current user", we use session.
        // If the goal is "dev auto-login seed", maybe that belongs elsewhere?
        // But presuming the existing logic was desired:

        let user;

        // 1. Try to get explicit session user
        if (session?.user?.email) {
            user = await db.user.findUnique({
                where: { email: session.user.email }
            });
        }

        // 2. Fallback for Dev Environment (if no session or specific dev email)
        if (!user && process.env.NODE_ENV === 'development') {
            const existingByEmail = await db.user.findUnique({
                where: { email: 'dev@insightengine.ai' }
            });

            user = existingByEmail;

            if (!user) {
                const existingById = await db.user.findUnique({
                    where: { id: DEV_USER_ID }
                });
                user = existingById;
            }

            if (!user) {
                console.warn('[DevAuth] User not found, seeding user_123...');
                // Auto-seed for dev convenience
                user = await db.user.create({
                    data: {
                        id: DEV_USER_ID,
                        email: 'dev@insightengine.ai',
                        name: 'Developer Mode',
                        password: 'dev_password_hash',
                    },
                });
            }
        }

        if (!user) {
            return NextResponse.json({ error: 'Unauthorized' }, { status: 401 });
        }

        return NextResponse.json({ success: true, user });
    } catch (error: unknown) {
        console.error('[DevAuth] Error fetching user:', error);
        return NextResponse.json(
            { error: error instanceof Error ? error.message : 'Internal Server Error' },
            { status: 500 }
        );
    }
}
