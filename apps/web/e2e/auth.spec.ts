import { test, expect } from "@playwright/test";
import { apiBaseURL } from "../playwright.config";

/**
 * Auth + dashboard flows. These need a live Go API (and its postgres/redis).
 *
 * They self-skip unless E2E_FULL_STACK=1 is set — CI runs the hermetic
 * landing specs only; use `npm run test:e2e:full` locally for these.
 */
const fullStack = process.env.E2E_FULL_STACK === "1";

test.describe("auth and dashboard", () => {
  test.skip(!fullStack, "Set E2E_FULL_STACK=1 with a live API to run");

  const suffix = Date.now().toString(36);
  const email = `e2e-${suffix}@example.com`;
  const password = "E2e-password-123!";
  let apiKey = "";

  test.beforeAll(async ({ request }) => {
    const res = await request.post(`${apiBaseURL}/v1/auth/register`, {
      data: { email, password, organization: `e2e-org-${suffix}` },
    });
    if (!res.ok()) {
      throw new Error(`register failed: ${res.status()} ${await res.text()}`);
    }
    const body = await res.json();
    apiKey = body.api_key ?? body.apiKey ?? "";
    if (!apiKey) throw new Error("register response had no api_key");
  });

  test("login page loads", async ({ page }) => {
    await page.goto("/login");
    await expect(
      page.getByRole("button", { name: /sign in|log in|login/i })
    ).toBeVisible();
  });

  test("dashboard loads with an API key", async ({ page }) => {
    await page.goto("/login");
    // The app stores the key in localStorage as flowx_api_key.
    await page.evaluate((k) => localStorage.setItem("flowx_api_key", k), apiKey);
    await page.goto("/wallets");
    await expect(page).toHaveURL(/wallets/);
    // Either wallets render or the empty state appears — both prove the API
    // round-tripped.
    await expect(
      page.getByText(/wallets|no wallets/i).first()
    ).toBeVisible();
  });

  test("webhook verify tool page loads", async ({ page }) => {
    await page.goto("/webhooks/verify");
    await expect(page.getByText(/verify/i).first()).toBeVisible();
  });
});
