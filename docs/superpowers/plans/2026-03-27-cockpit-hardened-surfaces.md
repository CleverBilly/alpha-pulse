# Cockpit Hardened Surfaces Implementation Plan

> **For agentic workers:** REQUIRED: Use superpowers:subagent-driven-development (if subagents available) or superpowers:executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Harden the K-line stage, evidence strip, and alerts page so the remaining soft-card feel disappears from the frontend command center.

**Architecture:** Keep business logic intact and refactor only the layout surfaces and component composition. Use TDD for the main structure landmarks first, then implement the chart screen shell, the continuous evidence strip, and the alerts hub in focused slices.

**Tech Stack:** Next.js App Router, React, TypeScript, CSS in `frontend/styles/globals.css`, Vitest, Testing Library

---

## Chunk 1: Structural Regression Tests

### Task 1: Lock the new dashboard and alerts landmarks

**Files:**
- Modify: `frontend/app/dashboard/page.test.tsx`
- Modify: `frontend/app/alerts/page.test.tsx`
- Modify: `frontend/components/dashboard/EvidenceRail.test.tsx`
- Modify: `frontend/components/chart/KlineChart.test.tsx`

- [ ] **Step 1: Write the failing tests**

Add assertions for:
- a dedicated dashboard chart stage surface and evidence strip surface
- a continuous evidence strip with segmented bays
- a K-line main screen header and technical control bar
- an alerts status band, watch desk, and history rail

- [ ] **Step 2: Run focused tests to verify they fail**

Run:
- `cd frontend && npm test -- app/dashboard/page.test.tsx`
- `cd frontend && npm test -- app/alerts/page.test.tsx`
- `cd frontend && npm test -- components/dashboard/EvidenceRail.test.tsx`
- `cd frontend && npm test -- components/chart/KlineChart.test.tsx`

Expected: FAIL on the new landmarks before implementation.

## Chunk 2: K-line Main Screen

### Task 2: Harden the chart stage into a main screen

**Files:**
- Modify: `frontend/components/chart/KlineChart.tsx`
- Modify: `frontend/styles/globals.css`

- [ ] **Step 1: Implement the minimal screen shell**

Add:
- a screen header landmark
- a technical controls landmark
- a darker instrument viewport landmark
- updated classes that support a harder, more terminal-like surface

- [ ] **Step 2: Run the focused chart test**

Run: `cd frontend && npm test -- components/chart/KlineChart.test.tsx`
Expected: PASS

## Chunk 3: Continuous Evidence Strip

### Task 3: Convert evidence cards into a segmented strip

**Files:**
- Modify: `frontend/components/dashboard/EvidenceRail.tsx`
- Modify: `frontend/styles/globals.css`

- [ ] **Step 1: Implement the strip shell and evidence bays**

Add:
- a strip shell landmark
- evidence bay landmarks
- denser summary and metric layout
- separators instead of standalone card feel

- [ ] **Step 2: Run the focused evidence test**

Run: `cd frontend && npm test -- components/dashboard/EvidenceRail.test.tsx`
Expected: PASS

## Chunk 4: Alerts Command Hub

### Task 4: Convert alerts page into a command hub

**Files:**
- Modify: `frontend/app/alerts/page.tsx`
- Modify: `frontend/components/alerts/AlertCenter.tsx`
- Modify: `frontend/components/alerts/AlertHistoryBoard.tsx`
- Modify: `frontend/styles/globals.css`

- [ ] **Step 1: Implement the status band, watch desk, and history rail**

Add:
- a denser page status band
- a watch desk shell around control actions
- a harder history rail summary and event flow

- [ ] **Step 2: Run the focused alerts tests**

Run:
- `cd frontend && npm test -- app/alerts/page.test.tsx`
- `cd frontend && npm test -- components/alerts/AlertCenter.test.tsx`
- `cd frontend && npm test -- components/alerts/AlertHistoryBoard.test.tsx`

Expected: PASS

## Chunk 5: Final Verification

### Task 5: Run final verification for the hardened surfaces pass

**Files:**
- Modify: none expected beyond implementation files above

- [ ] **Step 1: Run targeted unit tests**

Run:
- `cd frontend && npm test -- app/dashboard/page.test.tsx app/alerts/page.test.tsx components/chart/KlineChart.test.tsx components/dashboard/EvidenceRail.test.tsx components/alerts/AlertCenter.test.tsx components/alerts/AlertHistoryBoard.test.tsx`

Expected: PASS

- [ ] **Step 2: Run lint and build**

Run:
- `cd frontend && npm run lint`
- `cd frontend && npm run build`

Expected: PASS
