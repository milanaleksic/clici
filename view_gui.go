package main

import (
	"github.com/gotk3/gotk3/gtk"
	"log"
)

type GUIInterface struct {
	feedbackChannel chan Command
}

func (ui *GUIInterface) PresentState(state *State) {
	if state.Error != nil || len(state.JobStates) == 0 {
		// gui.errorDialog(state)
		// gui.bottomLine()
		return
	}
	if len(state.FailedTests) != 0 {
		// gui.informationDialogOfTests(state)
		// gui.bottomLine()
		return
	}

	return
}

func NewGUIInterface(feedbackChannel chan Command) (view *GUIInterface, err error) {
	view = &GUIInterface{
		feedbackChannel: feedbackChannel,
	}
	gtk.Init(nil)

	builder, err := gtk.BuilderNew()
	if err != nil {
		log.Fatal("Unable to create window:", err)
	}

	err = builder.AddFromString(string(MustAsset("data/test.glade")))
	if err != nil {
		log.Fatal("Unable to create window:", err)
	}

	win, err := builder.GetObject("window1")

	if w, ok := win.(*gtk.Window); ok {
		w.SetTitle("Add/Remove Widgets Example")
		w.Connect("destroy", func() {
			gtk.MainQuit()
		})

		go func() {
			w.ShowAll()
			gtk.Main()
			feedbackChannel <- CmdShutdown()
		}()

	} else {
		log.Fatal("Unable to create window:", err)
	}

	return
}

func (ui *GUIInterface) Close() {
}

func (ui *GUIInterface) windowWidget() *gtk.Widget {
	grid, err := gtk.GridNew()
	if err != nil {
		log.Fatal("Unable to create grid:", err)
	}
	grid.SetOrientation(gtk.ORIENTATION_VERTICAL)

	// Just as a demonstration, we create and destroy a Label without ever
	// adding it to a container.  In native GTK, this would result in a
	// memory leak, since gtk_widget_destroy() will not deallocate any
	// memory when passed a GtkWidget with a floating reference.
	//
	// gotk3 handles this situation by always sinking floating references
	// of any struct type embedding a glib.InitiallyUnowned, and by setting
	// a finalizer to unreference the object when Go has lost scope of the
	// variable.  Due to this design, widgets may be allocated freely
	// without worrying about handling memory incorrectly.
	//
	// The following code is not entirely useful (except to demonstrate
	// this point), but it is also not "incorrect" as the C equivalent
	// would be.
	unused, err := gtk.LabelNew("This label is never used")
	if err != nil {
		// Calling Destroy() is also unnecessary in this case.  The
		// memory will still be freed with or without calling it.
		unused.Destroy()
	}

	sw, err := gtk.ScrolledWindowNew(nil, nil)
	if err != nil {
		log.Fatal("Unable to create scrolled window:", err)
	}

	grid.Attach(sw, 0, 0, 2, 1)
	sw.SetHExpand(true)
	sw.SetVExpand(true)

	labelsGrid, err := gtk.GridNew()
	if err != nil {
		log.Fatal("Unable to create grid:", err)
	}
	labelsGrid.SetOrientation(gtk.ORIENTATION_VERTICAL)

	sw.Add(labelsGrid)
	labelsGrid.SetHExpand(true)

	insertBtn, err := gtk.ButtonNewWithLabel("Add a label")
	if err != nil {
		log.Fatal("Unable to create button:", err)
	}
	removeBtn, err := gtk.ButtonNewWithLabel("Remove a label")
	if err != nil {
		log.Fatal("Unable to create button:", err)
	}

	grid.Attach(insertBtn, 0, 1, 1, 1)
	grid.Attach(removeBtn, 1, 1, 1, 1)

	return &grid.Container.Widget
}