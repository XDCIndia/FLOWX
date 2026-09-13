'use client';

/**
 * Typed SWR hooks wrapping the FlowX API fetcher (lib/api.ts).
 *
 * Refresh policy:
 *  - 5s  for resources with in-flight activity (pending transfers, settling batches)
 *  - 10s for balances (they move often in demos)
 *  - 30s for everything else (wallets, schedules, rates, health)
 *
 * All mutation helpers revalidate the affected keys so pages auto-update
 * without a manual refresh button.
 */

import useSWR, { type SWRConfiguration, type KeyedMutator } from 'swr';
import { useSWRConfig } from 'swr';
import { useCallback } from 'react';
import {
  api,
  type Wallet,
  type WalletBalance,
  type WalletWithBalance,
  type Transaction,
  type BatchResponse,
  type ScheduleResponse,
  type RateResponse,
  type HealthResponse,
  type FeeSchedule,
  type PaymentQuoteRequest,
  type PaymentQuoteResponse,
  type PaymentExecuteRequest,
  type PaymentExecuteResponse,
} from '@/lib/api';

// --- Stable cache keys (exported so callers can target revalidation) ---
export const SWR_KEYS = {
  wallets: '/v1/wallets/',
  health: '/health',
  fees: '/v1/fees',
  schedules: '/v1/schedules/',
} as const;

const IN_FLIGHT_MS = 5_000;
const BALANCE_MS = 10_000;
const DEFAULT_MS = 30_000;

const BASE_OPTS: SWRConfiguration = {
  revalidateOnFocus: true,
  revalidateOnReconnect: true,
  dedupingInterval: 2_000,
  keepPreviousData: true,
  errorRetryCount: 2,
};

/** Statuses that mean money is still moving — poll these faster. */
const IN_FLIGHT_STATUSES = new Set(['pending', 'processing', 'submitted', 'queued']);

/** Terminal batch statuses — stop polling once reached. */
export const BATCH_TERMINAL_STATUSES = new Set([
  'completed',
  'partial',
  'failed',
  'compliance_hold',
]);

export interface ApiHook<T> {
  data: T | undefined;
  isLoading: boolean;
  error: Error | undefined;
  mutate: KeyedMutator<T>;
}

function useApiHook<T>(key: string | null, fetcher: () => Promise<T>, opts?: SWRConfiguration): ApiHook<T> {
  const { data, error, isLoading, mutate } = useSWR<T, Error>(key, fetcher, { ...BASE_OPTS, ...opts });
  return { data, isLoading, error, mutate };
}

// --- Health / fees ---

export function useHealth(): ApiHook<HealthResponse> {
  return useApiHook(SWR_KEYS.health, () => api.getHealth(), { refreshInterval: DEFAULT_MS });
}

export function useFeeSchedule(): ApiHook<FeeSchedule> {
  return useApiHook(SWR_KEYS.fees, () => api.getFeeSchedule(), { refreshInterval: DEFAULT_MS });
}

// --- Wallets ---

export interface WalletsData {
  wallets: Wallet[];
}

export function useWallets(): ApiHook<WalletsData> & { wallets: Wallet[] } {
  const hook = useApiHook<WalletsData>(SWR_KEYS.wallets, () => api.listWallets(), {
    refreshInterval: DEFAULT_MS,
  });
  return { ...hook, wallets: hook.data?.wallets ?? [] };
}

export function useWalletBalances(walletId: string | null): ApiHook<{ wallet_id: string; balances: WalletBalance[] }> & { balances: WalletBalance[] } {
  const key = walletId ? `/v1/wallets/${walletId}/balances` : null;
  const hook = useApiHook(key, () => api.getWalletBalances(walletId as string), {
    refreshInterval: BALANCE_MS,
  });
  return { ...hook, balances: hook.data?.balances ?? [] };
}

/**
 * Wallets joined with their balances — one hook for card-grid pages.
 * Falls back per-wallet on balance-fetch failure so a single error
 * never hides the wallet itself.
 */
