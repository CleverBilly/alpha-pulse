import { defineConfig, devices } from "@playwright/test";
import { DEFAULT_AUTH_COOKIE_NAME, DEFAULT_AUTH_SESSION_SECRET } from "./tests/e2e/support/mockAuthApi";

export default defineConfig({
  testDir: "./tests/e2e",
  timeout: 30_000,
  expect: {
    timeout: 8_000,
  },
  fullyParallel: true,
  reporter: "list",
  use: {
    baseURL: "http://127.0.0.1:3100",
    trace: "retain-on-failure",
  },
  projects: [
    {
      name: "chromium",
      use: { ...devices["Desktop Chrome"] },
    },
  ],
  webServer: {
    command: "npm run dev -- --hostname 127.0.0.1 --port 3100",
    port: 3100,
    reuseExistingServer: false,
    timeout: 120_000,
    env: {
      NEXT_PUBLIC_API_BASE_URL: "http://127.0.0.1:8080/api",
      NEXT_PUBLIC_MARKET_STREAM_BASE_URL: "http://127.0.0.1:8080",
      NEXT_PUBLIC_AUTH_ENABLED: "true",
      AUTH_COOKIE_NAME: DEFAULT_AUTH_COOKIE_NAME,
      AUTH_SESSION_SECRET: DEFAULT_AUTH_SESSION_SECRET,
    },
  },
});
