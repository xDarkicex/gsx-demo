@import "github.com/xDarkicex/gsx-demo/models"
@import "time"

// Dashboard is the authenticated view. It demos:
//   - stat cards (plain composition)
//   - LiveClock: @async + @fallback + @oob — the component loads via
//     hx-get through the ComponentRegistry, streams a skeleton inline,
//     then the worker's HTML arrives as an HTMX OOB swap.
//   - Following list + SuggestedUsers, both driven by libraVDB graph
//     queries; the follow buttons are @action components.
func Dashboard(p models.Page) {
    <section class="cards">
        <StatCard label="Users" value={p.Dashboard.Stats.Users} icon="👥" />
        <StatCard label="Following" value={p.Dashboard.Stats.Following} icon="→" />
        <StatCard label="Suggestions" value={p.Dashboard.Stats.Suggestions} icon="✨" />
    </section>

    <div id="live-clock" hx-get="/widgets/clock" hx-trigger="load" hx-swap="innerHTML"></div>

    <div class="grid-2">
        <Following page={p} />
        <SuggestedUsers page={p} />
    </div>
}

// StatCard is a plain composable — props are typed parameters,
// compiled to direct function calls.
func StatCard(label string, value int, icon string) {
    <div class="card stat-card">
        <span class="stat-icon">{icon}</span>
        <div>
            <div class="stat-value">{value}</div>
            <div class="stat-label muted">{label}</div>
        </div>
    </div>
}

// LiveClock renders in a worker goroutine (Async); the finished
// bytes replace the skeleton via an OOB swap into #live-clock.
// WithOOB("live-clock") sets the swap target.
@oob "live-clock"
@async
func LiveClock() {
    <div class="card clock">
        <h2>Live clock <span class="muted">(async + OOB)</span></h2>
        <p class="clock-time">{time.Now().Format("15:04:05")}</p>
        <p class="muted small">
            Rendered in a worker goroutine after the page streamed —
            the skeleton appeared instantly, this HTML swapped in
            via <code>hx-swap-oob</code>.
        </p>
    </div>
}

// ClockSkeleton is the immediate fallback for LiveClock. The
// decorator names the component it falls back for.
@fallback(LiveClock)
func ClockSkeleton() {
    <div class="card clock">
        <h2>Live clock</h2>
        <p class="clock-time skeleton-line">Loading…</p>
    </div>
}

// Following lists the current user's FOLLOWS graph neighbors. It is
// also rendered as an HTMX partial (/dashboard/partial/following)
// for the refresh button.
func Following(page models.Page) {
    <section class="card">
        <h2>
            Following ({len(page.Dashboard.Following)})
            <button class="btn btn-sm"
                hx-get="/dashboard/partial/following"
                hx-target="#following"
                hx-swap="outerHTML">
                ⟳ Refresh
            </button>
        </h2>
        <ul class="people-list">
            @if len(page.Dashboard.Following) == 0 {
                <li class="muted">Nobody yet — try following someone.</li>
            }
            @for _, f := range page.Dashboard.Following {
                <li>
                    <strong>{f.Name}</strong>
                    <span class="muted">@{f.ID}</span>
                </li>
            }
        </ul>
    </section>
}

// SuggestedUsers is a 2-hop graph traversal: users followed by the
// people I follow, ranked by mutual connections. Each row has a
// FollowButton whose @action toggles the edge and re-renders inline.
func SuggestedUsers(page models.Page) {
    <section class="card">
        <h2>Who to follow</h2>
        <ul class="people-list">
            @if len(page.Dashboard.Suggestions) == 0 {
                <li class="muted">No suggestions — you've met everyone!</li>
            }
            @for _, s := range page.Dashboard.Suggestions {
                <li>
                    <div class="suggest-info">
                        <strong>{s.Name}</strong>
                        <span class="muted">@{s.ID} · {s.Mutual} mutual</span>
                    </div>
                    <FollowButton props={map[string]any{"target_id": s.ID, "followed": s.Followed}} />
                </li>
            }
        </ul>
    </section>
}
