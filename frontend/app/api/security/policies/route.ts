export const dynamic = 'force-dynamic';

import { type NextRequest, NextResponse } from 'next/server';
import { db as prisma } from '@/lib/db';

export async function POST(req: NextRequest) {
    try {
        const body = await req.json();
        const { name, workspaceId, connectionId, tableName, condition, role, userId } = body;

        if (!name || !workspaceId || !connectionId || !tableName || !condition) {
            return new NextResponse('Missing required fields', { status: 400 });
        }

        const policy = await prisma.rLSPolicy.create({
            data: {
                name,
                workspaceId,
                connectionId,
                tableName,
                condition,
                role: role || null,
                userId: userId || null,
                isActive: true
            }
        });

        return NextResponse.json(policy);
    } catch (error) {
        console.error('[RLS_POST]', error);
        return new NextResponse('Internal Error', { status: 500 });
    }
}

export async function GET(req: NextRequest) {
    try {
        const { searchParams } = new URL(req.url);
        const workspaceId = searchParams.get('workspaceId');

        if (!workspaceId) {
            return new NextResponse('Workspace ID required', { status: 400 });
        }

        const policies = await prisma.rLSPolicy.findMany({
            where: { workspaceId },
            include: {
                user: { select: { name: true, email: true } }
            },
            orderBy: { createdAt: 'desc' }
        });

        return NextResponse.json(policies);
    } catch (error) {
        console.error('[RLS_GET]', error);
        return new NextResponse('Internal Error', { status: 500 });
    }
}
