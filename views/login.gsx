@import "github.com/xDarkicex/gsx-demo/models"

// Login is the public sign-in page — a compact, boxy auth card in
// the style of a database dashboard. A failed POST re-renders this
// view at 200 with the error banner populated from page data.
func Login(p models.Page) {
    <div class="auth-shell">
        <div class="auth-card">
            <div class="auth-brand">
                <span class="auth-logo">&lt;/&gt;</span>
                <span class="auth-brand-name">gsx-studio</span>
            </div>

            <h1 class="auth-title">Sign in</h1>
            <p class="auth-subtitle">Access your project dashboard</p>

            @if p.Login.Error != "" {
                <div class="notice notice-error"><span class="notice-icon">✕</span>{p.Login.Error}</div>
            }

            <form method="post" action="/login" class="auth-form">
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
                <button type="submit" class="uk-button uk-button-primary auth-submit">Sign in</button>
            </form>

            <div class="auth-footer">
                <span class="muted small">Demo account</span>
                <code class="small">alice&#64;demo.dev / demo123</code>
            </div>
        </div>
    </div>
}
