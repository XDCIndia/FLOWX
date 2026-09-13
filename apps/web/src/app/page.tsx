import Link from "next/link";
import type { Metadata } from "next";
import FluidCanvas from "@/components/landing/fluid-canvas";
import SmoothScroll from "@/components/landing/smooth-scroll";
import {
  AlertTriangle,
  ArrowRight,
  ArrowRightLeft,
  CalendarClock,
  CheckCircle2,
  ExternalLink,
  GitBranch,
  Layers,
  Send,
  ShieldCheck,
  Wallet,
  Webhook,
} from "lucide-react";
import { ArchitectureDiagram } from "@/components/landing/architecture";

export const metadata: Metadata = {
  title: "FlowX — Cross-border Payment Infrastructure on XDC Network",
  description:
    "Programmable cross-border payments on the XDC Network: wallets, transfers, FX, batch payouts, and a multi-leg route engine — live on Apothem testnet.",
};

const XDCSCAN = "https://testnet.xdcscan.com";

const contracts = [
  {
    label: "tUSDC Token",
    kind: "ERC-20 test token",
    address: "xdce069a90d55fdeca2cbec7793712aa3d807ffef9b",
    href: `${XDCSCAN}/address/xdce069a90d55fdeca2cbec7793712aa3d807ffef9b`,
    note: "Faucet-minted stablecoin used across the platform.",
  },
  {
    label: "FlowXPool AMM",
    kind: "Deployed pool contract",
    address: "xdc4a3a728562847a1fd9afdbd81356533eb05b12cd",
    href: `${XDCSCAN}/address/xdc4a3a728562847a1fd9afdbd81356533eb05b12cd`,
    note: "Seeded with 10 TXDC + 35,000 tUSDC of real testnet liquidity.",
  },
  {
    label: "Real swap — tx hash",
    kind: "Verified on-chain",
    address: "0x7f23e14a…aa2c771",
    href: `${XDCSCAN}/tx/7f23e14a6ae79db25aee72d3c9329bb17c8a4581ca9860dfc9edc1e80aa2c771`,
    note: "An AMM route executed end-to-end by the route engine.",
  },
];

const features = [
  {
    icon: Wallet,
    title: "Wallets",
    description: "Create XDC wallets with AES-encrypted private keys, custodial by default.",
  },
  {
    icon: Send,
    title: "Transfers",
    description: "Instant on-chain transfers with a returned tx hash you can verify yourself.",
  },
  {
    icon: ArrowRightLeft,
    title: "FX conversion",
    description: "Live rates with real tUSDC ⇄ TXDC swaps executed on the FlowXPool AMM.",
  },
  {
    icon: Layers,
    title: "Batch payments",
    description: "Fire up to 100 transfers in a single API call, settled atomically.",
  },
  {
    icon: CalendarClock,
    title: "Scheduled payouts",
    description: "Recurring daily payouts driven by the Redis-backed scheduler.",
  },
  {
    icon: GitBranch,
    title: "Route engine",
    description: "Every payment picks the best leg: direct on-chain, AMM swap, or Stripe bank.",
  },
  {
    icon: Webhook,
    title: "Signed webhooks",
    description: "HMAC-signed delivery with SSRF protection on callback URLs.",
  },
  {
    icon: ShieldCheck,
    title: "Compliance",
    description: "OFAC sanctions screening, velocity limits, and a human review queue.",
  },
];

const simulated = [
  {
    title: "Stripe rail is sandbox-only",
    body: "Fiat deposit and withdrawal run against Stripe's test mode. No real money ever moves — flipping to live keys is a config change, not a rewrite.",
  },
  {
    title: "Ripple ODL is adapter-ready",
    body: "The route engine's fourth leg is designed and typed, but activating real ODL settlement requires a Ripple partnership. We say that plainly.",
  },
  {
    title: "Testnet funds only",
    body: "Everything settles on XDC Apothem with tUSDC and TXDC from the faucet. The contracts, liquidity, and swap below are real and verifiable on-chain.",
  },
];

