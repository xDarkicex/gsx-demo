@import "github.com/xDarkicex/gsx-demo/models"
@import "github.com/xDarkicex/gsx-demo/internal/db"
@import "errors"

// Editor is the Table Editor — raw relational CRUD against the
// todos table, driven by HTMX: the search refetches the table
// (debounced), Save swaps the table, and each row toggles/deletes
// in place via colocated @actions.
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

    <div class="editor-toolbar">
        <form class="search-form" hx-get="/dashboard/editor/table" hx-target="#todo-table"
            hx-trigger="input changed delay:300ms" hx-swap="outerHTML">
            <label class="sr-only" for="todo-search">Search todos</label>
            <div class="search-input-wrap">
                <span class="search-icon">⌕</span>
                <input id="todo-search" class="uk-input search-input" type="search" name="q"
                    placeholder="Search by title…" value={p.Dash.TodoFilter} />
            </div>
            <button type="submit" class="uk-button uk-button-secondary search-button">Search</button>
        </form>
        <button type="button" class="uk-button uk-button-primary add-todo-button" uk-toggle="target: #new-todo-modal">
            <span class="button-plus">+</span> New todo
        </button>
    </div>

    <div id="new-todo-modal" uk-modal>
        <div class="uk-modal-dialog dashboard-card">
            <button class="uk-modal-close-default modal-close" type="button" uk-close aria-label="Close"></button>
            <div class="card-heading">
                <div>
                    <div class="card-kicker">Create record</div>
                    <h2>New todo</h2>
                </div>
            </div>
            <form
                hx-post="/_nano/action/TodoTable/Save"
                hx-target="#todo-table" hx-swap="outerHTML"
                hx-on::after-swap="UIkit.modal('#new-todo-modal').hide(); UIkit.notification({message: 'Todo saved', status: 'success'})"
                hx-on::response-error="UIkit.notification({message: 'Save failed — check the fields', status: 'danger'})">
                <div class="todo-form-grid">
                    <div class="field field-wide"><label class="uk-form-label" for="todo-title">Title</label>
                        <input id="todo-title" class="uk-input" type="text" name="title" placeholder="Ship the modal" required autofocus /></div>
                    <div class="field"><label class="uk-form-label" for="todo-priority">Priority</label>
                        <input id="todo-priority" class="uk-input" type="number" name="priority" value="3" min="1" max="5" /></div>
                    <div class="field"><label class="uk-form-label" for="todo-status">Status</label>
                        <select id="todo-status" class="uk-select" name="completed">
                            <option value="false">open</option>
                            <option value="true">closed</option>
                        </select></div>
                    <div class="field"><label class="uk-form-label" for="todo-opened">Opened date</label>
                        <input id="todo-opened" class="uk-input" type="text" name="opened_at" placeholder="2026-08-11 09:00:00" /></div>
                    <div class="field"><label class="uk-form-label" for="todo-due">Due date</label>
                        <input id="todo-due" class="uk-input" type="text" name="due_at" placeholder="2026-08-15 12:00:00" /></div>
                    <div class="field field-wide" x-data="{tags: [], draft: '', addTag() { var t = this.draft.trim().replace(/,+$/, ''); if (t && !this.tags.includes(t)) { this.tags.push(t); } this.draft = ''; }}">
                        <label class="uk-form-label">Tags</label>
                        <div class="tag-input-wrap">
                            <div class="tag-chips">
                                <template x-for="t in tags" :key="t">
                                    <span class="tag-chip" x-text="t" @click="tags = tags.filter(x => x !== t)"></span>
                                </template>
                            </div>
                            <input id="todo-tags" class="uk-input" type="text" x-model="draft"
                                @keydown.enter.prevent="addTag"
                                @keydown="if ($event.key === ',') { $event.preventDefault(); addTag() }"
                                placeholder="ops, docs" />
                        </div>
                        <input type="hidden" name="tags" :value="JSON.stringify(tags)" />
                    </div>
                </div>
                <div class="form-actions">
                    <button type="button" class="uk-button uk-button-secondary" uk-toggle="target: #new-todo-modal">Cancel</button>
                    <button type="submit" class="uk-button uk-button-primary">Save todo</button>
                </div>
            </form>
        </div>
    </div>

    <div id="todo-table">
        <TodoTable props={map[string]any{"todos": p.Dash.Todos, "filter": p.Dash.TodoFilter}} />
    </div>
}

