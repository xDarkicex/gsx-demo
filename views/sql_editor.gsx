@import "github.com/xDarkicex/gsx-demo/models"

// SQLPage is the SQL Editor — run arbitrary SQL against
// libraVDB. The Run button posts via HTMX and swaps the results
// region in place; the query stays in the textarea.
func SQLPage(p models.Page) {
    <h1 class="uk-h2">SQL Editor</h1>
    <p class="muted">run anything — <code>SELECT</code>, <code>INSERT</code>, <code>UPDATE</code>, <code>VERSIONS OF</code>…</p>

    <form hx-post="/dashboard/sql/run" hx-target="#sql-results" hx-swap="outerHTML" class="uk-margin-bottom">
        <textarea name="sql" rows="8" class="uk-textarea uk-font-mono"
            placeholder="SELECT id, title FROM todos">{p.Dash.SQLText}</textarea>
        <div class="uk-margin-top">
            <button class="uk-button uk-button-primary">Run</button>
        </div>
    </form>

    <div id="sql-results">
        <SQLResults props={map[string]any{
            "error": p.Dash.SQLError, "columns": p.Dash.SQLColumns,
            "rows": p.Dash.SQLRows, "text": p.Dash.SQLText,
        }} />
    </div>
}

// SQLResults is the results region: the error alert or the data
// grid. Swapped in place by the Run button.
func SQLResults(props map[string]any) {
    @if props["error"].(string) != "" {
        <div class="uk-alert uk-alert-danger">
            <pre class="uk-margin-remove">{props["error"].(string)}</pre>
        </div>
    }

    @if props["columns"] != nil {
        <div class="uk-card uk-card-default">
            <div class="uk-card-header">
                <div class="muted small">{len(props["rows"].([][]string))} rows</div>
            </div>
            <div class="uk-overflow-auto">
                <table class="uk-table uk-table-hover uk-table-divider uk-table-small">
                    <thead>
                        <tr>
                            @for _, c := range props["columns"].([]string) {
                                <th>{c}</th>
                            }
                        </tr>
                    </thead>
                    <tbody>
                        @for _, row := range props["rows"].([][]string) {
                            <tr>
                                @for _, cell := range row {
                                    <td class="small">{cell}</td>
                                }
                            </tr>
                        }
                        @if len(props["rows"].([][]string)) == 0 {
                            <tr><td colspan={len(props["columns"].([]string))} class="muted">Success. No rows returned.</td></tr>
                        }
                    </tbody>
                </table>
            </div>
        </div>
    }
}
