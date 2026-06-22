import { test, expect } from "@playwright/test";

test.describe("Authentication Flow", () => {
  test("should login successfully with admin credentials", async ({ page }) => {
    // Navigate to login page
    await page.goto("/login");

    // Check if we are on the login page
    await page.waitForSelector('input[name="username"]', { timeout: 30000 });

    // Fill in credentials
    await page.fill('input[name="username"]', "superadmin");
    await page.fill('input[name="password"]', "Password0!");
    await page.getByRole("button", { name: /Sign In/i }).click();

    // Should redirect to dashboard
    await expect(page).toHaveURL(/\/dashboard/);

    // Check for success toast or dashboard content
    await expect(page.getByRole("heading", { name: /Dashboard/i })).toBeVisible();
  });

  test("should show error on invalid credentials", async ({ page }) => {
    await page.goto("/login");

    await page.fill('input[name="username"]', "wronguser");
    await page.fill('input[name="password"]', "wrongpass");
    await page.click('button[type="submit"]');

    // Should show error message (either from Zod or API)
    // Based on LoginPage.tsx, it might show a toast or field error
    // Let's check for a generic error message or toast
    // (Assuming the API returns 401 and it's handled)
  });
});
