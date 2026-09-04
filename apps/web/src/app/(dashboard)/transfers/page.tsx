'use client';

import { useEffect, useState, useCallback, useMemo } from 'react';
import { api, type Transaction, type Wallet } from '@/lib/api';
import { useAuth } from '@/lib/auth-context';
import { useToast } from '@/lib/toast-context';
import { PageHeader } from '@/components/ui/page-header';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card';
import { Input } from '@/components/ui/input';
import { Select } from '@/components/ui/select';
import { Badge } from '@/components/ui/badge';
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table';
import { EmptyState } from '@/components/ui/empty-state';
import { Skeleton } from '@/components/ui/skeleton';
import { ArrowRightLeft, ExternalLink, Plus, X, Download, Copy } from 'lucide-react';

function statusBadge(status: string) {
  if (status === 'confirmed') return <Badge variant="success">{status}</Badge>;
  if (status === 'pending') return <Badge variant="warning">{status}</Badge>;
  return <Badge variant="danger">{status}</Badge>;
}

/** Convert xdc... prefix to 0x... for display. */
function toDisplayAddress(key: string) {
  return key.toLowerCase().startsWith('xdc') ? '0x' + key.slice(3) : key;
}

export default function TransfersPage() {
  const { getStoredWalletIds } = useAuth();
  const { toast } = useToast();
  const [transfers, setTransfers] = useState<Transaction[]>([]);
  const [wallets, setWallets] = useState<Map<string, Wallet>>(new Map());
  const [loading, setLoading] = useState(true);
  const [filter, setFilter] = useState('all');
  const [showForm, setShowForm] = useState(false);
  const [showReceive, setShowReceive] = useState(false);
  const [receiveWallet, setReceiveWallet] = useState('');
  const [showVerify, setShowVerify] = useState(false);
  const [verifyWallet, setVerifyWallet] = useState('');
  const [verifyTxHash, setVerifyTxHash] = useState('');
  const [verifying, setVerifying] = useState(false);
  const [submitting, setSubmitting] = useState(false);
  const [form, setForm] = useState({
    from_wallet_id: '',
    to_wallet_id: '',
    asset: 'TXDC',
    amount: '',
  });

  const walletIds = useMemo(() => getStoredWalletIds(), [getStoredWalletIds]);

  // Fetch wallet details to get 0x addresses
  const fetchWalletDetails = useCallback(async () => {
    const map = new Map<string, Wallet>();
    for (const id of walletIds) {
      try {
        const wallet = await api.getWallet(id);
        map.set(id, wallet);
      } catch {}
    }
    setWallets(map);
  }, [walletIds]);

  const fetchTransfers = useCallback(async () => {
    setLoading(true);
    try {
      const allTx: Transaction[] = [];
      for (const id of walletIds) {
        try {
          const res = await api.listTransactions(id, 50);
          allTx.push(...(res.transactions || []));
        } catch {}
      }
      const unique = Array.from(new Map(allTx.map((t) => [t.id, t])).values());
      unique.sort(
        (a, b) => new Date(b.created_at).getTime() - new Date(a.created_at).getTime()
      );
      setTransfers(unique);
    } catch {
      toast('Failed to load transfers', 'error');
    } finally {
      setLoading(false);
    }
  }, [walletIds, toast]);

  useEffect(() => {
    let cancelled = false;
    const run = async () => {
      if (cancelled) return;
      await Promise.all([fetchWalletDetails(), fetchTransfers()]);
    };
    run();
    return () => {
      cancelled = true;
    };
  }, [fetchWalletDetails, fetchTransfers]);

  const handleCreateTransfer = async (e: React.FormEvent) => {
    e.preventDefault();
    setSubmitting(true);
    try {
      await api.createTransfer(form);
      toast('Transfer initiated', 'success');
      setShowForm(false);
      setForm({ from_wallet_id: '', to_wallet_id: '', asset: 'TXDC', amount: '' });
      await fetchTransfers();
    } catch (err) {
      toast(err instanceof Error ? err.message : 'Transfer failed', 'error');
    } finally {
      setSubmitting(false);
    }
  };

  const handleVerifyDeposit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!verifyWallet || !verifyTxHash) return;
    setVerifying(true);
    try {
      const res = await api.verifyDeposit(verifyWallet, verifyTxHash);
      toast(`Deposit verified: ${res.amount} ${res.asset}`, 'success');
      setShowVerify(false);
      setVerifyTxHash('');
      await fetchTransfers();
      await fetchWalletDetails();
    } catch (err) {
      toast(err instanceof Error ? err.message : 'Verification failed', 'error');
    } finally {
      setVerifying(false);
    }
  };

  const copyAddress = (addr: string) => {
    navigator.clipboard.writeText(addr);
    toast('Address copied', 'info');
  };

  const getWalletLabel = (id: string) => {
    const wallet = wallets.get(id);
    if (wallet) {
      const addr = toDisplayAddress(wallet.public_key);
      return `${addr.slice(0, 10)}...${addr.slice(-6)}`;
    }
    return `${id.slice(0, 8)}...`;
  };

  const filtered =
    filter === 'all' ? transfers : transfers.filter((t) => t.status === filter);

  if (loading) {
    return (
      <div className="flex flex-col gap-8">
        <div className="flex items-center justify-between">
          <Skeleton className="h-10 w-48" />
          <Skeleton className="h-10 w-32" />
        </div>
        <Skeleton className="h-64" />
      </div>
    );
  }

  return (
    <div className="flex flex-col gap-8 animate-in fade-in slide-in-from-bottom-4 duration-500">
      <PageHeader
        title="Transfers"
        description="Send and receive TXDC on XDC Apothem testnet."
      >
        <div className="flex items-center gap-3">
          <Select
            value={filter}
            onChange={(e) => setFilter(e.target.value)}
            className="w-40"
          >
            <option value="all">All Statuses</option>
            <option value="confirmed">Confirmed</option>
            <option value="pending">Pending</option>
            <option value="failed">Failed</option>
          </Select>
          {walletIds.length >= 1 && (
            <>
              <Button
                variant={showReceive ? 'secondary' : 'primary'}
                onClick={() => { setShowReceive(!showReceive); setShowForm(false); setShowVerify(false); }}
              >
                {showReceive ? (
                  <>
                    <X className="h-4 w-4" /> Close
                  </>
                ) : (
                  <>
                    <Download className="h-4 w-4" /> Receive
                  </>
                )}
              </Button>
              <Button
                variant={showVerify ? 'secondary' : 'primary'}
                onClick={() => { setShowVerify(!showVerify); setShowForm(false); setShowReceive(false); }}
              >
                {showVerify ? (
                  <>
                    <X className="h-4 w-4" /> Close
                  </>
                ) : (
                  <>
                    <Plus className="h-4 w-4" /> Verify Deposit
                  </>
                )}
              </Button>
            </>
          )}
          {walletIds.length >= 2 && (
            <Button
              variant={showForm ? 'secondary' : 'primary'}
              onClick={() => { setShowForm(!showForm); setShowReceive(false); }}
            >
              {showForm ? (
                <>
                  <X className="h-4 w-4" /> Cancel
                </>
              ) : (
                <>
                  <Plus className="h-4 w-4" /> Send
                </>
              )}
            </Button>
          )}
        </div>
      </PageHeader>

      {/* Receive Card */}
      {showReceive && (
        <Card className="max-w-2xl">
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Download className="h-5 w-5" />
              Receive TXDC
            </CardTitle>
            <CardDescription>
              Share your wallet address to receive TXDC from MetaMask or any external wallet.
            </CardDescription>
          </CardHeader>
          <CardContent>
            <div className="flex flex-col gap-4">
              <div className="flex flex-col gap-1.5">
                <label className="text-sm font-medium">Select Wallet</label>
                <Select
                  value={receiveWallet}
                  onChange={(e) => setReceiveWallet(e.target.value)}
                >
                  <option value="">Select wallet</option>
                  {walletIds.map((id) => (
                    <option key={id} value={id}>
                      {getWalletLabel(id)}
                    </option>
                  ))}
                </Select>
              </div>
              {receiveWallet && wallets.get(receiveWallet) && (
                <div className="flex flex-col gap-2">
                  <label className="text-sm font-medium">Deposit Address (0x)</label>
                  <div className="flex items-center gap-2 rounded-lg border border-border bg-muted px-3 py-2">
                    <code className="flex-1 font-mono text-sm text-muted-foreground break-all">
                      {toDisplayAddress(wallets.get(receiveWallet)!.public_key)}
                    </code>
                    <Button
                      variant="ghost"
                      size="sm"
                      className="h-8 px-2"
                      onClick={() => copyAddress(toDisplayAddress(wallets.get(receiveWallet)!.public_key))}
                    >
                      <Copy className="h-4 w-4" />
                    </Button>
                  </div>
                  <p className="text-xs text-muted-foreground">
                    Copy this address and paste it in MetaMask to send TXDC to this wallet.
                    The transfer will appear here once confirmed on the XDC network.
                  </p>
                  <a
                    href={`https://apothem.blocksscan.io/address/${toDisplayAddress(wallets.get(receiveWallet)!.public_key)}`}
                    target="_blank"
                    rel="noreferrer"
                    className="inline-flex items-center gap-1 text-sm font-medium text-primary hover:text-primary-hover hover:underline"
                  >
                    View on BlocksScan
                    <ExternalLink className="h-3.5 w-3.5" />
                  </a>
                </div>
              )}
            </div>
          </CardContent>
        </Card>
      )}

      {/* Verify Deposit Card */}
      {showVerify && (
        <Card className="max-w-2xl">
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Plus className="h-5 w-5" />
              Verify Deposit
            </CardTitle>
            <CardDescription>
              Paste the tx hash from MetaMask to verify and credit the deposit.
            </CardDescription>
          </CardHeader>
          <CardContent>
            <form onSubmit={handleVerifyDeposit} className="flex flex-col gap-4">
              <div className="flex flex-col gap-1.5">
                <label className="text-sm font-medium">Select Wallet</label>
                <Select
                  value={verifyWallet}
                  onChange={(e) => setVerifyWallet(e.target.value)}
                  required
                >
                  <option value="">Select wallet</option>
                  {walletIds.map((id) => (
                    <option key={id} value={id}>
                      {getWalletLabel(id)}
                    </option>
                  ))}
                </Select>
              </div>
              <div className="flex flex-col gap-1.5">
                <label className="text-sm font-medium">Transaction Hash</label>
                <Input
                  value={verifyTxHash}
                  onChange={(e) => setVerifyTxHash(e.target.value)}
                  placeholder="0x..."
                  className="font-mono"
                  required
                />
              </div>
              <div className="flex justify-end">
                <Button
                  type="submit"
                  isLoading={verifying}
                  disabled={!verifyWallet || !verifyTxHash}
                >
                  Verify & Credit
                </Button>
              </div>
            </form>
          </CardContent>
        </Card>
      )}

      {/* Send Form */}
      {showForm && (
        <Card className="max-w-2xl">
          <CardHeader>
            <CardTitle>Send TXDC</CardTitle>
          </CardHeader>
          <CardContent>
            <form onSubmit={handleCreateTransfer} className="flex flex-col gap-5">
              <div className="grid grid-cols-1 gap-5 md:grid-cols-2">
                <div className="flex flex-col gap-1.5">
                  <label className="text-sm font-medium text-foreground">From Wallet</label>
                  <Select
                    value={form.from_wallet_id}
                    onChange={(e) =>
                      setForm({ ...form, from_wallet_id: e.target.value })
                    }
                    required
                  >
                    <option value="">Select wallet</option>
                    {walletIds.map((id) => (
                      <option key={id} value={id}>
                        {getWalletLabel(id)}
                      </option>
                    ))}
                  </Select>
                </div>
                <div className="flex flex-col gap-1.5">
                  <label className="text-sm font-medium text-foreground">To Wallet</label>
                  <Select
                    value={form.to_wallet_id}
                    onChange={(e) =>
                      setForm({ ...form, to_wallet_id: e.target.value })
                    }
                    required
                  >
                    <option value="">Select wallet</option>
                    {walletIds.map((id) => (
                      <option key={id} value={id}>
                        {getWalletLabel(id)}
                      </option>
                    ))}
                  </Select>
                </div>
              </div>
              <div className="grid grid-cols-1 gap-5 md:grid-cols-2">
                <div className="flex flex-col gap-1.5">
                  <label className="text-sm font-medium text-foreground">Asset</label>
                  <Input
                    value={form.asset}
                    onChange={(e) => setForm({ ...form, asset: e.target.value })}
                    required
                  />
                </div>
                <div className="flex flex-col gap-1.5">
                  <label className="text-sm font-medium text-foreground">Amount</label>
                  <Input
                    value={form.amount}
                    onChange={(e) => setForm({ ...form, amount: e.target.value })}
                    required
                    placeholder="0.0000000"
                    className="font-mono"
                  />
                </div>
              </div>
              <div className="flex justify-end">
                <Button
                  type="submit"
                  isLoading={submitting}
                  disabled={
                    !form.from_wallet_id || !form.to_wallet_id || !form.amount
                  }
                >
                  Send TXDC
                </Button>
              </div>
            </form>
          </CardContent>
        </Card>
      )}

      {/* Transfer History */}
      <Card>
        {filtered.length === 0 ? (
          <EmptyState
            icon={ArrowRightLeft}
            title="No transfers yet"
            description="Send TXDC from MetaMask to your FlowX wallet, or transfer between wallets."
          />
        ) : (
          <Table>
            <TableHead>
              <TableRow>
                <TableHeader>ID</TableHeader>
                <TableHeader>Amount</TableHeader>
                <TableHeader>From / To</TableHeader>
                <TableHeader>Date</TableHeader>
                <TableHeader>Status</TableHeader>
                <TableHeader>Tx Hash</TableHeader>
              </TableRow>
            </TableHead>
            <TableBody>
              {filtered.map((tr) => (
                <TableRow key={tr.id}>
                  <TableCell className="font-mono text-muted-foreground">
                    {tr.id.slice(0, 8)}
                  </TableCell>
                  <TableCell className="font-medium">
                    {tr.amount} {tr.asset}
                  </TableCell>
                  <TableCell className="font-mono text-xs text-muted-foreground">
                    {getWalletLabel(tr.from_wallet_id)} &rarr;{' '}
                    {getWalletLabel(tr.to_wallet_id)}
                  </TableCell>
                  <TableCell className="text-muted-foreground">
                    {new Date(tr.created_at).toLocaleString()}
                  </TableCell>
                  <TableCell>{statusBadge(tr.status)}</TableCell>
                  <TableCell>
                    {tr.tx_hash ? (
                      <a
                        href={`https://apothem.blocksscan.io/tx/${tr.tx_hash.replace(/^0x/, '')}`}
                        target="_blank"
                        rel="noreferrer"
                        className="inline-flex items-center gap-1 text-sm font-medium text-primary hover:text-primary-hover hover:underline"
                      >
                        {tr.tx_hash.slice(0, 10)}...
                        <ExternalLink className="h-3.5 w-3.5" />
                      </a>
                    ) : (
                      <span className="text-muted-foreground">—</span>
                    )}
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        )}
      </Card>
    </div>
  );
}
