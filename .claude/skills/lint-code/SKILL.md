---
name: lint-code
description: Run the linter, fix all reported issues, and verify tests still pass. Use whenever the user says "fix lints", "update lints", "run linter", "clean up lint issues", "golangci-lint errors", or any request to make the code pass linting — even if they just mention a specific lint rule by name.
---

# Lint Code

Run the linter, fix everything it reports, repeat until clean, then verify tests pass.

## Step 1: Run and collect

```bash
golangci-lint run 2>&1
```

If exit 0 — done, nothing to fix. Otherwise read every reported issue before touching any file.

## Step 2: Fix by rule

**`nlreturn`** — blank line required before `return`:

```go
// before
_, err := doSomething()
return err

// after
_, err := doSomething()

return err
```

**`wsl_v5`** — blank line required above a statement when too many statements precede it without separation:

```go
// before
baseName := strings.TrimSuffix(fileName, ".lua")
lastDot := strings.LastIndex(baseName, ".")

// after
baseName := strings.TrimSuffix(fileName, ".lua")

lastDot := strings.LastIndex(baseName, ".")
```

**`funlen`** — function too long. Extract a helper that has one clear job. Don't just shuffle lines to pass the counter.

**`gocyclo` / `gocognit`** — too many branches. Simplify the logic or extract sub-functions. Don't add blank lines to trick the counter.

**`gocritic`** — varied. Read the message; it's usually specific (e.g., "use `errors.As` instead of type assertion", "unneeded `else` after `return`"). Apply the suggestion literally.

**`unused` / `ineffassign`** — remove the dead variable or assignment. Don't assign to `_` unless the return value matters.

**`misspell`** — fix the typo in the string or comment.

**`errcheck`** — the return error is being silently dropped. Assign it and either handle or return it.

**`staticcheck`** — read the SA code in the message; fix as directed.

For any rule not listed here, read the linter message carefully — it usually tells you exactly what to change.

## Step 3: Re-run

```bash
golangci-lint run 2>&1
```

Repeat steps 2–3 until exit 0 and "0 issues."

## Step 4: Verify tests

```bash
go test ./...
```

All packages must pass. If a lint fix broke a test, the fix was wrong — revert it and find the correct approach.
