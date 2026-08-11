// Package db wraps libraVDB for the demo. All data access goes
// through the native SQL surface (db.Query / QueryWithParams) —
// relational tables plus the FOLLOWS graph via GRAPH_EDGES and
// JOIN MATCH. No vectors, no collections API after migration.
package db

import (
	"context"
	"errors"
	"fmt"
	"sort"
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

// Open creates the demo database. Storage is in-memory: the graph
// layer is process-local and can't be re-attached to a reopened
// collection, and the seed is deterministic anyway — every boot is
// a fresh, identical world.
func Open() (*DB, error) {
	raw, err := libravdb.Open(libravdb.WithStoragePath(":memory:demo"))
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

// migrate creates the graph-backed users collection. Called once
// per boot on the fresh in-memory database.
func (d *DB) migrate(ctx context.Context) error {
	// Register the FOLLOWS edge kind (process-wide, idempotent).
	libravdb.RegisterEdgeKind("FOLLOWS", 1)

	_, err := d.raw.CreateCollection(ctx, "users",
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
		"SELECT id, email, password_hash, name, created_at FROM users WHERE email = $email",
		libravdb.QueryParams{"email": email})
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
		"SELECT id, email, password_hash, name, created_at FROM users WHERE id = $id",
		libravdb.QueryParams{"id": id})
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
		"INSERT INTO users (id, email, password_hash, name, created_at) VALUES ($id, $email, $hash, $name, $created)",
		libravdb.QueryParams{
			"id": u.ID, "email": u.Email, "hash": u.PasswordHash,
			"name": u.Name, "created": u.CreatedAt,
		})
	return err
}

// Edge is one directed FOLLOWS edge.
type Edge struct{ From, To string }


// edges returns every FOLLOWS edge via the graph JOIN MATCH. The
// executor doesn't land projections for non-aggregate matches, so
// the composite ID ("src|tgt") carries the endpoints.
func (d *DB) edges(ctx context.Context) ([]Edge, error) {
	res, err := d.raw.Query(ctx,
		`SELECT src.id, tgt.id FROM users src JOIN MATCH (src)-[r:FOLLOWS]->(tgt)`)
	if err != nil {
		return nil, err
	}
	out := make([]Edge, 0, len(res.Results))
	for _, r := range res.Results {
		i := strings.LastIndexByte(r.ID, '|')
		if i <= 0 || i+1 >= len(r.ID) {
			continue
		}
		out = append(out, Edge{From: r.ID[:i], To: r.ID[i+1:]})
	}
	return out, nil
}

// Follow adds a FOLLOWS edge between two users. GRAPH_EDGES INSERTs
// take string literals (not bound params), so the ids are inlined
// with single-quote escaping. Ids are app-generated (xid/seed).
func (d *DB) Follow(ctx context.Context, me, target string) error {
	if me == target {
		return errors.New("cannot follow yourself")
	}
	_, err := d.raw.Query(ctx, fmt.Sprintf(
		"INSERT INTO GRAPH_EDGES VALUES ('%s','FOLLOWS','%s')",
		escapeLit(me), escapeLit(target)))
	return err
}

// Unfollow removes a FOLLOWS edge. SQL has no DELETE for
// GRAPH_EDGES (only INSERT resolves), so the graph's Go API does
// the removal: node ids are record ordinals + 1.
func (d *DB) Unfollow(ctx context.Context, me, target string) error {
	src, err := d.nodeID(ctx, me)
	if err != nil {
		return err
	}
	tgt, err := d.nodeID(ctx, target)
	if err != nil {
		return err
	}
	txn := d.gr.BeginTxn()
	if err := d.gr.RemoveEdge(txn, src, tgt, 1); err != nil {
		// Idempotent: an already-removed edge is success.
		return nil
	}
	return txn.Commit(ctx)
}

// nodeID maps a record id to its graph node id (ordinal + 1).
func (d *DB) nodeID(ctx context.Context, id string) (uint64, error) {
	res, err := d.raw.QueryWithParams(ctx,
		"SELECT id FROM users WHERE id = $id",
		libravdb.QueryParams{"id": id})
	if err != nil || len(res.Results) == 0 {
		return 0, fmt.Errorf("user %q not found", id)
	}
	return uint64(res.Results[0].Ordinal) + 1, nil
}

// escapeLit escapes a single-quoted SQL string literal.
func escapeLit(s string) string {
	return strings.ReplaceAll(s, "'", "''")
}

// Following returns the users the given user follows — the graph
// traversal (JOIN MATCH) with names resolved relationally.
func (d *DB) Following(ctx context.Context, me string) ([]models.Following, error) {
	es, err := d.edges(ctx)
	if err != nil {
		return nil, err
	}
	var out []models.Following
	for _, e := range es {
		if e.From != me {
			continue
		}
		u, err := d.UserByID(ctx, e.To)
		if err != nil || u == nil {
			continue
		}
		out = append(out, models.Following{ID: u.ID, Name: u.Name})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out, nil
}

// Followers returns follower counts for every user — the graph
// inverted, for the homepage leaderboard.
func (d *DB) Followers(ctx context.Context) ([]models.Follower, error) {
	es, err := d.edges(ctx)
	if err != nil {
		return nil, err
	}
	counts := make(map[string]int)
	for _, e := range es {
		counts[e.To]++
	}
	var out []models.Follower
	for id, n := range counts {
		u, err := d.UserByID(ctx, id)
		if err != nil || u == nil {
			continue
		}
		out = append(out, models.Follower{ID: u.ID, Name: u.Name, Followers: n})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Followers > out[j].Followers })
	return out, nil
}

// Suggest returns users followed by the people I follow (2 hops on
// the FOLLOWS graph), ranked by mutual connections, excluding people
// I already follow and myself.
func (d *DB) Suggest(ctx context.Context, me string) ([]models.Suggestion, error) {
	es, err := d.edges(ctx)
	if err != nil {
		return nil, err
	}
	following := make(map[string]bool)
	for _, e := range es {
		if e.From == me {
			following[e.To] = true
		}
	}
	mutual := make(map[string]int)
	for _, e := range es {
		if following[e.From] && e.To != me && !following[e.To] {
			mutual[e.To]++
		}
	}
	var out []models.Suggestion
	for id, n := range mutual {
		u, err := d.UserByID(ctx, id)
		if err != nil || u == nil {
			continue
		}
		out = append(out, models.Suggestion{ID: u.ID, Name: u.Name, Mutual: n})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Mutual > out[j].Mutual })
	return out, nil
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
func (d *DB) EdgeCount(ctx context.Context) (int, error) {
	es, err := d.edges(ctx)
	if err != nil {
		return 0, err
	}
	return len(es), nil
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

// Graph exposes the FOLLOWS graph for Go-level edge mutations.
func (d *DB) Graph() libravdb.Graph { return d.gr }
