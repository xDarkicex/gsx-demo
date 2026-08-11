# bug-bounty.md — rough edges found while building the supabase-style dashboard

Every issue found while building a real dashboard (supabase/studio as the
reference) against the nanite stack + libraVDB. Format per entry:

- **Issue** — what's wrong
- **Tried** — what we attempted
- **Happened** — the actual outcome
- **Desired** — what should happen
- **Error** — the exact error message (if any)
- **Repro** — minimal code
- **Status** — open / fixed

---

## Resolved (sent to libraVDB agent, fixed + regression-tested)

### R1. SELECT * column misalignment with JSON columns

- **Tried**: `SELECT * FROM todos` on a table with `(id, title, completed BOOLEAN, priority INTEGER, due_at TIMESTAMP, tags JSON)` after inserting all six columns.
- **Happened**: metadata came back shifted left by one — `completed:1`, `due_at:["docs","go"]` (the tags value), `priority:2026-08-12 10:00:00` (the due_at value), and `tags` missing entirely. Everything downstream (`GROUP BY`, `DISTINCT`, `SELECT *`-based reads) returned garbage.
- **Desired**: column-perfect projection.
- **Error**: none — silent data corruption.
- **Repro**:
  ```sql
  CREATE TABLE todos (id TEXT PRIMARY KEY, title TEXT NOT NULL, completed BOOLEAN DEFAULT false,
                      priority INTEGER DEFAULT 3, due_at TIMESTAMP, tags JSON)
  INSERT INTO todos (id, title, completed, priority, due_at, tags)
    VALUES ('t1','Write docs',false,1,'2026-08-12 10:00:00','["docs","go"]')
  SELECT * FROM todos
  ```
- **Status**: fixed — SELECT * now preserves JSON, timestamp, boolean, and integer columns.

### R2. Boolean true encodes as int 2; WHERE completed = true matches nothing

- **Tried**: `INSERT ... completed=true` then `SELECT id FROM todos WHERE completed = true`.
- **Happened**: `true` read back as `2` (int); `false` read back as `false`. The boolean WHERE predicate returned 0 rows.
- **Desired**: `true` round-trips as `true` and the predicate matches.
- **Repro**:
  ```sql
  INSERT INTO todos (id, title, completed, priority) VALUES ('t2','Ship demo',true,2)
  SELECT id FROM todos WHERE completed = true   -- 0 rows
  ```
- **Status**: fixed — TRUE/FALSE literals bind and round-trip; `WHERE completed = true` works.

### R3. IN and BETWEEN return 0 rows

- **Tried**: `WHERE priority IN (1, 2)` and `WHERE priority BETWEEN 1 AND 2` against rows with `priority` literally 1 and 2.
- **Happened**: 0 rows; no error. `LIKE` and `LIMIT/OFFSET` worked in the same probe.
- **Desired**: both predicates filter correctly.
- **Status**: fixed.

### R4. Multi-row INSERT breaks on JSON literals

- **Tried**: two-row `INSERT ... VALUES (...), (...)` where one value is `'["docs","go"]'`.
- **Happened**: `invalid JSON value for column "tags"` — the tuple parser split on the JSON array's internal commas.
- **Desired**: quoted literals (any content, including commas) survive tuple splitting.
- **Repro**:
  ```sql
  INSERT INTO todos (id, title, completed, priority, due_at, tags) VALUES
    ('t1','Write docs',false,1,'2026-08-12 10:00:00','["docs","go"]'),
    ('t2','Ship demo',true,2,'2026-08-11 09:00:00','["ship"]')
  ```
- **Status**: fixed.

### R5. AS OF metadata lost the latest UPDATE

- **Tried**: INSERT → `UPDATE ... SET completed = true` → `SELECT ... AS OF TIMESTAMP` at a later time.
- **Happened**: the updated value (`completed`) came back `<nil>` at a future timestamp.
- **Desired**: AS OF reflects the latest version at/after the write.
- **Status**: fixed — UPDATE persists correctly through AS OF; pgwire also emits correct types for boolean/timestamp/JSON.

