#!/usr/bin/env python3
"""FlowX demo tour — recorded with Playwright."""
import asyncio, sys
from playwright.async_api import async_playwright

BASE = "http://localhost:3001"
OUT = "/tmp/fluxa-webm"
VERIFY = "--verify" in sys.argv

W2_ADDR = "xdcc15160fbcee710749ab212c0c7f67290aef96e4a"

with open("/tmp/fluxa-demo-key") as f:
    API_KEY = f.read().strip()

seen = []

async def verify(page, name):
    if VERIFY:
        await page.screenshot(path=f"{OUT}/{len(seen):02d}-{name}.png")

async def click(page, name, selector, timeout=8000):
    try:
        await page.wait_for_selector(selector, timeout=timeout)
        await page.click(selector)
        await page.mouse.move(640, 400)
        await page.wait_for_timeout(1000)
    except Exception as exc:
        seen.append(f"MISS {name}: {exc.__class__.__name__}")
        return
    seen.append(f"OK   {name}")

async def fill(page, name, selector, value):
    try:
        el = page.locator(selector).first
        await el.wait_for(timeout=8000)
        await el.click()
        await page.wait_for_timeout(250)
        await el.fill(value)
        await page.wait_for_timeout(350)
    except Exception as exc:
        seen.append(f"MISS {name}: {exc.__class__.__name__}")
        return
    seen.append(f"OK   {name}")

async def pause(page, ms):
    await page.wait_for_timeout(ms)

async def settle(page, ms=1200):
    await page.mouse.move(640, 400)
    await page.wait_for_timeout(ms)

async def main():
    async with async_playwright() as p:
        browser = await p.chromium.launch(
            executable_path="/home/dhiraj/.cache/ms-playwright/chromium-1243/chrome-linux64/chrome",
            args=["--no-sandbox", "--disable-dev-shm-usage"],
        )
        ctx = await browser.new_context(
            viewport={"width": 1280, "height": 800},
            device_scale_factor=1,
            record_video_dir=OUT,
            record_video_size={"width": 1280, "height": 800},
        )
        page = await ctx.new_page()
        page.set_default_timeout(25000)

        # ---- 1. Landing page ----
        await page.goto(BASE + "/", wait_until="networkidle")
        await settle(page, 2500)
        await verify(page, "landing-hero")
        await pause(page, 4500)
        await page.evaluate("document.getElementById('architecture')?.scrollIntoView({behavior:'smooth'})")
        await settle(page, 2000); await pause(page, 3500)
        await page.evaluate("document.getElementById('proof')?.scrollIntoView({behavior:'smooth'})")
        await settle(page, 2000); await pause(page, 1500)
        try:
            link = page.locator("a[href*='xdcscan']").first
            await link.hover(); await pause(page, 2500)
        except Exception:
            pass
        await verify(page, "landing-proof")
        await page.evaluate("window.scrollBy({top: 750, behavior:'smooth'})")
        await settle(page, 1800); await pause(page, 2500)
        await page.evaluate("window.scrollBy({top: 750, behavior:'smooth'})")
        await settle(page, 1800); await pause(page, 4000)
        await page.evaluate("window.scrollTo({top: 0, behavior:'smooth'})")
        await settle(page, 2000)

        # ---- 2. Login with API key ----
        await click(page, "cta-launch", "text=Launch App")
        await settle(page, 2500)
        await verify(page, "login")
        await fill(page, "apikey", "input[placeholder='sk_live_...']", API_KEY)
        await pause(page, 900)
        await click(page, "login-submit", "form:has(input[placeholder='sk_live_...']) button[type='submit']")
        await pause(page, 6000)
        await verify(page, "after-login")

        # ---- 3. Overview ----
        await page.goto(BASE + "/overview", wait_until="networkidle")
        await settle(page, 2500); await pause(page, 5500)
        await verify(page, "overview")

        # ---- 4. Wallets ----
        await page.goto(BASE + "/wallets", wait_until="networkidle")
        await settle(page, 2500); await pause(page, 6500)
        await verify(page, "wallets")

        # ---- 5. Payments: compare + execute real swap ----
        await page.goto(BASE + "/payments", wait_until="networkidle")
        await settle(page, 2500)
        try:
            await page.locator("select").nth(0).select_option("TXDC"); await pause(page, 700)
            await page.locator("select").nth(1).select_option("USDC"); await pause(page, 700)
            seen.append("OK   selects")
        except Exception as exc:
            seen.append(f"MISS selects: {exc.__class__.__name__}")
        await fill(page, "amount", "input[placeholder='100000']", "0.2")
        await click(page, "compare", "button:has-text('Compare Routes')", timeout=10000)
        try:
            await page.wait_for_selector("text=Execute This Route", timeout=20000)
            seen.append("OK   routes-rendered")
        except Exception:
            seen.append("MISS routes-rendered")
        await pause(page, 4500)
        await verify(page, "routes")
        await pause(page, 2500)
        try:
            card = page.locator("div.border-green-500")
            await card.locator("button:has-text('Execute This Route')").click()
            seen.append("OK   execute-amm")
        except Exception as exc:
            seen.append(f"MISS execute-amm: {exc.__class__.__name__}")
        try:
            await page.wait_for_selector("text=Settlement in progress", timeout=20000)
            seen.append("OK   result-panel")
        except Exception:
            try:
                await page.wait_for_selector("code:has-text('0x')", timeout=8000)
                seen.append("OK   result-panel")
            except Exception:
                seen.append("MISS result-panel")
        try:
            await page.evaluate("window.scrollTo({top: document.body.scrollHeight, behavior:'smooth'})")
            await settle(page, 1500)
            await pause(page, 10000)
        except Exception:
            await pause(page, 8000)
        await pause(page, 1500)

        # ---- 6. Transfers history ----
        await page.goto(BASE + "/transfers", wait_until="networkidle")
        await settle(page, 2500); await pause(page, 6500)
        await verify(page, "transfers")

        # ---- 7. FX + Conversions ----
        await page.goto(BASE + "/fx", wait_until="networkidle")
        await settle(page, 2500); await pause(page, 5000)
        await page.goto(BASE + "/conversions", wait_until="networkidle")
        await settle(page, 2500); await pause(page, 4000)

        # ---- 8. Webhooks verify tool ----
        await page.goto(BASE + "/webhooks", wait_until="networkidle")
        await settle(page, 2500); await pause(page, 3000)
        await fill(page, "wh-secret", "input[placeholder='whsec_...']", "whsec_demo_secret")
        try:
            ta = page.locator("textarea")
            await ta.nth(0).click()
            await ta.nth(0).fill("X-Fluxa-Signature: sha256=abcdef1234567890\nX-Fluxa-Timestamp: 1700000000")
            await pause(page, 400)
            await ta.nth(1).click()
            await ta.nth(1).fill('{"event":"payment.completed","amount":"100"}')
            await pause(page, 500)
            seen.append("OK   wh-fields")
        except Exception as exc:
            seen.append(f"MISS wh-fields: {exc.__class__.__name__}")
        await click(page, "wh-verify", "button:has-text('Verify')")
        await pause(page, 3000)
        await verify(page, "webhooks")

        # ---- 9. End on overview ----
        await page.goto(BASE + "/overview", wait_until="networkidle")
        await settle(page, 2500); await pause(page, 3000)

        await ctx.close()
        await browser.close()

    with open(f"{OUT}/seen.txt", "w") as f:
        f.write("\n".join(seen) + "\n")
    misses = [l for l in seen if l.startswith("MISS")]
    print(f"steps: {len(seen)}, misses: {len(misses)}")
    for l in seen:
        print(l)

asyncio.run(main())
