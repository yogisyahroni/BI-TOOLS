import { NextRequest, NextResponse } from 'next/server';
import { getToken } from 'next-auth/jwt';

const API_BASE_URL = process.env.GO_BACKEND_URL || 'http://127.0.0.1:8080';

export async function POST(request: NextRequest) {
  try {
    // Get JWT token from request cookies
    const token = await getToken({ 
      req: request, 
      secret: process.env.NEXTAUTH_SECRET 
    });
    
    if (!token) {
      return NextResponse.json(
        { error: 'Unauthorized' },
        { status: 401 }
      );
    }

    const body = await request.json();

    // Get the raw JWT token from the session cookie
    // The encoded JWT is what the Go backend expects
    const cookieHeader = request.headers.get('cookie') || '';
    const sessionToken = cookieHeader
      .split(';')
      .find(c => c.trim().startsWith('next-auth.session-token='))
      ?.split('=')[1];

    if (!sessionToken) {
      return NextResponse.json(
        { error: 'No session token found' },
        { status: 401 }
      );
    }

    // Forward request to Go backend with the session token
    const response = await fetch(`${API_BASE_URL}/api/analytics/insights`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${sessionToken}`,
        'Cookie': cookieHeader,
      },
      body: JSON.stringify(body),
    });

    if (!response.ok) {
      const error = await response.text();
      return NextResponse.json(
        { error: error || 'Failed to generate insights' },
        { status: response.status }
      );
    }

    const data = await response.json();
    return NextResponse.json(data);
  } catch (error) {
    console.error('Analytics insights error:', error);
    return NextResponse.json(
      { error: 'Internal server error' },
      { status: 500 }
    );
  }
}
