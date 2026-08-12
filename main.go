// Command gsx-demo runs the full nanite stack showcase:
//
//	nanite router  → routes + auth middleware
//	nanite-render  → page pipeline (layout + view), components,
//	                 colocated actions, async/OOB, memoize
//	nanite-gsx     → every template is a .gsx component
//	libraVDB       → relational users + FOLLOWS graph, all via SQL
//
// Pages: / (public), /login (public), /dashboard (auth).
// Demo account: alice@demo.dev / demo123.
package main

import (
	"context"
	"crypto/rand"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/xDarkicex/nanite"
	"github.com/xDarkicex/nanite-gsx"
	"github.com/xDarkicex/nanite-render"
	"github.com/xDarkicex/nanite-render/nano"

	"github.com/xDarkicex/gsx-demo/internal/auth"
	"github.com/xDarkicex/gsx-demo/internal/db"
	"github.com/xDarkicex/gsx-demo/models"
	"github.com/xDarkicex/gsx-demo/views"
)

func main() {
	ctx := context.Background()

	// 1. libraVDB — relational + graph, seeded on first boot.
	d, err := db.Open("data")
	if err != nil {
		log.Fatalf("open libraVDB: %v", err)
	}
	defer d.Close()
	if err := d.Seed(ctx); err != nil {
		log.Fatalf("seed: %v", err)
	}

	// 2. nanite-gsx engine — every view is a .gsx component.
	gsxEngine := gsx.New()
	views.RegisterAppLayout(gsxEngine)
	views.RegisterNavbar(gsxEngine)
	views.RegisterHome(gsxEngine)
	views.RegisterLogin(gsxEngine)
	views.RegisterStatCard(gsxEngine)
	views.RegisterLiveClock(gsxEngine)
	views.RegisterClockSkeleton(gsxEngine)
	views.RegisterFollowButton(gsxEngine)
	views.RegisterCounterWidget(gsxEngine)
	views.RegisterDashLayout(gsxEngine)
	views.RegisterOverview(gsxEngine)
	views.RegisterEditor(gsxEngine)
	views.RegisterSQLPage(gsxEngine)
	views.RegisterTemporalPage(gsxEngine)
	views.RegisterGraphPage(gsxEngine)
	views.RegisterSettingsPage(gsxEngine)
	views.RegisterErrorPage(gsxEngine)
	views.RegisterTodoTable(gsxEngine)
	views.RegisterTodoRow(gsxEngine)
	views.RegisterSQLResults(gsxEngine)
	views.RegisterEdgeList(gsxEngine)
	views.RegisterEdgeRow(gsxEngine)

	// 3. ComponentRegistry — decorated components dispatched by
	// name: @action (Follow), @async/@fallback/@oob (LiveClock),
	// @memo (FollowButton).
	cr := render.NewComponentRegistry()
	views.RegisterFollowButtonComponent(cr)
	views.RegisterLiveClockComponent(cr)
	views.RegisterCounterWidgetComponent(cr)
	views.RegisterTodoTableComponent(cr)
	views.RegisterTodoRowComponent(cr)
	views.RegisterEdgeListComponent(cr)
	views.RegisterEdgeRowComponent(cr)
	reg := render.New(
		render.WithEngines(gsxEngine),
		// gsx compiles to direct Go calls — the source bytes are
		// never parsed at runtime, so the loader is a stub.
		render.WithDefaultLoader(func(name string) ([]byte, error) { return nil, nil }),
	)
	reg.AttachComponents(cr)

	// 4. Router.
	r := nanite.New(nanite.WithConfig(nanite.Config{
		RecoverPanics: true,
		NotFoundHandler: func(c *nanite.Context) {
			renderError(reg, c, http.StatusNotFound,
				"Not Found", "The page you're looking for doesn't exist.")
		},
	}))
	r.ServeStatic("/static", "./public")

	r.Get("/", home(reg))
	r.Get("/login", loginPage(reg))
	r.Post("/login", loginPost(reg))
	r.Post("/logout", logout)
	r.Get("/widgets/clock", clockWidget(reg))
	r.Get("/dashboard/partial/following", followingPartial(reg))

	// The dashboard group is guarded by session middleware.
	dash := r.Group("/dashboard", auth.RequireUser)
	dash.Get("/", overview(reg))
	dash.Get("/editor", editor(reg))
	dash.Get("/editor/table", editorTable(reg)) // HTMX search partial
	dash.Get("/sql", sqlPage(reg))
	dash.Post("/sql/run", sqlRun(reg))
	dash.Get("/temporal", temporal(reg))
	dash.Get("/graph", graph(reg))
	dash.Get("/settings", settings(reg))

	// Colocated server actions (@action in .gsx) — one mount point.
	r.Post("/_nano/action/*", func(c *nanite.Context) {
		reg.HandleAction(c.Writer, c.Request)
	})

	// Graceful shutdown: SIGINT/SIGTERM → router Shutdown runs the
	// shutdown hooks, which close libraVDB cleanly — the SQL
	// catalog and graph WAL flush to disk so clicks persist.
	r.AddShutdownHook(func() error { return d.Close() })
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sig
		_ = r.Shutdown(5 * time.Second)
	}()

	log.Printf("gsx-demo: http://localhost:3000 (alice@demo.dev / demo123)")
	if err := r.Start("3000"); err != nil {
		log.Fatal(err)
	}
}

