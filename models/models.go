// Package models defines the shared data types handed to every
// .gsx view. A single Page struct flows through the layout + view
// pipeline (nano.Page renders both with the same data).
package models

// User is a row in the users collection (libraVDB, relational).
type User struct {
	ID           string
	Name         string
	Email        string
	PasswordHash string
	CreatedAt    string
}

// Page is the one type every view accepts. Each page handler fills
// only the sub-struct for its route; views nil-check before use.
type Page struct {
	User      *User // set on every render; nil when anonymous
	Home      *HomeData
	Login     *LoginData
	Dashboard *DashData
}

// HomeData feeds the public homepage.
type HomeData struct {
	UserCount   int
	FollowerTop []Follower // top users by follower count (graph query)
	Feature     string     // value hydrated into the Alpine demo widget
}

// LoginData feeds the sign-in page. Error carries the flash banner
// from a failed POST (re-rendered at 200); Email keeps the typed
// value so the user doesn't retype it.
type LoginData struct {
	Error string
	Email string
}

// DashData feeds the authenticated dashboard.
type DashData struct {
	User        *User
	Stats       Stats
	Following   []Following // users I follow (graph traversal)
	Suggestions []Suggestion
}

// Stats are the memoized dashboard stat cards.
type Stats struct {
	Users      int
	Following  int
	Suggestions int
}

// Following is one row of the "people I follow" feed.
type Following struct {
	ID   string
	Name string
}

// Suggestion is a person the current user might want to follow —
// derived from the FOLLOWS graph (mutual connections).
type Suggestion struct {
	ID       string
	Name     string
	Mutual   int
	Followed bool
}

// Follower is one row of the homepage leaderboard.
type Follower struct {
	ID        string
	Name      string
	Followers int
}
