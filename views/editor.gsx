@import "github.com/xDarkicex/gsx-demo/models"

// Editor is the Table Editor — raw relational CRUD against the
// todos table: filter, insert, toggle, delete. Mutations are plain
// POST handlers (PRG) exercising libraVDB's SQL CRUD surface.
func Editor(p models.Page) {
    <div class="page-heading">
        <div>
            <div class="eyebrow">Data / Tables</div>
            <h1>Table editor</h1>
            <p class="page-subtitle">Manage the <code>todos</code> table with fast, direct CRUD actions.</p>
        </div>
        <div class="heading-chip"><span class="table-chip-icon">▦</span> todos</div>
    </div>

    @if p.Dash.TodoStat != "" {
        <div class="notice notice-success"><span class="notice-icon">✓</span>{p.Dash.TodoStat}</div>
    }

    <div class="editor-toolbar" x-data={map[string]any{"newTodo": false}}>
        <form method="get" action="/dashboard/editor" class="search-form">
            <label class="sr-only" for="todo-search">Search todos</label>
            <div class="search-input-wrap">
                <span class="search-icon">⌕</span>
                <input id="todo-search" class="uk-input search-input" type="search" name="q" placeholder="Search by title…"
                    value={p.Dash.TodoFilter} />
            </div>
            <button type="submit" class="uk-button uk-button-secondary search-button">Search</button>
        </form>
        <button type="button" class="uk-button uk-button-primary add-todo-button" @click="newTodo = !newTodo">
            <span class="button-plus">+</span> New todo
        </button>
    </div>

    <div class="new-todo-panel" x-show="newTodo" x-cloak>
        <form method="post" action="/dashboard/editor/save" class="dashboard-card">
            <div class="card-heading">
                <div>
                    <div class="card-kicker">Create record</div>
                    <h2>New todo</h2>
                </div>
                <span class="card-icon">+</span>
            </div>
            <div class="todo-form-grid">
                <div class="field field-wide"><label class="uk-form-label" for="todo-title">Title</label>
                    <input id="todo-title" class="uk-input" type="text" name="title" required /></div>
                <div class="field"><label class="uk-form-label" for="todo-priority">Priority</label>
                    <input id="todo-priority" class="uk-input" type="number" name="priority" value="3" min="1" max="5" /></div>
                <div class="field"><label class="uk-form-label" for="todo-due">Due date</label>
                    <input id="todo-due" class="uk-input" type="text" name="due_at" placeholder="2026-08-15 12:00:00" /></div>
                <div class="field field-wide"><label class="uk-form-label" for="todo-tags">Tags</label>
                    <input id="todo-tags" class="uk-input" type="text" name="tags" placeholder='["ops"]' /></div>
                <label class="checkbox-field"><input class="uk-checkbox" type="checkbox" name="completed" value="true" /> <span>Mark as completed</span></label>
            </div>
            <div class="form-actions">
                <button type="submit" class="uk-button uk-button-primary">Save todo</button>
            </div>
        </form>
    </div>

    <section class="dashboard-card table-card">
        <div class="table-card-header">
            <div>
                <div class="card-kicker">All records</div>
                <h2>Todo list</h2>
            </div>
            <span class="record-count">{len(p.Dash.Todos)} records</span>
        </div>
        <div class="table-scroll">
            <table class="uk-table todo-table">
                <thead>
                    <tr>
                        <th>Task</th>
                        <th>Status</th>
                        <th>Priority</th>
                        <th>Due</th>
                        <th>Tags</th>
                        <th class="actions-column">Actions</th>
                    </tr>
                </thead>
                <tbody>
                    @for _, t := range p.Dash.Todos {
                        <tr class={todoRowClass(t.Completed)}>
                            <td class="todo-title-cell"><span class={todoMarkerClass(t.Completed)}></span><strong>{t.Title}</strong></td>
                            <td>
                                @if t.Completed {
                                    <span class="status-pill status-done">done</span>
                                } @else {
                                    <span class="status-pill status-open">open</span>
                                }
                            </td>
                            <td><span class="priority-pill">P{t.Priority}</span></td>
                            <td class="muted">{t.DueAt}</td>
                            <td class="muted small">{t.Tags}</td>
                            <td class="row-actions">
                                <form method="post" action="/dashboard/editor/toggle" class="action-form">
                                    <input type="hidden" name="id" value={t.ID} />
                                    <input type="hidden" name="q" value={p.Dash.TodoFilter} />
                                    @if t.Completed {
                                        <button type="submit" class="uk-button uk-button-small uk-button-secondary action-button">Mark open</button>
                                    } @else {
                                        <button type="submit" class="uk-button uk-button-small uk-button-secondary action-button">Mark done</button>
                                    }
                                </form>
                                <form method="post" action="/dashboard/editor/delete" class="action-form">
                                    <input type="hidden" name="id" value={t.ID} />
                                    <input type="hidden" name="q" value={p.Dash.TodoFilter} />
                                    <button type="submit" class="uk-button uk-button-small uk-button-ghost action-button action-delete">Delete</button>
                                </form>
                            </td>
                        </tr>
                    }
                    @if len(p.Dash.Todos) == 0 {
                        <tr><td colspan="6" class="empty-table">No todos match your search.</td></tr>
                    }
                </tbody>
            </table>
        </div>
    </section>
}
