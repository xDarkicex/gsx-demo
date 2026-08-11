@import "github.com/xDarkicex/gsx-demo/models"

// TemporalPage explores libraVDB's temporal SQL: VERSIONS OF a
// table between two timestamps, showing every version of every
// row with its validity window.
func TemporalPage(p models.Page) {
    <h1 class="uk-h2">Temporal</h1>
    <p class="muted">
        <code>VERSIONS OF todos BETWEEN TIMESTAMP … AND TIMESTAMP …</code> —
        every version of every row, with <code>version</code>,
        <code>version_start</code>, <code>version_end</code>.
    </p>

    <form method="get" action="/dashboard/temporal" class="uk-grid-small" uk-grid>
        <div class="uk-width-1-4@m">
            <label class="uk-form-label">Start (RFC3339)</label>
            <input class="uk-input uk-font-mono" type="text" name="start" value={p.Dash.TemporalStart} />
        </div>
        <div class="uk-width-1-4@m">
            <label class="uk-form-label">End (RFC3339)</label>
            <input class="uk-input uk-font-mono" type="text" name="end" value={p.Dash.TemporalEnd} />
        </div>
        <div class="uk-width-1-4@m">
            <label class="uk-form-label">Table</label>
            <input class="uk-input uk-font-mono" type="text" name="table" value={p.Dash.TemporalTable} />
        </div>
        <div class="uk-width-auto uk-flex uk-flex-middle">
            <button class="uk-button uk-button-primary">Query</button>
        </div>
    </form>
    <p class="muted small">
        Tip: the range start must be within retention — query a range
        starting after the first write, or the executor errors with
        <code>retention expired</code>.
    </p>

    @if p.Dash.SQLError != "" {
        <div class="uk-alert uk-alert-danger uk-margin-top">
            <pre class="uk-margin-remove">{p.Dash.SQLError}</pre>
        </div>
    }

    @if len(p.Dash.Versions) > 0 {
        <div class="uk-overflow-auto uk-margin-top">
            <table class="uk-table uk-table-hover uk-table-divider uk-table-small">
                <thead>
                    <tr>
                        <th>id</th><th>version</th><th>title</th><th>completed</th>
                        <th>version_start</th><th>version_end</th>
                    </tr>
                </thead>
                <tbody>
                    @for _, v := range p.Dash.Versions {
                        <tr>
                            <td>{v.ID}</td>
                            <td>{v.Version}</td>
                            <td>{v.Title}</td>
                            <td>
                                @if v.Completed {
                                    <span class="uk-label uk-label-success">done</span>
                                } @else {
                                    <span class="uk-label">open</span>
                                }
                            </td>
                            <td class="muted small">{v.VersionStart}</td>
                            <td class="muted small">{v.VersionEnd}</td>
                        </tr>
                    }
                </tbody>
            </table>
        </div>
    } @else {
        @if p.Dash.SQLError == "" {
            <p class="muted uk-margin-top">No versions in range — try a range that overlaps the rows' lifetimes.</p>
        }
    }
}
