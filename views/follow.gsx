@import "fmt"
@import "errors"
@import "github.com/xDarkicex/gsx-demo/internal/auth"
@import "github.com/xDarkicex/gsx-demo/internal/db"

// Follow is a colocated server action (Next.js "use server" style).
// It runs via POST /_nano/action/FOLLOWBUTTON/Follow with HTMX
// headers; the component re-renders with the mutated props and the
// response swaps the button in place. A validation failure sets a
// flash form error and returns render.ErrValidation — the @error
// macro below displays it at 200.
@action Follow(rc *render.RenderContext, props map[string]any) error {
    me := auth.UserFromRequest(rc.Request)
    if me == nil {
        return errors.New("unauthenticated")
    }
    target, _ := props["target_id"].(string)
    if target == "" {
        rc.SetFormError("target", "Missing target user")
        return render.ErrValidation
    }
    if props["followed"] == true {
        props["followed"] = false
        return db.Default.Unfollow(rc.Request.Context(), me.ID, target)
    }
    props["followed"] = true
    return db.Default.Follow(rc.Request.Context(), me.ID, target)
}

// @memo caches the button's HTML by (target, state) — the render
// walk is bypassed for repeated keys on the action re-render path.
@memo(func(rc *render.RenderContext, props map[string]any) string {
    return fmt.Sprint(props["target_id"], "/", props["followed"])
})

// FollowButton is the toggling follow control. Its props are a
// map[string]any so the action's parsed body re-renders it directly
// (the registry adapter casts to the declared param type). The
// outer guard tolerates a missing target_id — the validation
// re-render after ErrValidation has only the flash error to show.
func FollowButton(props map[string]any) {
    @if v, ok := props["target_id"].(string); ok {
        @if props["followed"] == true {
            <form hx-post="/_nano/action/FollowButton/Follow" hx-target="closest li" hx-swap="outerHTML">
                <input type="hidden" name="target_id" value={v} />
                <input type="hidden" name="followed" value="true" />
                <button type="submit" class="btn btn-sm">✓ Following</button>
            </form>
        } @else {
            <form hx-post="/_nano/action/FollowButton/Follow" hx-target="closest li" hx-swap="outerHTML">
                <input type="hidden" name="target_id" value={v} />
                <input type="hidden" name="followed" value="false" />
                <button type="submit" class="btn btn-sm btn-primary">+ Follow</button>
            </form>
        }
    }
    @error("target")
}
