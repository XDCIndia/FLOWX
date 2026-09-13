import { test, expect } from "@playwright/test";

/**
 * Landing page smoke tests — fully hermetic (no API, no chain).
 * These run in CI on every push.
 */
test.describe("landing page", () => {
  test("renders hero pitch and testnet badge", async ({ page }) => {
    await page.goto("/");
    await expect(
      page.getByRole("heading", {
        level: 1,
        name: /Move money across borders/,
      })
    ).toBeVisible();
    await expect(page.getByText(/Apothem testnet/).first()).toBeVisible();
  });

  test("shows on-chain proof with xdcscan links", async ({ page }) => {
    await page.goto("/");
    // Both deployed contracts are linked on testnet.xdcscan.com (xdc-prefixed)
    const tUSDC = page.locator(
      'a[href*="testnet.xdcscan.com/address/xdce069a90d55fdeca2cbec7793712aa3d807ffef9b"]'
    );
    const pool = page.locator(
      'a[href*="testnet.xdcscan.com/address/xdc4a3a728562847a1fd9afdbd81356533eb05b12cd"]'
    );
    await expect(tUSDC.first()).toBeVisible();
    await expect(pool.first()).toBeVisible();

    // The real swap transaction is presented as evidence
    const swapTx = page.locator(
      'a[href*="testnet.xdcscan.com/tx/7f23e14a6ae79db25aee72d3c9329bb17c8a4581ca9860dfc9edc1e80aa2c771"]'
    );
    await expect(swapTx.first()).toBeVisible();
  });

  test("links to the GitHub repository", async ({ page }) => {
    await page.goto("/");
    const repo = page.locator('a[href="https://github.com/XDCIndia/FLOWX"]');
    await expect(repo.first()).toBeVisible();
  });

  test("has an honest simulation disclosure", async ({ page }) => {
    await page.goto("/");
    await expect(
      page.getByRole("heading", { name: /what is simulated/i })
    ).toBeVisible();
    await expect(page.getByText(/Stripe/i).first()).toBeVisible();
  });
});