const stack = ["Go 1.24 + Chi", "Next.js 14", "Tailwind CSS", "PostgreSQL", "Redis", "XDC Network"];

function GithubMark({ className }: { className?: string }) {
  return (
    <svg viewBox="0 0 24 24" fill="currentColor" aria-hidden="true" className={className}>
      <path d="M12 .5C5.65.5.5 5.65.5 12c0 5.08 3.29 9.39 7.86 10.91.58.11.79-.25.79-.55 0-.27-.01-1.17-.02-2.12-3.2.7-3.88-1.36-3.88-1.36-.52-1.33-1.28-1.68-1.28-1.68-1.04-.71.08-.7.08-.7 1.15.08 1.76 1.19 1.76 1.19 1.03 1.75 2.69 1.25 3.34.95.1-.74.4-1.25.72-1.53-2.55-.29-5.24-1.28-5.24-5.69 0-1.26.45-2.29 1.19-3.09-.12-.29-.52-1.46.11-3.05 0 0 .97-.31 3.17 1.18a11 11 0 0 1 5.77 0c2.2-1.49 3.16-1.18 3.16-1.18.63 1.59.24 2.76.12 3.05.74.8 1.18 1.83 1.18 3.09 0 4.43-2.69 5.4-5.26 5.68.41.36.78 1.06.78 2.14 0 1.55-.02 2.79-.02 3.17 0 .31.21.67.8.55A11.51 11.51 0 0 0 23.5 12C23.5 5.65 18.35.5 12 .5Z" />
    </svg>
  );
}

