# XDC Testnet Migration Plan (Stellar → XDC/Apothem)

**Status:** Draft v1 — awaiting Phase 0 go-ahead
**Scope:** Testnet only. All mainnet-affecting items are gated (see Risk Register).
**Target:** Fluxa payment rails (wallets, transfers, settlement, FX, reconciliation) running on XDC Apothem behind a backend flag, with Stellar remaining the default.

---

## 1. Background

Fluxa is currently Stellar-only. This plan migrates the chain backend to XDC Network testnet (Apothem, chain ID 51) using an abstraction-first approach: introduce a chain-agnostic interface, implement XDC as a second backend, flip by config. No big-bang rewrite.

Verified during planning (2026-09-04):
- Apothem RPC `https://rpc.apothem.network` — `eth_chainId` = `0x33` (51) ✅
- Chain active (block ~86.3M) ✅
- Faucet `https://faucet.apothem.network` reachable ✅

---

## 2. Coupling map (current state)

| Area | Files | Stellar-specific |
|------|-------|------------------|
| Chain client | `internal/stellar/client.go` | Horizon REST (account load, submit, payment stream, path payments) |
| Keypairs | `internal/stellar/keypair.go` | ed25519 (`G...` / `S...` seeds) |
| Smart wallets | `internal/wallet/contract_*.go`, `internal/stellar/scval.go` | Soroban contract wallets |
| Settlement | `internal/transfer/service.go`, `internal/wallet/service.go` | ops via `txnbuild` |
| FX | `internal/fx/*` | `FindPathsStrict` (DEX path payments) |
| Reconciliation | `internal/reconcile/*` | Horizon balance / tx queries |
| Fiat anchors | `internal/anchor/*` | SEP-6 / SEP-10 (Stellar-only protocol) |
| UI | `apps/web` | stellar.expert explorer links |
| Config | `internal/config` | 8 × `STELLAR_*` keys |

The integration seam already exists: `stellar.Client` is an interface consumed by wallet / transfer / fx / reconcile. It becomes the nucleus of the chain abstraction.

---

## 3. Approach

Refactor `stellar.Client` into a chain-agnostic `chain.ChainClient`; move `internal/stellar` → `internal/chain/stellar` adapted to it; add `internal/chain/xdc` (EVM implementation). Backend selection at boot via config.

**Why not fork-and-replace:**
- Diff stays reviewable per package instead of a 20-file rewrite
- A/B comparison of both backends during testing
- Rollback = config change, not revert storm

### Target shape

```
internal/chain/          ← chain-agnostic
  ChainClient interface
  asset.go               ← AssetRef { symbol, decimals, contractAddress | stellar issuer }
internal/chain/stellar/  ← moved + adapted (current behavior preserved)
internal/chain/xdc/      ← new EVM implementation (go-ethereum)
```

### ChainClient (initial signature — Phase 0 spike validates against real RPC)

- `GenerateKeypair() (address string, encryptedPrivateKey string, err error)`
- `GetBalance(ctx, address string, asset AssetRef) (decimal, error)`
- `BuildPayment(ctx, from, to, asset, amount, memo) (signedTx, txHash, error)`
- `SubmitTx(ctx, signedTx) (txHash, error)`
- `TxStatus(ctx, txHash) (confirmed bool, confirmations int, err error)`
- `StreamEvents(ctx, handler)` — payment events for reconciliation
- `EstimateFee(ctx, tx) (fee AssetRef-amount, error)`

---

## 4. Decision register

| # | Decision | Choice | Caveats / notes |
|---|----------|--------|-----------------|
| 1 | FX conversion | Oracle-priced internal conversion; settle in target token from treasury/liquidity wallet | **Business risk `FX-R1`** — treasury warehouses dual-asset inventory. Testnet-only until signed off |
| 2 | Smart wallets | Custodial-only for XDC backend (Soroban wallets not ported) | **Custody risk `CUSTODY-R2`** — see Risk Register |
| 3 | Key format | secp256k1, `0x…` addresses (EVM-native, MetaMask-compatible) | Wallets table schema unchanged (public_key is text) |
| 4 | Testnet USD token | Own `FluxaUSD` ERC-20, **6 decimals** (matches USDC convention; avoids a 18-dec translation class), plus mint faucet | Deployed in Phase 0 |
| 5 | SEP anchors | Dropped for XDC backend (Stellar-specific protocol) | Fiat rails = separate future track |
| 6 | Backend flag | `CHAIN_BACKEND` config key | Default `stellar` (current behaviour, zero migration risk). `xdc` = new opt-in backend. Explicit, documented values only — no auto-detection |

---

## 5. Risk register

