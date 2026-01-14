package main

import (
	"fmt"
	"log"
	"sync"

	"gioui.org/app"
	"gioui.org/font"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
)

// StateProvider interface for state types
type StateProvider[T any] interface {
	Copy() T
}

// State
type State struct {
	Count int
}

func (s State) Copy() State {
	return State{
		Count: s.Count,
	}
}

// Action
type Action[S any] interface {
	Apply(s S) S
}

// IncrementAction
type IncrementAction struct{}

func (a IncrementAction) Apply(s State) State {
	state := s.Copy()
	state.Count++
	return state
}

// DecrementAction
type DecrementAction struct{}

func (a DecrementAction) Apply(s State) State {
	state := s.Copy()
	state.Count--
	return state
}

// Reducer type
type Reducer[S any, A Action[S]] func(state S, action A) S

// Middleware type
type Middleware[S StateProvider[S], A Action[S]] func(store *Store[S, A], next Dispatch[A]) Dispatch[A]
type Dispatch[A any] func(action A)

// Logging Middleware
func LoggingMiddleware[S StateProvider[S], A Action[S]](store *Store[S, A], next Dispatch[A]) Dispatch[A] {
	return func(action A) {
		prevState := store.GetState()
		log.Printf("Action dispatched: %T, Previous State: %+v", action, prevState)
		next(action)
		newState := store.GetState()
		log.Printf("Action dispatched: %T, New State: %+v", action, newState)
	}
}

// Store
type Store[S StateProvider[S], A Action[S]] struct {
	mu          sync.RWMutex
	state       S
	reducer     Reducer[S, A]
	middleware  []Middleware[S, A]
	dispatch    Dispatch[A]
	subscribers []func()
}

func NewStore[S StateProvider[S], A Action[S]](
	reducer Reducer[S, A],
	initialState S,
	middleware ...Middleware[S, A],
) *Store[S, A] {
	store := &Store[S, A]{
		state:       initialState,
		reducer:     reducer,
		middleware:  middleware,
		subscribers: []func(){},
	}

	store.dispatch = store.applyMiddleware(store.dispatchInternal())
	return store
}

func (s *Store[S, A]) dispatchInternal() Dispatch[A] {
	return func(action A) {
		s.mu.Lock()
		s.state = s.reducer(s.state, action)
		s.mu.Unlock()

		// Notify subscribers after state update
		s.mu.RLock()
		subscribers := s.subscribers
		s.mu.RUnlock()

		for _, sub := range subscribers {
			sub()
		}
	}
}

func (s *Store[S, A]) applyMiddleware(dispatch Dispatch[A]) Dispatch[A] {
	// Apply in reverse order so first middleware is outermost
	for i := len(s.middleware) - 1; i >= 0; i-- {
		dispatch = s.middleware[i](s, dispatch)
	}
	return dispatch
}

func (s *Store[S, A]) Dispatch(action A) {
	s.dispatch(action)
}

func (s *Store[S, A]) GetState() S {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.state.Copy()
}

func (s *Store[S, A]) Subscribe(fn func()) func() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.subscribers = append(s.subscribers, fn)

	// Return unsubscribe function
	index := len(s.subscribers) - 1
	return func() {
		s.mu.Lock()
		defer s.mu.Unlock()
		if index < len(s.subscribers) {
			s.subscribers = append(s.subscribers[:index], s.subscribers[index+1:]...)
		}
	}
}

// AppAction is the union type for all actions in this app
type AppAction interface {
	Action[State]
}

// Reducer function
func reduce(state State, action AppAction) State {
	return action.Apply(state)
}

// ViewModel
type ViewModel struct {
	store *Store[State, AppAction]
}

func NewViewModel(store *Store[State, AppAction]) *ViewModel {
	return &ViewModel{
		store: store,
	}
}

func (v *ViewModel) Incre() {
	v.store.Dispatch(IncrementAction{})
}

func (v *ViewModel) Decre() {
	v.store.Dispatch(DecrementAction{})
}

func (v *ViewModel) CountLabel() string {
	return fmt.Sprintf("%d", v.store.GetState().Count)
}

func main() {
	go func() {
		w := &app.Window{}
		w.Option(
			app.Title("Counter App"),
			app.Size(unit.Dp(400), unit.Dp(200)),
			app.MinSize(unit.Dp(300), unit.Dp(100)),
		)
		if err := run(w); err != nil {
			log.Fatal(err)
		}
	}()
	app.Main()
}

func run(w *app.Window) error {
	th := material.NewTheme()
	store := NewStore(reduce, State{Count: 0}, LoggingMiddleware[State, AppAction])
	viewModel := NewViewModel(store)

	var ops op.Ops
	view := NewView(viewModel, th)

	// Subscribe to store changes and invalidate window
	store.Subscribe(func() {
		w.Invalidate()
	})

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

func (v *View) Layout(gtx layout.Context) layout.Dimensions {
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
				label := material.Body1(v.theme, v.viewModel.CountLabel())
				label.Font.Weight = font.Bold
				return label.Layout(gtx)
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
