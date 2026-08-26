import { Page, expect } from "@playwright/test";

export async function loginAsAdmin(page: Page) {
	await page.goto("/login");
	await page.waitForSelector('input[name="username"]', { timeout: 30000 });
	await page.fill('input[name="username"]', "superadmin");
	await page.fill('input[name="password"]', "Password0!");
	await page.getByRole("button", { name: /Sign In/i }).click();
	await expect(page).toHaveURL(/\/dashboard/, { timeout: 30000 });
}

export async function createResource(
	page: Page,
	name: string,
	description: string,
) {
	await page.goto("/resources");
	await page.getByRole("button", { name: /New Resource/i }).click();

	// Wait for dialog
	await expect(page.getByText(/Register Resource/i)).toBeVisible();

	await page.fill('input[name="name"]', name);
	await page.fill('textarea[name="description"]', description);

	// Click submit in dialog
	await page.getByRole("button", { name: "Create Resource" }).click();

	// Search for the resource to avoid pagination issues
	// The table search box is usually within the main content area
	await page.locator('main input[placeholder*="Search"]').fill(name);
	await page.keyboard.press("Enter");

	// Wait for the resource to appear in the table with longer timeout
	await expect(page.getByText(name)).toBeVisible({ timeout: 15000 });
}
