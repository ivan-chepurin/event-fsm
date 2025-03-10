package event_fsm

import (
	"context"
	"encoding/json"
	"fmt"
)

type ResultStatus int

const (
	Fail ResultStatus = -1
)

type StateType int

const (
	StateTypeTransition StateType = iota + 1
	StateTypeWaitEvent
)

type StateName string

var _stateNames = map[string]StateName{}

// State result statuses
func NewStateName(name string) StateName {
	if sn, ok := _stateNames[name]; ok {
		return sn
	}

	sn := StateName(name)
	_stateNames[name] = sn

	return sn
}

func GetStateName(name string) StateName {
	if sn, ok := _stateNames[name]; ok {
		return sn
	}

	return NewStateName("")
}

func (sn *StateName) String() string {
	return string(*sn)
}

// MarshalJSON check if state name is valid or not before marshalling with Alias
func (sn *StateName) MarshalJSON() ([]byte, error) {
	if _, ok := _stateNames[sn.String()]; !ok {
		return nil, fmt.Errorf("state name %s not found", sn.String())
	}

	return json.Marshal(sn.String())
}

// UnmarshalJSON unmarshal eventData to string, then check if it is a valid state name or not
func (sn *StateName) UnmarshalJSON(data []byte) error {
	var name string
	if err := json.Unmarshal(data, &name); err != nil {
		return err
	}

	if _, ok := _stateNames[name]; !ok {
		return fmt.Errorf("state name %s not found", name)
	}

	*sn = StateName(name)

	return nil
}

type Executor[T comparable] interface {
	Execute(ctx context.Context, e T) (ResultStatus, error)
}

type State[T comparable] struct {
	Name StateName

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

// StateDetector is a state detector
type StateDetector[T comparable] struct {
	states    map[string]*State[T]
	mainState *State[T]
}

func NewStateDetector[T comparable]() *StateDetector[T] {
	return &StateDetector[T]{
		states: make(map[string]*State[T]),
	}
}

func (sd *StateDetector[T]) NewState(name StateName, executor Executor[T], stateType StateType) *State[T] {
	if _, ok := _stateNames[name.String()]; !ok {
		panic(ErrStateNameNotFound)
	}

	state := &State[T]{
		Name:      name,
		Executor:  executor,
		Next:      make(map[ResultStatus]*State[T]),
		StateType: stateType,
	}
	sd.states[name.String()] = state

	return state
}

func (sd *StateDetector[T]) SetMainState(state *State[T]) {
	sd.mainState = state
}

func (sd *StateDetector[T]) getMainState() (*State[T], error) {
	if sd.mainState == nil {
		return nil, ErrMainStateNotFound
	}
	return sd.mainState, nil
}

func (sd *StateDetector[T]) GetStateByName(name StateName) (*State[T], error) {
	if state, ok := sd.states[name.String()]; ok {
		return state, nil
	}

	return nil, ErrStateNotFound
}

func (sd *StateDetector[T]) getNextState(state *State[T], response ResultStatus) (*State[T], bool) {
	if nextState, ok := state.Next[response]; ok {

		if sd.states[nextState.Name.String()] != nil {
			return nextState, true
		}
	}

	return nil, false
}
