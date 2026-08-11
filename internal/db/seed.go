package db

import (
	"context"
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
