package views

import "fmt"

// navActive returns "uk-active" when the sidebar item is the
// active page — used by the DashLayout class expressions.
func navActive(active, name string) string {
	if active == name {
		return "uk-active"
	}
	return ""
}

// sidebarNavActive keeps the sidebar links self-contained and gives the
// active item a stronger visual treatment than the framework default.
func sidebarNavActive(active, name string) string {
	if active == name {
		return "sidebar-link is-active"
	}
	return "sidebar-link"
}

// userInitial gives the signed-in profile a compact, stable avatar label.
func userInitial(name string) string {
	for _, r := range name {
		if r >= 'A' && r <= 'Z' || r >= 'a' && r <= 'z' {
			if r >= 'a' && r <= 'z' {
				return string(r - ('a' - 'A'))
			}
			return string(r)
		}
	}
	return "U"
}

func todoRowClass(completed bool) string {
	if completed {
		return "todo-row is-complete"
	}
	return "todo-row"
}

func todoMarkerClass(completed bool) string {
	if completed {
		return "todo-marker is-complete"
	}
	return "todo-marker"
}

// barWidth returns the CSS width for a chart bar (percentage of
// the max value), as an inline style string.
func barWidth(count, max int) string {
	if max <= 0 || count <= 0 {
		return "width:0%"
	}
	pct := count * 100 / max
	if pct < 4 {
		pct = 4
	}
	return fmt.Sprintf("width:%d%%", pct)
}
