export const dynamic = 'force-dynamic';

import { type NextRequest, NextResponse } from 'next/server';
import { getServerSession } from 'next-auth';
import { authOptions } from '@/lib/auth/auth-options';
import { db } from '@/lib/db';

export async function DELETE(
    request: NextRequest,
    props: { params: Promise<{ id: string }> }
) {
    try {
        const session = await getServerSession(authOptions);
        if (!session?.user?.id) {
            return NextResponse.json({ error: "Unauthorized" }, { status: 401 });
        }

        const params = await props.params;
        const { id } = params;

        // We check if the feed exists and if the current user created it OR has admin rights
        const feed = await db.queryFeed.findUnique({
            where: { id }
        });

        if (!feed) {
            return NextResponse.json({ error: "Feed not found" }, { status: 404 });
        }

        if (feed.createdBy !== session.user.id) {
            // Technically we should check workspace role too, but for MVP strict ownership is safer
            return NextResponse.json({ error: "Access denied" }, { status: 403 });
        }

        await db.queryFeed.delete({
            where: { id }
        });

        return NextResponse.json({ success: true });
    } catch (error: unknown) {
        return NextResponse.json(
            { error: error instanceof Error ? error.message : 'Internal Server Error' },
            { status: 500 }
        );
    }
}
