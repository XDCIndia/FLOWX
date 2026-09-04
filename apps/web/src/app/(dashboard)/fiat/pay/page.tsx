'use client';

// Local payment simulator for the testnet model. The Flutterwave rail runs
// in mock mode (no real checkout exists locally), so deposit links point
// here instead of a dead placeholder domain. Clicking "Complete Payment"
// fires the provider webhook that completes the deposit and triggers the
// on-chain TXDC credit.

import { useState, useEffect } from 'react';
import { PageHeader } from '@/components/ui/page-header';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card';

const API_BASE = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:3000';

export default function FiatPayPage() {
  const [params, setParams] = useState<{ ref: string; amount: string; currency: string }>({
    ref: '',
    amount: '',
    currency: '',
  });
  const [state, setState] = useState<'idle' | 'paying' | 'done' | 'error'>('idle');
  const [message, setMessage] = useState('');

  useEffect(() => {
    const q = new URLSearchParams(window.location.search);
    setParams({
      ref: q.get('ref') || '',
      amount: q.get('amount') || '',
      currency: q.get('currency') || '',
    });
  }, []);

  const completePayment = async () => {
    setState('paying');
    setMessage('');
    try {
      const res = await fetch(`${API_BASE}/v1/webhooks/fiat/flutterwave`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          // Local mock mode: FLUTTERWAVE_WEBHOOK_HASH=local-dev-hash
          'verif-hash': 'local-dev-hash',
        },
        body: JSON.stringify({
          event: 'charge.completed',
          data: {
            id: Math.floor(Math.random() * 1e9),
            tx_ref: params.ref,
            status: 'successful',
            amount: Number(params.amount),
            currency: params.currency,
          },
        }),
      });
      if (!res.ok) {
        const body = await res.text();
        throw new Error(`webhook rejected (${res.status}): ${body}`);
      }
      setState('done');
      setMessage('Payment confirmed. The deposit is being credited in TXDC on-chain (watch the Wallets page).');
    } catch (err) {
      setState('error');
      setMessage(err instanceof Error ? err.message : 'payment failed');
    }
  };

  return (
    <div className="flex flex-col gap-8 animate-in fade-in slide-in-from-bottom-4 duration-500 max-w-lg mx-auto">
      <PageHeader title="Checkout (mock)" description="Local payment simulator for the XDC testnet model." />

      <Card>
        <CardHeader>
          <CardTitle>Flutterwave · simulated</CardTitle>
          <CardDescription>No real money moves here. This stands in for the hosted checkout.</CardDescription>
        </CardHeader>
        <CardContent className="flex flex-col gap-4">
          <div className="grid grid-cols-2 gap-4 text-sm">
            <div>
              <div className="text-muted-foreground">Amount</div>
              <div className="text-lg font-semibold">
                {params.amount ? `${params.amount} ${params.currency}` : '—'}
              </div>
            </div>
            <div>
              <div className="text-muted-foreground">Reference</div>
              <div className="font-mono text-xs break-all">{params.ref || '—'}</div>
            </div>
          </div>

          {state === 'done' ? (
            <div className="rounded-md bg-green-500/10 border border-green-500/30 p-3 text-sm text-green-600 dark:text-green-400">
              {message}
            </div>
          ) : state === 'error' ? (
            <div className="rounded-md bg-red-500/10 border border-red-500/30 p-3 text-sm text-red-600 dark:text-red-400">
              {message}
            </div>
          ) : null}

          <div className="flex gap-2">
            <Button onClick={completePayment} isLoading={state === 'paying'} disabled={!params.ref || state === 'done'}>
              Complete Payment
            </Button>
            <Button variant="ghost" onClick={() => (window.location.href = '/fiat')}>Back to Fiat</Button>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