export function useWalletsWithBalances(walletIds: string[]): ApiHook<WalletWithBalance[]> {
  const joined = walletIds.join(',');
  const key = walletIds.length > 0 ? `/v1/wallets/with-balances:${joined}` : null;
  const fetcher = useCallback(async (): Promise<WalletWithBalance[]> => {
    return Promise.all(
      walletIds.map(async (id) => {
        try {
          const wallet = await api.getWallet(id);
          let balances: WalletBalance[] = [];
          try {
            const res = await api.getWalletBalances(id);
            balances = res.balances;
          } catch {
            // Balance fetch failure must not hide the wallet itself.
          }
          return { ...wallet, balances } as WalletWithBalance;
        } catch {
          return { id, public_key: id, created_at: '', balances: [] } as WalletWithBalance;
        }
      })
    );
  }, [joined]); // eslint-disable-line react-hooks/exhaustive-deps
  const hook = useApiHook(key, fetcher, {
    refreshInterval: BALANCE_MS,
  });
  return { ...hook, data: hook.data ?? [] };
}

// --- Transfers / transactions ---

export function useTransaction(id: string | null): ApiHook<Transaction> {
  const key = id ? `/v1/transfers/${id}` : null;
  return useApiHook(key, () => api.getTransaction(id as string), {
    refreshInterval: (data: Transaction | undefined) =>
      data && IN_FLIGHT_STATUSES.has(data.status) ? IN_FLIGHT_MS : DEFAULT_MS,
  });
}

export function useTransactions(
  walletId: string | null,
  limit = 50
): ApiHook<{ transactions: Transaction[] }> & { transactions: Transaction[] } {
  const key = walletId ? `/v1/transactions?wallet_id=${walletId}&limit=${limit}` : null;
  const hook = useApiHook(key, () => api.listTransactions(walletId as string, limit), {
    refreshInterval: (data: { transactions: Transaction[] } | undefined) =>
      data?.transactions?.some((t) => IN_FLIGHT_STATUSES.has(t.status)) ? IN_FLIGHT_MS : DEFAULT_MS,
  });
  return { ...hook, transactions: hook.data?.transactions ?? [] };
}

/**
 * Merged, de-duplicated transaction stream across every wallet the user
 * owns — what the Transfers page shows. Polls fast while anything is
 * still in flight, otherwise settles to the slow interval.
 */
export function useAllTransactions(
  walletIds: string[],
  limit = 50
): ApiHook<Transaction[]> & { transactions: Transaction[] } {
  const key = walletIds.length > 0 ? `/v1/transactions/all:${walletIds.join(',')}:${limit}` : null;
  const fetcher = useCallback(async (): Promise<Transaction[]> => {
    const all: Transaction[] = [];
    for (const id of walletIds) {
      try {
        const res = await api.listTransactions(id, limit);
        all.push(...(res.transactions || []));
      } catch {
        // A single failing wallet must not blank the whole history.
      }
    }
    const unique = Array.from(new Map(all.map((t) => [t.id, t])).values());
    unique.sort((a, b) => new Date(b.created_at).getTime() - new Date(a.created_at).getTime());
    return unique;
  }, [walletIds.join(',')]); // eslint-disable-line react-hooks/exhaustive-deps
  const hook = useApiHook(key, fetcher, {
    refreshInterval: (data: Transaction[] | undefined) =>
      data?.some((t) => IN_FLIGHT_STATUSES.has(t.status)) ? IN_FLIGHT_MS : DEFAULT_MS,
  });
  return { ...hook, transactions: hook.data ?? [] };
}

// --- Batches ---

/**
 * Batch detail with live status. Keeps polling at 5s while the batch is
 * settling and stops at a terminal status.
 */
export function useBatch(id: string | null): ApiHook<BatchResponse> {
  const key = id ? `/v1/transfers/batch/${id}` : null;
  return useApiHook(key, () => api.getBatch(id as string), {
    refreshInterval: (data: BatchResponse | undefined) => {
      if (!data || BATCH_TERMINAL_STATUSES.has(data.status)) return 0;
      return IN_FLIGHT_MS;
    },
  });
}

// --- Schedules ---

