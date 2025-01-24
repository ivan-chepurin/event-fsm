package event_fsm

import (
	"context"
	"encoding/json"
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

// UnmarshalJSON unmarshal data to string, then check if it is a valid state name or not
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

func ToStateName(name string) (StateName, error) {
	if sn, ok := _stateNames[name]; ok {
		return &sn, nil
	}

	return nil, ErrStateNameNotFound
}

type Executor[T comparable] interface {
	Execute(ctx context.Context, e Event[T]) (ResultStatus, []error)
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
	States map[StateName]*State[T]
}

func NewStateDetector[T comparable]() *StateDetector[T] {
	return &StateDetector[T]{
		States: make(map[StateName]*State[T]),
	}
}

func (sd *StateDetector[T]) NewState(name StateName, execotor Executor[T], stateType StateType) *State[T] {
	if _, ok := _stateNames[name.String()]; !ok {
		panic(ErrStateNameNotFound)
	}

	state := &State[T]{
		Name:      name,
		Executor:  execotor,
		Next:      make(map[ResultStatus]*State[T]),
		StateType: stateType,
	}
	sd.States[name] = state
	return state
}

func (sd *StateDetector[T]) GetStateByName(name StateName) (*State[T], error) {
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
