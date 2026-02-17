export const dynamic = 'force-dynamic';

import { type NextRequest, NextResponse } from 'next/server';
import { getServerSession } from 'next-auth';
import { authOptions } from '@/lib/auth/auth-options';
import { db } from '@/lib/db';
import {
    generateRegistrationOptions,
    verifyRegistrationResponse,
} from '@simplewebauthn/server';
import type { AuthenticatorTransport } from '@simplewebauthn/types';

// Domain (RP ID)
const rpID = process.env.NEXT_PUBLIC_RP_ID || 'localhost';
const origin = process.env.NEXT_PUBLIC_ORIGIN || 'http://localhost:3000';

export async function GET(_req: NextRequest) {
    const session = await getServerSession(authOptions);
    if (!session) {
        return new NextResponse('Unauthorized', { status: 401 });
    }

    try {
        const user = await db.user.findUnique({
            where: { id: session.user.id },
            include: { authenticators: true },
        });

        if (!user) {
            return new NextResponse('User not found', { status: 404 });
        }

        const options = await generateRegistrationOptions({
            rpName: 'InsightEngine',
            rpID,
            userID: new TextEncoder().encode(user.id),
            userName: user.email,
            // Don't re-register existing authenticators
            excludeCredentials: user.authenticators.map((authenticator) => ({
                id: authenticator.credentialID,
                transports: authenticator.transports ? (authenticator.transports.split(',') as AuthenticatorTransport[]) : undefined,
            })),
            authenticatorSelection: {
                residentKey: 'preferred',
                userVerification: 'preferred',
                authenticatorAttachment: 'platform', // Prefer built-in (TouchID/FaceID)
            },
        });

        const response = NextResponse.json(options);
        response.cookies.set('webauthn_challenge', options.challenge, {
            httpOnly: true,
            secure: process.env.NODE_ENV === 'production',
            sameSite: 'strict',
            maxAge: 60 * 5, // 5 minutes
        });

        return response;

    } catch (error) {
        console.error(error);
        return new NextResponse('Internal Server Error', { status: 500 });
    }
}

export async function POST(req: NextRequest) {
    const session = await getServerSession(authOptions);
    if (!session) {
        return new NextResponse('Unauthorized', { status: 401 });
    }

    const challenge = req.cookies.get('webauthn_challenge')?.value;
    if (!challenge) {
        return new NextResponse('Challenge expired', { status: 400 });
    }

    try {
        const body = await req.json();

        const verification = await verifyRegistrationResponse({
            response: body,
            expectedChallenge: challenge,
            expectedOrigin: origin,
            expectedRPID: rpID,
        });

        if (verification.verified && verification.registrationInfo) {
            const { credentialID, credentialPublicKey, counter, credentialDeviceType, credentialBackedUp } = verification.registrationInfo;

            await db.authenticator.create({
                data: {
                    credentialID,
                    credentialPublicKey: Buffer.from(credentialPublicKey),
                    counter: BigInt(counter),
                    credentialDeviceType,
                    credentialBackedUp,
                    transports: body.response.transports?.join(',') || 'internal',
                    userId: session.user.id,
                },
            });

            return NextResponse.json({ verified: true });
        } else {
            return new NextResponse('Verification failed', { status: 400 });
        }

    } catch (error) {
        console.error(error);
        return new NextResponse('Internal Error', { status: 500 });
    }
}
