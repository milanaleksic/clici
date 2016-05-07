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

// CallbackAsView is a simple wrapper for views that are simple and identified via a simple callback function
type CallbackAsView func(state *model.State)

// PresentState will just execute underlying function with updated state
func (callback CallbackAsView) PresentState(state *model.State) {
	callback(state)
}

// Close is ignored in this case
func (callback CallbackAsView) Close() {
}