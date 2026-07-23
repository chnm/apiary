# DB-Backed Parish ID Validation Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace PR #97's hardcoded parish ID range check (1–158) with per-request validation against the `bom.parishes` table, returning 400 with the offending IDs.

**Architecture:** The parser (`parseAPIParameters`) goes back to rejecting only malformed input; a new `Server.invalidParishIDs` method in `bom-parishes.go` checks requested IDs against the database with `SELECT id FROM bom.parishes WHERE id = ANY($1)`; `BillsHandler` calls it after parsing and 400s on unknown IDs. The requested-vs-found diff is a pure function so it is unit-testable without a database.

**Tech Stack:** Go, pgx/v4 (`s.DB` is a `*pgxpool.Pool`), gorilla/mux, standard `testing` package.

**Spec:** `docs/superpowers/specs/2026-07-23-parish-id-validation-design.md`

## Global Constraints

- Work on branch `refactor/parish-out-of-range-check` (PR #97). Do not create a new branch.
- `go vet` currently fails on an unrelated pre-existing bug in `bom-causes.go` (a `log.Printf` with 15 args for 14 verbs; a fix is in flight in another session). Until that merges, run root-package tests as `go test -vet=off .` — do NOT "fix" `bom-causes.go` in this branch.
- The 400 error message format is exactly: `invalid parish ID(s): 152, 999` (comma-space separated, request order).
- Integration tests in `cmd/apiary` need a live database (`APIARY_DB` set; CHNM database requires VPN). If unavailable, note it and let the user run them.
- Error responses use the file's existing pattern: `http.Error(w, msg, http.StatusBadRequest)` for client errors, `http.Error(w, "Database error", http.StatusInternalServerError)` + `log.Printf` for DB failures.
- Commits use conventional-commit style with `Co-Authored-By: Claude Fable 5 <noreply@anthropic.com>`.

---

### Task 1: Remove the hardcoded range check from the parser

**Files:**
- Modify: `bom-bills.go` (parish block inside `parseAPIParameters`, ~lines 262–276)
- Test: `bom-bills_test.go` (create, root package)

**Interfaces:**
- Consumes: `parseAPIParameters(r *http.Request) (APIParameters, error)` — existing, unexported, root package.
- Produces: parser behavior later tasks rely on — any integer parish IDs (including 0 or 99999) parse successfully into `params.Parish []int`; only non-integers error.

- [ ] **Step 1: Write the failing test**

Create `bom-bills_test.go`:

```go
package apiary

import (
	"net/http/httptest"
	"reflect"
	"testing"
)

func TestParseAPIParametersParish(t *testing.T) {
	// Non-integer parish IDs are rejected by the parser.
	r := httptest.NewRequest("GET", "/bom/bills?parish=abc", nil)
	if _, err := parseAPIParameters(r); err == nil {
		t.Error("expected error for non-integer parish ID, got nil")
	}

	// Comma-separated parish IDs are parsed in order, with whitespace trimmed.
	r = httptest.NewRequest("GET", "/bom/bills?parish=1,%205,151", nil)
	params, err := parseAPIParameters(r)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !reflect.DeepEqual(params.Parish, []int{1, 5, 151}) {
		t.Errorf("expected [1 5 151], got %v", params.Parish)
	}

	// The parser does not range-check IDs; existence is validated against
	// the database by Server.invalidParishIDs.
	r = httptest.NewRequest("GET", "/bom/bills?parish=0,99999", nil)
	params, err = parseAPIParameters(r)
	if err != nil {
		t.Fatalf("unexpected error for out-of-range IDs: %v", err)
	}
	if !reflect.DeepEqual(params.Parish, []int{0, 99999}) {
		t.Errorf("expected [0 99999], got %v", params.Parish)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test -vet=off -run TestParseAPIParametersParish -v .`
Expected: FAIL — the third case errors with `parish ID must be between 1 and 158` (the hardcoded check is still present).

- [ ] **Step 3: Remove the range check**

In `bom-bills.go`, inside the parish block of `parseAPIParameters`, delete these three lines (restoring the block to its state on `main`):

```go
			if parishInt < 1 || parishInt > 158 {
				return params, fmt.Errorf("parish ID must be between 1 and 158")
			}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test -vet=off -run TestParseAPIParametersParish -v .`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add bom-bills.go bom-bills_test.go
git commit -m "refactor: Remove hardcoded parish ID range check from parser

Parish ID existence will be validated against the database instead.

Co-Authored-By: Claude Fable 5 <noreply@anthropic.com>"
```

---

### Task 2: Pure helpers — missing-ID diff and int list formatting

**Files:**
- Modify: `bom-parishes.go` (add `missingIDs`)
- Modify: `helpers.go` (add `intsToString`; add `"strconv"` and `"strings"` to its import block if not present)
- Test: `bom-parishes_test.go` (create, root package), `helpers_test.go` (append)

**Interfaces:**
- Consumes: nothing from other tasks.
- Produces: `missingIDs(requested []int, found map[int]bool) []int` — returns members of `requested` absent from `found`, in request order, duplicates reported once, nil when nothing is missing. `intsToString(ids []int) string` — formats as `"152, 999"`.

- [ ] **Step 1: Write the failing tests**

Create `bom-parishes_test.go`:

```go
package apiary

import (
	"reflect"
	"testing"
)

func TestMissingIDs(t *testing.T) {
	found := map[int]bool{1: true, 5: true}
	cases := []struct {
		name      string
		requested []int
		want      []int
	}{
		{"all found", []int{1, 5}, nil},
		{"some missing", []int{1, 2, 5, 9}, []int{2, 9}},
		{"duplicates reported once", []int{2, 2}, []int{2}},
		{"empty request", nil, nil},
	}
	for _, c := range cases {
		got := missingIDs(c.requested, found)
		if !reflect.DeepEqual(got, c.want) {
			t.Errorf("%s: missingIDs(%v) = %v, want %v", c.name, c.requested, got, c.want)
		}
	}
}
```

Append to `helpers_test.go`:

```go
func TestIntsToString(t *testing.T) {
	if got := intsToString([]int{152, 999}); got != "152, 999" {
		t.Errorf(`intsToString([152 999]) = %q, want "152, 999"`, got)
	}
	if got := intsToString([]int{7}); got != "7" {
		t.Errorf(`intsToString([7]) = %q, want "7"`, got)
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test -vet=off -run 'TestMissingIDs|TestIntsToString' -v .`
Expected: FAIL to build — `undefined: missingIDs`, `undefined: intsToString`.

- [ ] **Step 3: Write minimal implementations**

Add to `bom-parishes.go`:

```go
// missingIDs returns the members of requested that are not present in found,
// in request order and without duplicates. A nil result means nothing is missing.
func missingIDs(requested []int, found map[int]bool) []int {
	var missing []int
	seen := make(map[int]bool, len(requested))
	for _, id := range requested {
		if !found[id] && !seen[id] {
			missing = append(missing, id)
			seen[id] = true
		}
	}
	return missing
}
```

Add to `helpers.go` (and add `"strconv"` and `"strings"` to its imports if absent):

```go
// intsToString formats a slice of ints as a comma-separated list, e.g. "152, 999".
func intsToString(ids []int) string {
	parts := make([]string, len(ids))
	for i, id := range ids {
		parts[i] = strconv.Itoa(id)
	}
	return strings.Join(parts, ", ")
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test -vet=off -run 'TestMissingIDs|TestIntsToString' -v .`
Expected: PASS (both tests)

- [ ] **Step 5: Commit**

```bash
git add bom-parishes.go bom-parishes_test.go helpers.go helpers_test.go
git commit -m "feat: Add missingIDs and intsToString helpers

Co-Authored-By: Claude Fable 5 <noreply@anthropic.com>"
```

---

### Task 3: Database-backed lookup — Server.invalidParishIDs

**Files:**
- Modify: `bom-parishes.go` (add method; its import block already has `context`)

**Interfaces:**
- Consumes: `missingIDs` from Task 2; `s.DB` (`*pgxpool.Pool`) from `server.go`.
- Produces: `(s *Server) invalidParishIDs(ctx context.Context, ids []int) ([]int, error)` — nil/empty result means all IDs exist in `bom.parishes`; error means the query itself failed.

This method is exercised through the handler integration test in Task 4 — it cannot be unit-tested without a database, so this task is implement + compile.

- [ ] **Step 1: Implement the method**

Add to `bom-parishes.go`:

```go
// invalidParishIDs returns the parish IDs from ids that do not exist in the
// bom.parishes table. A nil result means all IDs are valid. An error is
// returned only if the lookup query itself fails.
func (s *Server) invalidParishIDs(ctx context.Context, ids []int) ([]int, error) {
	if len(ids) == 0 {
		return nil, nil
	}

	query := `SELECT id FROM bom.parishes WHERE id = ANY($1);`
	rows, err := s.DB.Query(ctx, query, ids)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	found := make(map[int]bool, len(ids))
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		found[id] = true
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return missingIDs(ids, found), nil
}
```

- [ ] **Step 2: Verify it compiles and existing tests still pass**

Run: `go build ./... && go test -vet=off .`
Expected: build succeeds; `ok github.com/chnm/apiary`

- [ ] **Step 3: Commit**

```bash
git add bom-parishes.go
git commit -m "feat: Add Server.invalidParishIDs database lookup

Co-Authored-By: Claude Fable 5 <noreply@anthropic.com>"
```

---

### Task 4: Wire validation into BillsHandler

**Files:**
- Modify: `bom-bills.go` (`BillsHandler`, immediately after the `parseAPIParameters` error return, ~line 105; update the handler's godoc comment)
- Test: `cmd/apiary/bom_test.go` (append; integration, needs live DB)

**Interfaces:**
- Consumes: `invalidParishIDs` (Task 3), `intsToString` (Task 2), `apiParams.Parish []int` (Task 1).
- Produces: HTTP behavior — `GET /bom/bills?parish=<unknown id>` → 400 with body `invalid parish ID(s): <ids>`; valid IDs unchanged.

- [ ] **Step 1: Write the integration test**

Append to `cmd/apiary/bom_test.go`:

```go
func TestBomBillsInvalidParish(t *testing.T) {
	// Unknown parish IDs return 400 rather than an empty result set.
	req, _ := http.NewRequest("GET", "/bom/bills?start-year=1669&end-year=1754&parish=99999", nil)
	response := executeRequest(req)
	checkResponseCode(t, http.StatusBadRequest, response.Code)

	// A valid parish ID still returns 200.
	req, _ = http.NewRequest("GET", "/bom/bills?start-year=1669&end-year=1754&parish=1", nil)
	response = executeRequest(req)
	checkResponseCode(t, http.StatusOK, response.Code)
}
```

- [ ] **Step 2: Run the integration test to verify it fails (only if APIARY_DB is available)**

Run: `go test -run TestBomBillsInvalidParish -v ./cmd/apiary`
Expected: FAIL — the first request currently returns 200 with `[]`.
If `APIARY_DB` is not set (no VPN), record that this step was skipped and continue; the user will run it.

- [ ] **Step 3: Implement the handler wiring**

In `bom-bills.go`, inside `BillsHandler` directly after the `parseAPIParameters` error block (`return` at ~line 105), insert:

```go
		// Reject parish IDs that don't exist in the database
		if len(apiParams.Parish) > 0 {
			invalid, err := s.invalidParishIDs(r.Context(), apiParams.Parish)
			if err != nil {
				log.Printf("Error validating parish IDs: %v", err)
				http.Error(w, "Database error", http.StatusInternalServerError)
				return
			}
			if len(invalid) > 0 {
				http.Error(w, fmt.Sprintf("invalid parish ID(s): %s", intsToString(invalid)), http.StatusBadRequest)
				return
			}
		}
```

Update `BillsHandler`'s godoc comment to note the behavior, e.g. append the sentence: `Parish IDs are validated against the bom.parishes table; unknown IDs return 400 Bad Request.`

- [ ] **Step 4: Verify build and root tests; run integration test if DB available**

Run: `go build ./... && go test -vet=off .`
Expected: build succeeds; root tests pass.
If `APIARY_DB` is set: `go test -run TestBomBillsInvalidParish -v ./cmd/apiary` — Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add bom-bills.go cmd/apiary/bom_test.go
git commit -m "feat: Validate parish IDs against the database in BillsHandler

Unknown parish IDs now return 400 with the offending IDs listed,
instead of silently returning an empty array. Closes #44.

Co-Authored-By: Claude Fable 5 <noreply@anthropic.com>"
```

---

### Task 5: Push branch and update PR #97

**Files:** none (git/GitHub operations only)

**Interfaces:**
- Consumes: all commits from Tasks 1–4 plus the spec/plan doc commits already on the branch.
- Produces: updated PR #97 ready for the user to merge.

- [ ] **Step 1: Final verification before pushing**

Run: `go build ./... && go test -vet=off .`
Expected: build succeeds, root tests pass. (This is the superpowers:verification-before-completion gate — do not push on a red build.)

- [ ] **Step 2: Push the branch**

```bash
git push origin refactor/parish-out-of-range-check
```

- [ ] **Step 3: Update the PR description**

```bash
gh pr edit 97 --body "Closes #44

Replaces the original hardcoded parish ID range check with validation against the database, per the design in \`docs/superpowers/specs/2026-07-23-parish-id-validation-design.md\`.

* \`parseAPIParameters\` (in \`bom-bills.go\`) rejects only malformed (non-integer) parish IDs.
* A new \`Server.invalidParishIDs\` method (in \`bom-parishes.go\`) checks requested IDs against \`bom.parishes\` with a single \`WHERE id = ANY(\$1)\` lookup.
* \`BillsHandler\` returns \`400 invalid parish ID(s): <ids>\` for unknown IDs and \`500\` if the lookup query fails.

Chosen over a corrected hardcoded bound (drifts as the parish table grows — the original 158 was already stale at 151) and a TTL-cached ID set (staleness window rejects newly added parishes). The per-request lookup is a single primary-key query, and the endpoint sits behind the server's 1-hour response cache."
```

- [ ] **Step 4: Report status to the user**

State what was pushed, whether the integration test ran (or needs the VPN), and that the PR is ready for their review and merge.
