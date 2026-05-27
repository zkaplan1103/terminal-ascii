// Package page holds types shared between the router and the page models.
// Pages emit NavigateMsg to request a page swap; the router intercepts it.
// This package exists solely to break the router↔pages import cycle.
package page

// NavigateMsg is emitted by a page when it wants the router to swap the active
// page. The string is the target page key (e.g. "bio", "adopt", "nav").
type NavigateMsg struct {
	To string
}
