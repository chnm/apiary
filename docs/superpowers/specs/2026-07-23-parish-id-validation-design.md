# Parish ID Validation — Design

**Date:** 2026-07-23
**Related:** PR #97, Issue #44
**Status:** Approved

## Problem

The BOM bills endpoint accepts a `parish` query parameter (comma-separated parish
IDs). An ID outside the range of the `bom.parishes` table silently returns an
empty array instead of an error (Issue #44). PR #97 added a hardcoded range check
(1–158), but the actual table currently holds 151 parishes — the hardcoded bound
was already stale, and the parish table changes while the server is running, so
any hardcoded bound will drift again.

## Decision

Validate parish IDs against the database on each request rather than against a
hardcoded range. Chosen over (a) a corrected hardcoded constant, which drifts as
data changes, and (b) a TTL-cached ID set, which adds refresh machinery and a
staleness window where newly added parishes are wrongly rejected. A per-request
primary-key lookup is negligible, and the bills endpoint sits behind the server's
1-hour HTTP response cache.

## Design

### Changes

- **`bom-bills.go` / `parseAPIParameters`:** remove the hardcoded 1–158 range
  check. The parser returns to rejecting only non-integer parish values, as on
  `main`.
- **`bom-parishes.go`:** add `(s *Server) invalidParishIDs(ctx context.Context,
  ids []int) ([]int, error)`. Runs `SELECT id FROM bom.parishes WHERE id =
  ANY($1)`, collects found IDs, and returns the requested IDs not present in the
  table. The requested-vs-found diff is factored as a pure function so it is
  unit-testable without a database.
- **`bom-bills.go` / `BillsHandler`:** after successful parameter parsing, when
  `params.Parish` is non-empty, call `invalidParishIDs`. Non-empty result →
  `400` with `invalid parish ID(s): <list>` via the existing `http.Error`
  pattern. Query error → `500 Database error`, matching existing style.

### Data flow

Request → `parseAPIParameters` (400 on malformed input, unchanged) →
`invalidParishIDs` (400 on unknown IDs, 500 on DB failure) → bills query as
before. Zero and negative IDs need no special casing — they are absent from the
table and fall out of the same check.

### Testing

- Unit test the pure diff function (missing-ID computation).
- Unit tests for `parseAPIParameters` parish parsing (non-integer rejection,
  comma-separated lists) in the root package.
- End-to-end verification against the live database is optional (VPN), not
  required for merge confidence.

## Delivery

Implemented on the existing `refactor/parish-out-of-range-check` branch,
replacing the range check, so PR #97 remains the vehicle and closes #44. The PR
description will be updated to describe the DB-backed approach.
