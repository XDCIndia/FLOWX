import { NextResponse } from 'next/server';
import type { NextRequest } from 'next/server';

export function middleware(request: NextRequest) {
  // Rewrite root path to serve the landing page
  if (request.nextUrl.pathname === '/') {
    return NextResponse.rewrite(new URL('/landing.html', request.url));
  }
  return NextResponse.next();
}

export const config = {
  matcher: ['/'],
};
