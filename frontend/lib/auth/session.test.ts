import { describe, expect, it } from "vitest";
import { createSessionToken, verifySessionToken } from "./session";

describe("auth session helper", () => {
  it("creates and verifies a signed session token", async () => {
    const token = await createSessionToken({
      username: "alpha-admin",
      secret: "super-secret",
      expiresAt: 1_900_000_000,
    });

    const session = await verifySessionToken({
      token,
      secret: "super-secret",
      now: 1_800_000_000,
    });

    expect(session.valid).toBe(true);
    expect(session.username).toBe("alpha-admin");
    expect(session.expiresAt).toBe(1_900_000_000);
  });

  it("rejects tampered or expired tokens", async () => {
    const token = await createSessionToken({
      username: "alpha-admin",
      secret: "super-secret",
      expiresAt: 1_700_000_100,
    });

    const tampered = await verifySessionToken({
      token: `${token}tampered`,
      secret: "super-secret",
      now: 1_700_000_000,
    });
    const expired = await verifySessionToken({
      token,
      secret: "super-secret",
      now: 1_700_000_200,
    });

    expect(tampered.valid).toBe(false);
    expect(expired.valid).toBe(false);
  });
});
