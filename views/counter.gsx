@import "github.com/xDarkicex/gsx-demo/internal/db"

// Increment is a colocated server action: it bumps the durable
// counter in libraVDB and re-renders the widget with the new
// value, so the clicked count persists across refreshes and
// restarts.
@action Increment(rc *render.RenderContext, props map[string]any) error {
    n, err := db.Default.IncrementCounter(rc.Request.Context())
    if err != nil {
        return err
    }
    props["count"] = n
    if _, ok := props["feature"]; !ok {
        props["feature"] = "server-hydrated"
    }
    return nil
}

// CounterWidget is the homepage's Alpine demo: the initial state
// is serialized from Go via x-data auto-hydration, and every click
// posts to the @action above, which persists the count and swaps
// the card back with the re-hydrated value.
func CounterWidget(props map[string]any) {
    <div class="card" x-data={map[string]any{"count": props["count"], "feature": props["feature"]}}>
        <h2>Alpine.js, hydrated from Go</h2>
        <p class="muted">
            This widget's initial state was serialized by
            <code>c.WriteHydrateProps</code> — zero-alloc JSON from the
            server render. Clicks persist in libraVDB.
        </p>
        <div class="demo-row">
            <button class="btn" name="click" value="1"
                hx-post="/_nano/action/CounterWidget/Increment"
                hx-target="closest div.card"
                hx-swap="outerHTML">
                Clicked <span class="badge" x-text="count"></span> times
            </button>
            <span class="badge" x-text="feature"></span>
        </div>
    </div>
}
