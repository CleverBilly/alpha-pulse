import type { Page } from "@playwright/test";
import { createSessionToken } from "../../../lib/auth/session";

interface MockAuthApiOptions {
  username?: string;
  password?: string;
  cookieName?: string;
  secret?: string;
}

export const DEFAULT_AUTH_COOKIE_NAME = "alpha_pulse_session";
export const DEFAULT_AUTH_SESSION_SECRET =
  "d4575a42ee72cf727f577f9806537bda7f25ba2cd440143f0d392f3526384127";
const E2E_APP_ORIGIN = "http://127.0.0.1:3100";

export async function primeAuthSession(page: Page, options: MockAuthApiOptions = {}) {
  const username = options.username ?? "alpha-admin";
  const cookieName = options.cookieName ?? process.env.AUTH_COOKIE_NAME ?? DEFAULT_AUTH_COOKIE_NAME;
  const secret = options.secret
    ?? process.env.AUTH_SESSION_SECRET
    ?? DEFAULT_AUTH_SESSION_SECRET;
  const token = await createSessionToken({
    username,
    secret,
    expiresAt: Math.floor(Date.now() / 1000) + 24 * 60 * 60,
  });

  await page.context().addCookies([
    {
      name: cookieName,
      value: token,
      url: E2E_APP_ORIGIN,
      httpOnly: true,
      sameSite: "Lax",
    },
  ]);
}

export async function mockAuthApi(page: Page, options: MockAuthApiOptions = {}) {
  const username = options.username ?? "alpha-admin";
  const password = options.password ?? "alpha-pass";
  const cookieName = options.cookieName ?? process.env.AUTH_COOKIE_NAME ?? DEFAULT_AUTH_COOKIE_NAME;
  const secret = options.secret
    ?? process.env.AUTH_SESSION_SECRET
    ?? DEFAULT_AUTH_SESSION_SECRET;

  await page.route(/http:\/\/(127\.0\.0\.1|localhost):8080\/api\/auth\/login/, async (route) => {
    const body = route.request().postDataJSON() as { username?: string; password?: string };
    if (body.username !== username || body.password !== password) {
      await route.fulfill({
        status: 401,
        contentType: "application/json",
        body: JSON.stringify({
          code: 401,
          message: "invalid username or password",
        }),
      });
      return;
    }

    const token = await createSessionToken({
      username,
      secret,
      expiresAt: Math.floor(Date.now() / 1000) + 24 * 60 * 60,
    });
    await primeAuthSession(page, { username, cookieName, secret });

    await route.fulfill({
      status: 200,
      headers: {
        "Content-Type": "application/json",
        "Set-Cookie": `${cookieName}=${token}; Path=/; HttpOnly; SameSite=Lax`,
      },
      body: JSON.stringify({
        code: 0,
        message: "ok",
        data: {
          enabled: true,
          authenticated: true,
          username,
        },
      }),
    });
  });

  await page.route(/http:\/\/(127\.0\.0\.1|localhost):8080\/api\/auth\/session/, async (route) => {
    const cookies = await page.context().cookies();
    const sessionCookie = cookies.find((item) => item.name === cookieName);

    await route.fulfill({
      status: 200,
      contentType: "application/json",
      body: JSON.stringify({
        code: 0,
        message: "ok",
        data: {
          enabled: true,
          authenticated: Boolean(sessionCookie),
          username: sessionCookie ? username : undefined,
        },
      }),
    });
  });

  await page.route(/http:\/\/(127\.0\.0\.1|localhost):8080\/api\/auth\/logout/, async (route) => {
    await page.context().clearCookies();
    await route.fulfill({
      status: 200,
      headers: {
        "Content-Type": "application/json",
        "Set-Cookie": `${cookieName}=; Path=/; Max-Age=0; HttpOnly; SameSite=Lax`,
      },
      body: JSON.stringify({
        code: 0,
        message: "ok",
        data: {
          enabled: true,
          authenticated: false,
        },
      }),
    });
  });
}
