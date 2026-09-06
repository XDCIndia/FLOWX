# FlowX — How Each Feature Works (Simple English)

> One-line idea: **FlowX moves money across countries using the XDC blockchain as the highway.**
> Postgres is the notebook (ledger) where every movement is recorded first.
> The blockchain is the truck that actually carries the money.

---

## 1. Wallets 👛

**What it is:** Your account's pocket. Every user gets wallets that hold digital assets like TXDC and USDC.

**How it works:**
1. You sign up. FlowX creates a wallet for you and keeps the private key safe (custodial — like a bank keeping your money).
2. Your wallet has a public address (like an account number) on the XDC testnet.
3. You add money by buying TXDC through the Fiat feature (below) or receiving a transfer.
4. Every balance change is written in the ledger (Postgres) as double-entry bookkeeping — money never appears or disappears, it always moves from one wallet to another.

**API:** `POST /v1/wallets` (needs an `Idempotency-Key` header — a unique ID so retrying never double-creates)

---

## 2. FX Rates, Quotes & Conversions 💱

**What it is:** Changing one currency into another, at a fair live price.

**How it works — in 3 steps:**

1. **Rates** (`GET /v1/fx/rates`) — FlowX asks CoinGecko for the live market price (e.g., 1 USD = 83 INR). If CoinGecko is down, it falls back to the Stellar network, then to a built-in static table. Always answers.

2. **Quote** (`POST /v1/fx/quotes`) — Locks the price for about 30 minutes. Like a shopkeeper saying *"I'll sell you EUR at this rate if you decide within 30 min."* Nothing moves yet — it's just a price promise.

3. **Conversion** (`POST /v1/fx/conversions`) — Actually performs the swap using the locked quote. Happens instantly inside the ledger: your INR balance goes down, your EUR balance goes up. No blockchain needed for this step.

**Fee:** a small percentage (about 0.25%) is taken on each conversion.

---

## 3. Transfers (Sending Money) 💸

**What it is:** Sending value from your wallet to another wallet.

