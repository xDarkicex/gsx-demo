@import "github.com/xDarkicex/gsx-demo/models"

// SignupPage is the public registration page — a compact auth card
// matching the login. Successful signups auto-login and land on
// the new profile page.
func SignupPage(p models.Page) {
    <div class="auth-shell">
        <div class="auth-card">
            <div class="auth-brand">
                <span class="auth-logo">&lt;/&gt;</span>
                <span class="auth-brand-name">gsx-studio</span>
            </div>

            <h1 class="auth-title">Create account</h1>
            <p class="auth-subtitle">Join the demo project</p>

            @if p.Signup.Error != "" {
                <div class="notice notice-error"><span class="notice-icon">✕</span>{p.Signup.Error}</div>
            }

            <form method="post" action="/signup" class="auth-form">
                <div class="field">
                    <label class="uk-form-label" for="su-name">Name</label>
                    <input id="su-name" class="uk-input" type="text" name="name" placeholder="Ada Lovelace"
                        value={p.Signup.Name} autofocus />
                </div>
                <div class="field">
                    <label class="uk-form-label" for="su-email">Email</label>
                    <input id="su-email" class="uk-input" type="email" name="email" placeholder="you@example.com"
                        value={p.Signup.Email} />
                </div>
                <div class="field">
                    <label class="uk-form-label" for="su-password">Password</label>
                    <input id="su-password" class="uk-input" type="password" name="password" placeholder="at least 6 characters" />
                </div>
                <button type="submit" class="uk-button uk-button-primary auth-submit">Create account</button>
            </form>

            <div class="auth-footer">
                <span class="muted small">Already have an account?</span>
                <a href="/login" class="small">Sign in</a>
            </div>
        </div>
    </div>
}
