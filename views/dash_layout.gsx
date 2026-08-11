@import "github.com/xDarkicex/gsx-demo/models"

// DashLayout is the authenticated dashboard shell — a supabase-
// style sidebar with grouped navigation, the active page at
// @yield. The active item comes from p.Dash.Active.
func DashLayout(p models.Page) {
    <div class="uk-grid-collapse dash-shell" uk-grid>
        <aside class="dash-sidebar uk-width-1-4@m uk-width-1-1">
            <div class="uk-card uk-card-default uk-card-body">
                <div class="dash-brand">gsx-studio</div>
                <div class="muted small">project: demo</div>
                <hr class="uk-divider-small" />
                <ul class="uk-nav uk-nav-default">
                    <li class="uk-nav-header">Project</li>
                    <li class={navActive(p.Dash.Active, "overview")}>
                        <a href="/dashboard">Overview</a>
                    </li>
                    <li class="uk-nav-header">Tools</li>
                    <li class={navActive(p.Dash.Active, "editor")}>
                        <a href="/dashboard/editor">Table Editor</a>
                    </li>
                    <li class={navActive(p.Dash.Active, "sql")}>
                        <a href="/dashboard/sql">SQL Editor</a>
                    </li>
                    <li class={navActive(p.Dash.Active, "temporal")}>
                        <a href="/dashboard/temporal">Temporal</a>
                    </li>
                    <li class={navActive(p.Dash.Active, "graph")}>
                        <a href="/dashboard/graph">Graph</a>
                    </li>
                    <li class="uk-nav-header">Project</li>
                    <li class={navActive(p.Dash.Active, "settings")}>
                        <a href="/dashboard/settings">Settings</a>
                    </li>
                </ul>
                <hr class="uk-divider-small" />
                <div class="muted small">signed in as {p.User.Name}</div>
                <form method="post" action="/logout" class="uk-margin-small-top">
                    <button class="uk-button uk-button-small uk-button-default">Log out</button>
                </form>
            </div>
        </aside>
        <main class="uk-width-3-4@m dash-main">
            <div class="uk-container uk-container-large uk-padding">
                @yield
            </div>
        </main>
    </div>
}
