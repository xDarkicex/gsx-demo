@import "github.com/xDarkicex/gsx-demo/models"

// Login is the public sign-in page. A failed POST re-renders this
// view at 200 with the error banner populated from page data (the
// flash-error mechanism lives on the RenderContext — see the
// FollowButton action for the full @error flash loop).
func Login(p models.Page) {
    <section class="card login-card">
        <h1>Sign in</h1>
        <p class="muted">Demo account: <code>alice@demo.dev</code> / <code>demo123</code></p>

        @if p.Login.Error != "" {
            <p class="flash-error">{p.Login.Error}</p>
        }

        <form method="post" action="/login" class="form">
            <label>
                Email
                <input type="email" name="email" placeholder="you@example.com" value={p.Login.Email} autofocus />
            </label>
            <label>
                Password
                <input type="password" name="password" placeholder="••••••••" />
            </label>
            <button type="submit" class="btn btn-primary btn-block">Sign in</button>
        </form>

        <p class="muted small">
            Auth flow: bcrypt check against libraVDB, signed session
            cookie, nanite middleware guarding /dashboard.
        </p>
    </section>
}
