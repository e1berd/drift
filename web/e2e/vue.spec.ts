import { test, expect } from '@playwright/test'

test('unauthenticated visitors are sent to the auth pages', async ({ page }) => {
  await page.goto('/')
  await expect(page).toHaveURL(/\/(setup|login)/)
})
