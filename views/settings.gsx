@import "github.com/xDarkicex/gsx-demo/models"

// SettingsPage is a minimal project settings page.
func SettingsPage(p models.Page) {
    <h1 class="uk-h2">Settings</h1>
    <p class="muted">project demo</p>

    <div class="uk-card uk-card-default uk-card-body uk-width-1-2@m">
        <dl class="uk-description-list">
            <dt>Project</dt><dd>demo</dd>
            <dt>Signed in as</dt><dd>{p.User.Name} &#64;{p.User.Email}</dd>
            <dt>Stack</dt><dd>nanite · nanite-render · nanite-gsx · libraVDB</dd>
            <dt>Session cookie</dt><dd><code>gsx_session</code> (HttpOnly, 7 days)</dd>
        </dl>
        <form method="post" action="/logout">
            <button class="uk-button uk-button-danger">Log out</button>
        </form>
    </div>
}
