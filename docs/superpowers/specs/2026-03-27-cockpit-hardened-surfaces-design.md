# Cockpit Hardened Surfaces Design

## Goal

Push the redesigned frontend one step further away from "soft SaaS cards" and closer to a professional command center by hardening three remaining areas:

- the dashboard K-line stage
- the dashboard evidence chain
- the alerts page

## Problem

The shell and hero regions already feel more intentional, but these three zones still carry too much rounded-card behavior:

- the K-line area still reads like a chart placed inside a premium card
- the evidence chain still looks like three adjacent cards
- the alerts page still behaves like a page composed from cards instead of a true command hub

## Chosen Direction

Use a mixed hardening pass:

- **K-line stage:** treat it like a main screen with a darker, inset instrument viewport
- **Evidence chain:** treat it like a continuous analytical strip with partitioned bays
- **Alerts:** treat it like an event command hub with a live watch desk and a separate history rail

This preserves the light professional shell while removing the last soft-card bias inside the workspace.

## Design

### 1. K-line Main Screen

The K-line section becomes a true command surface instead of a white content card.

- add a structural frame around the chart stage
- replace the generic title row with a dedicated screen header
- turn the layer chip row into a technical control bar
- give the chart viewport its own darker instrument surface
- dock chart metadata and overlays so they feel attached to the screen rather than floating inside the card

The visual goal is "main trading screen", not "chart widget".

### 2. Continuous Evidence Strip

The evidence area becomes one continuous strip that is internally segmented into three evidence bays.

- keep the current data model and evidence types
- replace the three isolated cards with one strip container
- add per-bay indexing, state, summary line, metric grid, and deep-link action
- use separators and sectional rhythm rather than separate big rounded cards
- keep missing evidence inside the same strip in a dimmed "not ready" state

The visual goal is "analysis chain", not "three cards in a row".

### 3. Alerts Command Hub

The alerts page becomes a page-level command hub.

- replace the descriptive hero with a denser status band
- reshape the control area into a watch desk
- keep drawer behavior for global quick access, but make the page itself the primary control surface
- present live events as a continuous event flow instead of stacked soft cards
- reshape alert history into a rail-like review workspace with denser summary readouts

The visual goal is "dispatch center", not "alerts feature page".

## Boundaries

- keep existing API and store behavior intact
- do not change chart calculations, evidence generation logic, or alert polling mechanics
- focus on structure, visual hierarchy, surface language, and layout primitives
- add only focused regression tests for the new structural shells and key landmarks

## Testing

- update dashboard page structure test to cover the hardened chart stage and evidence strip landmarks
- update K-line chart test to assert the new screen header and control surface landmarks
- update evidence rail test to assert the continuous strip and segmented bays
- update alerts page test to assert the status band, watch desk, and history rail landmarks
- update alert history / alert center tests only as needed for structural changes
