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
type StateDetector struct {
	States map[string]*State
}

func NewStateDetector() *StateDetector {
	return &StateDetector{
		States: make(map[string]*State),
	}
}

func (sd *StateDetector) NewState(name string, usecase Executor, stateType StateType) *State {
	state := &State{
		Name:      name,
		Executor:  usecase,
		Next:      make(map[ResultStatus]*State),
		StateType: stateType,
	}
	sd.States[name] = state
	return state
}

func (sd *StateDetector) GetStateByName(name string) (*State, error) {
	if state, ok := sd.States[name]; ok {
		return state, nil
	}

	return nil, ErrStateNotFound
}

func (sd *StateDetector) getNextState(state *State, response ResultStatus) (*State, error) {
	if nextState, ok := state.Next[response]; ok {

		if sd.States[nextState.Name] != nil {
			return nextState, nil
		}
	}

	return nil, ErrStateNotFound
}

type Executor func(ctx context.Context, e Event) (ResultStatus, error)

type State struct {
	Name string

	StateType StateType

	Next map[ResultStatus]*State

	Executor Executor
}

func (s *State) SetNext(nextState *State, response ResultStatus) {
	s.Next[response] = nextState
}

func (s *State) getNext(response ResultStatus) (*State, error) {
	if state, ok := s.Next[response]; ok {
		return state, nil
	}
	return nil, ErrStateNotFound
}
