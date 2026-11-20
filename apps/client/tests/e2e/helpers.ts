import { Page, expect } from "@playwright/test";

export async function loginAsAdmin(page: Page) {
	await page.goto("/login");
	await page.waitForSelector('input[name="username"]', { timeout: 30000 });
	await page.fill('input[name="username"]', "superadmin");
	await page.fill('input[name="password"]', "Password0!");
	await page.getByRole("button", { name: /Sign In/i }).click();
	await expect(page).toHaveURL(/\/dashboard/, { timeout: 30000 });
	await expect(page.getByRole("heading", { name: /Dashboard/i })).toBeVisible({
		timeout: 30000,
	});
}

export async function createResource(
	page: Page,
	name: string,
	description: string,
) {
	await page.goto("/resources");
	await page.getByRole("button", { name: /New Access Right/i }).click();

	// Wait for dialog
	await expect(page.getByText(/Register Access Right/i)).toBeVisible();

	await page.fill('input[name="name"]', name);
	await page.fill('textarea[name="description"]', description);

	// Click submit in dialog
	await page.getByRole("button", { name: "Create Access Right" }).click();

	// Search for the resource to avoid pagination issues
	// The table search box is usually within the main content area
	await page.locator('main input[placeholder*="Search"]').fill(name);
	await page.keyboard.press("Enter");

	// Wait for the resource to appear in the table with longer timeout
	await expect(page.getByText(name)).toBeVisible({ timeout: 15000 });
}

export async function deleteResourceByName(page: Page, name: string) {
	await page.evaluate(async (targetName) => {
		const csrfCookie = document.cookie
			.split(";")
			.map((p) => p.trim())
			.find((p) => p.startsWith("csrf_token="));
		const csrfToken = csrfCookie
			? decodeURIComponent(csrfCookie.slice("csrf_token=".length))
			: "";
		const res = await fetch("/api/v1/access-rights", {
			credentials: "include",
		});
		if (!res.ok) return;
		const body = await res.json();
		const items = Array.isArray(body?.data) ? body.data : [];
		const match = items.find(
			(item: { name?: string }) => item.name === targetName,
		);
		if (!match?.id) return;
		await fetch(`/api/v1/access-rights/${match.id}`, {
			method: "DELETE",
			credentials: "include",
			headers: csrfToken ? { "X-CSRF-Token": csrfToken } : {},
		});
	}, name);
}

export async function logout(page: Page) {
	await page.evaluate(async () => {
		const csrfCookie = document.cookie
			.split(";")
			.map((p) => p.trim())
			.find((p) => p.startsWith("csrf_token="));
		const csrfToken = csrfCookie
			? decodeURIComponent(csrfCookie.slice("csrf_token=".length))
			: "";
		await fetch("/api/v1/auth/logout", {
			method: "POST",
			credentials: "include",
			headers: csrfToken ? { "X-CSRF-Token": csrfToken } : {},
		});
	});
}
