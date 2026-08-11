// Package db wraps libraVDB for the demo. All data access goes
// through the native SQL surface — relational tables plus the
// FOLLOWS graph (parameterized INSERT/DELETE GRAPH_EDGES, JOIN
// MATCH traversals with projections + source filtering).
//
// libraVDB also ships temporal and vector engines; this demo uses
// its relational + graph surfaces.
//
// Persistence: the SQL catalog and records survive reopen; the
// graph reattaches via SetGraph and its WAL replays the edges, so
// follow clicks persist across restarts with zero extra state.
package db

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/xDarkicex/libravdb/libravdb"

	"github.com/xDarkicex/gsx-demo/models"
)

// Default is the demo's data layer. main opens it at boot;
// @action bodies (static Go in generated views) use it directly.
var Default *DB

// DB is the demo's data layer: a libraVDB handle plus its graph.
type DB struct {
	raw *libravdb.Database
	gr  libravdb.Graph

	// BootTime anchors temporal queries — the earliest retained
	// version is the first write, which happens at boot.
	BootTime time.Time
}

// Open opens (or creates) the demo database on disk. Follow
// clicks persist across restarts (graph WAL replay on reopen).
func Open(dir string) (*DB, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}
	raw, err := libravdb.Open(libravdb.WithStoragePath(dir + "/demo.libravdb"))
	if err != nil {
		return nil, err
	}
	gr, err := libravdb.NewGraph(libravdb.GraphConfig{})
	if err != nil {
		raw.Close()
		return nil, err
	}
	d := &DB{raw: raw, gr: gr, BootTime: time.Now().UTC()}
	if err := d.migrate(context.Background()); err != nil {
		gr.Close()
		raw.Close()
		return nil, err
	}
	Default = d
	return d, nil
}

// Close releases the database and its graph.
func (d *DB) Close() error {
	d.gr.Close()
	return d.raw.Close()
}

// migrate ensures the graph-backed users collection. On first boot
// it is created (graph binding + metadata schema + unique email
// constraint); on reopen the catalog is intact and the graph is
// reattached, with the WAL replaying the FOLLOWS edges.
func (d *DB) migrate(ctx context.Context) error {
	// Register the FOLLOWS edge kind (process-wide, idempotent).
	libravdb.RegisterEdgeKind("FOLLOWS", 1)

	// The durable click counter (standard SQL CRUD — the catalog
	// persists, so the table and its rows survive reopen).
	if _, err := d.raw.Query(ctx,
		"CREATE TABLE counter (id TEXT PRIMARY KEY, value INTEGER)"); err != nil {
		if !strings.Contains(err.Error(), "exists") {
			return fmt.Errorf("create counter table: %w", err)
		}
	}
	if _, err := d.raw.Query(ctx,
		"INSERT INTO counter (id, value) VALUES ('home', 0)"); err != nil {
		// Row already present on reopen.
	}

	col, err := d.raw.GetCollection("users")
	if err == nil {
		// Reopened: reattach the graph; the WAL replays edges.
		col.SetGraph(d.gr)
		return nil
	}

	_, err = d.raw.CreateCollection(ctx, "users",
		libravdb.WithMetadataOnly(),
		libravdb.WithGraph(d.gr),
		libravdb.WithMetadataSchema(libravdb.MetadataSchema{
			"email":         libravdb.StringField,
			"password_hash": libravdb.StringField,
			"name":          libravdb.StringField,
			"created_at":    libravdb.StringField,
		}),
		libravdb.WithNamedUniqueConstraint("users_email_unique", "email"),
	)
	return err
}

// --- queries ---

// UserByEmail looks up a user by unique email.
func (d *DB) UserByEmail(ctx context.Context, email string) (*models.User, error) {
	res, err := d.raw.QueryWithParams(ctx,
		"SELECT id, email, password_hash, name, created_at FROM users WHERE email = $1",
		libravdb.QueryParams{"1": email})
	if err != nil {
		return nil, err
	}
	if len(res.Results) == 0 {
		return nil, nil
	}
	return scanUser(res.Results[0])
}

// UserByID looks up a user by primary key.
func (d *DB) UserByID(ctx context.Context, id string) (*models.User, error) {
	res, err := d.raw.QueryWithParams(ctx,
		"SELECT id, email, password_hash, name, created_at FROM users WHERE id = $1",
		libravdb.QueryParams{"1": id})
	if err != nil {
		return nil, err
	}
	if len(res.Results) == 0 {
		return nil, nil
	}
	return scanUser(res.Results[0])
}

