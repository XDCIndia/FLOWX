'use client';

import { useState, useEffect } from 'react';
import { api, type QuoteResponse, type RateResponse } from '@/lib/api';
import { useToast } from '@/lib/toast-context';
import { PageHeader } from '@/components/ui/page-header';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Select } from '@/components/ui/select';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Coins, ArrowRightLeft } from 'lucide-react';

export default function FXPage() {
  const { toast } = useToast();
  const [wallets, setWallets] = useState<{id: string; public_key: string}[]>([]);

  useEffect(() => {
    fetch((process.env.NEXT_PUBLIC_API_URL || 'http://localhost:3000') + '/v1/wallets', {
      headers: { Authorization: `Bearer ${localStorage.getItem('flowx_api_key')}` },
    })
      .then(r => r.json())
      .then(data => {
        const list = data.wallets || data || [];
        setWallets(list.map((w: any) => ({ id: w.id, public_key: w.public_key })));
      })
      .catch(() => {});
  }, []);

  // Rate lookup
  const [rateFrom, setRateFrom] = useState('USDC');
  const [rateTo, setRateTo] = useState('XLM');
  const [rate, setRate] = useState<RateResponse | null>(null);
  const [rateLoading, setRateLoading] = useState(false);

  // Quote -> Convert
  const [quoteForm, setQuoteForm] = useState({ from_asset: 'NGN', to_asset: 'TXDC', amount: '' });
  const [quote, setQuote] = useState<QuoteResponse | null>(null);
  const [quoteLoading, setQuoteLoading] = useState(false);
  const [convertWalletId, setConvertWalletId] = useState('');
  const [convertLoading, setConvertLoading] = useState(false);
  const [conversionResult, setConversionResult] = useState<any>(null);

  const handleGetRate = async () => {
    setRateLoading(true);
    try {
      const r = await api.getRates(rateFrom, rateTo);
      setRate(r);
      toast('Rate fetched', 'success');
    } catch (err) {
      toast(err instanceof Error ? err.message : 'Failed to fetch rate', 'error');
    } finally {
      setRateLoading(false);
    }
  };

  const handleGetQuote = async (e: React.FormEvent) => {
    e.preventDefault();
    setQuoteLoading(true);
    try {
      const q = await api.getQuote(quoteForm);
      setQuote(q);
      toast('Quote created — expires at ' + new Date(q.expires_at).toLocaleTimeString(), 'success');
    } catch (err) {
      toast(err instanceof Error ? err.message : 'Failed to create quote', 'error');
    } finally {
      setQuoteLoading(false);
    }
  };

  const handleConvert = async () => {
    if (!quote || !convertWalletId) return;
    setConvertLoading(true);
    try {
      const res = await api.convert({ wallet_id: convertWalletId, quote_id: quote.id });
      toast(`Converted via tx ${res.tx_hash?.slice(0, 8) || res.id.slice(0, 8)}`, 'success');
      setQuote(null);
    } catch (err) {
      toast(err instanceof Error ? err.message : 'Conversion failed', 'error');
    } finally {
      setConvertLoading(false);
    }
  };

  return (
    <div className="flex flex-col gap-8 animate-in fade-in slide-in-from-bottom-4 duration-500">
      <PageHeader title="FX & Conversion" description="Get live rates and execute cross-asset swaps via CoinGecko and on-chain settlement." />

      <div className="grid grid-cols-1 gap-6 lg:grid-cols-2">
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Coins className="h-5 w-5" />
              Live Rates
            </CardTitle>
            <CardDescription>Fetch a cached mid-market rate with spread.</CardDescription>
          </CardHeader>
          <CardContent className="flex flex-col gap-5">
            <div className="grid grid-cols-2 gap-4">
              <div className="flex flex-col gap-1.5">
                <label className="text-sm font-medium">From</label>
                <Input value={rateFrom} onChange={(e) => setRateFrom(e.target.value)} placeholder="USDC" />
              </div>
              <div className="flex flex-col gap-1.5">
                <label className="text-sm font-medium">To</label>
                <Input value={rateTo} onChange={(e) => setRateTo(e.target.value)} placeholder="XLM" />
              </div>
            </div>
            <Button onClick={handleGetRate} isLoading={rateLoading}>
              Get Rate
            </Button>
            {rate && (
              <div className="rounded-lg border border-border bg-muted p-4 flex flex-col gap-2">
                <div className="flex items-center justify-between">
                  <span className="text-sm text-muted-foreground">Rate</span>
                  <span className="font-mono font-medium">{rate.rate}</span>
                </div>
                <div className="flex items-center justify-between">
                  <span className="text-sm text-muted-foreground">Mid-market</span>
                  <span className="font-mono">{rate.mid_market_rate}</span>
                </div>
                <div className="flex items-center justify-between">
                  <span className="text-sm text-muted-foreground">Spread</span>
                  <span>{rate.spread_bps} bps</span>
                </div>
                <div className="flex items-center justify-between">
                  <span className="text-sm text-muted-foreground">Provider</span>
                  <Badge variant={rate.stale ? 'warning' : 'success'}>{rate.provider}</Badge>
                </div>
                {rate.stale && <p className="text-xs text-warning">Cached value is stale.</p>}
              </div>
            )}
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <ArrowRightLeft className="h-5 w-5" />
              Quote & Convert
            </CardTitle>
            <CardDescription>Quotes live for 30s — convert with a wallet and quote ID.</CardDescription>
          </CardHeader>
          <CardContent className="flex flex-col gap-5">
            <form onSubmit={handleGetQuote} className="flex flex-col gap-4">
              <div className="grid grid-cols-2 gap-4">
                <div className="flex flex-col gap-1.5">
                  <label className="text-sm font-medium">From Asset</label>
                  <Input
                    value={quoteForm.from_asset}
                    onChange={(e) => setQuoteForm({ ...quoteForm, from_asset: e.target.value })}
                    required
                  />
                </div>
                <div className="flex flex-col gap-1.5">
                  <label className="text-sm font-medium">To Asset</label>
                  <Input
                    value={quoteForm.to_asset}
                    onChange={(e) => setQuoteForm({ ...quoteForm, to_asset: e.target.value })}
                    required
                  />
                </div>
              </div>
              <div className="flex flex-col gap-1.5">
                <label className="text-sm font-medium">Amount</label>
                <Input
                  value={quoteForm.amount}
                  onChange={(e) => setQuoteForm({ ...quoteForm, amount: e.target.value })}
                  required
                  placeholder="100.00"
                  className="font-mono"
                />
              </div>
              <Button type="submit" isLoading={quoteLoading}>
                Get Quote
              </Button>
            </form>

            {quote && (
              <div className="rounded-lg border border-border bg-primary-subtle p-4 flex flex-col gap-3">
                <p className="text-sm font-medium">
                  {quote.from_amount} {quote.from_asset} → {quote.to_amount} {quote.to_asset}
                </p>
                <p className="text-xs text-muted-foreground">
                  Rate {quote.rate} · Fee {quote.fee} · Expires {new Date(quote.expires_at).toLocaleString()}
                </p>
                <div className="flex flex-col gap-1.5">
                  <label className="text-sm font-medium">Wallet for conversion</label>
                  <Select value={convertWalletId} onChange={(e) => setConvertWalletId(e.target.value)}>
                    <option value="">Select wallet</option>
                    {wallets.map((w) => (
                      <option key={w.id} value={w.id}>
                        {w.public_key ? '0x' + w.public_key.slice(-8) + '...' : w.id.slice(0, 16) + '...'}
                      </option>
                    ))}
                  </Select>
                </div>
                <Button onClick={handleConvert} isLoading={convertLoading} disabled={!convertWalletId}>
                  Execute Conversion
                </Button>
              </div>
            )}
          </CardContent>
        </Card>
      </div>

      {conversionResult && (
        <Card className="w-full max-w-xl">
          <CardHeader>
            <CardTitle className="text-lg">Conversion Complete</CardTitle>
          </CardHeader>
          <CardContent className="space-y-3">
            <div className="grid grid-cols-2 gap-4 text-sm">
              <div><span className="text-muted-foreground">Status: </span><span className="font-medium text-green-600">Success</span></div>
              <div><span className="text-muted-foreground">ID: </span><span className="font-mono text-xs">{conversionResult.id}</span></div>
              <div><span className="text-muted-foreground">Sent: </span><span className="font-medium">{conversionResult.source_amount} {conversionResult.source_asset}</span></div>
              <div><span className="text-muted-foreground">Received: </span><span className="font-medium">{conversionResult.dest_amount} {conversionResult.dest_asset}</span></div>
              <div><span className="text-muted-foreground">Rate: </span><span className="font-medium">{conversionResult.rate}</span></div>
              <div><span className="text-muted-foreground">Fee: </span><span className="font-medium">{conversionResult.fee_amount}</span></div>
            </div>
            {conversionResult.tx_hash && (
              <div className="text-sm"><span className="text-muted-foreground">Tx Hash: </span><span className="font-mono text-xs break-all">{conversionResult.tx_hash}</span></div>
            )}
            <Button variant="secondary" size="sm" onClick={() => setConversionResult(null)}>Dismiss</Button>
          </CardContent>
        </Card>
      )}
    </div>
  );
}
