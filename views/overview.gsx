@import "github.com/xDarkicex/gsx-demo/models"

// Overview is the dashboard home — supabase-style stat cards, a
// chart from a GROUP BY aggregate, and the live clock.
func Overview(p models.Page) {
    <h1 class="uk-h2">Overview</h1>
    <p class="muted">project demo · nanite stack</p>

    <div class="uk-grid uk-child-width-1-4@m uk-grid-small">
        <StatCard label="Users" value={p.Dash.Stats.Users} icon="👥" />
        <StatCard label="Follow edges" value={p.Dash.Stats.Following} icon="→" />
        <StatCard label="Todos" value={p.Dash.Stats.Todos} icon="✅" />
        <StatCard label="Clicks" value={p.Dash.Stats.Clicks} icon="🖱️" />
    </div>

    <div class="uk-grid uk-child-width-1-2@m uk-grid-medium uk-margin-top">
        <div class="uk-card uk-card-default uk-card-body">
            <h3 class="uk-h4">Todo priority distribution</h3>
            <p class="muted small">GROUP BY priority — libraVDB aggregate</p>
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
        </div>
        <div>
            <div id="liveclock" hx-get="/widgets/clock" hx-trigger="load" hx-swap="innerHTML"></div>
            <div class="uk-card uk-card-default uk-card-body uk-margin-top">
                <h3 class="uk-h4">Recently followed</h3>
                <ul class="uk-list uk-list-divider">
                    @for _, f := range p.Dash.Following {
                        <li>{f.Name} <span class="muted">&#64;{f.ID}</span></li>
                    }
                    @if len(p.Dash.Following) == 0 {
                        <li class="muted">Nobody yet.</li>
                    }
                </ul>
            </div>
        </div>
    </div>
}
