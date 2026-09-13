'use client';

import { createContext, useContext, useState, useCallback, useEffect, useRef, type ReactNode } from 'react';
import { CheckCircle2, AlertCircle, Info } from 'lucide-react';
import { cn } from '@/lib/utils';

interface Toast {
  id: number;
  message: string;
  type: 'success' | 'error' | 'info';
}

interface ToastContextValue {
  toasts: Toast[];
  toast: (message: string, type?: Toast['type']) => void;
  dismiss: (id: number) => void;
}

const ToastContext = createContext<ToastContextValue | null>(null);

let nextId = 0;

const icons = {
  success: CheckCircle2,
  error: AlertCircle,
  info: Info,
};

const styles = {
  success: 'border-success/20 text-success bg-success-subtle',
  error: 'border-danger/20 text-danger bg-danger-subtle',
  info: 'border-primary/20 text-primary bg-primary-subtle',
};

export function ToastProvider({ children }: { children: ReactNode }) {
  const [toasts, setToasts] = useState<Toast[]>([]);
  // Latest-toast ref so the session-expired listener (registered once)
  // always calls the current toast implementation.
  const toastRef = useRef<(message: string, type?: Toast['type']) => void>(() => {});

  const dismiss = useCallback((id: number) => {
    setToasts((prev) => prev.filter((t) => t.id !== id));
  }, []);

  // api.ts is not a React module, so on 401 it dispatches this event instead
  // of calling useToast directly. Surface the toast, then the redirect to
  // /login happens from the fetch layer.
  useEffect(() => {
    const handler = () => {
      toastRef.current?.('Session expired — please sign in again', 'error');
    };
    window.addEventListener('flowx:session-expired', handler);
    return () => window.removeEventListener('flowx:session-expired', handler);
  }, []);

  const toast = useCallback(
    (message: string, type: Toast['type'] = 'info') => {
      const id = nextId++;
      setToasts((prev) => [...prev, { id, message, type }]);
      setTimeout(() => dismiss(id), 4000);
    },
    [dismiss]
  );
  toastRef.current = toast;

  return (
    <ToastContext.Provider value={{ toasts, toast, dismiss }}>
      {children}
      <div className="fixed bottom-6 right-6 z-[200] flex flex-col gap-3 pointer-events-none">
        {toasts.map((t) => {
          const Icon = icons[t.type];
          return (
            <div
              key={t.id}
              className={cn(
                'pointer-events-auto flex items-center gap-3 rounded-xl border px-4 py-3 text-sm font-medium shadow-lg animate-in fade-in slide-in-from-bottom-4 duration-300 cursor-pointer',
                styles[t.type]
              )}
              onClick={() => dismiss(t.id)}
            >
              <Icon className="h-4 w-4 shrink-0" />
              {t.message}
            </div>
          );
        })}
      </div>
    </ToastContext.Provider>
  );
}

export function useToast() {
  const ctx = useContext(ToastContext);
  if (!ctx) throw new Error('useToast must be used within ToastProvider');
  return ctx;
}
