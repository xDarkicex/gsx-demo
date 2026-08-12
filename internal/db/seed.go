package db

import (
	"context"
	"strings"
	"time"

	"github.com/xDarkicex/gsx-demo/internal/auth"
	"github.com/xDarkicex/gsx-demo/models"
)

// Seed populates the demo with users + a FOLLOWS graph the first
// time it runs. Every seeded account uses password "demo123".
func (d *DB) Seed(ctx context.Context) error {
	n, err := d.Count(ctx)
	if err != nil {
		return err
	}
	if n > 0 {
		return nil // already seeded
	}

	hash, err := auth.HashPassword("demo123")
	if err != nil {
		return err
	}
	now := time.Now().UTC().Format(time.RFC3339)

	users := []models.User{
		{ID: "alice", Name: "Alice Nguyen", Email: "alice@demo.dev", PasswordHash: hash, CreatedAt: now},
		{ID: "bob", Name: "Bob Tanaka", Email: "bob@demo.dev", PasswordHash: hash, CreatedAt: now},
		{ID: "carol", Name: "Carol Okafor", Email: "carol@demo.dev", PasswordHash: hash, CreatedAt: now},
		{ID: "dave", Name: "Dave Kowalski", Email: "dave@demo.dev", PasswordHash: hash, CreatedAt: now},
		{ID: "eve", Name: "Eve Lindqvist", Email: "eve@demo.dev", PasswordHash: hash, CreatedAt: now},
		{ID: "frank", Name: "Frank Moreau", Email: "frank@demo.dev", PasswordHash: hash, CreatedAt: now},
	}
	for _, u := range users {
		if err := d.CreateUser(ctx, &u); err != nil {
			return err
		}
	}

	// The todos table — the table editor's CRUD subject.
	if _, err := d.raw.Query(ctx, `CREATE TABLE todos (
		id TEXT PRIMARY KEY, title TEXT NOT NULL, completed BOOLEAN DEFAULT false,
		priority INTEGER DEFAULT 3, opened_at TIMESTAMP, due_at TIMESTAMP, tags JSON)`); err != nil {
		if !strings.Contains(err.Error(), "exists") {
			return err
		}
	}
	today := time.Now().UTC()
	todos := []models.Todo{
		{ID: "todo-1", Title: "Write the dashboard shell", Completed: true, Priority: 1, OpenedAt: today.AddDate(0, 0, -3).Format("2006-01-02 15:04:05"), DueAt: today.AddDate(0, 0, -1).Format("2006-01-02 15:04:05"), Tags: `["dashboard","gsx"]`},
		{ID: "todo-2", Title: "Ship table editor CRUD", Completed: false, Priority: 1, DueAt: today.AddDate(0, 0, 1).Format("2006-01-02 15:04:05"), Tags: `["editor"]`},
		{ID: "todo-3", Title: "Test temporal VERSIONS OF", Completed: false, Priority: 2, DueAt: today.AddDate(0, 0, 2).Format("2006-01-02 15:04:05"), Tags: `["temporal","sql"]`},
		{ID: "todo-4", Title: "Graph traversal page", Completed: false, Priority: 2, DueAt: today.AddDate(0, 0, 3).Format("2006-01-02 15:04:05"), Tags: `["graph"]`},
		{ID: "todo-5", Title: "Polish the overview charts", Completed: true, Priority: 3, DueAt: today.AddDate(0, 0, -2).Format("2006-01-02 15:04:05"), Tags: `["ui"]`},
		{ID: "todo-6", Title: "Write bug-bounty findings", Completed: false, Priority: 3, DueAt: "", Tags: `["docs"]`},
	}
	for _, t := range todos {
		if err := d.SaveTodo(ctx, &t); err != nil {
			return err
		}
	}

	// FOLLOWS edges — a small social graph for the dashboard
	// traversal demos.
	edges := [][2]string{
		{"alice", "bob"}, {"alice", "carol"}, {"alice", "dave"},
		{"bob", "dave"}, {"bob", "eve"},
		{"carol", "eve"}, {"carol", "frank"},
		{"dave", "frank"},
		{"eve", "frank"},
	}
	for _, e := range edges {
		if err := d.Follow(ctx, e[0], e[1]); err != nil {
			return err
		}
	}
	return nil
}
