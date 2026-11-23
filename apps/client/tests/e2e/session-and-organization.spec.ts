import { randomUUID } from "node:crypto";
import { expect, type Page, test } from "@playwright/test";
import { loginAsAdmin, logout } from "./helpers";

async function browserRequest(page: Page, path: string, method = "GET") {
	return page.evaluate(
		async ({ path, method }) => {
			const prefix = "csrf_token=";
			const token = document.cookie
				.split("; ")
				.find((c) => c.startsWith(prefix));
			const response = await fetch(`/api/v1${path}`, {
				method,
				headers:
					method !== "GET" && token
						? { "X-CSRF-Token": token.slice(prefix.length) }
						: {},
			});
			return { status: response.status, body: await response.json() };
		},
		{ path, method },
	);
}

test("refresh rotates CSRF cookie and logout ends the session", async ({
	page,
}) => {
	await loginAsAdmin(page);
	try {
		const csrf = async () =>
			(await page.context().cookies()).find((c) => c.name === "csrf_token")
				?.value;
		const original = await csrf();
		expect(original).toBeTruthy();
		expect((await browserRequest(page, "/auth/refresh", "POST")).status).toBe(
			200,
		);
		const rotated = await csrf();
		expect(rotated).toBeTruthy();
		expect(rotated).not.toBe(original);
		expect((await browserRequest(page, "/auth/me")).status).toBe(200);
	} finally {
		await logout(page);
	}
	const cookies = (await page.context().cookies()).map((c) => c.name);
	for (const name of ["access_token", "refresh_token", "csrf_token"])
		expect(cookies).not.toContain(name);
	expect((await browserRequest(page, "/auth/me")).status).toBe(401);
});

test("browser proxy serves health and issues realtime ticket", async ({
	page,
}) => {
	await loginAsAdmin(page);
	try {
		expect((await browserRequest(page, "/health")).status).toBe(200);
		const result = await browserRequest(page, "/auth/ticket", "POST");
		expect(result.status).toBe(200);
		expect(result.body.data.ticket).toMatch(/\S+/);
	} finally {
		await logout(page);
	}
});

test("create, find, rename and delete an organization", async ({ page }) => {
	const slug = `e2e-${randomUUID()}`;
	const renamed = `Renamed ${slug}`;
	const findOwned = async () => {
		const { status, body } = await browserRequest(page, "/organizations/me");
		expect(status).toBe(200);
		return body.data.organizations.find(
			(org: { slug: string }) => org.slug === slug,
		);
	};
	await loginAsAdmin(page);
	try {
		await page.goto("/organizations");
		await page.getByRole("button", { name: "New Organization" }).click();
		const dialog = page.getByRole("dialog");
		await dialog.locator('[name="name"]').fill(slug);
		await dialog.locator('[name="slug"]').fill(slug);
		await dialog.getByRole("button", { name: "Create", exact: true }).click();
		await page
			.getByRole("main")
			.getByPlaceholder("Search…", { exact: true })
			.fill(slug);
		const row = page.getByRole("row").filter({ hasText: slug });
		await expect(row).toBeVisible();
		await row.getByRole("button", { name: "Row actions" }).click();
		await page.getByRole("menuitem", { name: "Edit" }).click();
		await dialog.locator('[name="name"]').fill(renamed);
		await dialog.getByRole("button", { name: "Save Changes" }).click();
		await expect(row.getByText(renamed, { exact: true })).toBeVisible();
		await row.getByRole("button", { name: "Row actions" }).click();
		await page.getByRole("menuitem", { name: "Delete" }).click();
		await page
			.getByRole("alertdialog")
			.getByRole("button", { name: "Delete", exact: true })
			.click();
		await expect(row).toHaveCount(0);
		expect(await findOwned()).toBeUndefined();
	} finally {
		try {
			const leftover = await findOwned();
			if (leftover) {
				expect(
					(
						await browserRequest(
							page,
							`/organizations/${leftover.id}`,
							"DELETE",
						)
					).status,
				).toBe(200);
				expect(await findOwned()).toBeUndefined();
			}
		} finally {
			await logout(page);
		}
	}
});
