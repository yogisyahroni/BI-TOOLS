import { type NextRequest, NextResponse } from "next/server";
import { getToken } from "next-auth/jwt";
import * as jose from "jose";

export const dynamic = "force-dynamic";

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
      return NextResponse.json({ error: "Server misconfiguration" }, { status: 500 });
    }

    // DEBUG LOGGING
    console.log("[API] /api/auth/token called");
    console.log(
      "[API] Cookies:",
      request.cookies.getAll().map((c) => c.name),
    );
    console.log("[API] NEXTAUTH_SECRET exists:", !!secret);

    // getToken (without raw: true) decodes the JWE and returns the claims object
    const decoded = await getToken({
      req: request,
      secret: secret,
    });

    console.log("[API] Decoded token:", decoded ? "FOUND" : "NULL");

    if (!decoded) {
      console.error("[API] /api/auth/token Unauthorized: No valid session token found");
      console.error("[API] Debug Info:");
      console.error(
        "[API] - Cookies Received:",
        request.cookies.getAll().map((c) => `${c.name}=${c.value.substring(0, 10)}...`),
      );
      console.error("[API] - NEXTAUTH_URL:", process.env.NEXTAUTH_URL);
      console.error(
        "[API] - Secure Cookie Prefix expected:",
        process.env.NEXTAUTH_URL?.startsWith("https") ? "__Secure-" : "None",
      );
      return NextResponse.json(
        { error: "Unauthorized", debug: "No valid session token" },
        { status: 401 },
      );
    }

    // Return the original backend access token if available
    if (decoded.accessToken) {
      return NextResponse.json({ token: decoded.accessToken });
    }

    console.warn("[API] No accessToken in JWT, falling back to resigned token");

    // Fallback: Re-sign as HS256 JWT (only if backend supports this, otherwise this path might still fail)
    const secretKey = new TextEncoder().encode(secret);
    const backendToken = await new jose.SignJWT({
      sub: decoded.sub || (decoded.id as string),
      id: decoded.id as string,
      email: decoded.email as string,
      name: decoded.name as string,
      iat: Math.floor(Date.now() / 1000),
    })
      .setProtectedHeader({ alg: "HS256" })
      .setExpirationTime("1h")
      .sign(secretKey);

    return NextResponse.json({ token: backendToken });
  } catch (error) {
    console.error("[API] Error getting token:", error);
    return NextResponse.json({ error: "Internal Server Error" }, { status: 500 });
  }
}
