import { clsx, type ClassValue } from 'clsx'
import { twMerge } from 'tailwind-merge'

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs))
}

// Enhanced fetch function that includes authentication
export async function fetchWithAuth(url: string, options: RequestInit = {}) {
  // Use the URL as-is since Next.js rewrites handle the forwarding to backend
  // The url should be something like '/api/go/...' which gets rewritten to backend
  let origin = 'http://localhost:3000';
  if (typeof window !== 'undefined') {
    origin = window.location.origin;
  } else if (process.env.NEXT_PUBLIC_APP_URL) {
    origin = process.env.NEXT_PUBLIC_APP_URL;
  }

  const fullUrl = url.startsWith('http') ? url : `${origin}${url}`;

  const headers = {
    'Content-Type': 'application/json',
    ...options.headers,
  };

  // Make the request with credentials included to send NextAuth session cookies
  const response = await fetch(fullUrl, {
    ...options,
    headers,
    credentials: 'include', // This sends NextAuth session cookies with the request
  });

  // Hanya tangani 401 jika ini bukan permintaan otentikasi
  if (response.status === 401 && !url.includes('/api/auth/')) {
    // Jangan redirect otomatis karena bisa menyebabkan infinite loop
    // Biarkan komponen yang menangani error ini sesuai kebutuhan
    console.warn(`Unauthorized access to ${url}. Status: ${response.status}`);
    // Log tambahan untuk debugging
    // console.log('Session cookies available:', document.cookie.includes('next-auth'));
  }

  return response;
}
