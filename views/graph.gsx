@import "github.com/xDarkicex/gsx-demo/models"

// GraphPage explores the FOLLOWS graph: the edge list (JOIN MATCH),
// an add-edge form (GRAPH_EDGES INSERT), and the 2-hop traversal
// with mutual counts.
func GraphPage(p models.Page) {
    <h1 class="uk-h2">Graph</h1>
    <p class="muted">FOLLOWS — <code>GRAPH_EDGES</code>, <code>JOIN MATCH</code>, 2-hop traversal</p>

    <div class="uk-grid uk-child-width-1-2@m" uk-grid>
        <div class="uk-card uk-card-default uk-card-body">
            <h3 class="uk-h4">Edges <span class="uk-badge">{len(p.Dash.Edges)}</span></h3>
            <form method="post" action="/dashboard/graph/add" class="uk-grid-small" uk-grid>
                <div class="uk-width-1-3"><input class="uk-input" type="text" name="from" placeholder="from" /></div>
                <div class="uk-width-1-3"><input class="uk-input" type="text" name="to" placeholder="to" /></div>
                <div class="uk-width-1-3"><button class="uk-button uk-button-small uk-button-primary">Add edge</button></div>
            </form>
            <ul class="uk-list uk-list-divider uk-margin-top">
                @for _, e := range p.Dash.Edges {
                    <li class="uk-flex uk-flex-between">
                        <span><strong>{e.From}</strong> → {e.To}</span>
                        <form method="post" action="/dashboard/graph/remove" class="uk-display-inline">
                            <input type="hidden" name="from" value={e.From} />
                            <input type="hidden" name="to" value={e.To} />
                            <button class="uk-button uk-button-small uk-button-danger">Remove</button>
                        </form>
                    </li>
                }
                @if len(p.Dash.Edges) == 0 {
                    <li class="muted">No edges.</li>
                }
            </ul>
        </div>

        <div class="uk-card uk-card-default uk-card-body">
            <h3 class="uk-h4">2-hop suggestions <span class="uk-badge">{len(p.Dash.Suggestions)}</span></h3>
            <p class="muted small">
                Chained <code>JOIN MATCH (me)-[f1]->(mid) JOIN MATCH (mid)-[f2]->(tgt)</code>
                with <code>GROUP BY</code> mutual counts.
            </p>
            <ul class="uk-list uk-list-divider">
                @for _, s := range p.Dash.Suggestions {
                    <li>
                        <strong>{s.Name}</strong>
                        <span class="muted">&#64;{s.ID} · {s.Mutual} mutual</span>
                    </li>
                }
                @if len(p.Dash.Suggestions) == 0 {
                    <li class="muted">No suggestions.</li>
                }
            </ul>
            <hr class="uk-divider-small" />
            <h3 class="uk-h4">Following <span class="uk-badge">{len(p.Dash.Following)}</span></h3>
            <ul class="uk-list uk-list-divider">
                @for _, f := range p.Dash.Following {
                    <li>{f.Name} <span class="muted">&#64;{f.ID}</span></li>
                }
                @if len(p.Dash.Following) == 0 {
                    <li class="muted">Nobody yet.</li>
                }
            </ul>
        </div>
    </div>
}
