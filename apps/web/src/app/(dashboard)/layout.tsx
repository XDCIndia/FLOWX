'use client';

import { useEffect } from 'react';
import { useRouter } from 'next/navigation';
import { hasStoredApiKey, useAuth } from '@/lib/auth-context';
import Sidebar from '@/components/Sidebar';

export default function DashboardLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  const { isAuthenticated, isReady } = useAuth();
  const router = useRouter();

  // Only redirect once auth has been read from localStorage (post-hydration);
  // the server always renders the loading state because localStorage does not
  // exist there. The key is read directly rather than from context so the
  // check can never observe a stale pre-hydration value.
  useEffect(() => {
    if (isReady && !hasStoredApiKey()) {
      router.push('/login');
    }
  }, [isReady, isAuthenticated, router]);

  if (!isReady || !isAuthenticated) {
    return (
      <div className="flex min-h-screen items-center justify-center">
        <div className="text-sm text-muted-foreground">Loading...</div>
      </div>
    );
  }

  return (
    <div className="min-h-screen">
      <Sidebar />
      <main className="ml-64 min-h-screen p-6 lg:p-10">
        <div className="mx-auto max-w-7xl">{children}</div>
      </main>
    </div>
  );
}
