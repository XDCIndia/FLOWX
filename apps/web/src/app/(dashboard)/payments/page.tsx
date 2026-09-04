'use client';

import { useState } from 'react';
import { useToast } from '@/lib/toast-context';
import { PageHeader } from '@/components/ui/page-header';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card';
import { Input } from '@/components/ui/input';
import { Select } from '@/components/ui/select';
import { Badge } from '@/components/ui/badge';
import { EmptyState } from '@/components/ui/empty-state';
import { Skeleton } from '@/components/ui/skeleton';
import { Route, Trophy, Clock, Shield, Coins, Zap, AlertTriangle, CheckCircle } from 'lucide-react';

interface RouteOption {
  route_id: string;
  route_name: string;
  score: number;
  cost_score: number;
  speed_score: number;
  reliability: number;
  compliance: number;
  liquidity: number;
  recommended: boolean;
  source_asset: string;
  dest_asset: string;
  source_amount: string;
  dest_amount: string;
  rate: string;
  fee: string;
  fee_asset: string;
  settlement_time: string;
  provider: string;
  warnings?: string[];
}

interface QuoteResponse {
  source_asset: string;
  dest_asset: string;
  amount: string;
  ranking_mode: string;
  routes: RouteOption[];
  total_routes: number;
}

interface ExecutionResult {
  route_id: string;
  route_name: string;
  reference: string;
  source_asset: string;
  dest_asset: string;
  amount: string;
  dest_amount: string;
}

function ScoreBar({ label, score, icon: Icon }: { label: string; score: number; icon: React.ElementType }) {
  const color = score >= 80 ? 'bg-green-500' : score >= 50 ? 'bg-yellow-500' : 'bg-red-500';
  return (
    <div className="flex items-center gap-2">
      <Icon className="h-4 w-4 text-muted-foreground" />
      <span className="text-xs text-muted-foreground w-20">{label}</span>
      <div className="flex-1 h-2 bg-muted rounded-full overflow-hidden">
        <div className={`h-full ${color} rounded-full`} style={{ width: `${Math.min(score, 100)}%` }} />
      </div>
      <span className="text-xs font-mono w-12 text-right">{score.toFixed(1)}</span>
    </div>
  );
}

function SettlementBadge({ time }: { time: string }) {
  const match = time.match(/(\d+)([smh])/);
  if (!match) return <Badge>{time}</Badge>;
  const [, val, unit] = match;
  const minutes = unit === 'm' ? parseInt(val) : unit === 'h' ? parseInt(val) * 60 : parseInt(val) / 60;
  if (minutes <= 1) return <Badge variant="success">{time}</Badge>;
  if (minutes <= 10) return <Badge variant="warning">{time}</Badge>;
  return <Badge variant="danger">{time}</Badge>;
}

