package event_fsm

import (
	"context"
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

type ResultStatus string

var _resultStatuses = map[string]ResultStatus{}

// NewResultStatus creates a new ResultStatus instance
func NewResultStatus(status string) ResultStatus {
	if rs, ok := _resultStatuses[status]; ok {
		return rs
	}

	rs := ResultStatus(status)
	_resultStatuses[status] = rs

	return rs
}

func (r *ResultStatus) String() string {
	return string(*r)
}

func (r *ResultStatus) MarshalJSON() ([]byte, error) {
	if _, ok := _resultStatuses[r.String()]; !ok {
		return nil, fmt.Errorf("result status %s not found", r.String())
	}

	return json.Marshal(r.String())
}

func (r *ResultStatus) UnmarshalJSON(data []byte) error {
	var status string
	if err := json.Unmarshal(data, &status); err != nil {
		return err
	}

	if status == "" {
		return fmt.Errorf("result status is empty")
	}

	*r = ResultStatus(status)

	return nil
}

func (r *ResultStatus) Scan(value interface{}) error {
	if str, ok := value.(string); ok {
		*r = ResultStatus(str)
		return nil
	}

	return fmt.Errorf("cannot convert %v to string", value)
}

func (r *ResultStatus) Value() (driver.Value, error) {
	if r.String() == "" {
		return nil, fmt.Errorf("result status is empty")
	}

	return r.String(), nil
}

var (
	ResultStatusFail = NewResultStatus("fail")
	ResultStatusOk   = NewResultStatus("ok")
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

func ToStateName(name string) StateName {
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

func (sn *StateName) Scan(value interface{}) error {
	if str, ok := value.(string); ok {
		*sn = StateName(str)
		return nil
	}

	return fmt.Errorf("cannot convert %v to string", value)
}

func (sn *StateName) Value() (driver.Value, error) {
	if sn.String() == "" {
		return nil, fmt.Errorf("state name is empty")
	}

	return sn.String(), nil
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
