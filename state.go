package event_fsm

import "context"

type ResultStatus int

const (
	Fail ResultStatus = -1
)

type StateType int

const (
	StateTypeTransition StateType = iota + 1
	StateTypeWaitEvent
)

// StateDetector is a state detector
type StateDetector[T comparable] struct {
	States map[string]*State[T]
}

func NewStateDetector[T comparable]() *StateDetector[T] {
	return &StateDetector[T]{
		States: make(map[string]*State[T]),
	}
}

func (sd *StateDetector[T]) NewState(name string, usecase Executor[T], stateType StateType) *State[T] {
	state := &State[T]{
		Name:      name,
		Executor:  usecase,
		Next:      make(map[ResultStatus]*State[T]),
		StateType: stateType,
	}
	sd.States[name] = state
	return state
}

func (sd *StateDetector[T]) GetStateByName(name string) (*State[T], error) {
	if state, ok := sd.States[name]; ok {
		return state, nil
	}

	return nil, ErrStateNotFound
}

func (sd *StateDetector[T]) getNextState(state *State[T], response ResultStatus) (*State[T], error) {
	if nextState, ok := state.Next[response]; ok {

		if sd.States[nextState.Name] != nil {
			return nextState, nil
		}
	}

	return nil, ErrStateNotFound
}

//	type Event [T comparable]struct {
//		id string
//
//		stateName string
//		state     *State
//		prevState *State
//
//		data T
//
//		et EventType
//	}
type Executor[T comparable] func(ctx context.Context, e Event[T]) (ResultStatus, error)

type State[T comparable] struct {
	Name string

	StateType StateType

	Next map[ResultStatus]*State[T]

	Executor Executor[T]
}

func (s *State[T]) SetNext(nextState *State[T], response ResultStatus) {
	s.Next[response] = nextState
}

func (s *State[T]) getNext(response ResultStatus) (*State[T], error) {
	if state, ok := s.Next[response]; ok {
		return state, nil
	}
	return nil, ErrStateNotFound
}
