package main

import (
	"fmt"
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
	state   State
	reducer func(state State, action Action) State
}

func NewStore(reducer func(State, Action) State, initialState State) *Store {
	return &Store{
		state:   initialState,
		reducer: reducer,
	}
}

func (s *Store) Dispatch(action Action) {
	s.state = s.reducer(s.state, action)
}

func (s *Store) GetState() State {
	return s.state
}

// ViewModel
type ViewModel struct {
	store *Store // Add store to ViewModel
}

func NewViewModel(store *Store) *ViewModel {
	vm := &ViewModel{
		store: store, // Store the store
	}

	return vm
}

func (v *ViewModel) Incre() {
	v.store.Dispatch(IncrementAction{})
}

func (v *ViewModel) Decre() {
	v.store.Dispatch(DecrementAction{})
}

func (v *ViewModel) CountLabel() string {
	return fmt.Sprintf("%v", v.store.state.Count)
}

func main() {
	go func() {
		w := &app.Window{}
		w.Option(app.Title("Counter App"))
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

	var ops op.Ops
	view := NewView(viewModel, th) // Pass ViewModel to View

	for {
		switch e := w.Event().(type) {
		case app.DestroyEvent:
			return e.Err
		case app.FrameEvent:
			gtx := app.NewContext(&ops, e)

			view.Layout(gtx, th) // Pass theme to Layout
			e.Frame(gtx.Ops)
		}
	}
}

type View struct {
	viewModel       *ViewModel // Use ViewModel
	theme           *material.Theme
	incrementButton widget.Clickable
	decrementButton widget.Clickable
}

func NewView(vm *ViewModel, theme *material.Theme) *View { // Accept ViewModel
	return &View{
		viewModel:       vm,
		theme:           theme,
		incrementButton: widget.Clickable{}, // Initialize buttons here
		decrementButton: widget.Clickable{}, // Initialize buttons here
	}
}

// Layout accepts theme
func (v *View) Layout(gtx layout.Context, th *material.Theme) layout.Dimensions {
	// Event handling in Layout
	return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{
			Axis:      layout.Horizontal,
			Alignment: layout.Middle,
			Spacing:   layout.SpaceEvenly,
		}.Layout(gtx,
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				if v.incrementButton.Clicked(gtx) {
					v.viewModel.Incre()
				}
				return material.Button(th, &v.incrementButton, "Increment").Layout(gtx)
			}),
			layout.Rigid(layout.Spacer{Width: unit.Dp(20)}.Layout),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				paint.ColorOp{Color: color.NRGBA{R: 0, G: 0, B: 0, A: 255}}.Add(gtx.Ops)
				m := material.Body1(th, v.viewModel.CountLabel()) // Use ViewModel's CountLabel
				m.Font.Weight = font.Bold
				return m.Layout(gtx)
			}),
			layout.Rigid(layout.Spacer{Width: unit.Dp(20)}.Layout),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				if v.decrementButton.Clicked(gtx) {
					v.viewModel.Decre()
				}
				return material.Button(th, &v.decrementButton, "Decrement").Layout(gtx)
			}),
		)
	})
}
