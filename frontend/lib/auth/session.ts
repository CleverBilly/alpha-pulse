type CreateSessionTokenInput = {
  username: string;
  secret: string;
  expiresAt: number;
};

type VerifySessionTokenInput = {
  token: string | undefined;
  secret: string | undefined;
  now?: number;
};

type SessionVerificationResult = {
  valid: boolean;
  username?: string;
  expiresAt?: number;
};

const encoder = new TextEncoder();
const decoder = new TextDecoder();

export async function createSessionToken({
  username,
  secret,
  expiresAt,
}: CreateSessionTokenInput): Promise<string> {
  const payload = `${username}:${expiresAt}`;
  const payloadBytes = encoder.encode(payload);
  const signature = await sign(payloadBytes, secret);
  return `${toBase64Url(payloadBytes)}.${toBase64Url(signature)}`;
}

export async function verifySessionToken({
  token,
  secret,
  now = Math.floor(Date.now() / 1000),
}: VerifySessionTokenInput): Promise<SessionVerificationResult> {
  if (!token || !secret) {
    return { valid: false };
  }

  const parts = token.split(".");
  if (parts.length !== 2) {
    return { valid: false };
  }

  try {
    const payloadBytes = fromBase64Url(parts[0]);
    const signatureBytes = fromBase64Url(parts[1]);
    const verified = await verify(payloadBytes, signatureBytes, secret);
    if (!verified) {
      return { valid: false };
    }

    const payload = decoder.decode(payloadBytes);
    const [username, expiryValue] = payload.split(":");
    const expiresAt = Number(expiryValue);
    if (!username || !Number.isFinite(expiresAt) || now > expiresAt) {
      return { valid: false };
    }

    return {
      valid: true,
      username,
      expiresAt,
    };
  } catch {
    return { valid: false };
  }
}

async function sign(payload: Uint8Array, secret: string) {
  const key = await importHmacKey(secret, ["sign"]);
  const signature = await crypto.subtle.sign("HMAC", key, payload);
  return new Uint8Array(signature);
}

async function verify(payload: Uint8Array, signature: Uint8Array, secret: string) {
  const key = await importHmacKey(secret, ["verify"]);
  return crypto.subtle.verify("HMAC", key, signature, payload);
}

async function importHmacKey(secret: string, usages: KeyUsage[]) {
  return crypto.subtle.importKey(
    "raw",
    encoder.encode(secret),
    { name: "HMAC", hash: "SHA-256" },
    false,
    usages,
  );
}

function toBase64Url(bytes: Uint8Array) {
  if (typeof Buffer !== "undefined") {
    return Buffer.from(bytes)
      .toString("base64")
      .replace(/\+/g, "-")
      .replace(/\//g, "_")
      .replace(/=+$/g, "");
  }

  let binary = "";
  for (const byte of bytes) {
    binary += String.fromCharCode(byte);
  }
  return btoa(binary).replace(/\+/g, "-").replace(/\//g, "_").replace(/=+$/g, "");
}

function fromBase64Url(value: string) {
  const normalized = value.replace(/-/g, "+").replace(/_/g, "/");
  const padding = normalized.length % 4 === 0 ? "" : "=".repeat(4 - (normalized.length % 4));
  if (typeof Buffer !== "undefined") {
    return Uint8Array.from(Buffer.from(normalized + padding, "base64"));
  }

  const binary = atob(normalized + padding);
  return Uint8Array.from(binary, (char) => char.charCodeAt(0));
}
