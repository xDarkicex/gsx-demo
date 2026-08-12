@import "github.com/xDarkicex/gsx-demo/models"

// ErrorPage renders an HTTP error (404/500) as a centered card
// with the status code.
func ErrorPage(p models.Page) {
    <div class="auth-shell">
        <div class="auth-card error-card">
            <div class="auth-brand">
                <span class="auth-logo">&lt;/&gt;</span>
                <span class="auth-brand-name">gsx-studio</span>
            </div>
            <div class="error-code">{p.Error.Code}</div>
            <h1 class="auth-title">{p.Error.Title}</h1>
            <p class="auth-subtitle">{p.Error.Message}</p>
            <div class="form-actions">
                <a class="uk-button uk-button-primary" href="/dashboard">Back to dashboard</a>
            </div>
        </div>
    </div>
}