// TodoTable renders the records table. Its Save action inserts a
// row and re-renders the table, which HTMX swaps in place.
@action Save(rc *render.RenderContext, props map[string]any) error {
    title, _ := props["title"].(string)
    if title == "" {
        return errors.New("title is required")
    }
    openedAt, _ := props["opened_at"].(string)
    dueAt, _ := props["due_at"].(string)
    tags, _ := props["tags"].(string)
    t := &models.Todo{
        Title:     title,
        Priority:  3,
        OpenedAt:  openedAt,
        DueAt:     dueAt,
        Tags:      tags,
        Completed: props["completed"] == true,
    }
    if p, ok := props["priority"].(int); ok && p >= 1 && p <= 5 {
        t.Priority = p
    }
    if err := db.Default.SaveTodo(rc.Request.Context(), t); err != nil {
        return err
    }
    ts, err := db.Default.Todos(rc.Request.Context(), "")
    if err != nil {
        return err
    }
    props["todos"] = ts
    props["filter"] = ""
    return nil
}

func TodoTable(props map[string]any) {
    <section class="dashboard-card table-card">
        <div class="table-card-header">
            <div>
                <div class="card-kicker">All records</div>
                <h2>Todo list</h2>
            </div>
            <span class="record-count">{len(props["todos"].([]models.Todo))} records</span>
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
                    @for _, t := range props["todos"].([]models.Todo) {
                        <TodoRow props={map[string]any{
                            "id": t.ID, "title": t.Title, "completed": t.Completed,
                            "priority": t.Priority, "due_at": t.DueAt, "tags": t.Tags,
                        }} />
                    }
                    @if len(props["todos"].([]models.Todo)) == 0 {
                        <tr><td colspan="6" class="empty-table">No todos match your search.</td></tr>
                    }
                </tbody>
            </table>
        </div>
    </section>
}

// TodoRow is one table row. Its actions toggle and delete the
// record in place — HTMX swaps just the row.
@action Toggle(rc *render.RenderContext, props map[string]any) error {
    id := props["id"].(string)
    if err := db.Default.ToggleTodo(rc.Request.Context(), id); err != nil {
        return err
    }
    // Reload the full record — the re-render needs every field,
    // not just the id from the form.
    t, err := db.Default.TodoByID(rc.Request.Context(), id)
    if err != nil {
        return err
    }
    props["id"] = t.ID
    props["title"] = t.Title
    props["completed"] = t.Completed
    props["priority"] = t.Priority
    props["due_at"] = t.DueAt
    props["tags"] = t.Tags
    return nil
}

@action Delete(rc *render.RenderContext, props map[string]any) error {
    id := props["id"].(string)
    if err := db.Default.DeleteTodo(rc.Request.Context(), id); err != nil {
        return err
    }
    return nil // htmx swap:delete removes the row; body ignored
}

func TodoRow(props map[string]any) {
    <tr class={todoRowClass(props["completed"] == true)}>
        <td class="todo-title-cell"><span class={todoMarkerClass(props["completed"] == true)}></span><strong>{props["title"]}</strong></td>
        <td>
            @if props["completed"] == true {
                <span class="status-pill status-done">done</span>
            } @else {
                <span class="status-pill status-open">open</span>
            }
        </td>
        <td><span class="priority-pill">P{props["priority"]}</span></td>
        <td class="muted">{props["due_at"]}</td>
        <td class="muted small">{props["tags"]}</td>
        <td class="row-actions">
            <form class="action-form"
                hx-post="/_nano/action/TodoRow/Toggle"
                hx-target="closest tr" hx-swap="outerHTML">
                <input type="hidden" name="id" value={props["id"].(string)} />
                <input type="hidden" name="completed" value={props["completed"]} />
                <button type="submit" class="uk-button uk-button-small uk-button-secondary action-button">
                    @if props["completed"] == true {
                        Mark open
                    } @else {
                        Mark done
                    }
                </button>
            </form>
            <form class="action-form"
                hx-post="/_nano/action/TodoRow/Delete"
                hx-target="closest tr" hx-swap="delete">
                <input type="hidden" name="id" value={props["id"].(string)} />
                <button type="submit" class="uk-button uk-button-small uk-button-ghost action-button action-delete">Delete</button>
            </form>
        </td>
    </tr>
}
