@import "github.com/xDarkicex/gsx-demo/models"

// SQLPage is the SQL Editor — run arbitrary SQL against
// libraVDB, see the results grid or the error.
func SQLPage(p models.Page) {
    <h1 class="uk-h2">SQL Editor</h1>
    <p class="muted">run anything — <code>SELECT</code>, <code>INSERT</code>, <code>UPDATE</code>, <code>VERSIONS OF</code>…</p>

    <form method="post" action="/dashboard/sql/run" class="uk-margin-bottom">
        <textarea name="sql" rows="8" class="uk-textarea uk-font-mono"
            placeholder="SELECT id, title FROM todos">{p.Dash.SQLText}</textarea>
        <div class="uk-margin-top">
            <button class="uk-button uk-button-primary">Run</button>
        </div>
    </form>

    @if p.Dash.SQLError != "" {
        <div class="uk-alert uk-alert-danger">
            <pre class="uk-margin-remove">{p.Dash.SQLError}</pre>
        </div>
    }

    @if p.Dash.SQLColumns != nil {
        <div class="uk-card uk-card-default">
            <div class="uk-card-header">
                <div class="muted small">{len(p.Dash.SQLRows)} rows</div>
            </div>
            <div class="uk-overflow-auto">
                <table class="uk-table uk-table-hover uk-table-divider uk-table-small">
                    <thead>
                        <tr>
                            @for _, col := range p.Dash.SQLColumns {
                                <th>{col}</th>
                            }
                        </tr>
                    </thead>
                    <tbody>
                        @for _, row := range p.Dash.SQLRows {
                            <tr>
                                @for _, cell := range row {
                                    <td class="small">{cell}</td>
                                }
                            </tr>
                        }
                        @if len(p.Dash.SQLRows) == 0 {
                            <tr><td colspan={len(p.Dash.SQLColumns)} class="muted">Success. No rows returned.</td></tr>
                        }
                    </tbody>
                </table>
            </div>
        </div>
    }
}
