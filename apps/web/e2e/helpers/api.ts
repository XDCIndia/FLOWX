import { request, test, type APIRequestContext } from "@playwright/test";

/**
 * Test-state bootstrap against the Go API.
 *
 * Registering a tenant through direct API calls (instead of clicking the
 * signup form) keeps specs fast and immune to UI refactors. The UI login
 * flow itself is still exercised in auth.spec.ts.
 */

export interface TestTenant {
  email: string;
  password: string;
  apiKey: string;
  accessToken: string;
}

export async function registerTenant(
  api: APIRequestContext,
  suffix = Date.now().toString(36)
): Promise<TestTenant> {
  const email = `e2e-${suffix}-${Math.random().toString(36).slice(2, 8)}@example.com`;
  const password = "e2e-password-123";
  const name = `E2E Tenant ${suffix}`;

  const reg = await api.post("/v1/auth/register", {
    data: { name, email, password },
  });
  if (!reg.ok()) {
    throw new Error(`register failed: ${reg.status()} ${await reg.text()}`);
  }

  const login = await api.post("/v1/auth/login", { data: { email, password } });
  if (!login.ok()) {
    throw new Error(`login failed: ${login.status()} ${await login.text()}`);
  }
  const { access_token } = await login.json();

  const key = await api.post("/v1/keys/", {
    data: { label: "e2e" },
    headers: { Authorization: `Bearer ${access_token}` },
  });
  if (!key.ok()) {
    throw new Error(`create api key failed: ${key.status()} ${await key.text()}`);
  }
  const { key: apiKey } = await key.json();

  return { email, password, apiKey, accessToken: access_token };
}

/** True when a live API is reachable at the configured base. */
export async function apiIsUp(api: APIRequestContext): Promise<boolean> {
  try {
    const res = await api.get("/health", { timeout: 3_000 });
    return res.ok();
  } catch {
    return false;
  }
}

/**
 * Skip the enclosing spec file when no live API is available, so a plain
 * `npx playwright test` (web only) stays green and hermetic.
 */
export async function requireApi(api: APIRequestContext): Promise<boolean> {
  if (await apiIsUp(api)) return true;
  test.skip(true, "live API not reachable — run `npm run test:e2e:full` for full-stack");
  return false;
}

export async function createWallet(
  api: APIRequestContext,
  apiKey: string
): Promise<{ id: string; public_key: string }> {
  // Hermetic: with XDC_TREASURY_SECRET_KEY unset the API skips on-chain
  // funding — keypair generation is local secp256k1.
  const res = await api.post("/v1/wallets/", {
    data: {},
    headers: { Authorization: `Bearer ${apiKey}` },
  });
  if (!res.ok()) {
    throw new Error(`create wallet failed: ${res.status()} ${await res.text()}`);
  }
  return res.json();
}

/** Seed the browser with a logged-in session before the app hydrates. */
export function seedSession(page: import("@playwright/test").Page, apiKey: string, walletIds: string[] = []) {
  return page.addInitScript(
    ({ key, ids }: { key: string; ids: string[] }) => {
      localStorage.setItem("flowx_api_key", key);
      localStorage.setItem("flowx_wallet_ids", JSON.stringify(ids));
      document.cookie = `flowx_api_key=${key}; path=/; max-age=31536000; SameSite=Lax`;
    },
    { key: apiKey, ids: walletIds }
  );
}

export async function newApiContext(apiBaseURL: string): Promise<APIRequestContext> {
  return request.newContext({ baseURL: apiBaseURL });
}
