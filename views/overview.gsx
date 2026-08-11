@import "github.com/xDarkicex/gsx-demo/models"

// Overview is the dashboard home — supabase-style stat cards, a
// chart from a GROUP BY aggregate, and the live clock.
func Overview(p models.Page) {
    <div class="page-heading">
        <div>
            <div class="eyebrow">Workspace overview</div>
            <h1>Good morning, {p.User.Name}</h1>
            <p class="page-subtitle">A quick read on your demo project and the work moving through it.</p>
        </div>
        <div class="heading-chip"><span class="status-dot"></span> Live data</div>
    </div>

    <div class="stat-grid">
        <StatCard label="Users" value={p.Dash.Stats.Users} icon="US" />
        <StatCard label="Follow edges" value={p.Dash.Stats.Following} icon="→" />
        <StatCard label="Todos" value={p.Dash.Stats.Todos} icon="✓" />
        <StatCard label="Clicks" value={p.Dash.Stats.Clicks} icon="⊙" />
    </div>

    <div class="overview-grid">
        <section class="dashboard-card chart-card">
            <div class="card-heading">
                <div>
                    <div class="card-kicker">Task health</div>
                    <h2>Todo priority distribution</h2>
                </div>
                <span class="card-icon">▥</span>
            </div>
            <p class="card-description">A live view of how work is distributed across priorities.</p>
            <div class="bars">
                @for _, b := range p.Dash.Bars {
                    <div class="bar-row">
                        <span class="bar-label muted">{b.Label}</span>
                        <div class="bar-track">
                            <div class="bar-fill" style={barWidth(b.Count, p.Dash.MaxBar)}></div>
                        </div>
                        <span class="bar-count">{b.Count}</span>
                    </div>
                }
                @if len(p.Dash.Bars) == 0 {
                    <p class="muted">No todos yet.</p>
                }
            </div>
        </section>
        <div>
            <div id="liveclock" hx-get="/widgets/clock" hx-trigger="load" hx-swap="innerHTML"></div>
            <section class="dashboard-card following-card">
                <div class="card-heading">
                    <div>
                        <div class="card-kicker">Social graph</div>
                        <h2>Recently followed</h2>
                    </div>
                    <span class="card-icon">↗</span>
                </div>
                <ul class="clean-list">
                    @for _, f := range p.Dash.Following {
                        <li><span class="list-avatar">{userInitial(f.Name)}</span><span class="list-copy"><strong>{f.Name}</strong><span>&#64;{f.ID}</span></span></li>
                    }
                    @if len(p.Dash.Following) == 0 {
                        <li class="empty-state">Nobody yet.</li>
                    }
                </ul>
            </section>
        </div>
    </div>
}
