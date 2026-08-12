@import "github.com/xDarkicex/gsx-demo/models"

// Home is the public landing page. The hero CTAs switch on the
// session (signed in → open the dashboard), the stats come from
// libraVDB, and the sections give the stack its hierarchy.
func Home(p models.Page) {
    <div class="home-hero">
        <div class="home-hero-inner">
            <div class="eyebrow">The nanite stack</div>
            <h1 class="home-title">Server-rendered apps.<br />Compiled, not interpreted.</h1>
            <p class="home-lead">
                React's mental model, Go's runtime. Components in <code>.gsx</code> compile to
                direct Go calls — <code>0 B/op</code> on the hot path, no virtual DOM, no
                JavaScript build step.
            </p>
            <div class="home-ctas">
                @if p.User != nil {
                    <a class="uk-button uk-button-primary home-cta" href="/dashboard">Open dashboard</a>
                    <a class="uk-button uk-button-secondary home-cta" href="/profile">View profile</a>
                } @else {
                    <a class="uk-button uk-button-primary home-cta" href="/login">Sign in</a>
                    <a class="uk-button uk-button-secondary home-cta" href="/signup">Create account</a>
                }
            </div>
            <p class="home-meta muted small">demo account — <code>alice&#64;demo.dev</code> / <code>demo123</code></p>
        </div>
    </div>

    <div class="home-stats">
        <div class="home-stat"><div class="home-stat-value">{p.Home.UserCount}</div><div class="home-stat-label">users</div></div>
        <div class="home-stat"><div class="home-stat-value">{len(p.Home.FollowerTop)}</div><div class="home-stat-label">graph nodes</div></div>
        <div class="home-stat"><div class="home-stat-value">{p.Home.TodoCount}</div><div class="home-stat-label">todos</div></div>
        <div class="home-stat"><div class="home-stat-value">{p.Home.ClickCount}</div><div class="home-stat-label">clicks</div></div>
    </div>

    <div class="home-features">
        <div class="feature-card">
            <div class="feature-icon">▦</div>
            <h3>Relational SQL</h3>
            <p class="muted small">Tables, constraints, aggregates — libraVDB's native SQL surface, exercised by the table editor.</p>
        </div>
        <div class="feature-card">
            <div class="feature-icon">↗</div>
            <h3>Graph traversals</h3>
            <p class="muted small">FOLLOWS edges via <code>GRAPH_EDGES</code>, 2-hop suggestions via chained <code>JOIN MATCH</code>.</p>
        </div>
        <div class="feature-card">
            <div class="feature-icon">◷</div>
            <h3>Temporal history</h3>
            <p class="muted small"><code>VERSIONS OF</code> and <code>AS OF</code> — every row's history, queryable.</p>
        </div>
        <div class="feature-card">
            <div class="feature-icon">⚡</div>
            <h3>Server actions</h3>
            <p class="muted small">Colocated <code>@action</code>s mutate state and re-render — HTMX swaps the page in place.</p>
        </div>
        <div class="feature-card">
            <div class="feature-icon">∞</div>
            <h3>Async streaming</h3>
            <p class="muted small">Skeletons stream instantly, workers render in the background, OOB swaps land the result.</p>
        </div>
        <div class="feature-card">
            <div class="feature-icon">✦</div>
            <h3>Hydrated Alpine</h3>
            <p class="muted small">Go state serialized to the client via <code>WriteHydrateProps</code> — zero-alloc JSON.</p>
        </div>
    </div>

    <div class="home-grid">
        <CounterWidget props={map[string]any{"count": p.Home.Counter, "feature": p.Home.Feature}} />

        <section class="dashboard-card">
            <div class="card-heading">
                <div>
                    <div class="card-kicker">Social graph</div>
                    <h2>Leaderboard</h2>
                </div>
                <span class="record-count">{len(p.Home.FollowerTop)}</span>
            </div>
            <p class="card-description">Top users by follower count — libraVDB <code>JOIN MATCH</code>.</p>
            <ol class="clean-list leaderboard">
                @for _, f := range p.Home.FollowerTop {
                    <li>
                        <span class="list-avatar">{userInitial(f.Name)}</span>
                        <span class="list-copy">
                            <strong>{f.Name}</strong>
                            <span><a href={"/profile/" + f.ID}>&#64;{f.ID}</a> · {f.Followers} followers</span>
                        </span>
                    </li>
                }
            </ol>
        </section>
    </div>
}
