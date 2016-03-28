package view

import "github.com/milanaleksic/clici/model"

// AvoidUnicode being set will ask UI NOT to use nice Unicode characters for Job statuses, like "✓"
var AvoidUnicode bool

// View is a very simple representation of a view that is able to represent current state
// of the application state model
type View interface {
	PresentState(state *model.State)
	Close()
}
