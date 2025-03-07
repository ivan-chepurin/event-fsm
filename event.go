package event_fsm

import "fmt"

type EventData[T comparable] interface {
	Data() T
	IsNull() bool
	StateName() (StateName, error)
	SetStateName(StateName)
}

// Event is a struct that represents an event that can be processed by usecase
type Event[T comparable] struct {
	id string

	state     *State[T]
	prevState *State[T]

	data EventData[T]
}

func NewEvent[T comparable](id string, data EventData[T]) Event[T] {
	return Event[T]{
		id:   id,
		data: data,
	}
}

func (e *Event[T]) Data() (T, bool) {
	return e.data.Data(), e.data.IsNull()
}

func (e *Event[T]) ID() string {
	return e.id
}

func (e *Event[T]) GetLog() string {
	log := fmt.Sprintf(
		"id: %s, prevStateName: %s, currentStateName: %s",
		e.id,
		e.prevState.Name,
		e.state.Name,
	)
	return log
}

func (e *Event[T]) CurrentStateName() StateName {
	return e.state.Name
}

func (e *Event[T]) NextStateName(status ResultStatus) (StateName, error) {
	if state, err := e.state.getNext(status); err == nil {
		return state.Name, nil
	} else {
		return nil, err
	}
}

type Result struct {
	Status ResultStatus
}
