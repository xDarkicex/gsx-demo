@import "github.com/xDarkicex/gsx-demo/models"
@import "github.com/xDarkicex/gsx-demo/internal/auth"
@import "github.com/xDarkicex/gsx-demo/internal/db"
@import "strings"

// GraphPage explores the FOLLOWS graph. Edges add and remove in
// place via HTMX: the Add action re-renders both panels, each
// edge row's Remove action deletes just its <li>.
func GraphPage(p models.Page) {
    <div class="page-heading">
        <div>
            <div class="eyebrow">Social graph</div>
            <h1>Graph explorer</h1>
            <p class="page-subtitle">
                The <code>FOLLOWS</code> graph — edges via
                <code>GRAPH_EDGES</code>, traversals via <code>JOIN MATCH</code>.
            </p>
        </div>
        <div class="heading-chip"><span class="status-dot"></span> FOLLOWS</div>
    </div>

    <div id="graph-panels">
        <EdgeList props={map[string]any{
            "edges": p.Dash.Edges, "suggestions": p.Dash.Suggestions,
            "following": p.Dash.Following, "msg": p.Dash.GraphMsg,
            "ok": p.Dash.GraphMsgOk, "me": p.User.ID,
        }} />
    </div>
}

// EdgeList renders the edges + suggestions panels. Its Add action
// creates endpoints when missing and re-renders both panels.
@action Add(rc *render.RenderContext, props map[string]any) error {
    from := strings.ToLower(props["from"].(string))
    to := strings.ToLower(props["to"].(string))
    if from == "" || to == "" {
        props["msg"] = "Both endpoints are required."
        props["ok"] = false
        return nil
    }
    var created []string
    for _, id := range []string{from, to} {
        u, err := db.Default.UserByID(rc.Request.Context(), id)
        if err != nil {
            return err
        }
        if u == nil {
            created = append(created, id)
        }
    }
    for _, id := range created {
        if err := db.Default.EnsureUser(rc.Request.Context(), id); err != nil {
            return err
        }
    }
    if err := db.Default.Follow(rc.Request.Context(), from, to); err != nil {
        return err
    }
    props["msg"] = ""
    if len(created) > 0 {
        props["msg"] = "created " + strings.Join(created, ", ") + " — edge added."
        props["ok"] = true
    }
    // Reload the panels' data.
    es, err := db.Default.Edges(rc.Request.Context())
    if err != nil {
        return err
    }
    props["edges"] = es
    me := auth.UserFromRequest(rc.Request)
    sug, err := db.Default.SuggestFresh(rc.Request.Context(), me.ID)
    if err != nil {
        return err
    }
    props["suggestions"] = sug
    following, err := db.Default.Following(rc.Request.Context(), me.ID)
    if err != nil {
        return err
    }
    props["following"] = following
    return nil
}

func EdgeList(props map[string]any) {
    <div class="overview-grid">
        <section class="dashboard-card">
            <div class="card-heading">
                <div>
                    <div class="card-kicker">Connections</div>
                    <h2>Edges</h2>
                </div>
                <span class="record-count">{len(props["edges"].([]models.Edge))}</span>
            </div>

            @if props["msg"].(string) != "" {
                @if props["ok"] == true {
                    <div class="notice notice-success"><span class="notice-icon">✓</span>{props["msg"].(string)}</div>
                } @else {
                    <div class="notice notice-error"><span class="notice-icon">✕</span>{props["msg"].(string)}</div>
                }
            }

            <form class="graph-add-form"
                hx-post="/_nano/action/EdgeList/Add"
                hx-target="#graph-panels" hx-swap="outerHTML">
                <input class="uk-input" type="text" name="from" placeholder="from" />
                <span class="graph-arrow">→</span>
                <input class="uk-input" type="text" name="to" placeholder="to" />
                <button class="uk-button uk-button-primary">Add edge</button>
            </form>

            <ul class="clean-list">
                @for _, e := range props["edges"].([]models.Edge) {
                    <EdgeRow props={map[string]any{"from": e.From, "to": e.To}} />
                }
                @if len(props["edges"].([]models.Edge)) == 0 {
                    <li class="empty-state">No edges yet — add one above.</li>
                }
            </ul>
        </section>

        <section class="dashboard-card">
            <div class="card-heading">
                <div>
                    <div class="card-kicker">2-hop traversal</div>
                    <h2>Suggestions</h2>
                </div>
                <span class="record-count">{len(props["suggestions"].([]models.Suggestion))}</span>
            </div>
            <p class="card-description">
                Chained <code>JOIN MATCH (me)-[f1]-&gt;(mid)-[f2]-&gt;(tgt)</code>,
                grouped by mutual count.
            </p>

            <ul class="clean-list">
                @for _, s := range props["suggestions"].([]models.Suggestion) {
                    <li>
                        <span class="list-avatar">{userInitial(s.Name)}</span>
                        <span class="list-copy">
                            <strong>{s.Name}</strong>
                            <span><a href={"/profile/" + s.ID}>&#64;{s.ID}</a> · {s.Mutual} mutual</span>
                        </span>
                    </li>
                }
                @if len(props["suggestions"].([]models.Suggestion)) == 0 {
                    <li class="empty-state">No suggestions right now.</li>
                }
            </ul>
        </section>
    </div>
}

// EdgeRow is one edge list item. Its Remove action deletes the
// edge; HTMX removes the <li> in place.
@action Remove(rc *render.RenderContext, props map[string]any) error {
    if err := db.Default.Unfollow(rc.Request.Context(),
        props["from"].(string), props["to"].(string)); err != nil {
        return err
    }
    return nil // htmx swap:delete removes the row
}

func EdgeRow(props map[string]any) {
    <li>
        <span class="list-avatar">{userInitial(props["from"].(string))}</span>
        <span class="list-copy">
            <strong>{props["from"].(string)}</strong>
            <span>→ {props["to"].(string)}</span>
        </span>
        <form class="action-form list-action"
            hx-post="/_nano/action/EdgeRow/Remove"
            hx-target="closest li" hx-swap="delete">
            <input type="hidden" name="from" value={props["from"].(string)} />
            <input type="hidden" name="to" value={props["to"].(string)} />
            <button class="uk-button uk-button-small uk-button-ghost action-button action-delete">Remove</button>
        </form>
    </li>
}
