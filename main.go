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
	"log"
	"net/http"
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
	d, err := db.Open()
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
	views.RegisterDashboard(gsxEngine)
	views.RegisterStatCard(gsxEngine)
	views.RegisterLiveClock(gsxEngine)
	views.RegisterClockSkeleton(gsxEngine)
	views.RegisterFollowing(gsxEngine)
	views.RegisterSuggestedUsers(gsxEngine)
	views.RegisterFollowButton(gsxEngine)

	// 3. ComponentRegistry — decorated components dispatched by
	// name: @action (Follow), @async/@fallback/@oob (LiveClock),
	// @memo (FollowButton).
	cr := render.NewComponentRegistry()
	views.RegisterFollowButtonComponent(cr)
	views.RegisterLiveClockComponent(cr)
	reg := render.New(
		render.WithEngines(gsxEngine),
		// gsx compiles to direct Go calls — the source bytes are
		// never parsed at runtime, so the loader is a stub.
		render.WithDefaultLoader(func(name string) ([]byte, error) { return nil, nil }),
	)
	reg.AttachComponents(cr)

	// 4. Router.
	r := nanite.New(nanite.WithPanicRecovery(true))
	r.ServeStatic("/static", "./public")

	r.Get("/", home(reg))
	r.Get("/login", loginPage(reg))
	r.Post("/login", loginPost(reg))
	r.Post("/logout", logout)
	r.Get("/widgets/clock", clockWidget(reg))
	r.Get("/dashboard/partial/following", followingPartial(reg))

	// The dashboard group is guarded by session middleware.
	dash := r.Group("/dashboard", auth.RequireUser)
	dash.Get("/", dashboard(reg))

	// Colocated server actions (@action in .gsx) — one mount point.
	r.Post("/_nano/action/*", func(c *nanite.Context) {
		reg.HandleAction(c.Writer, c.Request)
	})

	log.Printf("gsx-demo: http://localhost:3000 (alice@demo.dev / demo123)")
	if err := r.Start("3000"); err != nil {
		log.Fatal(err)
	}
}

// fail logs the error and writes a plain 500. c.Error alone only
// stores the error; without error middleware nothing is written.
func fail(c *nanite.Context, err error) {
	log.Printf("handler error: %v", err)
	c.String(http.StatusInternalServerError, "Internal Server Error")
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
			fail(c, err)
			return
		}
		top, err := db.Default.Followers(c.Request.Context())
		if err != nil {
			fail(c, err)
			return
		}
		page := models.Page{
			User: auth.CurrentUser(c),
			Home: &models.HomeData{
				UserCount:   n,
				FollowerTop: top,
				Feature:     "server-hydrated",
			},
		}
		if err := renderPage(reg, c, "Home", page); err != nil {
			fail(c, err)
		}
	}
}

func loginPage(reg *render.Registry) nanite.HandlerFunc {
	return func(c *nanite.Context) {
		page := models.Page{User: auth.CurrentUser(c), Login: &models.LoginData{}}
		if err := renderPage(reg, c, "Login", page); err != nil {
			fail(c, err)
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
				fail(c, err)
			}
			return
		}

		tok, err := auth.Default.Create(u)
		if err != nil {
			fail(c, err)
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

// dashboard builds the authenticated view: graph-derived stats,
// the follow graph, and 2-hop suggestions.
func dashboard(reg *render.Registry) nanite.HandlerFunc {
	return func(c *nanite.Context) {
		u := auth.CurrentUser(c)
		ctx := c.Request.Context()

		following, err := db.Default.Following(ctx, u.ID)
		if err != nil {
			fail(c, err)
			return
		}
		followingSet := make(map[string]bool, len(following))
		for _, f := range following {
			followingSet[f.ID] = true
		}
		suggestions, err := db.Default.Suggest(ctx, u.ID)
		if err != nil {
			fail(c, err)
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
			fail(c, err)
			return
		}
		edges, err := db.Default.EdgeCount(ctx)
		if err != nil {
			fail(c, err)
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
			fail(c, err)
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
			fail(c, err)
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
			fail(c, err)
			return
		}
		page := models.Page{User: u, Dashboard: &models.DashData{Following: following}}
		if err := nano.Render(c, reg, "gsx", "Following", page); err != nil {
			fail(c, err)
		}
	}
}
