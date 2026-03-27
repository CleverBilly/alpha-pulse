# Login Redirect Loop Dev Auth Design

**Problem**

Local development can enter a login loop: the backend accepts credentials and issues a session cookie, but the frontend middleware still rejects the session and redirects back to `/login`.

**Root Cause**

The frontend middleware validates the session cookie with its own `AUTH_SESSION_SECRET` and `NEXT_PUBLIC_AUTH_ENABLED` environment. During `./scripts/dev.sh`, the backend loads `backend/.env`, but the frontend process only receives shell environment plus `frontend/.env.local`. When `frontend/.env.local` is missing or incomplete, the frontend can run with a different auth configuration from the backend.

**Approaches**

1. Relax frontend middleware validation when config is missing.
   This avoids the redirect loop but weakens route protection and hides configuration errors.
2. Make local dev startup propagate auth configuration from `backend/.env` into the frontend process.
   This fixes the configuration split at the source and keeps middleware protection intact.

**Chosen Design**

Use approach 2. Extract a small shell helper that:

- reads selected auth variables from `frontend/.env.local` and `backend/.env`
- keeps explicit precedence for `./scripts/dev.sh`: shell env, then `backend/.env`, then `frontend/.env.local`
- maps `ENABLE_SINGLE_USER_AUTH` to `NEXT_PUBLIC_AUTH_ENABLED` when the frontend value is absent
- fails fast when frontend auth is enabled but `AUTH_SESSION_SECRET` is still missing

`scripts/dev.sh` will source the helper before starting the frontend so local development gets a consistent auth configuration by default.

**Testing**

- add a focused shell regression test for auth env resolution precedence
- verify the test fails before the helper exists
- run the focused regression after the fix
- run frontend lint, build, and auth e2e checks
