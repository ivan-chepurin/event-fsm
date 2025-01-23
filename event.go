package event_fsm

import "fmt"

// EventType is a type of event that can be processed by usecase
type EventType string

type EventData[T comparable] interface {
	Data() T
	IsNull() bool
}

// Event is a struct that represents an event that can be processed by usecase
type Event[T comparable] struct {
	id string

	stateName StateName
	state     *State[T]
	prevState *State[T]

	data EventData[T]

	et EventType
}

func NewEvent[T comparable](id string, stateName StateName, data EventData[T], eventType EventType) Event[T] {
	return Event[T]{
		id:        id,
		stateName: stateName,
		data:      data,
		et:        eventType,
	}
}

func (e *Event[T]) Data() (T, bool) {
	return e.data.Data(), e.data.IsNull()
}

func (e *Event[T]) ID() string {
	return e.id
}

func (e *Event[T]) Type() EventType {
	return e.et
}

func (e *Event[T]) GetLog() string {
	log := fmt.Sprintf("id: %s, stateName: %s, type: %s", e.id, e.state.Name, e.et)
	return log
}

func (e *Event[T]) CurrentStateName() StateName {
	return e.state.Name
}

func (e *Event[T]) NextStateName(status ResultStatus) (StateName, error) {
	if state, err := e.state.getNext(status); err != nil {
		return "", err
	} else {
		return state.Name, nil
	}
}

type Result struct {
	Status ResultStatus
}