---

## Verified working (build on this)

- Full `CREATE TABLE` DDL: `TEXT PRIMARY KEY`, `NOT NULL`, `DEFAULT`, `BOOLEAN`, `INTEGER`, `TIMESTAMP`, `JSON`.
- Single- and multi-row INSERTs (incl. JSON), `SELECT *`, `WHERE` (=, LIKE, IN, BETWEEN), `LIMIT/OFFSET`, `ORDER BY`, `UPDATE`, `DELETE` (returns the row), `COUNT`, `DISTINCT`, `GROUP BY`.
- Temporal: `AS OF TIMESTAMP`, `VERSIONS OF <table> BETWEEN TIMESTAMP ... AND TIMESTAMP ...` with `version`, `version_start`, `version_end`.
  - Note: the range start must be ≥ the oldest retained version; earlier starts error with
    `resolve temporal range start: retention expired: requested <t>, oldest retained <t>` (expected).
- Graph: parameterized `INSERT INTO GRAPH_EDGES VALUES ($1,'FOLLOWS',$2)` / `DELETE`, single and chained `JOIN MATCH` with projections + WHERE + GROUP BY, `SetGraph` reattach + WAL replay.

---

## Open (sent to the relevant agent / under investigation)

### G1. [nanite-gsx] Static attribute values containing quotes break generated Go — FIXED

- **Tried**: `placeholder='["ops"]'` on an input.
- **Happened**: the codegen emitted `c.WriteString(" placeholder=\"["ops"]\"")` — the value's `"` chars
  landed raw inside a double-quoted Go string literal → `syntax error: unexpected name ops`.
- **Desired**: attribute values escape correctly.
- **Repro**: `<input type="text" placeholder='["ops"]' />`
- **Status**: fixed — static attr values now pass through `goStr` (escapes `\` and `"`). Regression test
  `TestE2E_QuotedAttrValue`.

### G2. [nanite-gsx] @for loop variables shadow the generated context param `c` — OPEN

- **Tried**: `@for _, c := range p.Dash.SQLColumns { <th>{c}</th> }`.
- **Happened**: the codegen emits `for _, c := range ... { c.WriteString(...) }` — `c` is now a string,
  shadowing the `*render.ComponentContext` param → `c.WriteString undefined (type string has no field...)`.
- **Desired**: loop variables can't collide with the context param (rename one side automatically, or
  error with a clear message).
- **Repro**: any `@for` whose variables include `c`.
- **Status**: open — demo workaround: rename the loop variable (`col`, `cell`).

### L6. [libraVDB] NOT in UPDATE SET is unsupported — OPEN

- **Tried**: `UPDATE todos SET completed = NOT completed WHERE id = $1`.
- **Happened**: `optimize error: UPDATE SET expression: expression kind 13 is unsupported`.
- **Desired**: unary NOT (and presumably other boolean ops) work in UPDATE SET expressions.
- **Repro**:
  ```sql
  CREATE TABLE t (id TEXT PRIMARY KEY, done BOOLEAN DEFAULT false)
  INSERT INTO t (id) VALUES ('a')
  UPDATE t SET done = NOT done WHERE id = 'a'
  ```
- **Status**: open — demo workaround: read-modify-write (SELECT then UPDATE with the flipped literal).

### L7. [libraVDB] VERSIONS OF metadata value types — NOTE

- `version` comes back as a uint type (not int), and `version_start`/`version_end` as `time.Time`
  (not strings) in the result metadata. Not a bug — but consumers must handle the types
  (`fmt.Sprint` renders the timestamps; int conversion needs the uint cases).
- The temporal range start must be ≥ the oldest retained version, else
  `resolve temporal range start: retention expired: requested <t>, oldest retained <t>` (expected).
