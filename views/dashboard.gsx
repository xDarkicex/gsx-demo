@import "time"
// StatCard is a plain composable — props are typed parameters,
// compiled to direct function calls.
func StatCard(label string, value int, icon string) {
    <article class="stat-card">
        <div class="stat-card-top"><span class="stat-icon">{icon}</span><span class="stat-trend">+ live</span></div>
        <div class="stat-value">{value}</div>
        <div class="stat-label">{label}</div>
    </article>
}

// LiveClock renders in a worker goroutine (Async); the finished
// bytes replace the skeleton via an OOB swap into #liveclock.
// WithOOB("live-clock") sets the swap target. The server time is
// hydrated into Alpine, which then ticks it every second — a
// genuinely live clock.
@oob "live-clock"
@async
func LiveClock() {
    <div class="card clock">
        <h2>Live clock <span class="muted">(async + OOB + Alpine tick)</span></h2>
        <p class="clock-time"
            x-data={map[string]any{"now": time.Now().Format("15:04:05")}}
            x-init="setInterval(() => now = new Date().toLocaleTimeString('en-GB', {hour12: false}), 1000)"
            x-text="now">
            {time.Now().Format("15:04:05")}
        </p>
        <p class="muted small">
            Rendered in a worker goroutine after the page streamed —
            the skeleton appeared instantly, this HTML swapped in
            via <code>hx-swap-oob</code>, and the time ticks via
            Alpine from the server-hydrated start value.
        </p>
    </div>
}

// ClockSkeleton is the immediate fallback for LiveClock. The
// decorator names the component it falls back for.
@fallback(LiveClock)
func ClockSkeleton() {
    <div class="card clock">
        <h2>Live clock</h2>
        <p class="clock-time skeleton-line">Loading…</p>
    </div>
}
