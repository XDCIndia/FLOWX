import { defineConfig, devices } from "@playwright/test";

/**
 * E2E configuration for the FlowX web dashboard.
 *
 * Two bases are in play:
 *  - Web app (the system under test): default http://localhost:3001
 *  - Go API (state setup / health):   default http://localhost:3000
 *
 * Env overrides:
 *  - PLAYWRIGHT_BASE_URL  -> web base (falls back to NEXT_PUBLIC_WEB_URL, then :3001)
 *  - PLAYWRIGHT_API_URL   -> API base (falls back to NEXT_PUBLIC_API_URL, then :3000)
 *  - E2E_FULL_STACK=1     -> run specs that need a live API; without it, API-backed
 *                            specs self-skip so the suite stays hermetic.
 *
 * CI (GitHub Actions) boots postgres+redis, the Go API, and the built web app
 * before invoking `npx playwright test` — see .github/workflows/ci.yml.
 * For the full-stack local run, see e2e/README.md (`npm run test:e2e:full`).
 */

const webBaseURL =
  process.env.PLAYWRIGHT_BASE_URL ||
  process.env.NEXT_PUBLIC_WEB_URL ||
  "http://localhost:3001";

export const apiBaseURL =
  process.env.PLAYWRIGHT_API_URL ||
  process.env.NEXT_PUBLIC_API_URL ||
  "http://localhost:3000";

export default defineConfig({
  testDir: "./e2e",
  // Deterministic ordering; the suite manages its own tenant state.
  workers: 1,
  retries: process.env.CI ? 1 : 0,
  timeout: 30_000,
  expect: { timeout: 10_000 },
  reporter: process.env.CI ? "github" : "list",
  use: {
    baseURL: webBaseURL,
    trace: "on-first-retry",
    screenshot: "only-on-failure",
  },
  projects: [{ name: "chromium", use: { ...devices["Desktop Chrome"] } }],
  // In CI the workflow starts the production server itself after `next build`.
  // Locally, spin up `next dev` on :3001 (reusing an already-running server).
  webServer: process.env.PLAYWRIGHT_BASE_URL
    ? undefined
    : {
        command: process.env.CI
          ? "npm run start -- --port 3001"
          : "npm run dev -- --port 3001",
        url: webBaseURL,
        reuseExistingServer: !process.env.CI,
        timeout: 180_000,
        env: { NEXT_PUBLIC_API_URL: apiBaseURL },
      },
});
