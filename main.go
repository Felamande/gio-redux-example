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

// Middleware type
type Middleware func(store *Store, next Dispatch) Dispatch
type Dispatch func(action Action)

// Logging Middleware
func LoggingMiddleware(store *Store, next Dispatch) Dispatch {
	return func(action Action) {
		prevState := store.GetState()
		log.Printf("Action dispatched: %T, Previous State: %+v", action, prevState)
		next(action)
		newState := store.GetState()
		log.Printf("Action dispatched: %T, New State: %+v", action, newState)
	}
}

// Store
type Store struct {
	state      State
	reducer    func(state State, action Action) State
	middleware []Middleware
	dispatch   Dispatch
}

func NewStore(reducer func(State, Action) State, initialState State, middleware ...Middleware) *Store {
	store := &Store{
		state:      initialState,
		reducer:    reducer,
		middleware: middleware,
	}

	store.dispatch = store.applyMiddleware(store.dispatchInternal())
	return store
}

func (s *Store) dispatchInternal() Dispatch {
	return func(action Action) {
		s.state = s.reducer(s.state, action)
	}
}

func (s *Store) applyMiddleware(dispatch Dispatch) Dispatch {
	for _, middleware := range s.middleware {
		dispatch = middleware(s, dispatch)
	}
	return dispatch
}

func (s *Store) Dispatch(action Action) {
	s.dispatch(action)
}

func (s *Store) GetState() State {
	return s.state
}

// ViewModel
type ViewModel struct {
	store *Store
}

func NewViewModel(store *Store) *ViewModel {
	vm := &ViewModel{
		store: store,
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
	store := NewStore(reduce, State{Count: 0}, LoggingMiddleware)
	viewModel := NewViewModel(store)

	var ops op.Ops
	view := NewView(viewModel, th)

	// log.Printf("Initial state: %+v", store.GetState())

	for {
		switch e := w.Event().(type) {
		case app.DestroyEvent:
			return e.Err
		case app.FrameEvent:
			gtx := app.NewContext(&ops, e)

			view.Layout(gtx)
			e.Frame(gtx.Ops)
		}
	}
}

type View struct {
	viewModel       *ViewModel
	theme           *material.Theme
	incrementButton widget.Clickable
	decrementButton widget.Clickable
}

func NewView(vm *ViewModel, theme *material.Theme) *View {
	return &View{
		viewModel:       vm,
		theme:           theme,
		incrementButton: widget.Clickable{},
		decrementButton: widget.Clickable{},
	}
}

// Layout accepts theme
func (v *View) Layout(gtx layout.Context) layout.Dimensions {
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
				return material.Button(v.theme, &v.incrementButton, "Increment").Layout(gtx)
			}),
			layout.Rigid(layout.Spacer{Width: unit.Dp(20)}.Layout),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				paint.ColorOp{Color: color.NRGBA{R: 0, G: 0, B: 0, A: 255}}.Add(gtx.Ops)
				m := material.Body1(v.theme, v.viewModel.CountLabel())
				m.Font.Weight = font.Bold
				return m.Layout(gtx)
			}),
			layout.Rigid(layout.Spacer{Width: unit.Dp(20)}.Layout),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				if v.decrementButton.Clicked(gtx) {
					v.viewModel.Decre()
				}
				return material.Button(v.theme, &v.decrementButton, "Decrement").Layout(gtx)
			}),
		)
	})
}