// CreateUser inserts a new user. The unique email constraint is
// enforced by libraVDB; a duplicate returns an error.
func (d *DB) CreateUser(ctx context.Context, u *models.User) error {
	_, err := d.raw.QueryWithParams(ctx,
		"INSERT INTO users (id, email, password_hash, name, created_at) VALUES ($1, $2, $3, $4, $5)",
		libravdb.QueryParams{
			"1": u.ID, "2": u.Email, "3": u.PasswordHash,
			"4": u.Name, "5": u.CreatedAt,
		})
	return err
}

// Follow adds a FOLLOWS edge — parameterized GRAPH_EDGES INSERT.
func (d *DB) Follow(ctx context.Context, me, target string) error {
	if me == target {
		return errors.New("cannot follow yourself")
	}
	_, err := d.raw.QueryWithParams(ctx,
		"INSERT INTO GRAPH_EDGES VALUES ($1, 'FOLLOWS', $2)",
		libravdb.QueryParams{"1": me, "2": target})
	return err
}

// Unfollow removes a FOLLOWS edge — parameterized GRAPH_EDGES
// DELETE.
func (d *DB) Unfollow(ctx context.Context, me, target string) error {
	_, err := d.raw.QueryWithParams(ctx,
		"DELETE FROM GRAPH_EDGES WHERE source = $1 AND type = 'FOLLOWS' AND target = $2",
		libravdb.QueryParams{"1": me, "2": target})
	return err
}

// Following returns the users the given user follows — a JOIN
// MATCH traversal with projections and source filtering.
func (d *DB) Following(ctx context.Context, me string) ([]models.Following, error) {
	res, err := d.raw.QueryWithParams(ctx,
		`SELECT tgt.id, tgt.name FROM users src
		 JOIN MATCH (src)-[r:FOLLOWS]->(tgt)
		 WHERE src.id = $1 ORDER BY tgt.name`,
		libravdb.QueryParams{"1": me})
	if err != nil {
		return nil, err
	}
	var out []models.Following
	for _, r := range res.Results {
		out = append(out, models.Following{
			ID:   str(r.Metadata["id"]),
			Name: str(r.Metadata["name"]),
		})
	}
	return out, nil
}

// Followers returns follower counts for every user — the graph
// inverted, for the homepage leaderboard.
func (d *DB) Followers(ctx context.Context) ([]models.Follower, error) {
	res, err := d.raw.Query(ctx,
		`SELECT tgt.id, tgt.name, COUNT(*) AS followers FROM users src
		 JOIN MATCH (src)-[r:FOLLOWS]->(tgt)
		 GROUP BY tgt.id, tgt.name ORDER BY followers DESC`)
	if err != nil {
		return nil, err
	}
	var out []models.Follower
	for _, r := range res.Results {
		out = append(out, models.Follower{
			ID:        str(r.Metadata["id"]),
			Name:      str(r.Metadata["name"]),
			Followers: atoi(r.Metadata["followers"]),
		})
	}
	return out, nil
}

// Suggest returns users followed by the people I follow (2 hops on
// the FOLLOWS graph), ranked by mutual connections. A single
// chained JOIN MATCH traversal with projections and filtering.
// The handler excludes people already followed.
func (d *DB) Suggest(ctx context.Context, me string) ([]models.Suggestion, error) {
	res, err := d.raw.QueryWithParams(ctx,
		`SELECT tgt.id, tgt.name, COUNT(*) AS mutual FROM users me
		 JOIN MATCH (me)-[f1:FOLLOWS]->(mid)
		 JOIN MATCH (mid)-[f2:FOLLOWS]->(tgt)
		 WHERE me.id = $1 AND tgt.id <> $1
		 GROUP BY tgt.id, tgt.name ORDER BY mutual DESC`,
		libravdb.QueryParams{"1": me})
	if err != nil {
		return nil, err
	}
	var out []models.Suggestion
	for _, r := range res.Results {
		out = append(out, models.Suggestion{
			ID:     str(r.Metadata["id"]),
			Name:   str(r.Metadata["name"]),
			Mutual: atoi(r.Metadata["mutual"]),
		})
	}
	return out, nil
}

// GetCounter returns the durable homepage click count.
func (d *DB) GetCounter(ctx context.Context) (int, error) {
	res, err := d.raw.Query(ctx, "SELECT value FROM counter WHERE id = 'home'")
	if err != nil {
		return 0, err
	}
	if len(res.Results) == 0 {
		return 0, nil
	}
	return atoi(res.Results[0].Metadata["value"]), nil
}

