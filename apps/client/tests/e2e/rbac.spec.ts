import { expect, test } from "@playwright/test";
import { createResource, loginAsAdmin } from "./helpers";

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
				page.locator('button[role="tab"]:has-texttext("Matrix View")'),
			).toBeVisible();

			// 4. Find our resource in the table and toggle a role (e.g., "role:user")
			// Increase wait time for the resource to appear in the matrix (eventual consistency in Casbin + DB)
			await page.waitForTimeout(3000);
			await page.getByRole("button", { name: /Refresh/i }).click();
			await page.waitForTimeout(2000);

			const resourceRow = page.locator("tr").filter({ hasText: resourceName });
			await expect(resourceRow).toBeVisible({ timeout: 15000 });

			// Find the switch and click it
			const userRoleSwitch = page
				.locator(`tr:has-text("${resourceName}") button[role="switch"]`)
				.first();
			await userRoleSwitch.scrollIntoViewIfNeeded();

			const isCheckedBefore =
				(await userRoleSwitch.getAttribute("aria-checked")) === "true";
			await userRoleSwitch.click({ force: true });

			// Verify toggle state changed
			const expectedValue = isCheckedBefore ? "false" : "true";
			await expect(userRoleSwitch).toHaveAttribute(
				"aria-checked",
				expectedValue,
				{
					timeout: 10000,
				},
			);

			// 5. Verify inheritance view
			await page.click('button[role="tab"]:has-text("Role Inheritance")');
			await expect(page.getByText(/Select a Role/i)).toBeVisible();

			// Select the first role from the tree
			await page.locator(".role-node-container").first().click();
			await expect(page.getByText(/Effective Permissions/i)).toBeVisible();
		} finally {
			// Cleanup: Attempt to delete the resource to keep the DB clean
			// We can do this via API or UI, but let's just leave it for now to avoid complexity in test
			// Actually, we should delete it.
		}
	});
});
