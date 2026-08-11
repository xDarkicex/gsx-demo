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
