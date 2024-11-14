import { expect, test } from "@playwright/test";
const baseUrl = "http://localhost:1239";

test("has title", async ({ page }) => {
  await page.goto(`${baseUrl}`);

  // Expect a title "to contain" a substring.
  await expect(page).toHaveTitle(/Crystal Observability/);
});

test("get started link", async ({ page }) => {
  await page.goto(`${baseUrl}`);

  await page
    .getByRole("navigation")
    .locator("section")
    .getByRole("link")
    .click();

  // Expects page to have a heading with the name of Installation.
  await expect(page.getByRole("heading", { name: "All Charts" })).toBeVisible();
});
