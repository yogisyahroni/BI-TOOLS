export const dynamic = 'force-dynamic';

import { type NextRequest, NextResponse } from 'next/server';
import { getServerSession } from 'next-auth';
import { authOptions } from '@/lib/auth/auth-options';
import { db } from '@/lib/db';

export async function GET(_req: NextRequest) {
    const session = await getServerSession(authOptions);
    if (!session?.user?.id) {
        return new NextResponse('Unauthorized', { status: 401 });
    }

    try {
        const authenticators = await db.authenticator.findMany({
            where: { userId: session.user.id },
            select: {
                credentialID: true,
                credentialDeviceType: true,
                credentialBackedUp: true,
                transports: true,
            }
        });

        const friendlyAuths = authenticators.map(auth => ({
            id: auth.credentialID,
            type: auth.credentialDeviceType === 'singleDevice' ? 'Device-Bound' : 'Synced (Passkey)',
            backedUp: auth.credentialBackedUp,
            transports: auth.transports,
            name: 'Passkey',
            createdAt: new Date(),
        }));

        return NextResponse.json(friendlyAuths);
    } catch (error) {
        console.error(error);
        return new NextResponse('Internal Error', { status: 500 });
    }
}

export async function DELETE(req: NextRequest) {
    const session = await getServerSession(authOptions);
    if (!session?.user?.id) {
        return new NextResponse('Unauthorized', { status: 401 });
    }

    try {
        const { credentialID } = await req.json();

        const auth = await db.authenticator.findUnique({
            where: { credentialID },
        });

        if (!auth || auth.userId !== session.user.id) {
            return new NextResponse('Not found or unauthorized', { status: 404 });
        }

        await db.authenticator.delete({
            where: { credentialID },
        });

        return NextResponse.json({ success: true });
    } catch (error) {
        console.error(error);
        return new NextResponse('Internal Error', { status: 500 });
    }
}
