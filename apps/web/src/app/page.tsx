import Link from "next/link";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { ArrowRight, ArrowRightLeft, Wallet, Send, Repeat, Shield, Globe } from "lucide-react";

const features = [
  {
    icon: Wallet,
    title: "Wallets",
    description: "Create XDC wallets with encrypted private keys. Custodial by default.",
  },
  {
    icon: Send,
    title: "Transfers",
    description: "Instant on-chain XDC transfers with tx hash and compliance screening.",
  },
  {
    icon: ArrowRightLeft,
    title: "FX Conversion",
    description: "Real-time rates from CoinGecko. Swap USDC and TXDC seamlessly.",
  },
  {
    icon: Repeat,
    title: "Batch & Schedules",
    description: "Send up to 100 transfers in one call. Set up recurring daily payouts.",
  },
  {
    icon: Shield,
    title: "Compliance",
    description: "OFAC sanctions screening, velocity checks, and human review queue.",
  },
  {
    icon: Globe,
    title: "Fiat Rails",
    description: "Deposit and withdraw via Flutterwave or Stripe. Multi-currency support.",
  },
];

export default function Home() {
  return (
    <div className="flex min-h-screen flex-col">
      {/* Header */}
      <header className="border-b border-border">
        <div className="mx-auto flex max-w-6xl items-center justify-between px-6 py-4">
          <div className="flex items-center gap-2">
            <div className="flex h-8 w-8 items-center justify-center rounded-lg bg-primary text-primary-foreground font-bold text-sm">
              F
            </div>
            <span className="text-lg font-semibold">FlowX</span>
          </div>
          <div className="flex items-center gap-3">
            <Link href="/login">
              <Button variant="ghost" size="sm">Sign In</Button>
            </Link>
            <Link href="/login">
              <Button size="sm">Get Started</Button>
            </Link>
          </div>
        </div>
      </header>

      {/* Hero */}
      <section className="flex flex-1 flex-col items-center justify-center px-6 py-24 text-center">
        <div className="mx-auto max-w-3xl">
          <h1 className="text-4xl font-bold tracking-tight sm:text-5xl lg:text-6xl">
            Cross-border payments
            <br />
            <span className="text-primary">for emerging markets</span>
          </h1>
          <p className="mt-6 text-lg text-muted-foreground max-w-2xl mx-auto">
            A programmable payments API built on the XDC Network. Move value across borders
            with wallets, transfers, FX, batch payments, and fiat rails — behind a clean REST API.
          </p>
          <div className="mt-8 flex items-center justify-center gap-4">
            <Link href="/login">
              <Button size="lg" className="gap-2">
                Start Building <ArrowRight className="h-4 w-4" />
              </Button>
            </Link>
            <a href="https://github.com/XDCIndia/FLOWX" target="_blank" rel="noopener noreferrer">
              <Button variant="outline" size="lg">View on GitHub</Button>
            </a>
          </div>
        </div>
      </section>

      {/* Code snippet */}
      <section className="border-t border-border bg-muted/50 px-6 py-16">
        <div className="mx-auto max-w-2xl">
          <p className="text-center text-sm font-medium text-muted-foreground mb-4">
            Get started in 3 lines
          </p>
          <div className="rounded-lg border border-border bg-background p-6 font-mono text-sm">
            <pre className="overflow-x-auto"><code>{`import { FluxaClient } from "@savitura/fluxa";

const client = new FluxaClient({ apiKey: "fx_sk_..." });

const wallet = await client.wallets.create();
const tx = await client.transfers.create({
  from_wallet_id: wallet.id,
  to_wallet_id: "recipient-id",
  asset: "TXDC",
  amount: "10.0000000",
});`}</code></pre>
          </div>
        </div>
      </section>

      {/* Features */}
      <section className="px-6 py-20">
        <div className="mx-auto max-w-6xl">
          <h2 className="text-center text-3xl font-bold">Everything you need</h2>
          <p className="mt-3 text-center text-muted-foreground">
            Payment primitives behind a clean API
          </p>
          <div className="mt-12 grid gap-6 sm:grid-cols-2 lg:grid-cols-3">
            {features.map((f) => (
              <Card key={f.title}>
                <CardHeader>
                  <div className="flex h-10 w-10 items-center justify-center rounded-lg bg-primary/10 text-primary">
                    <f.icon className="h-5 w-5" />
                  </div>
                  <CardTitle className="mt-3 text-lg">{f.title}</CardTitle>
                </CardHeader>
                <CardContent>
                  <CardDescription>{f.description}</CardDescription>
                </CardContent>
              </Card>
            ))}
          </div>
        </div>
      </section>

      {/* Footer */}
      <footer className="border-t border-border px-6 py-8">
        <div className="mx-auto flex max-w-6xl items-center justify-between text-sm text-muted-foreground">
          <p>FlowX — Cross-border payment infrastructure</p>
          <p>Built on XDC Network</p>
        </div>
      </footer>
    </div>
  );
}
