import { expect, test } from "@playwright/test";
import {
	createResource,
	deleteResourceByName,
	loginAsAdmin,
	logout,
} from "./helpers";

test.describe("RBAC Flow - End to End", () => {
	test.beforeEach(async ({ page }) => {
		await loginAsAdmin(page);
	});

	test("should create a resource and manage permissions", async ({ page }) => {
		const resourceName = `TestResource_${Date.now()}`;

		try {
			// 1. Create Resource
			await createResource(page, resourceName, "A temporary test resource");

			// 2. Go to Permissions
			await page.goto("/permissions");

			// 3. Ensure we are on the Matrix tab (default)
			await expect(
				page.locator('button[role="tab"]:has-text("Matrix View")'),
			).toBeVisible();

			// 4. Find our resource in the table and toggle a role (e.g., "role:user")
			// Increase wait time for the resource to appear in the matrix (eventual consistency in Casbin + DB)
			await page.waitForTimeout(3000);
			await page.getByRole("button", { name: /Refresh/i }).click();
			await page.waitForTimeout(2000);

			const resourceRow = page.locator("tr").filter({ hasText: resourceName });
			await expect(resourceRow).toBeVisible({ timeout: 15000 });

			// Find an unassigned switch on an access right that has endpoints.
			// A freshly created access right has no endpoints, so it cannot be assigned.
			const candidateSwitch = page
				.locator("tbody tr")
				.filter({ hasNot: page.getByText("0 Endpoints", { exact: true }) })
				.locator(
					'[role="switch"][aria-checked="false"]:not([aria-disabled="true"])',
				)
				.first();
			await expect(candidateSwitch).toBeVisible({ timeout: 15000 });
			const switchId = await candidateSwitch.getAttribute("id");
			const userRoleSwitch = page.locator(`[id="${switchId}"]`);
			await userRoleSwitch.scrollIntoViewIfNeeded();

			await userRoleSwitch.click({ force: true });
			await expect(userRoleSwitch).toHaveAttribute("aria-checked", "true", {
				timeout: 10000,
			});

			// Restore original state to keep the DB clean
			await userRoleSwitch.click({ force: true });
			await expect(userRoleSwitch).toHaveAttribute("aria-checked", "false", {
				timeout: 10000,
			});

			// 5. Verify inheritance view
			await page.click('button[role="tab"]:has-text("Role Inheritance")');
			await expect(
				page.getByRole("heading", { name: /Select a Role/i }),
			).toBeVisible();

			// Select the first role from the tree
			await page.getByRole("button", { name: /perms/i }).first().click();
			await expect(page.getByText(/Effective Permissions/i)).toBeVisible();
		} finally {
			await deleteResourceByName(page, resourceName).catch(() => {});
			await logout(page).catch(() => {});
		}
	});
});
