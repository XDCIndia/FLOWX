# FlowX

**Cross-border payment infrastructure for emerging markets.**

FlowX is a programmable payments API built on the [XDC Network](https://xdc.org). It gives fintech products and developers the primitives to move value across borders — wallet management, internal transfers, FX conversion, batch payments, scheduled payouts, and fiat on/off-ramps — behind a clean REST API.

> **Status**: Active development — testnet only.

---

## What it does

- ✅ **Wallets** — create XDC wallets with encrypted private keys
- ✅ **Transfers** — instant on-chain XDC transfers with tx hash
- ✅ **FX / Conversion** — real-time rates from CoinGecko, execute USDC ↔ TXDC swaps
- ✅ **Batch Payments** — send up to 100 transfers in one API call
- ✅ **Scheduled Payouts** — recurring daily/weekly/monthly transfers
- ✅ **Fiat Rails** — deposit/withdraw via Flutterwave/Stripe (mock mode)
- ✅ **Multi-currency** — NGN, USD, EUR, GBP, INR, KES, GHS, ZAR
- ✅ **Webhooks** — signed delivery of payment events to developer endpoints
- ✅ **Route Engine** — compare bank (Stripe), payment network (Ripple ODL), and blockchain (XDC) routes
- ✅ **Compliance** — velocity checks, spending limits, guardian wallets
- ✅ **Fee Management** — configurable transfer and conversion fees
- ✅ **Test Faucets** — get test USDC and real on-chain TXDC from treasury

---

## Tech Stack

| Component | Technology |
|-----------|------------|
| **Backend** | Go 1.22 + Chi router |
| **Frontend** | Next.js 14 + TypeScript + Tailwind CSS |
| **Database** | PostgreSQL 15 |
| **Cache** | Redis 7 |
| **Blockchain** | XDC Network (Apothem Testnet) |
| **FX Rates** | CoinGecko API |
| **Payment Rails** | Stripe, Flutterwave, Ripple ODL |

---

## Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                    FlowX REST API (Go)                           │
│  Chi router │ JWT + API key auth │ Rate limiting │ Multi-tenant │
└─────────────────────────────────────────────────────────────────┘
        │
        ├── Wallet Service       ──► postgres: wallets, balances
        ├── Transfer Service     ──► postgres: transactions
        ├── FX Service           ──► CoinGecko + rate cache (Redis)
        ├── Batch Service        ──► postgres: batches, batch_items
        ├── Schedule Service     ──► postgres: schedules
        ├── Fiat Service         ──► Flutterwave/Stripe adapters
        ├── Routing Service      ──► bank/payment/blockchain routes
        ├── Fee Service          ──► postgres: fees, fee_collections
        └── Webhook Dispatcher   ──► postgres: webhook_endpoints
                │
                ▼
        XDC Network (Apothem Testnet)
        RPC: https://rpc.apothem.network
        Explorer: https://testnet.xdcscan.com
```

---

## Getting Started

### Prerequisites

- Go 1.22+
- Node.js 18+
- PostgreSQL 15+
- Redis 7+
- Docker & Docker Compose

### Quick Start with Docker

```bash
git clone https://github.com/XDCIndia/FLOWX
cd FLOWX

# Start all services
docker-compose up -d

# Run migrations
docker exec -i fluxa-postgres-1 psql -U fluxa -d fluxa < db/migrations/*.sql

# Open the app
open http://localhost:3001
```

### Manual Setup

```bash
# 1. Clone and install
git clone https://github.com/XDCIndia/FLOWX
cd FLOWX
go mod tidy

# 2. Configure environment
cp .env.example .env
# Edit .env with your settings

# 3. Start PostgreSQL and Redis
docker-compose up -d postgres redis

# 4. Run migrations
make migrate

# 5. Start the API
make run-api

# 6. Start the worker
make run-worker

# 7. Start the frontend
cd apps/web
npm install
npm run dev
```

---

## Environment Variables

| Variable | Description | Required |
|----------|-------------|----------|
| `PORT` | API listen port (default: 3000) | No |
| `DATABASE_URL` | PostgreSQL connection string | Yes |
| `REDIS_URL` | Redis connection string | Yes |
| `XDC_RPC_URL` | XDC RPC endpoint | Yes |
| `XDC_TREASURY_SECRET_KEY` | Treasury wallet private key | Yes |
| `MASTER_ENCRYPTION_KEY` | 32-byte hex key for encryption | Yes |
| `JWT_SECRET` | JWT signing secret | Yes |
| `STRIPE_SECRET_KEY` | Stripe API key (for bank payments) | No |
| `FLUTTERWAVE_SECRET_KEY` | Flutterwave API key (for fiat) | No |

---

## API Reference

All endpoints are prefixed `/v1`. Auth: `Authorization: Bearer <api_key>`.

### Authentication

```http
POST /v1/auth/register     Create account
POST /v1/auth/login        Login — returns JWT
POST /v1/keys              Create API key → sk_live_...
```

### Wallets

```http
GET  /v1/wallets                   List wallets
POST /v1/wallets                   Create wallet
GET  /v1/wallets/:id               Get wallet details
GET  /v1/wallets/:id/balances      Get balances
POST /v1/wallets/:id/faucet        Get test tokens (USDC/TXDC)
DELETE /v1/wallets/:id             Delete wallet
```

### Transfers

```http
POST /v1/transfers                 Create transfer
GET  /v1/transfers/:id             Get transfer status
GET  /v1/transactions              List transactions
```

### FX / Conversion

```http
GET  /v1/fx/rates?from=USDC&to=TXDC    Get exchange rate
POST /v1/fx/quote                       Get conversion quote
POST /v1/fx/convert                     Execute conversion
```

### Batch Payments

```http
POST /v1/transfers/batch           Create batch (up to 100 transfers)
GET  /v1/transfers/batch/:id       Get batch status
GET  /v1/transfers/batch/:id/export  Export as CSV
```

### Scheduled Payouts

```http
POST /v1/schedules                 Create schedule
GET  /v1/schedules                 List schedules
PATCH /v1/schedules/:id            Pause/resume schedule
DELETE /v1/schedules/:id           Cancel schedule
```

### Fiat Rails

```http
POST /v1/wallets/:id/deposit/fiat       Initiate fiat deposit
POST /v1/wallets/:id/deposit/fiat/simulate  Simulate deposit (demo)
POST /v1/wallets/:id/withdraw/fiat      Initiate fiat withdrawal
```

### Webhooks

```http
POST /v1/webhooks                  Register webhook endpoint
GET  /v1/webhooks                  List endpoints
DELETE /v1/webhooks/:id            Delete endpoint
GET  /v1/webhooks/:id/deliveries   List delivery logs
POST /v1/webhooks/verify           Verify webhook signature
```

### Payments (Multi-route)

```http
POST /v1/payments/compare          Compare routes (bank/XDC/blockchain)
POST /v1/payments/execute          Execute payment via selected route
```

### Other

```http
GET  /v1/fees                      Fee schedule
GET  /v1/usage                     Usage statistics
GET  /health                       Health check
```

---

## Demo Flow

1. **Create Account** → Get API key at `/login`
2. **Create Wallet** → Get XDC wallet address
3. **Get Test Tokens** → Click "Get TXDC" or "Get USDC" buttons
4. **Convert Assets** → Swap USDC ↔ TXDC at live rates
5. **Send Payment** → Transfer TXDC to another wallet
6. **Batch Payment** → Send to multiple recipients at once
7. **Schedule Payout** → Set up recurring transfers
8. **Compare Routes** → See bank vs blockchain vs payment network options

---

## Frontend Pages

| Page | URL | Description |
|------|-----|-------------|
| Landing | `/` | Marketing page |
| Login | `/login` | Sign in / Create account |
| Dashboard | `/overview` | System health overview |
| Wallets | `/wallets` | Create/manage wallets, faucet |
| Transfers | `/transfers` | Send XDC, view history |
| FX | `/fx` | Live rates, convert assets |
| Conversions | `/conversions` | Conversion history |
| Batch | `/batch` | Batch transfers |
| Schedules | `/schedules` | Recurring payouts |
| Payments | `/payments` | Multi-route payment engine |
| Fiat | `/fiat` | Deposit/withdraw fiat |
| Webhooks | `/webhooks` | Manage webhook endpoints |
| Usage | `/usage` | API usage & billing |

---

## XDC Testnet

- **Network**: Apothem Testnet
- **Chain ID**: 51
- **RPC**: https://rpc.apothem.network
- **Explorer**: https://testnet.xdcscan.com
- **Faucet**: https://faucet.apothem.network

### Getting Test TXDC

1. Create a wallet on the Wallets page
2. Click "Get TXDC" button
3. Treasury wallet sends real TXDC to your wallet
4. View on [testnet.xdcscan.com](https://testnet.xdcscan.com)

---

## License

MIT License - see [LICENSE](LICENSE) for details.