// IncrementCounter bumps the durable click count and returns the
// new value. Read-modify-write with a parameterized literal —
// arithmetic in UPDATE SET is not supported yet in libraVDB.
func (d *DB) IncrementCounter(ctx context.Context) (int, error) {
	n, err := d.GetCounter(ctx)
	if err != nil {
		return 0, err
	}
	n++
	if _, err := d.raw.QueryWithParams(ctx,
		"UPDATE counter SET value = $1 WHERE id = 'home'",
		libravdb.QueryParams{"1": n}); err != nil {
		return 0, err
	}
	return n, nil
}

// Count returns the number of users (homepage + dashboard stats).
func (d *DB) Count(ctx context.Context) (int, error) {
	res, err := d.raw.Query(ctx, "SELECT COUNT(*) AS n FROM users")
	if err != nil {
		return 0, err
	}
	if len(res.Results) == 0 {
		return 0, nil
	}
	return atoi(res.Results[0].Metadata["n"]), nil
}

// EdgeCount returns the number of FOLLOWS edges (dashboard stats).
// GRAPH_EDGES resolves for INSERT/DELETE; counting goes through
// the aggregate JOIN MATCH instead.
func (d *DB) EdgeCount(ctx context.Context) (int, error) {
	res, err := d.raw.Query(ctx,
		`SELECT COUNT(*) AS n FROM users src JOIN MATCH (src)-[r:FOLLOWS]->(tgt)`)
	if err != nil {
		return 0, err
	}
	if len(res.Results) == 0 {
		return 0, nil
	}
	return atoi(res.Results[0].Metadata["n"]), nil
}

// --- helpers ---

func scanUser(r *libravdb.SearchResult) (*models.User, error) {
	id, _ := r.Metadata["id"].(string)
	if id == "" {
		return nil, fmt.Errorf("malformed user row: %+v", r.Metadata)
	}
	return &models.User{
		ID:           id,
		Email:        str(r.Metadata["email"]),
		PasswordHash: str(r.Metadata["password_hash"]),
		Name:         str(r.Metadata["name"]),
		CreatedAt:    str(r.Metadata["created_at"]),
	}, nil
}

func str(v any) string {
	s, _ := v.(string)
	return s
}

func atoi(v any) int {
	switch n := v.(type) {
	case int:
		return n
	case int8:
		return int(n)
	case int16:
		return int(n)
	case int32:
		return int(n)
	case int64:
		return int(n)
	case uint:
		return int(n)
	case uint8:
		return int(n)
	case uint16:
		return int(n)
	case uint32:
		return int(n)
	case uint64:
		return int(n)
	case float32:
		return int(n)
	case float64:
		return int(n)
	case string:
		i, _ := strconv.Atoi(n)
		return i
	}
	// Unknown numeric type — fall back to rendering the value.
	if s := fmt.Sprint(v); s != "" && s != "<nil>" {
		if i, err := strconv.Atoi(s); err == nil {
			return i
		}
	}
	return 0
}

// --- dashboard surfaces ---

// Todos returns the todos table rows, optionally filtered by a
// title substring (LIKE).
func (d *DB) Todos(ctx context.Context, filter string) ([]models.Todo, error) {
	sql := `SELECT id, title, completed, priority, due_at, tags FROM todos`
	params := libravdb.QueryParams{}
	if filter != "" {
		sql += ` WHERE title LIKE $1 ORDER BY priority, id`
		params["1"] = "%" + filter + "%"
	} else {
		sql += ` ORDER BY priority, id`
	}
	res, err := d.raw.QueryWithParams(ctx, sql, params)
	if err != nil {
		return nil, err
	}
	var out []models.Todo
	for _, r := range res.Results {
		out = append(out, models.Todo{
			ID:        str(r.Metadata["id"]),
			Title:     str(r.Metadata["title"]),
			Completed: r.Metadata["completed"] == true,
			Priority:  atoi(r.Metadata["priority"]),
			DueAt:     str(r.Metadata["due_at"]),
			Tags:      fmt.Sprint(r.Metadata["tags"]),
		})
	}
	return out, nil
}

// SaveTodo inserts a new todo row (parameterized CRUD).
func (d *DB) SaveTodo(ctx context.Context, t *models.Todo) error {
	_, err := d.raw.QueryWithParams(ctx,
		`INSERT INTO todos (id, title, completed, priority, due_at, tags)
		 VALUES ($1, $2, $3, $4, $5, $6)`,
		libravdb.QueryParams{
			"1": t.ID, "2": t.Title, "3": t.Completed, "4": t.Priority,
			"5": t.DueAt, "6": t.Tags,
		})
	return err
}

