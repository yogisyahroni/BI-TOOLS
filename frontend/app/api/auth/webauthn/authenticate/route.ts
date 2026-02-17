export const dynamic = 'force-dynamic';

import { generateAuthenticationOptions, verifyAuthenticationResponse } from '@simplewebauthn/server';
import { type NextRequest, NextResponse } from 'next/server';
import { db } from '@/lib/db';
import { getServerSession } from 'next-auth';
import { authOptions } from '@/lib/auth/auth-options';
import { type AuthenticatorTransport } from '@simplewebauthn/types';

export async function GET(_req: NextRequest) {
    const session = await getServerSession(authOptions);
    if (!session?.user?.email) {
        return NextResponse.json({ error: 'Unauthorized' }, { status: 401 });
    }

    const user = await db.user.findUnique({
        where: { email: session.user.email },
        include: { authenticators: true },
    });

    if (!user) {
        return NextResponse.json({ error: 'User not found' }, { status: 404 });
    }

    const options = await generateAuthenticationOptions({
        rpID: process.env.NEXT_PUBLIC_RP_ID || 'localhost',
        allowCredentials: user.authenticators.map((authenticator) => ({
            id: authenticator.credentialID,
            type: 'public-key',
            transports: (authenticator.transports ? authenticator.transports.split(',') : []) as AuthenticatorTransport[],
        })),
        userVerification: 'preferred',
    });

    const response = NextResponse.json(options);
    response.cookies.set('webauthn_challenge', options.challenge, {
        httpOnly: true,
        secure: process.env.NODE_ENV === 'production',
        sameSite: 'strict',
        maxAge: 60 * 5, // 5 minutes
    });

    return response;
}

export async function POST(req: NextRequest) {
    const challenge = req.cookies.get('webauthn_challenge')?.value;
    if (!challenge) {
        return NextResponse.json({ error: 'Challenge not found' }, { status: 400 });
    }

    const body = await req.json();
    const { email } = body;

    // Note: In a real flow, you might identify the user by the credential ID in the response
    // rather than relying on an email in the body, or rely on a temporary session. 
    // For now, assuming email is passed or we can look it up.

    if (!email) {
        return NextResponse.json({ error: 'Email required' }, { status: 400 });
    }

    const user = await db.user.findUnique({
        where: { email },
        include: { authenticators: true },
    });

    if (!user) {
        return NextResponse.json({ error: 'User not found' }, { status: 404 });
    }

    const authenticator = user.authenticators.find(
        (auth) => auth.credentialID === body.id
    );

    if (!authenticator) {
        return NextResponse.json({ error: 'Authenticator not registered' }, { status: 400 });
    }

    let verification;
    try {
        verification = await verifyAuthenticationResponse({
            response: body,
            expectedChallenge: challenge,
            expectedOrigin: process.env.NEXT_PUBLIC_ORIGIN || 'http://localhost:3000',
            expectedRPID: process.env.NEXT_PUBLIC_RP_ID || 'localhost',

            credential: {
                id: authenticator.credentialID,
                publicKey: new Uint8Array(authenticator.credentialPublicKey),
                counter: Number(authenticator.counter),
                // eslint-disable-next-line no-nested-ternary
                transports: (authenticator.transports ? authenticator.transports.split(',') : []) as AuthenticatorTransport[],
            },
        });
    } catch (error) {
        const message = error instanceof Error ? error.message : 'Verification failed';
        return NextResponse.json({ error: message }, { status: 400 });
    }

    const { verified } = verification;

    if (verified) {
        const { authenticationInfo } = verification;
        await db.authenticator.update({
            where: { credentialID: authenticator.credentialID },
            data: {
                counter: BigInt(authenticationInfo.newCounter),
            },
        });

        // Here we would typically set a session cookie or return a JWT
        // Since we are using NextAuth, the client would use the result to sign in via signIn('credentials', ...)
        // providing a custom token or similar. 
        // For this implementation, we just return success: true.

        return NextResponse.json({ verified: true });
    }

    return NextResponse.json({ verified: false }, { status: 400 });
}
