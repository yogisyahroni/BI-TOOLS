import { type NextRequest, NextResponse } from 'next/server';
import { getToken } from 'next-auth/jwt';
import * as jose from 'jose';

export const dynamic = 'force-dynamic';

/**
 * GET /api/auth/token
 *
 * Returns an HS256-signed JWT that the Go backend can verify.
 * NextAuth stores tokens as JWE (encrypted), which the Go backend
 * cannot decode. This endpoint decodes the JWE, extracts the claims,
 * and re-signs them as a standard HS256 JWT using NEXTAUTH_SECRET.
 */
export async function GET(request: NextRequest) {
    try {
        const secret = process.env.NEXTAUTH_SECRET;
        if (!secret) {
            return NextResponse.json({ error: 'Server misconfiguration' }, { status: 500 });
        }

        // getToken (without raw: true) decodes the JWE and returns the claims object
        const decoded = await getToken({
            req: request as any,
            secret: secret,
        });

        if (!decoded) {
            return NextResponse.json({ error: 'Unauthorized' }, { status: 401 });
        }

        // Re-sign as HS256 JWT that the Go backend can parse
        const secretKey = new TextEncoder().encode(secret);
        const backendToken = await new jose.SignJWT({
            sub: decoded.sub || decoded.id as string,
            id: decoded.id as string,
            email: decoded.email as string,
            name: decoded.name as string,
            iat: Math.floor(Date.now() / 1000),
        })
            .setProtectedHeader({ alg: 'HS256' })
            .setExpirationTime('1h')
            .sign(secretKey);

        return NextResponse.json({ token: backendToken });
    } catch (error) {
        console.error('[API] Error getting token:', error);
        return NextResponse.json({ error: 'Internal Server Error' }, { status: 500 });
    }
}
