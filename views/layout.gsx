@import "github.com/xDarkicex/gsx-demo/models"

// AppLayout is the single page shell: document head (Alpine + HTMX
// CDNs), navbar, the view injected at @yield, and the footer.
func AppLayout(p models.Page) {
    <html lang="en">
        <head>
            <meta charset="utf-8" />
            <meta name="viewport" content="width=device-width, initial-scale=1" />
            <title>gsx-demo — nanite · nanite-render · nanite-gsx · libraVDB</title>
            <link rel="stylesheet" href="/static/vendor/uikit.min.css" />
            <link rel="stylesheet" href="/static/app.css" />
            <script defer src="/static/vendor/uikit.min.js"></script>
            <script defer src="/static/vendor/alpine.min.js"></script>
            <script defer src="/static/vendor/htmx.min.js"></script>
        </head>
        <body>
            <Navbar page={p} />
            <main class="container">
                @yield
            </main>
            <footer class="footer">
                <p class="muted">gsx-demo · compiled to Go, rendered by nanite-render, stored in libraVDB</p>
            </footer>
        </body>
    </html>
}

// Navbar shows the session state. Alpine drives the dropdown menu;
// the user object arrives server-side via x-data hydration.
func Navbar(page models.Page) {
    <nav class="navbar" x-data="{open: false}">
        <a href="/" class="brand">&lt;/&gt;&nbsp;gsx-demo</a>
        <div class="nav-right">
            @if page.User != nil {
                <a class="nav-link muted" href="/profile">Signed in as {page.User.Name}</a>
                <a class="nav-link" href="/dashboard">Dashboard</a>
                <button class="btn btn-sm" @click="open = !open">
                    Menu ▾
                </button>
                <div class="nav-menu" x-show="open" x-cloak @click.away="open = false">
                    <form method="post" action="/logout">
                        <button type="submit" class="btn btn-sm btn-block">Log out</button>
                    </form>
                </div>
            } @else {
                <a class="nav-link" href="/login">Sign in</a>
            }
        </div>
    </nav>
}