// fail logs the error and renders the error page at 500.
func fail(reg *render.Registry, c *nanite.Context, err error) {
	log.Printf("handler error: %v", err)
	renderError(reg, c, http.StatusInternalServerError,
		"Internal Server Error", err.Error())
}

// renderError renders the ErrorPage view at the given status.
// The status is set first — writing it after the body would leave
// the default 200 on the wire.
func renderError(reg *render.Registry, c *nanite.Context, status int, title, msg string) {
	c.Status(status)
	page := models.Page{User: auth.CurrentUser(c), Error: &models.ErrorData{
		Code: status, Title: title, Message: msg,
	}}
	if err := nano.Page(reg, c).
		Engine(render.CustomEngine("gsx")).
		Layout("AppLayout").
		View("ErrorPage").
		With(page).
		Render(); err != nil {
		c.String(status, title)
	}
}

// renderPage renders a view inside the gsx AppLayout.
func renderPage(reg *render.Registry, c *nanite.Context, view string, data any) error {
	return nano.Page(reg, c).
		Engine(render.CustomEngine("gsx")).
		Layout("AppLayout").
		View(view).
		With(data).
		Render()
}

func home(reg *render.Registry) nanite.HandlerFunc {
	return func(c *nanite.Context) {
		n, err := db.Default.Count(c.Request.Context())
		if err != nil {
			fail(reg, c, err)
			return
		}
		top, err := db.Default.Followers(c.Request.Context())
		if err != nil {
			fail(reg, c, err)
			return
		}
		clicks, err := db.Default.GetCounter(c.Request.Context())
		if err != nil {
			fail(reg, c, err)
			return
		}
		page := models.Page{
			User: auth.CurrentUser(c),
			Home: &models.HomeData{
				UserCount:   n,
				FollowerTop: top,
				Feature:     "server-hydrated",
				Counter:     clicks,
			},
		}
		if err := renderPage(reg, c, "Home", page); err != nil {
			fail(reg, c, err)
		}
	}
}

func loginPage(reg *render.Registry) nanite.HandlerFunc {
	return func(c *nanite.Context) {
		page := models.Page{User: auth.CurrentUser(c), Login: &models.LoginData{}}
		if err := renderPage(reg, c, "Login", page); err != nil {
			fail(reg, c, err)
		}
	}
}

// loginPost checks credentials against libraVDB, issues a session
// cookie, and redirects. Failures re-render the page at 200 with
// the flash banner populated (the value survives in LoginData).
func loginPost(reg *render.Registry) nanite.HandlerFunc {
	return func(c *nanite.Context) {
		email := c.FormValue("email")
		password := c.FormValue("password")

		u, err := db.Default.UserByEmail(c.Request.Context(), email)
		if err != nil || u == nil || !auth.CheckPassword(u.PasswordHash, password) {
			page := models.Page{
				Login: &models.LoginData{Error: "Invalid email or password.", Email: email},
			}
			bw := render.AcquireWriter(c.Writer)
			defer render.ReleaseWriter(bw)
			rc := render.AcquireContext(bw, c.Request)
			rc.Loader = reg.DefaultLoader()
			defer render.ReleaseContext(rc)
			if err := reg.Page(rc).Engine(render.CustomEngine("gsx")).
				Layout("AppLayout").View("Login").With(page).Render(); err != nil {
				fail(reg, c, err)
			}
			return
		}

		tok, err := auth.Default.Create(u)
		if err != nil {
			fail(reg, c, err)
			return
		}
		http.SetCookie(c.Writer, &http.Cookie{
			Name:     auth.SessionCookie,
			Value:    tok,
			Path:     "/",
			HttpOnly: true,
			MaxAge:   int(auth.SessionTTL.Seconds()),
		})
		c.Redirect(http.StatusFound, "/dashboard")
	}
}

