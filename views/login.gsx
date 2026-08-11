@import "github.com/xDarkicex/gsx-demo/models"

// Login is the public sign-in page. A failed POST re-renders this
// view at 200 with the error banner populated from page data.
func Login(p models.Page) {
    <div class="auth-shell">
        <div class="auth-card">
            <div class="card-heading">
                <div>
                    <div class="card-kicker">Welcome back</div>
                    <h2>Sign in to gsx-studio</h2>
                </div>
            </div>
            <p class="page-subtitle">
                Demo account: <code>alice&#64;demo.dev</code> / <code>demo123</code>
            </p>

            @if p.Login.Error != "" {
                <div class="notice notice-error"><span class="notice-icon">✕</span>{p.Login.Error}</div>
            }

            <form method="post" action="/login">
                <div class="field">
                    <label class="uk-form-label" for="login-email">Email</label>
                    <input id="login-email" class="uk-input" type="email" name="email"
                        placeholder="you@example.com" value={p.Login.Email} autofocus />
                </div>
                <div class="field">
                    <label class="uk-form-label" for="login-password">Password</label>
                    <input id="login-password" class="uk-input" type="password" name="password"
                        placeholder="••••••••" />
                </div>
                <div class="form-actions">
                    <button type="submit" class="uk-button uk-button-primary">Sign in</button>
                </div>
            </form>

            <p class="muted small">
                bcrypt against libraVDB · signed session cookie ·
                nanite middleware guards /dashboard
            </p>
        </div>
    </div>
}
