import { type NextAuthOptions } from 'next-auth';
import CredentialsProvider from 'next-auth/providers/credentials';
import GoogleProvider from 'next-auth/providers/google';

export const authOptions: NextAuthOptions = {
    providers: [
        // Google OAuth2 Provider (TASK-007) - DISABLED until credentials are configured
        // Uncomment and configure when you have valid Google OAuth credentials
        /*
        GoogleProvider({
            clientId: process.env.GOOGLE_CLIENT_ID || '',
            clientSecret: process.env.GOOGLE_CLIENT_SECRET || '',
            authorization: {
                params: {
                    prompt: "consent",
                    access_type: "offline",
                    response_type: "code"
                }
            },
            profile(profile) {
                return {
                    id: profile.sub,
                    name: profile.name,
                    email: profile.email,
                    image: profile.picture,
                };
            },
        }),
        */
        // Credentials Provider (Password-based login)
        CredentialsProvider({
            name: 'Credentials',
            credentials: {
                email: { label: 'Email', type: 'email' },
                password: { label: 'Password', type: 'password' },
                webauthn_token: { label: 'WebAuthn Token', type: 'text' },
            },
            async authorize(credentials) {
                console.warn('[AUTH] Authorize called for:', credentials?.email);

                // 1. WebAuthn Flow
                if (credentials?.webauthn_token) {
                    try {
                        const { decode } = await import('next-auth/jwt');
                        const decoded = await decode({
                            token: credentials.webauthn_token,
                            secret: process.env.NEXTAUTH_SECRET || '',
                        });
                        if (decoded && decoded.usage === 'webauthn_login' && decoded.id && decoded.email) {
                            return {
                                id: decoded.id as string,
                                email: decoded.email as string,
                                name: 'User',
                            };
                        }
                        return null;
                    } catch (e) {
                        console.error('[AUTH] WebAuthn token validation failed', e);
                        return null;
                    }
                }

                // 2. Standard Email/Password Flow via Go Backend
                if (!credentials?.email || !credentials?.password) {
                    return null;
                }

                try {
                    const res = await fetch('http://127.0.0.1:8080/api/auth/login', {
                        method: 'POST',
                        body: JSON.stringify({
                            email: credentials.email,
                            password: credentials.password,
                        }),
                        headers: { 'Content-Type': 'application/json' },
                    });

                    // Handle non-JSON responses (e.g., 500 HTML error pages from backend)
                    const text = await res.text();
                    let data;
                    try {
                        data = JSON.parse(text);
                    } catch (parseError) {
                        console.error('[AUTH] Failed to parse backend response:', text);
                        return null;
                    }

                    if (!res.ok) {
                        console.error('[AUTH] Login failed:', data);
                        // If backend returns 401/403, return null to signal invalid credentials
                        // Do NOT throw, otherwise NextAuth returns 500
                        return null;
                    }

                    if (data.user) {
                        console.warn('[AUTH] Login successful for:', data.user.email);
                        return {
                            id: String(data.user.id),
                            email: data.user.email,
                            name: data.user.name || data.user.email.split('@')[0],
                        };
                    }

                    console.error('[AUTH] No user data in response:', data);
                    return null;
                } catch (error) {
                    console.error('[AUTH] Error during authorization:', error);
                    return null;
                }
            },
        }),
    ],
    session: {
        strategy: 'jwt',
        maxAge: 30 * 24 * 60 * 60, // 30 days
        updateAge: 24 * 60 * 60,   // Refresh session every 24 hours
    },

    pages: {
        signIn: '/auth/signin',
        signOut: '/auth/signout',
        error: '/auth/error',
    },
    callbacks: {
        async redirect({ url, baseUrl }) {
            if (url.startsWith('/')) {
                return `${baseUrl}${url}`;
            }
            if (url.startsWith(baseUrl)) {
                return url;
            }
            try {
                if (new URL(url).origin === baseUrl) {
                    return url;
                }
            } catch {
                // Invalid URL, fallback to default
            }
            return `${baseUrl}/dashboards`;
        },
        async jwt({ token, user, account }) {
            if (user) {
                token.id = user.id;
            }
            if (account?.access_token) {
                token.accessToken = account.access_token;
            }
            return token;
        },
        async session({ session, token }) {
            if (session.user) {
                (session.user as any).id = token.id as string;
            }
            if (token.accessToken) {
                (session as any).accessToken = token.accessToken;
            }
            return session;
        },
    },

    secret: process.env.NEXTAUTH_SECRET,
    debug: false,
};
