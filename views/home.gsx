@import "github.com/xDarkicex/gsx-demo/models"

// Home is the public landing page. It demos server data (user count,
// graph-derived leaderboard) and an Alpine widget whose initial state
// is hydrated from Go via x-data auto-hydration.
func Home(p models.Page) {
    <section class="hero">
        <h1>The nanite stack, in one page</h1>
        <p class="lead">
            <code>nanite</code> router · <code>nanite-render</code> engine ·
            <code>nanite-gsx</code> templates · <code>libraVDB</code> relational + graph.
        </p>
        <div class="hero-actions">
            @if p.User != nil {
                <a class="btn btn-primary btn-lg" href="/dashboard">Open dashboard</a>
            } @else {
                <a class="btn btn-primary btn-lg" href="/login">Sign in</a>
            }
        </div>
        <p class="muted">{p.Home.UserCount} users in libraVDB</p>
    </section>

    <CounterWidget props={map[string]any{"count": p.Home.Counter, "feature": p.Home.Feature}} />

    <section class="card">
        <h2>Leaderboard — a graph query</h2>
        <p class="muted">
            Top users by follower count, computed with libraVDB
            <code>JOIN MATCH</code> over the FOLLOWS graph.
        </p>
        <ol class="leaderboard">
            @for _, f := range p.Home.FollowerTop {
                <li>
                    <strong>{f.Name}</strong>
                    <span class="muted">@{f.ID} · {f.Followers} followers</span>
                </li>
            }
        </ol>
    </section>
}
