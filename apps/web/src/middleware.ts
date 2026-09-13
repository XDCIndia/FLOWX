import { NextResponse } from 'next/server';
import type { NextRequest } from 'next/server';

// Routes that don't require authentication
const PUBLIC_PATHS = ['/', '/login', '/fiat/pay'];

export function middleware(request: NextRequest) {
  const { pathname } = request.nextUrl;

  // Allow public paths ('/' serves the marketing landing in app/page.tsx)
  if (PUBLIC_PATHS.includes(pathname)) {
    return NextResponse.next();
  }

  // Protect everything else — require an API key cookie
  const apiKey = request.cookies.get('flowx_api_key')?.value;
  if (!apiKey) {
    return NextResponse.redirect(new URL('/login', request.url));
  }

  return NextResponse.next();
}

export const config = {
  matcher: ['/((?!_next|api|favicon.ico).*)'],
};
