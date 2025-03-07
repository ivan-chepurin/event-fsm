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

type StateName interface {
	json.Marshaler
	json.Unmarshaler
	String() string
}

type stateName string

var _stateNames = map[string]stateName{}

// State result statuses
func NewStateName(name string) StateName {
	if sn, ok := _stateNames[name]; !ok {
		sn = stateName(name)
		_stateNames[name] = sn
		return &sn
	}

	panic(ErrStateNameAlreadyExists)
}

func GetStateName(name string) (StateName, error) {
	if sn, ok := _stateNames[name]; ok {
		return &sn, nil
	}

	if name == "" {
		return nil, ErrStateNameEmpty
	}

	return nil, ErrStateNameNotFound
}

func (sn stateName) String() string {
	return string(sn)
}

// MarshalJSON check if state name is valid or not before marshalling with Alias
func (sn stateName) MarshalJSON() ([]byte, error) {
	if _, ok := _stateNames[string(sn)]; ok {
		type Alias stateName
		return json.Marshal(&struct {
			*Alias
		}{
			Alias: (*Alias)(&sn),
		})
	}

	return nil, ErrStateNameNotFound
}

// UnmarshalJSON unmarshal eventData to string, then check if it is a valid state name or not
func (sn *stateName) UnmarshalJSON(data []byte) error {
	var name string
	if err := json.Unmarshal(data, &name); err != nil {
		return err
	}

	if msn, ok := _stateNames[name]; ok {
		*sn = msn
		return nil
	}

	return ErrStateNameNotFound
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

	fmt.Printf("State %s created\n", name.String())

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
