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
	"strconv"
	"strings"

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
	d := &DB{raw: raw, gr: gr}
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
		return int(n)
	case int64:
		return int(n)
	case float64:
		return int(n)
	case string:
		i, _ := strconv.Atoi(n)
		return i
	}
	return 0
}
