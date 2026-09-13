import type { ReactElement } from "react";

/**
 * Hand-crafted inline SVG architecture diagram for FlowX.
 * Dark-optimized, no external images. Purely presentational — SSR-safe.
 *
 * Layout: Client -> Next.js Dashboard -> Go API (Chi) -> Route Engine
 * (3 legs: direct on-chain, FlowXPool AMM, Stripe bank) -> XDC Apothem,
 * with PostgreSQL / Redis sidecars under the Go API.
 */
export function ArchitectureDiagram(): ReactElement {
  return (
    <svg
      viewBox="0 0 960 620"
      role="img"
      aria-labelledby="fx-arch-title fx-arch-desc"
      className="h-auto w-full min-w-[680px]"
      fontFamily="ui-sans-serif, system-ui, sans-serif"
    >
      <title id="fx-arch-title">FlowX system architecture</title>
      <desc id="fx-arch-desc">
        Request flow from a web client through the Next.js dashboard and Go API
        into a route engine that settles directly on-chain, via the FlowXPool
        AMM, or through Stripe bank rails, with PostgreSQL and Redis as sidecars.
      </desc>

      <style>{`
        .fx-box { fill: rgba(255,255,255,0.03); stroke: rgba(255,255,255,0.12); }
        .fx-leg { fill: rgba(99,91,255,0.06); stroke: rgba(129,140,248,0.45); }
        .fx-chain { fill: rgba(52,211,153,0.05); stroke: rgba(52,211,153,0.55); }
        .fx-title { fill: #e4e4e7; font-size: 13px; font-weight: 600; }
        .fx-sub { fill: #71717a; font-size: 10.5px; }
        .fx-note { fill: #52525b; font-size: 10px; }
        .fx-flow {
          stroke: #6366f1; stroke-width: 1.5; fill: none;
          stroke-dasharray: 6 6; animation: fx-dash 1.4s linear infinite;
        }
        .fx-sidecar {
          stroke: #52525b; stroke-width: 1.25; fill: none; stroke-dasharray: 4 4;
        }
        .fx-static { stroke: #3f3f46; stroke-width: 1.5; fill: none; }
        @keyframes fx-dash { to { stroke-dashoffset: -12; } }
        @media (prefers-reduced-motion: reduce) {
          .fx-flow { animation: none; }
        }
      `}</style>

      <defs>
        <pattern id="fx-grid" width="40" height="40" patternUnits="userSpaceOnUse">
          <path d="M40 0 H0 V40" fill="none" stroke="rgba(255,255,255,0.035)" strokeWidth="1" />
        </pattern>
        <marker id="fx-arrow" viewBox="0 0 10 10" refX="8" refY="5" markerWidth="6.5" markerHeight="6.5" orient="auto-start-reverse">
          <path d="M0 0 L10 5 L0 10 z" fill="#6366f1" />
        </marker>
        <marker id="fx-arrow-dim" viewBox="0 0 10 10" refX="8" refY="5" markerWidth="6" markerHeight="6" orient="auto-start-reverse">
          <path d="M0 0 L10 5 L0 10 z" fill="#52525b" />
        </marker>
      </defs>

      <rect x="0" y="0" width="960" height="620" fill="url(#fx-grid)" rx="12" />

      {/* ---- Main pipeline row (y center 266) ---- */}

      {/* Client */}
      <rect className="fx-box" x="20" y="230" width="140" height="72" rx="10" />
      <text className="fx-title" x="90" y="262" textAnchor="middle">Client</text>
      <text className="fx-sub" x="90" y="280" textAnchor="middle">SDK · Dashboard</text>

      <path className="fx-flow" d="M165 266 H212" markerEnd="url(#fx-arrow)" />

      {/* Next.js dashboard */}
      <rect className="fx-box" x="220" y="230" width="150" height="72" rx="10" />
      <text className="fx-title" x="295" y="262" textAnchor="middle">Next.js 14</text>
      <text className="fx-sub" x="295" y="280" textAnchor="middle">Dashboard · App Router</text>

      <path className="fx-flow" d="M375 266 H422" markerEnd="url(#fx-arrow)" />

      {/* Go API */}
      <rect className="fx-box" x="430" y="230" width="150" height="72" rx="10" />
      <text className="fx-title" x="505" y="262" textAnchor="middle">Go API</text>
      <text className="fx-sub" x="505" y="280" textAnchor="middle">Go 1.24 · Chi</text>

      <path className="fx-flow" d="M585 266 H612" markerEnd="url(#fx-arrow)" />

      {/* ---- Sidecars ---- */}
      <path className="fx-sidecar" d="M505 306 V364" markerEnd="url(#fx-arrow-dim)" />
      <rect className="fx-box" x="430" y="370" width="150" height="56" rx="10" />
      <text className="fx-title" x="505" y="395" textAnchor="middle">PostgreSQL</text>
      <text className="fx-sub" x="505" y="411" textAnchor="middle">ledger · tenants</text>

      <path className="fx-sidecar" d="M505 430 V454" markerEnd="url(#fx-arrow-dim)" />
      <rect className="fx-box" x="430" y="460" width="150" height="56" rx="10" />
      <text className="fx-title" x="505" y="485" textAnchor="middle">Redis</text>
      <text className="fx-sub" x="505" y="501" textAnchor="middle">queues · cache</text>

      {/* ---- Route engine container ---- */}
      <rect className="fx-box" x="620" y="50" width="170" height="460" rx="14" strokeWidth="1.25" />
      <text className="fx-title" x="705" y="82" textAnchor="middle">Route Engine</text>
      <text className="fx-sub" x="705" y="98" textAnchor="middle">smart routing</text>

      {/* Leg 1: direct on-chain */}
      <rect className="fx-leg" x="635" y="110" width="140" height="88" rx="10" />
      <text className="fx-title" x="705" y="145" textAnchor="middle">Direct on-chain</text>
      <text className="fx-sub" x="705" y="163" textAnchor="middle">XDC transfer</text>
      <text className="fx-note" x="705" y="180" textAnchor="middle">wallet to wallet</text>

      {/* Leg 2: AMM */}
      <rect className="fx-leg" x="635" y="228" width="140" height="88" rx="10" />
      <text className="fx-title" x="705" y="263" textAnchor="middle">FlowXPool AMM</text>
      <text className="fx-sub" x="705" y="281" textAnchor="middle">tUSDC ⇄ TXDC swap</text>
      <text className="fx-note" x="705" y="298" textAnchor="middle">deployed pool</text>

      {/* Leg 3: Stripe bank */}
      <rect className="fx-leg" x="635" y="346" width="140" height="88" rx="10" />
      <text className="fx-title" x="705" y="381" textAnchor="middle">Stripe Bank</text>
      <text className="fx-sub" x="705" y="399" textAnchor="middle">fiat rail</text>
      <text className="fx-note" x="705" y="416" textAnchor="middle">sandbox mode</text>

      <text className="fx-note" x="705" y="474" textAnchor="middle">ODL-ready 4th leg</text>
      <text className="fx-note" x="705" y="489" textAnchor="middle">partnership required</text>

      {/* ---- Settlement targets ---- */}

      {/* XDC Apothem */}
      <path className="fx-flow" d="M780 154 H800 V175 H812" />
      <path className="fx-flow" d="M780 272 H812" />
      <rect className="fx-chain" x="820" y="160" width="120" height="160" rx="12" strokeWidth="1.5" />
      <text x="880" y="214" textAnchor="middle" fill="#34d399" fontSize="20" fontWeight="700">XDC</text>
      <text className="fx-title" x="880" y="238" textAnchor="middle">Apothem Testnet</text>
      <text className="fx-sub" x="880" y="256" textAnchor="middle">tUSDC · TXDC</text>
      <text className="fx-note" x="880" y="298" textAnchor="middle">on-chain settlement</text>

      {/* Fiat rails (Stripe target) */}
      <path className="fx-flow" d="M780 390 H800 V455 H812" />
      <rect className="fx-box" x="820" y="420" width="120" height="70" rx="12" />
      <text className="fx-title" x="880" y="450" textAnchor="middle">Fiat rails</text>
      <text className="fx-sub" x="880" y="468" textAnchor="middle">Stripe · sandbox</text>

      {/* ---- Legend ---- */}
      <path className="fx-flow" d="M40 585 H80" />
      <text className="fx-note" x="90" y="589">live request path</text>
      <path className="fx-sidecar" d="M220 585 H260" />
      <text className="fx-note" x="270" y="589">sidecar (dashed)</text>
    </svg>
  );
}