| ID | Risk | Owner | Gate / mitigation |
|----|------|-------|-------------------|
| FX-R1 | Oracle-priced FX puts dual-asset inventory + spread risk on treasury | **Dheeraj** (business sign-off required) | Mainnet blocked until: position limits, spread policy, and liquidity source documented and approved. Testnet proceeds unrestricted |
| CUSTODY-R2 | Custodial-only = treasury signer holds unilateral custody; no on-chain spending limits | **Owner: Dheeraj; co-owner: infra lead TBD** | Timeline: HSM / MPC / multisig evaluation decision doc due **end of Phase 5**; implementation complete before testnet-exit review. Not "if product demands" — a hard gate |
| ENG-R3 | Decimals mismatch (Stellar 7 vs EVM 18) → silent rounding bugs | Eng | Single shared unit-conversion layer at the interface boundary (`AssetRef.decimals` is authoritative). `FluxaUSD` uses 6 decimals to shrink the blast radius |
| ENG-R4 | Nonce management under concurrent worker instances | Eng | Single-flight nonce allocator per key, designed in Phase 4 (not discovered in Phase 8) |
| ENG-R5 | Gas estimation edge cases (EIP-1559 vs legacy on XDPoS) | Eng | Empirical validation in spike; fee oracle in Phase 2 |
| ENG-R6 | Reorgs / finality assumptions wrong | Eng | **Decided now, validated in spike** — see §6 |

---

## 6. Finality & confirmation policy (decided up front, not in Phase 4)

**Policy: do not trust Apothem's finality claims at face value — validate empirically during Phase 0, start conservative, harden before mainnet.**

- Phase 0 spike includes: measuring reorg behaviour on Apothem over a sample window (subsequent-block hash checks on recent history via RPC), plus a published XDPoS finality-mechanism review.
- **Initial testnet default:** `XDC_CONFIRMATIONS_REQUIRED = 6` (~12s at 2s blocks) — conservative, still fast. Configurable per environment.
- Mainnet gate: confirmation policy re-decided with real-chain data + exchange/industry practice for XDC before any mainnet deployment.

This assumption underpins `TxStatus`, transfer state machine, and reconciliation — hence it is a Phase 0 deliverable, not a late-phase surprise.

---

## 7. Phases

| Phase | Scope | Deliverable | Est. (2 devs) |
|-------|-------|-------------|----------------|
| 0 — Spike | Faucet funding; deploy `FluxaUSD`; Go keygen → sign → ERC-20 transfer → receipt poll; **finality validation (§6)**; RPC rate-limit probing | Go spike program + deployed token + go/no-go | 1–2 d |
| 1 — Chain interface | Refactor to `chain.ChainClient`; adapt Stellar impl; no behavior change | Compiles; all existing tests green | 2–3 d |
| 2 — XDC client | `internal/chain/xdc`: keygen, balances (native + ERC-20), tx lifecycle, receipt polling, event stream + **fee/gas estimation + pending-tx tracking (observability baseline)** | XDC passes same contract tests as Stellar impl | 3–5 d |
| 3 — Wallet service | Keypair swap, faucet funding bot, custodial path | Wallet create/fund/balance e2e on Apothem | 2–4 d |
| 4 — Transfers | Settlement → EVM tx, confirmations (§6 policy), memo→calldata, tx_hash tracking; **nonce allocator (ENG-R4)** | Transfers settle on Apothem; idempotency preserved | 4–6 d |
| 5 — FX | Decision #1 implementation (oracle-priced conversion + treasury settle) | FX quotes + conversion on testnet | 3–5 d |
| 6 — Reconciliation | Horizon → RPC lookups; discrepancy detection ported | Reconcile job green against Apothem | 2–3 d |
| 7 — Config + UI + SDK | `CHAIN_BACKEND`, `XDC_RPC_URL`, `XDC_CHAIN_ID`, token registry; explorer links → apothem.blocksscan.io; **EVM observability suite (gas tracker, chain-head lag vs local view, RPC health as /health component)**; SDK parity; Postman env; docs | Full stack on XDC testnet | 3–4 d |
| 8 — Hardening | e2e on Apothem; failure modes (RPC down, reorgs, stuck tx); ops runbook | Release-ready testnet | 4–6 d |

**Revised total: ~5–7 weeks** (Phases 4 + 8 carry explicit buffer for ENG-R4/R5/R6, per review).

---

## 8. Observability (EVM-specific, new vs. Stellar backend)

Stellar's flat-fee model meant Fluxa never tracked gas. XDC adds:

- Gas price tracker + alerting (fees spiking → settlement delays)
- Pending-tx pool visibility (stuck tx detection before user impact)
- Chain-head lag vs. local view (RPC provider lag detection)
- RPC provider health as a first-class `/health` component (replaces the `horizon` component; note the horizon health probe has a pre-existing unmarshal bug at the current commit)
- Nonce-gap alarms

Baseline lands in Phase 2; dashboarding/alerts in Phase 7.

---

## 9. Open items before "go"

- [ ] Sign-off on §4 decision register (esp. #1 FX, #2 custody)
- [ ] Confirm CUSTODY-R2 co-owner (infra lead)
- [ ] Pick spike RPC endpoint: public `rpc.apothem.network` vs. private/paid endpoint (public is rate-limited; acceptable for spike)

---

*Prepared by Hermes; review input incorporated from team review (formatting fix on CHAIN_BACKEND row, CUSTODY-R2 ownership, finality-policy timing).*
