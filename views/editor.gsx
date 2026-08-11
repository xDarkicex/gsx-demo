@import "github.com/xDarkicex/gsx-demo/models"

// Editor is the Table Editor — raw relational CRUD against the
// todos table: filter, insert, toggle, delete. Mutations are plain
// POST handlers (PRG) exercising libraVDB's SQL CRUD surface.
func Editor(p models.Page) {
    <h1 class="uk-h2">Table Editor</h1>
    <p class="muted">todos — <code>libraVDB</code> relational CRUD</p>

    @if p.Dash.TodoStat != "" {
        <div class="uk-alert uk-alert-success">{p.Dash.TodoStat}</div>
    }

    <div class="uk-grid uk-grid-small" uk-grid>
        <div class="uk-width-expand">
            <form method="get" action="/dashboard/editor" class="uk-form-stacked">
                <input class="uk-input" type="text" name="q" placeholder="Filter by title…"
                    value={p.Dash.TodoFilter} />
                <button class="uk-button uk-button-default uk-margin-small-left">Filter</button>
            </form>
        </div>
        <div class="uk-width-auto">
            <a class="uk-button uk-button-primary" href="#new-todo" uk-toggle>+ New todo</a>
        </div>
    </div>

    <div id="new-todo" hidden>
        <form method="post" action="/dashboard/editor/save" class="uk-card uk-card-default uk-card-body uk-margin-top">
            <div class="uk-grid-small uk-child-width-1-3@m" uk-grid>
                <div><label class="uk-form-label">Title</label>
                    <input class="uk-input" type="text" name="title" required /></div>
                <div><label class="uk-form-label">Priority</label>
                    <input class="uk-input" type="number" name="priority" value="3" min="1" max="5" /></div>
                <div><label class="uk-form-label">Due (YYYY-MM-DD HH:MM:SS)</label>
                    <input class="uk-input" type="text" name="due_at" placeholder="2026-08-15 12:00:00" /></div>
                <div class="uk-margin-small-top"><label class="uk-form-label">Tags (JSON)</label>
                    <input class="uk-input" type="text" name="tags" placeholder='["ops"]' /></div>
                <div class="uk-margin-small-top"><label class="uk-form-label">Completed</label>
                    <input class="uk-checkbox" type="checkbox" name="completed" value="true" /></div>
            </div>
            <button class="uk-button uk-button-primary uk-margin-top">Save</button>
        </form>
    </div>

    <div class="uk-overflow-auto uk-margin-top">
        <table class="uk-table uk-table-hover uk-table-divider uk-table-small">
            <thead>
                <tr>
                    <th>Title</th>
                    <th>Status</th>
                    <th>Priority</th>
                    <th>Due</th>
                    <th>Tags</th>
                    <th></th>
                </tr>
            </thead>
            <tbody>
                @for _, t := range p.Dash.Todos {
                    <tr>
                        <td>{t.Title}</td>
                        <td>
                            @if t.Completed {
                                <span class="uk-label uk-label-success">done</span>
                            } @else {
                                <span class="uk-label">open</span>
                            }
                        </td>
                        <td>P{t.Priority}</td>
                        <td class="muted">{t.DueAt}</td>
                        <td class="muted small">{t.Tags}</td>
                        <td>
                            <form method="post" action="/dashboard/editor/toggle" class="uk-display-inline">
                                <input type="hidden" name="id" value={t.ID} />
                                <button class="uk-button uk-button-small uk-button-default">Toggle</button>
                            </form>
                            <form method="post" action="/dashboard/editor/delete" class="uk-display-inline">
                                <input type="hidden" name="id" value={t.ID} />
                                <button class="uk-button uk-button-small uk-button-danger">Delete</button>
                            </form>
                        </td>
                    </tr>
                }
                @if len(p.Dash.Todos) == 0 {
                    <tr><td colspan="6" class="muted">No todos match.</td></tr>
                }
            </tbody>
        </table>
    </div>
}