export function useSchedules(): ApiHook<{ schedules: ScheduleResponse[] }> & { schedules: ScheduleResponse[] } {
  const hook = useApiHook(SWR_KEYS.schedules, () => api.listSchedules(), {
    refreshInterval: DEFAULT_MS,
  });
  return { ...hook, schedules: hook.data?.schedules ?? [] };
}

// --- FX ---

export function useFxRates(from: string | null, to: string | null): ApiHook<RateResponse> {
  const key = from && to ? `/v1/fx/rates?from=${encodeURIComponent(from)}&to=${encodeURIComponent(to)}` : null;
  return useApiHook(key, () => api.getRates(from as string, to as string), {
    refreshInterval: DEFAULT_MS,
  });
}

// --- Payment routes ---

export interface PaymentQuoteVariables extends PaymentQuoteRequest {}

export interface PaymentExecuteVariables extends PaymentExecuteRequest {}

export interface PaymentQuoteHook {
  quote: PaymentQuoteResponse | null;
  isLoading: boolean;
  error: Error | undefined;
  getQuote: (vars: PaymentQuoteVariables) => Promise<PaymentQuoteResponse>;
  reset: () => void;
}

/**
 * On-demand payment-route quote. Quotes expire server-side (~30s) and
 * are user-triggered, so this is a manual mutation rather than a poll.
 */
export function usePaymentQuote(): PaymentQuoteHook {
  const key = '/v1/payments/quote';
  // No fetcher: quotes are user-triggered and expire server-side, so SWR
  // only stores the result — getQuote drives the network call via mutate.
  const { data, error, isLoading, mutate } = useSWR<PaymentQuoteResponse | undefined, Error>(
    key,
    {
      revalidateOnFocus: false,
      revalidateOnReconnect: false,
      revalidateOnMount: false,
    }
  );
  const getQuote = useCallback(
    async (vars: PaymentQuoteVariables) => {
      const result = await mutate(() => api.getPaymentQuote(vars), { revalidate: false });
      return result as PaymentQuoteResponse;
    },
    [mutate]
  );
  const reset = useCallback(() => mutate(undefined, { revalidate: false }), [mutate]);
  return { quote: data ?? null, isLoading, error, getQuote, reset };
}

export interface PaymentExecuteHook {
  result: PaymentExecuteResponse | null;
  isLoading: boolean;
  error: Error | undefined;
  execute: (vars: PaymentExecuteVariables) => Promise<PaymentExecuteResponse>;
  reset: () => void;
}

/**
 * Execute a chosen route, then revalidate wallets and transactions so
 * balances and history reflect the payment immediately.
 */
export function useExecuteRoute(): PaymentExecuteHook {
  const { mutate: globalMutate } = useSWRConfig();
  const key = '/v1/payments/send';
  // No fetcher: execution is user-triggered; see usePaymentQuote above.
  const { data, error, isLoading, mutate } = useSWR<PaymentExecuteResponse | undefined, Error>(
    key,
    {
      revalidateOnFocus: false,
      revalidateOnReconnect: false,
      revalidateOnMount: false,
    }
  );
  const execute = useCallback(
    async (vars: PaymentExecuteVariables) => {
      const result = await mutate(() => api.executePaymentRoute(vars), { revalidate: false });
      // Money moved — refresh balances and history everywhere.
      await globalMutate((k) => typeof k === 'string' && k.startsWith('/v1/wallets'));
      await globalMutate((k) => typeof k === 'string' && k.startsWith('/v1/transactions'));
      return result as PaymentExecuteResponse;
    },
    [mutate, globalMutate]
  );
  const reset = useCallback(() => mutate(undefined, { revalidate: false }), [mutate]);
  return { result: data ?? null, isLoading, error, execute, reset };
}

/**
 * Revalidate every core resource after any mutation. Call after
 * create/delete/update operations that the API doesn't push to us.
 */
export function useRevalidateCore(): () => Promise<void> {
  const { mutate } = useSWRConfig();
  return useCallback(async () => {
    await mutate((k) => typeof k === 'string' && (k.startsWith('/v1/wallets') || k.startsWith('/v1/transactions') || k.startsWith(SWR_KEYS.schedules)));
  }, [mutate]);
}
