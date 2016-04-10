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

type CallbackAsView func(state *model.State)

func (callback CallbackAsView) PresentState(state *model.State) {
	callback(state)
}

func (callback CallbackAsView) Close() {
}