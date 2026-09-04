'use client';

import { createContext, useContext, useCallback, useSyncExternalStore, type ReactNode } from 'react';

interface AuthContextValue {
  apiKey: string | null;
  isAuthenticated: boolean;
  /** True once the client store has been read after hydration. */
  isReady: boolean;
  login: (apiKey: string) => void;
  logout: () => void;
  getStoredWalletIds: () => string[];
  addStoredWalletId: (id: string) => void;
}

const AuthContext = createContext<AuthContextValue | null>(null);

export const API_KEY_KEY = 'fluxa_api_key';
const WALLET_IDS_KEY = 'fluxa_wallet_ids';

/** Reads the current key straight from localStorage (client-only). */
export function hasStoredApiKey(): boolean {
  if (typeof window === 'undefined') return false;
  return localStorage.getItem(API_KEY_KEY) !== null;
}

// The API key lives in localStorage, which only exists on the client. Reading
// it synchronously on the server (or during hydration's first client render)
// is impossible, so useSyncExternalStore renders a null server snapshot and
// lets React re-check the real value right after hydration. That keeps the
// server HTML identical to the first client render (no hydration mismatch)
// and flips to the stored key in a second, post-hydration pass.
const localListeners = new Set<() => void>();

function subscribe(onStoreChange: () => void): () => void {
  // Same-tab writes (login/logout) notify through localListeners; other tabs
  // notify through the native storage event.
  window.addEventListener('storage', onStoreChange);
  localListeners.add(onStoreChange);
  return () => {
    window.removeEventListener('storage', onStoreChange);
    localListeners.delete(onStoreChange);
  };
}

function notifyLocalListeners() {
  localListeners.forEach((listener) => listener());
}

function readApiKey(): string | null {
  return localStorage.getItem(API_KEY_KEY);
}

// Server and hydration snapshot: never run on the client for the first pass.
function readApiKeyOnServer(): string | null {
  return null;
}

export function AuthProvider({ children }: { children: ReactNode }) {
  const apiKey = useSyncExternalStore(subscribe, readApiKey, readApiKeyOnServer);

  // True on the client, false on the server/hydration. React re-checks it
  // right after hydration, so guards can tell "still hydrating" from
  // "authenticated".
  const isReady = useSyncExternalStore(
    () => () => {},
    () => true,
    () => false
  );

  const login = useCallback((key: string) => {
    localStorage.setItem(API_KEY_KEY, key);
    notifyLocalListeners();
  }, []);

  const logout = useCallback(() => {
    localStorage.removeItem(API_KEY_KEY);
    localStorage.removeItem(WALLET_IDS_KEY);
    notifyLocalListeners();
  }, []);

  const getStoredWalletIds = useCallback((): string[] => {
    try {
      const raw = localStorage.getItem(WALLET_IDS_KEY);
      return raw ? JSON.parse(raw) : [];
    } catch {
      return [];
    }
  }, []);

  const addStoredWalletId = useCallback((id: string) => {
    const existing = getStoredWalletIds();
    if (!existing.includes(id)) {
      localStorage.setItem(WALLET_IDS_KEY, JSON.stringify([...existing, id]));
    }
  }, [getStoredWalletIds]);

  return (
    <AuthContext.Provider
      value={{
        apiKey,
        isAuthenticated: !!apiKey,
        isReady,
        login,
        logout,
        getStoredWalletIds,
        addStoredWalletId,
      }}
    >
      {children}
    </AuthContext.Provider>
  );
}

export function useAuth() {
  const ctx = useContext(AuthContext);
  if (!ctx) throw new Error('useAuth must be used within AuthProvider');
  return ctx;
}