func logout(c *nanite.Context) {
	if cookie, err := c.Request.Cookie(auth.SessionCookie); err == nil {
		auth.Default.Delete(cookie.Value)
	}
	http.SetCookie(c.Writer, &http.Cookie{
		Name: auth.SessionCookie, Value: "", Path: "/", HttpOnly: true, MaxAge: -1,
	})
	c.Redirect(http.StatusFound, "/")
}

// renderDash renders a page inside the DashLayout sidebar shell.
func renderDash(reg *render.Registry, c *nanite.Context, view string, data any) error {
	return nano.Page(reg, c).
		Engine(render.CustomEngine("gsx")).
		Layout("DashLayout").
		View(view).
		With(data).
		Render()
}

// dashBase returns the authenticated page skeleton with the nav
// item active.
func dashBase(c *nanite.Context, active string) models.Page {
	return models.Page{
		User: auth.CurrentUser(c),
		Dash: &models.DashData{Active: active},
	}
}

// overview renders the dashboard home: stat cards, the priority
// chart, and the live clock.
func overview(reg *render.Registry) nanite.HandlerFunc {
	return func(c *nanite.Context) {
		ctx := c.Request.Context()
		u := auth.CurrentUser(c)

		users, err := db.Default.Count(ctx)
		if err != nil {
			fail(reg, c, err)
			return
		}
		edges, err := db.Default.EdgeCount(ctx)
		if err != nil {
			fail(reg, c, err)
			return
		}
		todos, err := db.Default.TodoCount(ctx)
		if err != nil {
			fail(reg, c, err)
			return
		}
		clicks, err := db.Default.GetCounter(ctx)
		if err != nil {
			fail(reg, c, err)
			return
		}
		bars, err := db.Default.PriorityBars(ctx)
		if err != nil {
			fail(reg, c, err)
			return
		}
		following, err := db.Default.Following(ctx, u.ID)
		if err != nil {
			fail(reg, c, err)
			return
		}
		maxBar := 0
		for _, b := range bars {
			if b.Count > maxBar {
				maxBar = b.Count
			}
		}

		page := dashBase(c, "overview")
		page.Dash.Stats = models.Stats{
			Users: users, Following: edges, Todos: todos, Clicks: clicks,
		}
		page.Dash.Bars = bars
		page.Dash.MaxBar = maxBar
		page.Dash.Following = following
		if err := renderDash(reg, c, "Overview", page); err != nil {
			fail(reg, c, err)
		}
	}
}

// editor renders the Table Editor (todos CRUD).
func editor(reg *render.Registry) nanite.HandlerFunc {
	return func(c *nanite.Context) {
		ctx := c.Request.Context()
		filter := c.Query("q")
		todos, err := db.Default.Todos(ctx, filter)
		if err != nil {
			fail(reg, c, err)
			return
		}
		page := dashBase(c, "editor")
		page.Dash.Todos = todos
		page.Dash.TodoFilter = filter
		switch c.Query("notice") {
		case "created":
			page.Dash.TodoStat = "Todo created."
		case "updated":
			page.Dash.TodoStat = "Todo status updated."
		case "deleted":
			page.Dash.TodoStat = "Todo deleted."
		}
		if err := renderDash(reg, c, "Editor", page); err != nil {
			fail(reg, c, err)
		}
	}
}

// editorTable renders just the TodoTable — the debounced search
// input swaps it in place via HTMX.
func editorTable(reg *render.Registry) nanite.HandlerFunc {
	return func(c *nanite.Context) {
		filter := c.Query("q")
		todos, err := db.Default.Todos(c.Request.Context(), filter)
		if err != nil {
			fail(reg, c, err)
			return
		}
		if err := nano.Render(c, reg, "gsx", "TodoTable",
			map[string]any{"todos": todos, "filter": filter}); err != nil {
			fail(reg, c, err)
		}
	}
}

// editorSave inserts a todo (PRG — mutate, redirect back).
func editorSave(reg *render.Registry) nanite.HandlerFunc {
	return func(c *nanite.Context) {
		t := &models.Todo{
			ID:        "todo-" + randHex(6),
			Title:     c.FormValue("title"),
			Priority:  parsePriority(c.FormValue("priority")),
			DueAt:     c.FormValue("due_at"),
			Tags:      c.FormValue("tags"),
			Completed: c.FormValue("completed") == "true",
		}
		if err := db.Default.SaveTodo(c.Request.Context(), t); err != nil {
			fail(reg, c, err)
			return
		}
		editorRedirect(c, "created")
	}
}

