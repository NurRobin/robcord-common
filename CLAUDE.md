# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What Is This?

**robcord-common** contains shared Go packages and frontend assets used by multiple services in the Robcord monorepo. It is not a standalone service — it is imported as a Go module via `replace ../robcord-common` directives in service `go.mod` files.

## Go Packages

```
robcord-common/
├── auth/           claims.go         JWT claims structures (shared between Zentrale + Workspace)
├── doctor/         doctor.go         Pre-flight check framework (DB, network, env validation)
│                   checks.go         Reusable check implementations
├── envparse/       envparse.go       Typed env var parsing (string, int, bool, duration)
├── errors/         errors.go         Shared error types and helpers
├── httputil/       httputil.go       WriteJSON, WriteError, RateLimit, GetClientIP, CORS helpers
└── sqliteutil/     sqliteutil.go     SQLite connection helpers (WAL mode, busy timeout, pragmas)
```

## Frontend Assets

```
├── tailwind-theme.js     Shared Tailwind CSS theme (colors, spacing, typography)
└── tailwind-theme.d.ts   TypeScript declarations for the theme
```

The theme uses **kebab-case CSS custom properties** (`bg-bg-darkest`, `text-text-primary`). This is intentional — see FINDINGS.md IP-002.

## Usage

### Go Services

Both `robcord-zentrale` and `robcord-workspace` import common via:

```go
// go.mod
require github.com/NurRobin/robcord-common v0.0.0
replace github.com/NurRobin/robcord-common => ../robcord-common
```

### Tailwind Theme

All Next.js apps import the theme:

```js
// tailwind.config.ts
const theme = require('../../robcord-common/tailwind-theme');
```

## When to Add Code Here

**Add to common when:**
- The same utility exists (or will exist) in both Go services
- A pattern is service-agnostic (error formatting, HTTP helpers, env parsing)
- Frontend assets must be identical across apps (theme, shared types)

**Do NOT add to common when:**
- Logic is specific to one service (e.g., workspace claiming, zentrale proxy)
- It would create a circular dependency
- It's a one-off helper used in exactly one place

## Key Rules

- **Extract early**: When a pattern appears in two services, move it here immediately. Don't defer.
- **No service-specific imports**: Common must not import from zentrale, workspace, or sfu-hub.
- **Keep packages small**: Each package should have a single, clear purpose.
- **Test alongside consumers**: Common has no standalone test suite. Its code is tested via the services that import it.
