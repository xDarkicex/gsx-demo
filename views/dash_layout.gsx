@import "github.com/xDarkicex/gsx-demo/models"

// DashLayout is the authenticated dashboard shell — a supabase-
// style sidebar with grouped navigation, the active page at
// @yield. It is a complete document: head with the framework
// stylesheets (getuikit, Alpine, HTMX) and the sidebar shell.
func DashLayout(p models.Page) {
    <html lang="en">
        <head>
            <meta charset="utf-8" />
            <meta name="viewport" content="width=device-width, initial-scale=1" />
            <title>gsx-studio — dashboard</title>
            <link rel="preconnect" href="https://fonts.googleapis.com">
            <link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
            <link href="https://fonts.googleapis.com/css2?family=Outfit:wght@300;400;500;600;700;800&display=swap" rel="stylesheet">
            <link rel="stylesheet" href="/static/vendor/uikit.min.css" />
            <link rel="stylesheet" href="/static/app.css?v=4" />
            <script defer src="/static/vendor/uikit.min.js"></script>
            <script defer src="/static/vendor/alpine.min.js"></script>
            <script defer src="/static/vendor/htmx.min.js"></script>
        </head>
        <body class="dash-body">
            <div class="dash-shell">
                <aside class="dash-sidebar">
                    <div class="dash-brand-row">
                        <span class="brand-mark">g</span>
                        <div>
                            <div class="dash-brand">gsx-studio</div>
                            <div class="sidebar-caption">data workspace</div>
                        </div>
                    </div>

                    <div class="workspace-switcher">
                        <div class="sidebar-caption">Workspace</div>
                        <div class="workspace-name">Demo project <span class="workspace-chevron">⌄</span></div>
                        <div class="workspace-meta"><span class="status-dot"></span> Local environment</div>
                    </div>

                    <nav class="sidebar-nav" aria-label="Dashboard navigation">
                        <div class="nav-section-label">Workspace</div>
                        <a class={sidebarNavActive(p.Dash.Active, "overview")} href="/dashboard">
                            <span class="nav-icon">⌂</span><span>Overview</span>
                        </a>
                        <div class="nav-section-label nav-section-spaced">Build</div>
                        <a class={sidebarNavActive(p.Dash.Active, "editor")} href="/dashboard/editor">
                            <span class="nav-icon">▦</span><span>Table editor</span>
                        </a>
                        <a class={sidebarNavActive(p.Dash.Active, "sql")} href="/dashboard/sql">
                            <span class="nav-icon nav-icon-text">SQL</span><span>SQL editor</span>
                        </a>
                        <a class={sidebarNavActive(p.Dash.Active, "temporal")} href="/dashboard/temporal">
                            <span class="nav-icon">◷</span><span>Temporal</span>
                        </a>
                        <a class={sidebarNavActive(p.Dash.Active, "graph")} href="/dashboard/graph">
                            <span class="nav-icon">⌘</span><span>Graph</span>
                        </a>
                        <div class="nav-section-label nav-section-spaced">Account</div>
                        <a class={sidebarNavActive(p.Dash.Active, "settings")} href="/dashboard/settings">
                            <span class="nav-icon">⚙</span><span>Settings</span>
                        </a>
                    </nav>

                    <div class="sidebar-footer">
                        <div class="profile-row">
                            <span class="profile-avatar">{userInitial(p.User.Name)}</span>
                            <div class="profile-copy">
                                <a href="/profile"><strong>{p.User.Name}</strong></a>
                                <span>{p.User.Email}</span>
                            </div>
                        </div>
                        <form method="post" action="/logout">
                            <button class="logout-button" type="submit"><span>↪</span> Log out</button>
                        </form>
                    </div>
                </aside>
                <main class="dash-main">
                    <header class="dash-topbar">
                        <div class="topbar-context"><span class="topbar-kicker">GSX / DEMO</span><span class="topbar-separator">/</span><span>Dashboard</span></div>
                        <div class="topbar-status"><span class="status-dot"></span> All systems operational</div>
                    </header>
                    <div class="dashboard-content">
                        @yield
                    </div>
                </main>
            </div>
        </body>
    </html>
}
