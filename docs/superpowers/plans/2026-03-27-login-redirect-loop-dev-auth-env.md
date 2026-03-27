# Login Redirect Loop Dev Auth Env Implementation Plan

> **For agentic workers:** REQUIRED: Use superpowers:subagent-driven-development (if subagents available) or superpowers:executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make local development start the frontend with the same auth configuration as the backend so successful logins stop bouncing back to `/login`.

**Architecture:** Extract auth env resolution into a small shell helper used by `scripts/dev.sh`. The helper reads only the auth-related keys we need, treats the local backend as the auth source of truth with precedence `shell env > backend/.env > frontend/.env.local`, and fails fast on invalid local auth configuration.

**Tech Stack:** Bash, existing repo scripts, Playwright, Next.js

---

## Chunk 1: Regression Test

### Task 1: Add a failing shell regression for auth env resolution

**Files:**
- Create: `scripts/dev_env_test.sh`
- Test: `scripts/dev_env_test.sh`

- [ ] **Step 1: Write the failing test**

Create a shell test that expects a helper script to:
- derive `NEXT_PUBLIC_AUTH_ENABLED=true` from `ENABLE_SINGLE_USER_AUTH=true`
- fall back to backend `AUTH_COOKIE_NAME` and `AUTH_SESSION_SECRET`
- use frontend auth values only when backend auth values are absent

- [ ] **Step 2: Run test to verify it fails**

Run: `bash scripts/dev_env_test.sh`
Expected: FAIL because the helper script does not exist yet.

## Chunk 2: Minimal Fix

### Task 2: Implement auth env resolution for local dev startup

**Files:**
- Create: `scripts/dev_env.sh`
- Modify: `scripts/dev.sh`
- Test: `scripts/dev_env_test.sh`

- [ ] **Step 1: Write minimal implementation**

Add shell helpers that:
- read one value from an env file without executing the file
- resolve auth env with precedence `shell > frontend/.env.local > backend/.env > defaults`
- export resolved values before starting the frontend
- exit early with a clear error if auth is enabled but the frontend secret is still empty

- [ ] **Step 2: Run the focused test to verify it passes**

Run: `bash scripts/dev_env_test.sh`
Expected: PASS

## Chunk 3: Verification

### Task 3: Re-run relevant verification

**Files:**
- Modify: `README.md` if the new startup behavior needs clarification

- [ ] **Step 1: Run targeted UI auth verification**

Run: `cd frontend && npm run test:e2e -- auth.spec.ts`
Expected: PASS

- [ ] **Step 2: Run broader frontend safety checks**

Run: `cd frontend && npm run lint && npm run build`
Expected: all commands succeed
