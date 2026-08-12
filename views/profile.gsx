@import "github.com/xDarkicex/gsx-demo/models"

// ProfilePage shows a user's profile — self (via the dashboard
// user) or any @handle on the dash. Stats come from the FOLLOWS
// graph; other users get a follow button.
func ProfilePage(p models.Page) {
    <div class="profile-head">
        <span class="list-avatar profile-avatar">{userInitial(p.Profile.User.Name)}</span>
        <div class="profile-meta">
            <h1>{p.Profile.User.Name}</h1>
            <p class="muted">&#64;{p.Profile.User.ID} · {p.Profile.User.Email}</p>
            <p class="muted small">member since {p.Profile.User.CreatedAt}</p>
        </div>
        @if !p.Profile.IsSelf {
            <li class="profile-follow">
                <FollowButton props={map[string]any{
                    "target_id": p.Profile.User.ID, "followed": p.Profile.Followed,
                }} />
            </li>
        }
    </div>

    <div class="stat-grid">
        <StatCard label="Following" value={p.Profile.FollowingCount} icon="→" />
        <StatCard label="Followers" value={p.Profile.FollowerCount} icon="←" />
    </div>

    <section class="dashboard-card">
        <div class="card-heading">
            <div>
                <div class="card-kicker">FOLLOWS graph</div>
                <h2>Following</h2>
            </div>
            <span class="record-count">{p.Profile.FollowingCount}</span>
        </div>
        <ul class="clean-list">
            @for _, f := range p.Profile.Following {
                <li>
                    <span class="list-avatar">{userInitial(f.Name)}</span>
                    <span class="list-copy">
                        <strong>{f.Name}</strong>
                        <span><a href={"/profile/" + f.ID}>&#64;{f.ID}</a></span>
                    </span>
                    <a class="uk-button uk-button-small uk-button-ghost action-button" href={"/profile/" + f.ID}>View profile</a>
                </li>
            }
            @if len(p.Profile.Following) == 0 {
                <li class="empty-state">Nobody yet.</li>
            }
        </ul>
    </section>
}
