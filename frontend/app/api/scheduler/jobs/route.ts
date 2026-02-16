import { type NextRequest, NextResponse } from 'next/server';
import { getToken } from 'next-auth/jwt';

const API_BASE_URL = process.env.GO_BACKEND_URL || 'http://127.0.0.1:8080';

// Proxy scheduler jobs API dengan fallback
export async function GET(request: NextRequest) {
  try {
    const token = await getToken({ req: request, secret: process.env.NEXTAUTH_SECRET });
    
    if (!token) {
      return NextResponse.json({ error: 'Unauthorized' }, { status: 401 });
    }

    // Try to fetch from backend
    try {
      const response = await fetch(`${API_BASE_URL}/api/scheduler/jobs`, {
        headers: {
          'Authorization': `Bearer ${token.sub}`,
        },
        signal: AbortSignal.timeout(5000),
      });

      if (response.ok) {
        const data = await response.json();
        return NextResponse.json(data);
      }
    } catch (error) {
      console.warn('Backend scheduler not available, returning empty array');
    }

    // Return empty array as fallback
    return NextResponse.json([]);
  } catch (error) {
    console.error('Scheduler API error:', error);
    return NextResponse.json([]);
  }
}
