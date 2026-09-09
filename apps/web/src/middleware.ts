import { NextResponse } from 'next/server';
import type { NextRequest } from 'next/server';

// Routes that don't require authentication
const PUBLIC_PATHS = ['/', '/login', '/landing.html', '/fiat/pay'];

export function middleware(request: NextRequest) {
  const { pathname } = request.nextUrl;

  // Rewrite root path to serve the landing page
  if (pathname === '/') {
    return NextResponse.rewrite(new URL('/landing.html', request.url));
  }

  // Allow public paths
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
