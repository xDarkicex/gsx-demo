# gsx-demo

The nanite stack, end to end. A social dashboard demo showing every layer working together:

```
┌─────────────────────────────────────────────────────────────┐
│  nanite router   → routes, groups, auth middleware          │
│  nanite-render   → page pipeline, components, actions,      │
│                    async/OOB streaming, memoization         │
│  nanite-gsx      → every template is a .gsx component       │
│  libraVDB        → relational users + FOLLOWS graph,        │
│                    accessed entirely through SQL            │
│  HTMX + Alpine   → CDN-loaded, server-hydrated interactivity│
└─────────────────────────────────────────────────────────────┘
```

## Pages

| Route | Auth | What it demos |
|---|---|---|
| `/` | public | Homepage — graph-derived leaderboard (`JOIN MATCH`), Alpine widget hydrated from Go via `x-data` auto-hydration |
| `/login` | public | bcrypt against libraVDB, session cookie, flash error banner, `value=` echo |
| `/dashboard` | required | Stats cards, `@async` LiveClock (skeleton → OOB swap), FOLLOWS graph traversal, `@action` follow buttons, `@memo` cache, HTMX partial refresh |
| `/_nano/action/*` | session | Colocated server actions (`@action` in `.gsx`) — CSRF baseline, flash validation at 200 |

Demo account: **alice@demo.dev / demo123** (six seeded users, a FOLLOWS graph, `demo123` for all).

## Run

```bash
# Requires the sibling repos (see go.mod replace directives):
#   ../nanite  ../nanite-render  ../nanite-gsx  ../libraVDB  ../lexer
go run .            # or: go build -o gsx-demo . && ./gsx-demo
```

Then open http://localhost:3000. The database is in-memory with a
deterministic seed — every boot is a fresh world.

## The gsx views

`views/*.gsx` are AOT-compiled to Go (`views/*_gsx.go`) by the
nanite-gsx compiler:

```bash
go run github.com/xDarkicex/nanite-gsx/cmd/gsx compile -dir views
```

Each view is a Go function taking typed props. The compiler emits a
`RenderX` function plus a `RegisterX` / `RegisterXComponent` pair.

### Superpowers on display

- **`@action`** — `follow.gsx` has a colocated server action
  (`Follow`) that mutates the FOLLOWS graph and re-renders the
  component; the button toggles in place via HTMX. Validation
  failures set a flash error and re-render at 200 (`@error`).
- **`@async` + `@fallback` + `@oob`** — `dashboard.gsx`'s LiveClock
  streams a skeleton instantly, then the worker's HTML arrives as a
  trailing HTMX OOB swap.
- **`@memo`** — the follow button's HTML is cached by
  (target, state); repeated keys skip the render walk.
- **`@yield`** — the layout composes the view purely in gsx.
- **`@for` / `@if` / `@else`** — control flow in the template.
- **Alpine hydration** — `x-data={goExpr}` auto-serializes Go state
  to the client via `c.WriteHydrateProps` (zero-alloc JSON).
- **HTMX partials** — `/dashboard/partial/following` re-renders a
  single component for the refresh button.

### The libraVDB layer

All data access is SQL (`db.Query` / `QueryWithParams`) — no
collections API after migration:

- Relational: the `users` collection (metadata schema, named unique
  constraint on email) — CRUD with `$1/$2/...` parameters.
- Graph edges: parameterized `INSERT INTO GRAPH_EDGES VALUES ($1,
  'FOLLOWS', $2)` and `DELETE FROM GRAPH_EDGES WHERE source = $1 ...`.
- Traversal: `SELECT tgt.id, tgt.name FROM users src JOIN MATCH
  (src)-[r:FOLLOWS]->(tgt) WHERE src.id = $1` — projections and
  source filtering included.
- 2-hop suggestions: two single-hop traversals aggregated in Go
  (chained `JOIN MATCH ... JOIN MATCH` target filtering is not
  reliable yet — a known libraVDB quirk).
- Persistence: the catalog, records, and graph WAL survive reopen;
  the graph reattaches with `col.SetGraph(gr)` and the WAL replays
  the edges — follow clicks persist across restarts, and the server
  flushes them on graceful shutdown (`SIGTERM` → router shutdown
  hook → `db.Close()`).

libraVDB also ships temporal and vector engines — this demo uses
its relational + graph surfaces.

## Layout

```
main.go              — wiring: engine, registry, routes, handlers
models/models.go     — Page/User types shared by every view
internal/db/         — libraVDB: open, migrate, seed, queries
internal/auth/       — bcrypt, sessions, nanite middleware
views/*.gsx          — all templates (compiled to *_gsx.go)
public/app.css       — styles (no framework)
```