export default function Home() {
  return (
    <div className="flex min-h-screen flex-col bg-white text-zinc-600">
      <SmoothScroll />
      {/* White hero — original landing.html theme (fluid ink canvas) */}
      <section className="relative overflow-hidden bg-white text-zinc-900">
        <FluidCanvas />
        {/* white scrim keeps the ink behind the text legible */}
        <div
          aria-hidden="true"
          className="pointer-events-none absolute inset-0 bg-[radial-gradient(ellipse_75%_65%_at_50%_45%,rgba(255,255,255,0.85),rgba(255,255,255,0.35)_60%,rgba(255,255,255,0))]"
        />
        <header className="relative mx-auto flex w-full max-w-6xl items-center justify-between px-6 py-5">
          <Link href="/" className="flex items-center gap-2.5">
            <div className="flex h-8 w-8 items-center justify-center rounded-lg bg-zinc-900 text-sm font-bold text-white">
              F
            </div>
            <span className="text-lg font-semibold text-zinc-900">FlowX</span>
          </Link>
          <nav className="hidden items-center gap-1 rounded-full border border-zinc-900/10 bg-white/60 px-2 py-1.5 text-sm text-zinc-500 backdrop-blur md:flex">
            <a href="#architecture" className="rounded-full px-3 py-1 transition-colors hover:text-zinc-900">How it works</a>
            <a href="#proof" className="rounded-full px-3 py-1 transition-colors hover:text-zinc-900">Proof</a>
            <a href="#features" className="rounded-full px-3 py-1 transition-colors hover:text-zinc-900">Features</a>
          </nav>
          <div className="flex items-center gap-3">
            <a
              href="https://github.com/XDCIndia/FLOWX"
              target="_blank"
              rel="noopener noreferrer"
              className="rounded-lg p-2 text-zinc-500 transition-colors hover:bg-zinc-900/5 hover:text-zinc-900"
              aria-label="FlowX on GitHub"
            >
              <GithubMark className="h-5 w-5" />
            </a>
            <Link
              href="/login"
              className="rounded-full bg-zinc-900 px-4 py-2 text-sm font-medium text-white transition-colors hover:bg-zinc-700"
            >
              Launch App
            </Link>
          </div>
        </header>

        <div className="relative mx-auto max-w-4xl px-6 pb-28 pt-14 text-center sm:pt-20">
          <p className="inline-flex items-center gap-2 rounded-full border border-zinc-900/10 bg-white/60 px-3 py-1 text-xs font-medium text-zinc-500 backdrop-blur">
            <span className="relative flex h-2 w-2">
              <span className="absolute inline-flex h-full w-full animate-ping rounded-full bg-emerald-500 opacity-60" />
              <span className="relative inline-flex h-2 w-2 rounded-full bg-emerald-500" />
            </span>
            Live on the XDC Apothem testnet · Chain ID 51
          </p>
          <h1 className="mt-6 text-4xl font-bold tracking-tight text-zinc-900 animate-in fade-in slide-in-from-bottom-4 duration-700 sm:text-6xl lg:text-7xl">
            Move money across borders on the XDC Network.
          </h1>
          <p className="mx-auto mt-6 max-w-2xl text-lg text-zinc-500 animate-in fade-in slide-in-from-bottom-4 duration-700">
            An open, programmable payments platform — EVM-native wallets, cross-border transfers,
            FX conversion and on-chain settlement behind one clean REST API.
          </p>
          <div className="mx-auto mt-10 flex max-w-md items-center gap-2 rounded-full border border-zinc-900/10 bg-white/60 p-1.5 shadow-sm backdrop-blur animate-in fade-in slide-in-from-bottom-4 duration-700">
            <input
              type="email"
              placeholder="you@company.com"
              aria-label="Work email"
              className="w-full flex-1 bg-transparent px-4 text-sm text-zinc-900 outline-none placeholder:text-zinc-400"
            />
            <Link
              href="/login"
              className="shrink-0 rounded-full bg-zinc-900 px-5 py-2.5 text-sm font-semibold text-white transition-colors hover:bg-zinc-700"
            >
              Launch App →
            </Link>
          </div>
        </div>
      </section>

      {/* Architecture */}
      <section id="architecture" className="border-t border-zinc-900/10 px-6 py-20">
        <div className="mx-auto max-w-6xl">
          <h2 className="text-center text-3xl font-bold text-zinc-900">How money moves</h2>
          <p className="mx-auto mt-3 max-w-2xl text-center text-zinc-500">
            Every request flows through the dashboard into the Go API, where the route engine
            selects a settlement leg. On-chain legs settle on XDC Apothem; the Stripe leg runs in
            sandbox and is built to flip to production keys.
          </p>
          <div className="mt-12 overflow-x-auto rounded-2xl border border-zinc-900/10 bg-white p-4 shadow-sm sm:p-6">
            <ArchitectureDiagram />
          </div>
        </div>
      </section>

      {/* Proof it's real */}
      <section id="proof" className="border-t border-zinc-900/10 px-6 py-20">
        <div className="mx-auto max-w-6xl">
          <div className="flex flex-col items-start justify-between gap-4 sm:flex-row sm:items-end">
            <div>
              <h2 className="text-3xl font-bold text-zinc-900">Proof it&apos;s real</h2>
              <p className="mt-3 max-w-2xl text-zinc-500">
                No mock addresses. These contracts are deployed on XDC Apothem and the swap below
                was executed by the route engine — click through and verify on xdcscan.
              </p>
            </div>
            <span className="inline-flex items-center gap-1.5 rounded-full border border-emerald-600/30 bg-emerald-500/10 px-3 py-1 text-xs font-medium text-emerald-700">
              <CheckCircle2 className="h-3.5 w-3.5" /> Verified on-chain
            </span>
          </div>
          <div className="mt-10 grid gap-4 lg:grid-cols-3">
            {contracts.map((c) => (
              <a
                key={c.address}
                href={c.href}
                target="_blank"
                rel="noopener noreferrer"
                className="group rounded-xl border border-zinc-900/10 bg-white p-6 shadow-sm transition-colors hover:border-indigo-400 hover:shadow-md"
              >
                <div className="flex items-center justify-between">
                  <p className="text-sm font-semibold text-zinc-900">{c.label}</p>
                  <ExternalLink className="h-4 w-4 text-zinc-400 transition-colors group-hover:text-indigo-500" />
                </div>
                <p className="mt-1 text-xs uppercase tracking-wide text-zinc-400">{c.kind}</p>
                <p className="mt-4 break-all rounded-lg bg-zinc-900/[0.04] px-3 py-2 font-mono text-xs text-indigo-600">
                  {c.address}
                </p>
                <p className="mt-4 text-sm leading-relaxed text-zinc-500">{c.note}</p>
              </a>
            ))}
          </div>
        </div>
      </section>

      {/* Features */}
      <section id="features" className="border-t border-zinc-900/10 px-6 py-20">
        <div className="mx-auto max-w-6xl">
          <h2 className="text-center text-3xl font-bold text-zinc-900">Everything behind the API</h2>
          <p className="mx-auto mt-3 max-w-2xl text-center text-zinc-500">
            Eight payment primitives, one REST surface.
          </p>
          <div className="mt-12 grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
            {features.map((f) => (
              <div
                key={f.title}
                className="rounded-xl border border-zinc-900/10 bg-white p-6 shadow-sm transition-colors hover:border-indigo-400"
              >
                <div className="flex h-10 w-10 items-center justify-center rounded-lg bg-indigo-500/10 text-indigo-600">
                  <f.icon className="h-5 w-5" />
                </div>
                <h3 className="mt-4 font-semibold text-zinc-900">{f.title}</h3>
                <p className="mt-2 text-sm leading-relaxed text-zinc-500">{f.description}</p>
              </div>
            ))}
          </div>
        </div>
      </section>

      {/* What is simulated */}
      <section className="border-t border-zinc-900/10 px-6 py-20">
        <div className="mx-auto max-w-6xl">
          <h2 className="text-3xl font-bold text-zinc-900">What is simulated</h2>
          <p className="mt-3 max-w-2xl text-zinc-500">
            A demo you can trust starts with saying what isn&apos;t production yet.
          </p>
          <div className="mt-10 grid gap-4 md:grid-cols-3">
            {simulated.map((s) => (
              <div key={s.title} className="rounded-xl border border-amber-500/30 bg-amber-500/[0.06] p-6">
                <div className="flex items-center gap-2 text-amber-600">
                  <AlertTriangle className="h-4 w-4" />
                  <h3 className="font-semibold">{s.title}</h3>
                </div>
                <p className="mt-3 text-sm leading-relaxed text-zinc-500">{s.body}</p>
              </div>
            ))}
          </div>
        </div>
      </section>

      {/* Footer */}
      <footer className="mt-auto border-t border-zinc-900/10 px-6 py-10">
        <div className="mx-auto flex max-w-6xl flex-col gap-6 sm:flex-row sm:items-center sm:justify-between">
          <div>
            <div className="flex items-center gap-2.5">
              <div className="flex h-7 w-7 items-center justify-center rounded-lg bg-zinc-900 text-xs font-bold text-white">
                F
              </div>
              <span className="font-semibold text-zinc-900">FlowX</span>
            </div>
            <p className="mt-2 text-sm text-zinc-500">
              Cross-border payment infrastructure on XDC Network.
            </p>
          </div>
          <div className="flex flex-wrap items-center gap-2">
            {stack.map((s) => (
              <span
                key={s}
                className="rounded-full border border-zinc-900/10 bg-zinc-900/[0.03] px-3 py-1 text-xs text-zinc-500"
              >
                {s}
              </span>
            ))}
          </div>
          <a
            href="https://github.com/XDCIndia/FLOWX"
            target="_blank"
            rel="noopener noreferrer"
            className="inline-flex items-center gap-2 text-sm text-zinc-500 transition-colors hover:text-zinc-900"
          >
            <GithubMark className="h-4 w-4" /> github.com/XDCIndia/FLOWX
          </a>
        </div>
      </footer>
    </div>
  );
}