export default function PaymentsPage() {
  const { toast } = useToast();
  const [sourceAsset, setSourceAsset] = useState('INR');
  const [destAsset, setDestAsset] = useState('EUR');
  const [amount, setAmount] = useState('');
  const [rankingMode, setRankingMode] = useState('balanced');
  const [loading, setLoading] = useState(false);
  const [quote, setQuote] = useState<QuoteResponse | null>(null);
  const [executingRoute, setExecutingRoute] = useState<string | null>(null);
  const [executionResult, setExecutionResult] = useState<ExecutionResult | null>(null);

  const handleQuote = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!amount || parseFloat(amount) <= 0) return;
    setLoading(true);
    try {
      const res = await fetch(`${process.env.NEXT_PUBLIC_API_URL || 'http://localhost:3000'}/v1/payments/quote`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          Authorization: `Bearer ${localStorage.getItem('flowx_api_key')}`,
        },
        body: JSON.stringify({ source_asset: sourceAsset, dest_asset: destAsset, amount, ranking_mode: rankingMode }),
      });
      if (!res.ok) { const err = await res.json(); throw new Error(err.error?.message || 'Quote failed'); }
      const data = await res.json();
      setQuote(data);
      setExecutionResult(null);
      toast(`Found ${data.total_routes} route(s)`, 'success');
    } catch (err) {
      toast(err instanceof Error ? err.message : 'Failed to get quote', 'error');
    } finally { setLoading(false); }
  };

  const handleExecute = async (route: RouteOption) => {
    setExecutingRoute(route.route_id);
    try {
      const res = await fetch(`${process.env.NEXT_PUBLIC_API_URL || 'http://localhost:3000'}/v1/payments/send`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          Authorization: `Bearer ${localStorage.getItem('flowx_api_key')}`,
        },
        body: JSON.stringify({
          source_asset: route.source_asset,
          dest_asset: route.dest_asset,
          amount: route.source_amount,
          route_id: route.route_id,
        }),
      });
      if (!res.ok) { const err = await res.json(); throw new Error(err.error?.message || 'Execution failed'); }
      const data = await res.json();
      setExecutionResult(data);
      toast(`${route.route_name} initiated! Ref: ${data.reference}`, 'success');
    } catch (err) {
      toast(err instanceof Error ? err.message : 'Failed to execute route', 'error');
    } finally {
      setExecutingRoute(null);
    }
  };

  return (
    <div className="flex flex-col gap-8 animate-in fade-in slide-in-from-bottom-4 duration-500">
      <PageHeader title="Payments" description="Compare routes, fees, and settlement times across payment rails." />
      <Card className="max-w-3xl">
        <CardHeader>
          <CardTitle className="flex items-center gap-2"><Route className="h-5 w-5" />Find Best Route</CardTitle>
          <CardDescription>Enter payment details to compare routes with scoring.</CardDescription>
        </CardHeader>
        <CardContent>
          <form onSubmit={handleQuote} className="flex flex-col gap-4">
            <div className="grid grid-cols-2 gap-4 md:grid-cols-4">
              <div className="flex flex-col gap-1.5">
                <label className="text-sm font-medium">From</label>
                <Select value={sourceAsset} onChange={(e) => setSourceAsset(e.target.value)}>
                  <option value="INR">INR (Rupee)</option>
                  <option value="EUR">EUR (Euro)</option>
                  <option value="NGN">NGN (Naira)</option>
                  <option value="KES">KES (Shilling)</option>
                  <option value="USDC">USDC</option>
                  <option value="TXDC">TXDC</option>
                  <option value="XLM">XLM</option>
                </Select>
              </div>
              <div className="flex flex-col gap-1.5">
                <label className="text-sm font-medium">To</label>
                <Select value={destAsset} onChange={(e) => setDestAsset(e.target.value)}>
                  <option value="EUR">EUR (Euro)</option>
                  <option value="INR">INR (Rupee)</option>
                  <option value="TXDC">TXDC</option>
                  <option value="USDC">USDC</option>
                  <option value="XLM">XLM</option>
                  <option value="NGN">NGN (Naira)</option>
                  <option value="KES">KES (Shilling)</option>
                </Select>
              </div>
              <div className="flex flex-col gap-1.5">
                <label className="text-sm font-medium">Amount</label>
                <Input value={amount} onChange={(e) => setAmount(e.target.value)} placeholder="100000" className="font-mono" required />
              </div>
              <div className="flex flex-col gap-1.5">
                <label className="text-sm font-medium">Optimize For</label>
                <Select value={rankingMode} onChange={(e) => setRankingMode(e.target.value)}>
                  <option value="balanced">Balanced</option>
                  <option value="cheapest">Cheapest</option>
                  <option value="fastest">Fastest</option>
                  <option value="most_reliable">Most Reliable</option>
                </Select>
              </div>
            </div>
            <div className="flex justify-end">
              <Button type="submit" isLoading={loading} disabled={!amount}>Compare Routes</Button>
            </div>
          </form>
        </CardContent>
      </Card>

      {loading && <div className="grid grid-cols-1 gap-6 lg:grid-cols-2"><Skeleton className="h-64" /><Skeleton className="h-64" /></div>}
      {!loading && quote && quote.routes.length === 0 && (
        <EmptyState icon={AlertTriangle} title="No routes available" description={`No routes for ${sourceAsset} → ${destAsset}.`} />
      )}
      {!loading && quote && quote.routes.length > 0 && (<>
        <div className="flex items-center gap-2 text-sm text-muted-foreground">
          <Trophy className="h-4 w-4 text-yellow-500" />
          <span>Found <strong>{quote.total_routes}</strong> route(s) for <strong>{quote.amount} {quote.source_asset}</strong> → <strong>{quote.dest_asset}</strong> <Badge className="ml-1">{quote.ranking_mode}</Badge></span>
        </div>
        <div className="grid grid-cols-1 gap-6 lg:grid-cols-2">
          {quote.routes.map((route, idx) => (
            <Card key={route.route_id + "-" + idx} className={`relative ${route.recommended ? 'border-green-500 border-2' : ''}`}>
              {route.recommended && (<div className="absolute -top-3 right-4"><Badge variant="success" className="flex items-center gap-1"><Trophy className="h-3 w-3" /> Recommended</Badge></div>)}
              <CardHeader className="pb-3">
                <div className="flex items-start justify-between">
                  <div><CardTitle className="text-lg">{route.route_name}</CardTitle><p className="text-sm text-muted-foreground mt-1">{route.provider}</p></div>
                  <div className="text-right"><div className="text-2xl font-bold font-mono">{route.score.toFixed(1)}</div><div className="text-xs text-muted-foreground">/ 100</div></div>
                </div>
              </CardHeader>
              <CardContent className="flex flex-col gap-4">
                <div className="flex items-center justify-between rounded-lg border border-border bg-muted p-3">
                  <div className="text-center"><div className="text-lg font-mono font-bold">{route.source_amount}</div><div className="text-xs text-muted-foreground">{route.source_asset}</div></div>
                  <div className="text-2xl px-4">&rarr;</div>
                  <div className="text-center"><div className="text-lg font-mono font-bold text-green-600">{route.dest_amount}</div><div className="text-xs text-muted-foreground">{route.dest_asset}</div></div>
                </div>
                <div className="flex flex-col gap-2">
                  <ScoreBar label="Cost" score={route.cost_score} icon={Coins} />
                  <ScoreBar label="Speed" score={route.speed_score} icon={Zap} />
                  <ScoreBar label="Reliability" score={route.reliability} icon={Shield} />
                  <ScoreBar label="Compliance" score={route.compliance} icon={CheckCircle} />
                </div>
                <div className="grid grid-cols-2 gap-2 text-sm">
                  <div className="flex items-center gap-1.5 text-muted-foreground"><Coins className="h-4 w-4" />Fee: {route.fee} {route.fee_asset}</div>
                  <div className="flex items-center gap-1.5 text-muted-foreground"><Clock className="h-4 w-4" />Settlement: <SettlementBadge time={route.settlement_time} /></div>
                </div>
                <div className="text-xs text-muted-foreground">Rate: 1 {route.source_asset} = {route.rate} {route.dest_asset}</div>
                {route.warnings && route.warnings.length > 0 && (<div className="flex flex-col gap-1">{route.warnings.map((w, i) => (<div key={i} className="text-xs text-yellow-600 flex items-center gap-1"><AlertTriangle className="h-3 w-3" />{w}</div>))}</div>)}
                <Button
                  className="w-full"
                  variant={route.recommended ? 'primary' : 'secondary'}
                  disabled={executingRoute !== null}
                  isLoading={executingRoute === route.route_id}
                  onClick={() => handleExecute(route)}
                >
                  {executingRoute === route.route_id ? 'Executing...' : 'Execute This Route'}
                </Button>
              </CardContent>
            </Card>
          ))}
        </div>

        {executionResult && (() => {
          const isStripeUrl = executionResult.reference.startsWith('http');
          return (
            <Card className="mt-2 border-green-500 border-2">
              <CardHeader>
                <CardTitle className="flex items-center gap-2 text-green-600">
                  <CheckCircle className="h-5 w-5" />{isStripeUrl ? 'Payment Ready' : 'Payment Initiated'}
                </CardTitle>
                <CardDescription>{isStripeUrl ? 'Click below to complete payment on Stripe.' : 'Your payment has been submitted and is being processed.'}</CardDescription>
              </CardHeader>
              <CardContent className="flex flex-col gap-3">
                <div className="grid grid-cols-2 gap-4 text-sm">
                  <div><span className="text-muted-foreground">Route:</span> <span className="font-medium">{executionResult.route_name}</span></div>
                  <div><span className="text-muted-foreground">Reference:</span> <code className="bg-muted px-2 py-0.5 rounded font-mono text-xs">{isStripeUrl ? 'Stripe Checkout' : executionResult.reference}</code></div>
                  <div><span className="text-muted-foreground">Sent:</span> <span className="font-mono font-bold">{executionResult.amount} {executionResult.source_asset}</span></div>
                  <div><span className="text-muted-foreground">Receiving:</span> <span className="font-mono font-bold text-green-600">{executionResult.dest_amount} {executionResult.dest_asset}</span></div>
                </div>
                {isStripeUrl ? (
                  <div className="flex flex-col gap-2">
                    <div className="flex items-center gap-2 text-xs text-muted-foreground bg-muted rounded-lg p-3">
                      <Clock className="h-4 w-4" />
                      Complete payment on Stripe to finish settlement.
                    </div>
                    <a href={executionResult.reference} target="_blank" rel="noopener noreferrer" className="w-full">
                      <Button className="w-full">Open Stripe Checkout &rarr;</Button>
                    </a>
                  </div>
                ) : executionResult.reference.startsWith('0x') ? (
                  <div className="flex items-center gap-2 text-xs text-success bg-success/10 rounded-lg p-3">
                    <CheckCircle className="h-4 w-4" />
                    Payment completed on XDC blockchain!
                    <a href={`https://testnet.xdcscan.com/tx/${executionResult.reference}`} target="_blank" rel="noopener noreferrer" className="underline">
                      View on Explorer
                    </a>
                  </div>
                ) : (
                  <div className="flex items-center gap-2 text-xs text-muted-foreground bg-muted rounded-lg p-3">
                    <Clock className="h-4 w-4" />
                    Settlement in progress. Track using reference: <code className="font-mono">{executionResult.reference}</code>
                  </div>
                )}
                <Button variant="secondary" onClick={() => setExecutionResult(null)}>Dismiss</Button>
              </CardContent>
            </Card>
          );
        })()}
      </>)}
    </div>
  );
}