// editorToggle flips a todo's completed flag.
func editorToggle(reg *render.Registry) nanite.HandlerFunc {
	return func(c *nanite.Context) {
		if err := db.Default.ToggleTodo(c.Request.Context(), c.FormValue("id")); err != nil {
			fail(reg, c, err)
			return
		}
		editorRedirect(c, "updated")
	}
}

// editorDelete removes a todo.
func editorDelete(reg *render.Registry) nanite.HandlerFunc {
	return func(c *nanite.Context) {
		if err := db.Default.DeleteTodo(c.Request.Context(), c.FormValue("id")); err != nil {
			fail(reg, c, err)
			return
		}
		editorRedirect(c, "deleted")
	}
}

func editorRedirect(c *nanite.Context, notice string) {
	location := "/dashboard/editor?notice=" + url.QueryEscape(notice)
	if filter := c.FormValue("q"); filter != "" {
		location += "&q=" + url.QueryEscape(filter)
	}
	c.Redirect(http.StatusFound, location)
}

// sqlPage renders the SQL editor (results from the last run).
func sqlPage(reg *render.Registry) nanite.HandlerFunc {
	return func(c *nanite.Context) {
		page := dashBase(c, "sql")
		page.Dash.SQLText = "SELECT id, title, completed, priority FROM todos ORDER BY priority"
		if err := renderDash(reg, c, "SQLPage", page); err != nil {
			fail(reg, c, err)
		}
	}
}

// sqlRun executes the submitted SQL and swaps in the results
// region (HTMX partial).
func sqlRun(reg *render.Registry) nanite.HandlerFunc {
	return func(c *nanite.Context) {
		text := c.FormValue("sql")
		cols, rows, err := db.Default.RunSQL(c.Request.Context(), text)
		page := dashBase(c, "sql")
		page.Dash.SQLText = text
		if err != nil {
			page.Dash.SQLError = err.Error()
		} else {
			page.Dash.SQLColumns = cols
			page.Dash.SQLRows = rows
		}
		if err := nano.Render(c, reg, "gsx", "SQLResults",
			map[string]any{"error": page.Dash.SQLError,
				"columns": page.Dash.SQLColumns,
				"rows": page.Dash.SQLRows, "text": page.Dash.SQLText}); err != nil {
			fail(reg, c, err)
		}
	}
}

// temporal renders the VERSIONS OF explorer.
func temporal(reg *render.Registry) nanite.HandlerFunc {
	return func(c *nanite.Context) {
		ctx := c.Request.Context()
		table := c.Query("table")
		if table == "" {
			table = "todos"
		}
		start := c.Query("start")
		end := c.Query("end")
		if start == "" {
			// Default: from boot (the first write — the retention
			// floor) to an hour from now.
			start = db.Default.BootTime.Add(-time.Second).Format(time.RFC3339)
		}
		if end == "" {
			end = time.Now().UTC().Add(time.Hour).Format(time.RFC3339)
		}
		var versions []models.Version
		var terr error
		if table != "" {
			versions, terr = db.Default.Versions(ctx, table, start, end)
		}
		page := dashBase(c, "temporal")
		page.Dash.TemporalTable = table
		page.Dash.TemporalStart = start
		page.Dash.TemporalEnd = end
		page.Dash.Versions = versions
		if terr != nil {
			page.Dash.SQLError = terr.Error()
		}
		if err := renderDash(reg, c, "TemporalPage", page); err != nil {
			fail(reg, c, err)
		}
	}
}

// graph renders the FOLLOWS graph explorer.
func graph(reg *render.Registry) nanite.HandlerFunc {
	return func(c *nanite.Context) {
		ctx := c.Request.Context()
		u := auth.CurrentUser(c)
		edges, err := db.Default.Edges(ctx)
		if err != nil {
			fail(reg, c, err)
			return
		}
		following, err := db.Default.Following(ctx, u.ID)
		if err != nil {
			fail(reg, c, err)
			return
		}
		followingSet := make(map[string]bool, len(following))
		for _, f := range following {
			followingSet[f.ID] = true
		}
		suggestions, err := db.Default.Suggest(ctx, u.ID)
		if err != nil {
			fail(reg, c, err)
			return
		}
		var fresh []models.Suggestion
		for _, s := range suggestions {
			if !followingSet[s.ID] {
				fresh = append(fresh, s)
			}
		}
		page := dashBase(c, "graph")
		page.Dash.Edges = edges
		page.Dash.Following = following
		page.Dash.Suggestions = fresh
		page.Dash.GraphMsg = c.Query("msg")
		page.Dash.GraphMsgOk = c.Query("ok") == "1"
		if err := renderDash(reg, c, "GraphPage", page); err != nil {
			fail(reg, c, err)
		}
	}
}

