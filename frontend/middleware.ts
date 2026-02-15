import { withAuth } from "next-auth/middleware"
import { getToken } from "next-auth/jwt"
import { jwtVerify } from "jose"

export default withAuth(
    async function middleware(req) {
        // Custom logging or logic if needed
    },
    {
        callbacks: {
            async authorized({ token, req }) {
                // 1. If NextAuth successfully decoded it (unlikely with JWE mismatch), accept it
                if (token) return true;

                // 2. Fallback: Manually verify HS256 token using jose (Edge compatible)
                try {
                    const secret = process.env.NEXTAUTH_SECRET;
                    if (!secret) return false;

                    // Get raw token string
                    const rawToken = await getToken({ req, secret, raw: true });
                    if (!rawToken) return false;

                    // Verify signature
                    const encodedSecret = new TextEncoder().encode(secret);
                    await jwtVerify(rawToken as string, encodedSecret, { algorithms: ['HS256'] });

                    console.log(`[MIDDLEWARE] HS256 Token verified for path: ${req.nextUrl.pathname}`);
                    return true;
                } catch (error) {
                    console.error(`[MIDDLEWARE] Token validation failed:`, error);
                    return false;
                }
            }
        },
        pages: {
            signIn: '/auth/signin',
        },
    }
)

export const config = {
    // Protect all routes except auth, api, static files, and root
    matcher: [
        /*
         * Match all request paths except for the ones starting with:
         * - / (root path - handles its own redirect logic)
         * - auth (authentication routes)
         * - api (API routes, handled separately or by backend)
         * - _next/static (static files)
         * - _next/image (image optimization files)
         * - favicon.ico (favicon file)
         * - manifest.webmanifest (PWA manifest)
         * - icon-*.png, icon.svg (PWA icons)
         * 
         * CRITICAL: /dashboards IS NOW PROTECTED (removed from exclusion)
         */
        "/((?!auth|api|_next/static|_next/image|favicon.ico|manifest.webmanifest|icon-|$).*)"
    ]
}

