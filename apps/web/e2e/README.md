# FlowX web e2e tests (Playwright)

## Hermetic run (CI default)

Only needs the web app — no API, database, or chain:

```bash
cd apps/web
npm install
npx playwright install chromium   # once
npm run test:e2e                  # landing + static specs
```

## Full-stack run (local)

Needs postgres + redis + Go API + web. From the repo root:

```bash
docker compose up -d postgres redis
# in one terminal: boot the API (loads .env)
go run ./cmd/api
# in another:
cd apps/web && E2E_FULL_STACK=1 npm run test:e2e
```

Specs that need the API self-skip unless `E2E_FULL_STACK=1` is set.
Steps touching testnet RPC are intentionally left to manual demos.
