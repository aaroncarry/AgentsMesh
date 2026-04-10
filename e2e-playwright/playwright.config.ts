import { defineConfig, devices } from "@playwright/test";
import { getWebBaseUrl } from "./helpers/env";

const isCI = !!process.env.CI;

export default defineConfig({
  testDir: "./tests",
  fullyParallel: false,
  forbidOnly: isCI,
  retries: isCI ? 1 : 0,
  workers: 1,
  timeout: isCI ? 90_000 : 60_000,

  expect: {
    timeout: 10_000,
  },

  reporter: isCI
    ? [["list"], ["junit", { outputFile: "report.xml" }]]
    : [["html", { open: "never" }]],

  use: {
    baseURL: getWebBaseUrl(),
    trace: "on-first-retry",
    screenshot: "only-on-failure",
    actionTimeout: 10_000,
    navigationTimeout: 30_000,
  },

  projects: [
    {
      name: "setup",
      testMatch: /global\.setup\.ts/,
    },
    {
      name: "admin-setup",
      testMatch: /admin\.setup\.ts/,
    },
    {
      name: "chromium",
      use: {
        ...devices["Desktop Chrome"],
        storageState: ".auth/user.json",
      },
      dependencies: ["setup"],
      testIgnore: /.*admin.*\.spec\.ts/,
    },
    {
      name: "admin",
      use: {
        ...devices["Desktop Chrome"],
        storageState: ".auth/admin.json",
      },
      dependencies: ["admin-setup"],
      testMatch: /.*admin.*\.spec\.ts/,
    },
  ],
});
