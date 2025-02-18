package main

import (
	"image/color"
	"log"

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

func (s *Store) Dispatch(action Action) {
	s.state = s.reducer(s.state, action)
	for _, listener := range s.listeners {
		listener(s.state)
	}
}

func (s *Store) GetState() State {
	return s.state
}

// ViewModel
type ViewModel struct {
	s          *Store
	countLabel string
}

func NewViewModel(store *Store) *ViewModel {
	vm := &ViewModel{
		s: store,
	}

	return vm
}

func (v *ViewModel) Increment() {
	v.s.Dispatch(IncrementAction{})
}

func (v *ViewModel) Decrement() {
	v.s.Dispatch(DecrementAction{})
}

func (vm *ViewModel) CountLabel() string {
	return vm.countLabel
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
	// gofont.Register()
	th := material.NewTheme()
	store := NewStore(reduce, State{Count: 0})
	viewModel := NewViewModel(store) // Create ViewModel

	// UI elements

	var ops op.Ops
	view := NewView(viewModel, th) // Pass ViewModel to View

	for {
		switch e := w.Event().(type) {
		case app.DestroyEvent:
			return e.Err
		case app.FrameEvent:
			gtx := app.NewContext(&ops, e)

			view.Layout(gtx) // Pass theme to Layout

			e.Frame(gtx.Ops)
		}
	}
}

type View struct {
	viewModel       *ViewModel // Use ViewModel
	th              *material.Theme
	incrementButton widget.Clickable
	decrementButton widget.Clickable
}

func NewView(vm *ViewModel, theme *material.Theme) *View { // Accept ViewModel
	return &View{
		viewModel:       vm,
		th:              theme,
		incrementButton: widget.Clickable{}, // Initialize buttons here
		decrementButton: widget.Clickable{}, // Initialize buttons here
	}
}

// Layout accepts theme
func (v *View) Layout(gtx layout.Context) layout.Dimensions {
	return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{
			Axis:      layout.Horizontal,
			Alignment: layout.Middle,
			Spacing:   layout.SpaceEvenly,
		}.Layout(gtx,
			layout.Rigid(material.Button(v.th, &v.incrementButton, "Increment").Layout),
			layout.Rigid(layout.Spacer{Width: unit.Dp(20)}.Layout),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				paint.ColorOp{Color: color.NRGBA{R: 0, G: 0, B: 0, A: 255}}.Add(gtx.Ops)
				m := material.Body1(v.th, v.viewModel.CountLabel()) // Use ViewModel's CountLabel
				m.Font.Weight = font.Bold
				return m.Layout(gtx)
			}),
			layout.Rigid(layout.Spacer{Width: unit.Dp(20)}.Layout),
			layout.Rigid(material.Button(v.th, &v.decrementButton, "Decrement").Layout),
		)
	})
}
