package main

import (
	"image/color"
	"log"
	"strconv"

	"gioui.org/app"
	"gioui.org/font"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
)

// State
type State struct {
	Count int
}

// Action
type Action interface {
	Apply(s State) State
}

// IncrementAction
type IncrementAction struct{}

func (a IncrementAction) Apply(s State) State {
	s.Count++
	return s
}

// DecrementAction
type DecrementAction struct{}

func (a DecrementAction) Apply(s State) State {
	s.Count--
	return s
}

// Reducer
func reduce(state State, action Action) State {
	return action.Apply(state)
}

// Store
type Store struct {
	state     State
	reducer   func(state State, action Action) State
	listeners []func(State)
}

func NewStore(reducer func(State, Action) State, initialState State) *Store {
	return &Store{
		state:     initialState,
		reducer:   reducer,
		listeners: []func(State){},
	}
}

func (s *Store) Subscribe(listener func(State)) {
	s.listeners = append(s.listeners, listener)
}

func (s *Store) Dispatch(action Action) {
	s.state = s.reducer(s.state, action)
	for _, listener := range s.listeners {
		listener(s.state)
	}
}

func (s *Store) GetState() State {
	return s.state
}

func main() {
	go func() {
		w := &app.Window{}
		if err := run(w); err != nil {
			log.Fatal(err)
		}
	}()
	app.Main()
}

func run(w *app.Window) error {
	// gofont.Register() // gofont.Register is automatically called by material.NewTheme
	th := material.NewTheme()

	// Redux Store
	store := NewStore(reduce, State{Count: 0})

	// UI elements
	incrementButton := widget.Clickable{} // Corrected widget.Button
	decrementButton := widget.Clickable{} // Corrected widget.Button
	countLabel := ""

	// Subscribe to store changes
	store.Subscribe(func(state State) {
		countLabel = "Count: " + strconv.Itoa(state.Count)
		w.Invalidate() // Corrected w.Invalidate()
	})

	var ops op.Ops
	for {
		switch e := w.Event().(type) {
		case app.DestroyEvent: // Corrected event.DestroyEvent
			return e.Err
		case app.FrameEvent: // Corrected event.FrameEvent
			gtx := app.NewContext(&ops, e) // Corrected layout.NewContext

			// Event handling
			if incrementButton.Clicked(gtx) {
				store.Dispatch(IncrementAction{})
			}
			if decrementButton.Clicked(gtx) {
				store.Dispatch(DecrementAction{})
			}

			layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{
					Axis:      layout.Horizontal, // Corrected layout.Row
					Alignment: layout.Middle,
					Spacing:   layout.SpaceEvenly,
				}.Layout(gtx,
					layout.Rigid(material.Button(th, &incrementButton, "Increment").Layout),
					layout.Rigid(layout.Spacer{Width: unit.Dp(20)}.Layout), // Use unit.Dp for spacing
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						paint.ColorOp{Color: color.NRGBA{R: 0, G: 0, B: 0, A: 255}}.Add(gtx.Ops) // Corrected color.Black to color.NRGBA
						m := material.Body1(th, countLabel)
						m.Font.Weight = font.Bold
						return m.Layout(gtx)
					}),
					layout.Rigid(layout.Spacer{Width: unit.Dp(20)}.Layout), // Use unit.Dp for spacing
					layout.Rigid(material.Button(th, &decrementButton, "Decrement").Layout),
				)
			})

			e.Frame(gtx.Ops)
		}
	}
}
