'use client';

import { useState, useMemo, useEffect } from 'react';
import { api } from '@/lib/api';
import { useAuth } from '@/lib/auth-context';
import { useToast } from '@/lib/toast-context';
import { PageHeader } from '@/components/ui/page-header';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Select } from '@/components/ui/select';
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card';
import { Banknote } from 'lucide-react';

export default function FiatPage() {
  const { getStoredWalletIds } = useAuth();
  const { toast } = useToast();
  const walletIds = useMemo(() => getStoredWalletIds(), [getStoredWalletIds]);

  // Resolve each stored wallet ID to its on-chain address so dropdowns
  // show xdc… addresses (not internal UUIDs). Falls back to the ID prefix.
  const [walletOptions, setWalletOptions] = useState<{ id: string; label: string }[]>([]);
  useEffect(() => {
    let cancelled = false;
    Promise.all(
      walletIds.map(async (id) => {
        try {
          const w = await api.getWallet(id);
          return { id, label: `${w.public_key.slice(0, 10)}…${w.public_key.slice(-4)}` };
        } catch {
          return { id, label: `${id.slice(0, 8)}…` };
        }
      })
    ).then((opts) => {
      if (!cancelled) setWalletOptions(opts);
    });
    return () => { cancelled = true; };
  }, [walletIds]);

  const [tab, setTab] = useState<'deposit' | 'withdraw'>('deposit');
  
  // Exchange rates: 1 USDC = X local currency
  const currencyRates: Record<string, number> = {
    'NGN': 1500,
    'USD': 1,
    'EUR': 0.92,
    'GBP': 0.79,
    'INR': 83,
    'KES': 153,
    'GHS': 15.8,
    'ZAR': 18.5,
  };

  // Deposit
  const [deposit, setDeposit] = useState({ wallet_id: '', amount: '', currency: 'NGN', email: '', name: '' });
  const [depositLoading, setDepositLoading] = useState(false);
  const [depositLink, setDepositLink] = useState<string | null>(null);

  // Withdraw
  const [withdraw, setWithdraw] = useState({
    wallet_id: '',
    amount: '',
    currency: 'NGN',
    asset: 'USDC',
    account_bank: '',
    account_number: '',
  });
  const [withdrawLoading, setWithdrawLoading] = useState(false);
  const [withdrawRef, setWithdrawRef] = useState<string | null>(null);

  const handleDeposit = async (e: React.FormEvent) => {
    e.preventDefault();
    setDepositLoading(true);
    try {
      const res = await api.fiatDeposit(deposit.wallet_id, {
        amount: deposit.amount,
        currency: deposit.currency,
        email: deposit.email,
        name: deposit.name,
      });
      setDepositLink(res.payment_link);
      toast(`Deposit reference ${res.reference}`, 'success');
    } catch (err) {
      toast(err instanceof Error ? err.message : 'Deposit failed', 'error');
    } finally {
      setDepositLoading(false);
    }
  };

  const handleWithdraw = async (e: React.FormEvent) => {
    e.preventDefault();
    setWithdrawLoading(true);
    try {
      const res = await api.fiatWithdraw(withdraw.wallet_id, {
        amount: withdraw.amount,
        currency: withdraw.currency,
        account_bank: withdraw.account_bank,
        account_number: withdraw.account_number,
      });
      setWithdrawRef(res.reference);
      toast(`Withdrawal ${res.status === 'completed' ? 'completed' : 'submitted'} — ${res.reference}`, 'success');
    } catch (err) {
      toast(err instanceof Error ? err.message : 'Withdrawal failed', 'error');
    } finally {
      setWithdrawLoading(false);
    }
  };

  return (
    <div className="flex flex-col gap-8 animate-in fade-in slide-in-from-bottom-4 duration-500">
      <PageHeader title="Fiat Rails" description="Local currency ↔ Stellar USDC via Flutterwave (mock when keys absent)." />

      <div className="flex gap-2 border-b border-border">
        <button
          onClick={() => setTab('deposit')}
          className={`px-4 py-2 text-sm font-medium border-b-2 transition-colors ${tab === 'deposit' ? 'border-primary text-primary' : 'border-transparent text-muted-foreground hover:text-foreground'}`}
        >
          Deposit
        </button>
        <button
          onClick={() => setTab('withdraw')}
          className={`px-4 py-2 text-sm font-medium border-b-2 transition-colors ${tab === 'withdraw' ? 'border-primary text-primary' : 'border-transparent text-muted-foreground hover:text-foreground'}`}
        >
          Withdraw
        </button>
      </div>

      {tab === 'deposit' ? (
        <Card className="max-w-xl">
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Banknote className="h-5 w-5" />
              Fiat Deposit
            </CardTitle>
            <CardDescription>Returns a payment link — redirect the customer to complete funding.</CardDescription>
          </CardHeader>
          <CardContent>
            <form onSubmit={handleDeposit} className="flex flex-col gap-4">
              <div className="flex flex-col gap-1.5">
                <label className="text-sm font-medium">Wallet</label>
                <Select value={deposit.wallet_id} onChange={(e) => setDeposit({ ...deposit, wallet_id: e.target.value })} required>
                  <option value="">Select wallet</option>
                  {walletOptions.map((w) => (
                    <option key={w.id} value={w.id}>
                      {w.label}
                    </option>
                  ))}
                </Select>
              </div>
              <div className="grid grid-cols-2 gap-4">
                <div className="flex flex-col gap-1.5">
                  <label className="text-sm font-medium">Amount</label>
                  <Input value={deposit.amount} onChange={(e) => setDeposit({ ...deposit, amount: e.target.value })} required className="font-mono" />
                </div>
                <div className="flex flex-col gap-1.5">
                  <label className="text-sm font-medium">Currency</label>
                  <select value={deposit.currency} onChange={(e) => setDeposit({ ...deposit, currency: e.target.value })} className="flex h-10 w-full rounded-md border border-border bg-background px-3 py-2 text-sm">
                    <option value="NGN">NGN - Nigerian Naira</option>
                    <option value="USD">USD - US Dollar</option>
                    <option value="EUR">EUR - Euro</option>
                    <option value="GBP">GBP - British Pound</option>
                    <option value="INR">INR - Indian Rupee</option>
                    <option value="KES">KES - Kenyan Shilling</option>
                    <option value="GHS">GHS - Ghanaian Cedi</option>
                    <option value="ZAR">ZAR - South African Rand</option>
                  </select>
                </div>
              </div>
              <div className="flex flex-col gap-1.5">
                <label className="text-sm font-medium">Customer Email</label>
                <Input type="email" value={deposit.email} onChange={(e) => setDeposit({ ...deposit, email: e.target.value })} required />
              </div>
              <div className="flex flex-col gap-1.5">
                <label className="text-sm font-medium">Customer Name</label>
                <Input value={deposit.name} onChange={(e) => setDeposit({ ...deposit, name: e.target.value })} required />
              </div>
              <Button type="submit" isLoading={depositLoading}>
                Initiate Deposit
              </Button>
              <Button
                type="button"
                variant="secondary"
                onClick={async () => {
                  if (!deposit.wallet_id || !deposit.amount) {
                    toast('Select wallet and enter amount', 'error');
                    return;
                  }
                  const rate = currencyRates[deposit.currency.toUpperCase()] || 1500;
                  const usdcAmount = parseFloat(deposit.amount) / rate;
                  try {
                    const res = await api.faucet(deposit.wallet_id, 'USDC', usdcAmount);
                    toast(`Simulated! Added ${usdcAmount.toFixed(2)} USDC (${deposit.amount} ${deposit.currency} = 1 USDC @ ${rate})`, 'success');
                  } catch (err) {
                    toast(err instanceof Error ? err.message : 'Simulate failed', 'error');
                  }
                }}
              >
                Simulate Deposit (Demo)
              </Button>
            </form>
            {depositLink && (
              <div className="mt-4 rounded-lg border border-primary/20 bg-primary-subtle p-4">
                <p className="text-sm font-medium">Payment link</p>
                <a href={depositLink} target="_blank" rel="noreferrer" className="text-sm text-primary hover:underline break-all">
                  {depositLink}
                </a>
              </div>
            )}
          </CardContent>
        </Card>
      ) : (
        <Card className="max-w-xl">
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Banknote className="h-5 w-5" />
              Fiat Withdrawal
            </CardTitle>
            <CardDescription>Converts USDC/TXDC → fiat and triggers a bank payout.</CardDescription>
          </CardHeader>
          <CardContent>
            <form onSubmit={handleWithdraw} className="flex flex-col gap-4">
              <div className="flex flex-col gap-1.5">
                <label className="text-sm font-medium">Wallet</label>
                <Select value={withdraw.wallet_id} onChange={(e) => setWithdraw({ ...withdraw, wallet_id: e.target.value })} required>
                  <option value="">Select wallet</option>
                  {walletOptions.map((w) => (
                    <option key={w.id} value={w.id}>
                      {w.label}
                    </option>
                  ))}
                </Select>
              </div>
              <div className="grid grid-cols-3 gap-4">
                <div className="flex flex-col gap-1.5">
                  <label className="text-sm font-medium">Amount</label>
                  <Input value={withdraw.amount} onChange={(e) => setWithdraw({ ...withdraw, amount: e.target.value })} required className="font-mono" />
                </div>
                <div className="flex flex-col gap-1.5">
                  <label className="text-sm font-medium">Asset</label>
                  <select value={withdraw.asset} onChange={(e) => setWithdraw({ ...withdraw, asset: e.target.value })} className="flex h-10 w-full rounded-md border border-border bg-background px-3 py-2 text-sm">
                    <option value="USDC">USDC</option>
                    <option value="TXDC">TXDC</option>
                  </select>
                </div>
                <div className="flex flex-col gap-1.5">
                  <label className="text-sm font-medium">Currency</label>
                  <select value={withdraw.currency} onChange={(e) => setWithdraw({ ...withdraw, currency: e.target.value })} className="flex h-10 w-full rounded-md border border-border bg-background px-3 py-2 text-sm">
                    <option value="NGN">NGN - Nigerian Naira</option>
                    <option value="USD">USD - US Dollar</option>
                    <option value="EUR">EUR - Euro</option>
                    <option value="GBP">GBP - British Pound</option>
                    <option value="INR">INR - Indian Rupee</option>
                    <option value="KES">KES - Kenyan Shilling</option>
                    <option value="GHS">GHS - Ghanaian Cedi</option>
                    <option value="ZAR">ZAR - South African Rand</option>
                  </select>
                </div>
              </div>
              <div className="flex flex-col gap-1.5">
                <label className="text-sm font-medium">Account Bank</label>
                <Input value={withdraw.account_bank} onChange={(e) => setWithdraw({ ...withdraw, account_bank: e.target.value })} required placeholder="044" />
              </div>
              <div className="flex flex-col gap-1.5">
                <label className="text-sm font-medium">Account Number</label>
                <Input value={withdraw.account_number} onChange={(e) => setWithdraw({ ...withdraw, account_number: e.target.value })} required />
              </div>
              <Button type="submit" isLoading={withdrawLoading}>
                Initiate Withdrawal
              </Button>
            </form>
            {withdrawRef && (
              <div className="mt-4 rounded-lg border border-success/20 bg-success-subtle p-4">
                <p className="text-sm font-medium text-success">Reference: {withdrawRef}</p>
                <p className="text-xs text-muted-foreground">Withdrawal submitted to Flutterwave for processing.</p>
              </div>
            )}
          </CardContent>
        </Card>
      )}
    </div>
  );
}
