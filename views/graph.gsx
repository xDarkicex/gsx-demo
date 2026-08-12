@import "github.com/xDarkicex/gsx-demo/models"

// GraphPage explores the FOLLOWS graph: the edge list (JOIN MATCH),
// an add-edge form (GRAPH_EDGES INSERT), and the 2-hop traversal
// with mutual counts.
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

    @if p.Dash.GraphMsg != "" {
        @if p.Dash.GraphMsgOk {
            <div class="notice notice-success"><span class="notice-icon">✓</span>{p.Dash.GraphMsg}</div>
        } @else {
            <div class="notice notice-error"><span class="notice-icon">✕</span>{p.Dash.GraphMsg}</div>
        }
    }

    <div class="overview-grid">
        <section class="dashboard-card">
            <div class="card-heading">
                <div>
                    <div class="card-kicker">Connections</div>
                    <h2>Edges</h2>
                </div>
                <span class="record-count">{len(p.Dash.Edges)}</span>
            </div>

            <form method="post" action="/dashboard/graph/add" class="graph-add-form">
                <input class="uk-input" type="text" name="from" placeholder="from" />
                <span class="graph-arrow">→</span>
                <input class="uk-input" type="text" name="to" placeholder="to" />
                <button class="uk-button uk-button-primary">Add edge</button>
            </form>

            <ul class="clean-list">
                @for _, e := range p.Dash.Edges {
                    <li>
                        <span class="list-avatar">{userInitial(e.From)}</span>
                        <span class="list-copy">
                            <strong>{e.From}</strong>
                            <span>→ {e.To}</span>
                        </span>
                        <form method="post" action="/dashboard/graph/remove" class="action-form list-action">
                            <input type="hidden" name="from" value={e.From} />
                            <input type="hidden" name="to" value={e.To} />
                            <button class="uk-button uk-button-small uk-button-ghost action-button action-delete">Remove</button>
                        </form>
                    </li>
                }
                @if len(p.Dash.Edges) == 0 {
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
                <span class="record-count">{len(p.Dash.Suggestions)}</span>
            </div>
            <p class="card-description">
                Chained <code>JOIN MATCH (me)-[f1]-&gt;(mid)-[f2]-&gt;(tgt)</code>,
                grouped by mutual count.
            </p>

            <ul class="clean-list">
                @for _, s := range p.Dash.Suggestions {
                    <li>
                        <span class="list-avatar">{userInitial(s.Name)}</span>
                        <span class="list-copy">
                            <strong>{s.Name}</strong>
                            <span>&#64;{s.ID} · {s.Mutual} mutual</span>
                        </span>
                    </li>
                }
                @if len(p.Dash.Suggestions) == 0 {
                    <li class="empty-state">No suggestions right now.</li>
                }
            </ul>
        </section>
    </div>
}