// ToggleTodo flips a todo's completed flag — a single
// parameterized UPDATE with a unary NOT in the SET expression.
func (d *DB) ToggleTodo(ctx context.Context, id string) error {
	_, err := d.raw.QueryWithParams(ctx,
		`UPDATE todos SET completed = NOT completed WHERE id = $1`,
		libravdb.QueryParams{"1": id})
	return err
}

// DeleteTodo removes a todo row.
func (d *DB) DeleteTodo(ctx context.Context, id string) error {
	_, err := d.raw.QueryWithParams(ctx,
		`DELETE FROM todos WHERE id = $1`,
		libravdb.QueryParams{"1": id})
	return err
}

// TodoCount returns the number of todos (overview stat).
func (d *DB) TodoCount(ctx context.Context) (int, error) {
	res, err := d.raw.Query(ctx, "SELECT COUNT(*) AS n FROM todos")
	if err != nil {
		return 0, err
	}
	if len(res.Results) == 0 {
		return 0, nil
	}
	return atoi(res.Results[0].Metadata["n"]), nil
}

// PriorityBars returns todos per priority — the overview chart
// (GROUP BY aggregate).
func (d *DB) PriorityBars(ctx context.Context) ([]models.Bar, error) {
	res, err := d.raw.Query(ctx,
		`SELECT priority, COUNT(*) AS n FROM todos GROUP BY priority ORDER BY priority`)
	if err != nil {
		return nil, err
	}
	var out []models.Bar
	for _, r := range res.Results {
		out = append(out, models.Bar{
			Label: "P" + str(r.Metadata["priority"]),
			Count: atoi(r.Metadata["n"]),
		})
	}
	return out, nil
}

// Versions runs a temporal VERSIONS OF query over a table for the
// given time range and returns the version history.
func (d *DB) Versions(ctx context.Context, table, start, end string) ([]models.Version, error) {
	res, err := d.raw.Query(ctx, fmt.Sprintf(
		`SELECT id, version, title, completed, version_start, version_end
		 FROM VERSIONS OF %s BETWEEN TIMESTAMP '%s' AND TIMESTAMP '%s'
		 ORDER BY id, version`, table, start, end))
	if err != nil {
		return nil, err
	}
	var out []models.Version
	for _, r := range res.Results {
		out = append(out, models.Version{
			ID:           str(r.Metadata["id"]),
			Version:      atoi(r.Metadata["version"]),
			Title:        str(r.Metadata["title"]),
			Completed:    r.Metadata["completed"] == true,
			// version_start/end come back as time.Time values —
			// fmt.Sprint renders them.
			VersionStart: fmt.Sprint(r.Metadata["version_start"]),
			VersionEnd:   fmt.Sprint(r.Metadata["version_end"]),
		})
	}
	return out, nil
}

// RunSQL executes arbitrary SQL and returns a renderable result:
// columns + stringified rows, or the error text.
func (d *DB) RunSQL(ctx context.Context, sql string) ([]string, [][]string, error) {
	res, err := d.raw.Query(ctx, sql)
	if err != nil {
		return nil, nil, err
	}
	// Column order from the result metadata.
	var cols []string
	for _, c := range res.Columns {
		cols = append(cols, c)
	}
	if len(cols) == 0 {
		// Fall back to sorted metadata keys when the executor
		// didn't report columns.
		seen := map[string]bool{}
		for _, r := range res.Results {
			for k := range r.Metadata {
				if !seen[k] {
					seen[k] = true
					cols = append(cols, k)
				}
			}
		}
		sort.Strings(cols)
	}
	var rows [][]string
	for _, r := range res.Results {
		var row []string
		for _, c := range cols {
			row = append(row, fmt.Sprint(r.Metadata[c]))
		}
		rows = append(rows, row)
	}
	return cols, rows, nil
}

// Edges returns every FOLLOWS edge (graph page).
func (d *DB) Edges(ctx context.Context) ([]models.Edge, error) {
	res, err := d.raw.Query(ctx,
		`SELECT src.id, tgt.id FROM users src JOIN MATCH (src)-[r:FOLLOWS]->(tgt)`)
	if err != nil {
		return nil, err
	}
	var out []models.Edge
	for _, r := range res.Results {
		i := strings.LastIndexByte(r.ID, '|')
		if i <= 0 || i+1 >= len(r.ID) {
			continue
		}
		out = append(out, models.Edge{From: r.ID[:i], To: r.ID[i+1:]})
	}
	return out, nil
}

// RawQuery exposes the raw result for debugging.
func (d *DB) RawQuery(ctx context.Context, sql string) (*libravdb.SearchResults, error) {
	return d.raw.Query(ctx, sql)
}