**How it works:**
1. You request a transfer with an **Idempotency-Key** (a unique random ID). If your network fails and you retry, the key prevents sending the money twice.
2. FlowX checks your balance and takes the transfer fee (about 0.30%).
3. It writes the movement in the ledger (your balance down, receiver's balance up).
4. If the transfer must settle on the blockchain, a transaction is sent from FlowX's **treasury wallet** and waits for 6 confirmations (≈ 2 seconds on XDC).
5. Status changes: `pending` → `processing` → `completed` (or `failed`).

**API:** `POST /v1/transfers`

> ⚠️ Current status: the treasury wallet needs TXDC (testnet coins) for on-chain settlement. Quotes and ledger steps work; the final on-chain step succeeds only when the treasury is funded.

---

## 4. Batches 📦

**What it is:** Paying many people in one action — like a company running payroll for 200 employees at once.

**How it works:**
1. You create a batch with a list of transfers, all sharing one `batch_id`.
2. FlowX creates every transfer inside one database transaction — either **all succeed or none do** (atomic).
3. Each transfer then settles through the normal transfer pipeline.
4. You can watch the batch's overall status: how many items are done, pending, or failed.

**API:** `POST /v1/batches`

---

## 5. Schedules ⏰

**What it is:** Automatic recurring payments — "send 50 TXDC every Monday."

**How it works:**
1. You create a schedule with an amount, a destination, and a timing rule (daily / weekly / monthly).
2. A background **worker** (a separate program, `cmd/worker`) checks Redis every minute for due schedules.
3. When a schedule is due, the worker claims it (so two workers never fire it twice), runs the transfer, and moves `next_run_at` to the next date.
4. Each firing appears as a normal transfer in your history.

**API:** `POST /v1/schedules` — needs the worker running.

---

## 6. Fiat On-Ramp (Buying Crypto with Normal Money) 💳

**What it is:** Getting TXDC into your wallet using regular money (Naira, Dollars) — the "entrance door" from the normal world.

**How it works:**
1. You request a fiat payment (e.g., 10,000 NGN) through a checkout (Flutterwave in production; a local simulator in development).
2. You "pay" in the checkout.
3. A **webhook** (an automatic notification from the payment provider) arrives at FlowX confirming the payment.
4. FlowX credits TXDC to your wallet at the current FX rate.

**API:** `POST /v1/fiat/payments` + webhook at `POST /v1/webhooks/flutterwave`

---

## 7. Payment Routing (The Intelligence Engine) 🧠🏆

**What it is:** The star feature. When you want to send money abroad, there are several possible "roads." FlowX asks every road for its price, scores them, and recommends the best one.

**The three roads (routes):**

| Route | Type | Settlement time | Example fee |
|---|---|---|---|
| `stripe_bank` | Traditional bank transfer | ~3 days | ₹2,003 |
| `payment_network` | Modern payment network (Ripple-style) | ~5 hours | ₹1,200 |
| `xdc_bridge` | Blockchain (XDC) | ~12 seconds | ₹500 |

**How it works:**
1. **Evaluate** — FlowX's *evaluator* asks every route that supports your corridor (e.g., INR → EUR): *"How much for 100,000 INR?"* Each route answers with fee, exchange rate, and settlement time.
2. **Score** — the *scorer* grades each option 0–100 on cost, speed, reliability, liquidity, and compliance. The weights depend on your preference: **balanced**, **cheapest**, **fastest**, or **most_reliable**.
3. **Compliance check** — a route can be blocked or warned if the payment looks risky.
4. **Recommend** — the highest score gets the 🏆. Blocked routes are removed.
5. **Send** — you can execute the recommended route (or pick another one manually) with `POST /v1/payments/send`.

**Real example (tested live):**

```
$100,000 INR → EUR
🏆 xdc_bridge       fee=₹500    12 seconds   score 83.45
   payment_network  fee=₹1,200  5 hours      score 60.83
   stripe_bank      fee=₹2,003  3 days       score 51.54
```

**API:** `POST /v1/payments/quote` · `POST /v1/payments/send` · `GET /v1/payments/routes`

---

## 8. Fees 💰

**What it is:** Small charges on movements — how the platform earns.

**How it works:**
- Fee schedules are stored per asset pair, with a wildcard default (e.g., 0.30% for transfers, 0.25% for conversions).
- When you quote or transfer, the fee is calculated and shown to you *before* you confirm.
- The fee is taken as part of the ledger entry — everything is transparent and recorded.

---

## 9. Compliance & Risk 🛡️

**What it is:** Safety checks on payments — limits, sanctions screening, and risk scoring.

**How it works:**
- Payment limits are configurable per tenant (e.g., max $10,000 per transfer without approval).
- The routing engine runs a compliance check on each route before recommending it.
- A risk page in the dashboard shows the current risk posture.
- In development, compliance is relaxed so you can test freely.

---

## 10. API Keys 🔑

**What it is:** Programmatic access — like a password for scripts and apps.

**How it works:**
1. After logging in, you mint a key: `POST /v1/keys` → returns a secret starting with `sk_live_`.
2. Send it as `Authorization: Bearer sk_live_...` on API calls.
3. Keys can be revoked at any time.

---

## The Big Picture (One Diagram)

```
                    ┌─────────────────────────────────────────┐
   You (UI / API)   │                FLOWX                     │
        │           │                                          │
        ▼           │  ┌─────────┐   ┌──────────┐  ┌────────┐ │
   ┌─────────┐      │  │ Wallets │──▶│   FX     │  │Transfer│ │
   │  Login  │      │  │(custody)│   │(convert) │  │(send)  │ │
   │ / Keys  │      │  └─────────┘   └──────────┘  └───┬────┘ │
   └─────────┘      │       │                │          │      │
                    │       ▼                ▼          ▼      │
                    │  ┌──────────────────────────────────┐   │
                    │  │         LEDGER (Postgres)        │   │
                    │  │   every movement recorded here   │   │
                    │  └──────────────────────────────────┘   │
                    │       ▲                                   │
                    │  ┌────┴───────────┐   ┌───────────────┐  │
                    │  │ Payments       │   │ Schedules     │  │
                    │  │ Routing Engine │   │ (worker cron) │  │
                    │  └────────────────┘   └───────────────┘  │
                    │              │                            │
                    └──────────────┼────────────────────────────┘
                                   ▼
                    ┌──────────────────────────┐
                    │  XDC Blockchain (settle) │
                    │  + Fiat rail (enter)     │
                    └──────────────────────────┘
```

**Golden rule of the system:** *Record first in the ledger, settle second on the chain.*
If the blockchain step fails, the ledger shows exactly what happened — money is never lost or double-spent.
