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
	Dash      *DashData // authenticated dashboard pages
	Error     *ErrorData // HTTP error page
	Signup    *SignupData
	Profile   *ProfileData
}

// SignupData feeds the registration page.
type SignupData struct {
	Error string
	Name  string
	Email string
}

// ProfileData feeds the profile page (self or another user).
type ProfileData struct {
	User           *User
	IsSelf         bool
	Followed       bool
	FollowingCount int
	FollowerCount  int
	Following      []Following
}

// ErrorData feeds the error page (404/500).
type ErrorData struct {
	Code    int
	Title   string
	Message string
}

// HomeData feeds the public homepage.
type HomeData struct {
	UserCount   int
	FollowerTop []Follower // top users by follower count (graph query)
	Feature     string     // value hydrated into the Alpine demo widget
	Counter     int        // durable click counter (libraVDB)
}

// LoginData feeds the sign-in page. Error carries the flash banner
// from a failed POST (re-rendered at 200); Email keeps the typed
// value so the user doesn't retype it.
type LoginData struct {
	Error string
	Email string
}

// DashData feeds the authenticated dashboard pages. Active names
// the sidebar item: overview|editor|sql|temporal|graph|settings.
type DashData struct {
	User        *User
	Active      string
	Stats       Stats
	Following   []Following // users I follow (graph traversal)
	Suggestions []Suggestion

	// Table editor
	Todos      []Todo
	TodoStat   string // feedback line after a mutation
	TodoFilter string

	// Overview chart
	MaxBar int

	// SQL editor
	SQLText    string
	SQLColumns []string
	SQLRows    [][]string
	SQLError   string

	// Temporal
	Versions []Version
	TemporalTable string
	TemporalStart string
	TemporalEnd   string

	// Graph
	Edges      []Edge
	GraphMsg   string // flash notice on the graph page
	GraphMsgOk bool   // notice is a success (green)

	// Overview
	Bars []Bar
}

// Stats are the dashboard stat cards.
type Stats struct {
	Users       int
	Following   int
	Suggestions int
	Todos       int
	Clicks      int
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

// Todo is a row in the todos table — the table editor's CRUD
// subject (raw relational surface).
type Todo struct {
	ID        string
	Title     string
	Completed bool
	Priority  int
	DueAt     string
	Tags      string
}

// Version is one row of a VERSIONS OF temporal query.
type Version struct {
	ID           string
	Version      int
	Title        string
	Completed    bool
	VersionStart string
	VersionEnd   string
}

// Edge is one directed FOLLOWS edge (graph page).
type Edge struct {
	From, To string
}

// Bar is one bar of the overview chart (GROUP BY result).
type Bar struct {
	Label string
	Count int
}

// Follower is one row of the homepage leaderboard.
type Follower struct {
	ID        string
	Name      string
	Followers int
}