// settings renders the project settings page.
func settings(reg *render.Registry) nanite.HandlerFunc {
	return func(c *nanite.Context) {
		page := dashBase(c, "settings")
		if err := renderDash(reg, c, "SettingsPage", page); err != nil {
			fail(reg, c, err)
		}
	}
}

// oldDashboard is kept for reference — the new pages above replace it.
func dashboard(reg *render.Registry) nanite.HandlerFunc {
	return func(c *nanite.Context) {
		u := auth.CurrentUser(c)
		ctx := c.Request.Context()

		following, err := db.Default.Following(ctx, u.ID)
		if err != nil {
			fail(reg, c, err)
			return
		}
		followingSet := make(map[string]bool, len(following))
		for _, f := range following {
			followingSet[f.ID] = true
		}
		suggestions, err := db.Default.Suggest(ctx, u.ID)
		if err != nil {
			fail(reg, c, err)
			return
		}
		var fresh []models.Suggestion
		for _, s := range suggestions {
			if !followingSet[s.ID] {
				fresh = append(fresh, s)
			}
		}
		n, err := db.Default.Count(ctx)
		if err != nil {
			fail(reg, c, err)
			return
		}
		edges, err := db.Default.EdgeCount(ctx)
		if err != nil {
			fail(reg, c, err)
			return
		}

		page := models.Page{
			User: u,
			Dashboard: &models.DashData{
				User:      u,
				Following: following,
				Stats: models.Stats{
					Users:       n,
					Following:   edges,
					Suggestions: len(fresh),
				},
				Suggestions: fresh,
			},
		}
		if err := renderPage(reg, c, "Dashboard", page); err != nil {
			fail(reg, c, err)
		}
	}
}

// clockWidget dispatches LiveClock through the ComponentRegistry —
// the @async worker forks, the skeleton streams, the real HTML
// arrives as an HTMX OOB swap.
func clockWidget(reg *render.Registry) nanite.HandlerFunc {
	return func(c *nanite.Context) {
		bw := render.AcquireWriter(c.Writer)
		defer render.ReleaseWriter(bw)
		rc := render.AcquireContext(bw, c.Request)
		rc.Loader = reg.DefaultLoader()
		defer render.ReleaseContext(rc)
		if err := reg.RenderComponent(bw, rc, "LiveClock", nil); err != nil {
			fail(reg, c, err)
			return
		}
		// CloseSuspense tells the coordinator no more workers are
		// coming; it drains asynchronously. Hold the handler open
		// so the worker's trailing OOB chunk hits the wire before
		// the response completes — the skeleton streams first,
		// the real clock swaps in via hx-swap-oob.
		rc.CloseSuspense()
		time.Sleep(200 * time.Millisecond)
	}
}

// followingPartial re-renders the Following card for HTMX refresh.
func followingPartial(reg *render.Registry) nanite.HandlerFunc {
	return func(c *nanite.Context) {
		u := auth.UserFromRequest(c.Request)
		if u == nil {
			c.Redirect(http.StatusFound, "/login")
			return
		}
		following, err := db.Default.Following(c.Request.Context(), u.ID)
		if err != nil {
			fail(reg, c, err)
			return
		}
		page := models.Page{User: u, Dashboard: &models.DashData{Following: following}}
		if err := nano.Render(c, reg, "gsx", "Following", page); err != nil {
			fail(reg, c, err)
		}
	}
}

// randHex returns n random hex chars (for todo ids).
func randHex(n int) string {
	const hex = "0123456789abcdef"
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "000000"
	}
	for i := range b {
		b[i] = hex[b[i]&0xf]
	}
	return string(b)
}

// parsePriority coerces a form priority to 1..5, defaulting to 3.
func parsePriority(s string) int {
	p := 3
	if n, err := strconv.Atoi(s); err == nil && n >= 1 && n <= 5 {
		p = n
	}
	return p
}
