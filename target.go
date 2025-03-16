package event_fsm

import (
	"encoding/json"
	"time"
)

type TargetData[T comparable] interface {
	// Data returns the data of the object, whose state is being processed
	Data() T

	// IsNull true if the data is null
	IsNull() bool

	// ID unique ID of the object, whose state is being processed
	// You need to use the user ID, or any other object whose state is tracked by the FSM.
	ID() string

	// MetaInfo additional information about the event of the object, JSON format
	MetaInfo() json.RawMessage
}

// Target is a struct that represents an event that can be processed by usecase
type Target[T comparable] struct {
	id      string
	eventID string

	state       *State[T]
	stateResult ResultStatus

	data TargetData[T]
}

func NewTarget[T comparable](data TargetData[T]) Target[T] {
	return Target[T]{
		id:   data.ID(),
		data: data,
	}
}

func (e *Target[T]) Data() (T, bool) {
	return e.data.Data(), e.data.IsNull()
}

func (e *Target[T]) ID() string {
	return e.id
}

func (e *Target[T]) event() Event {
	return Event{
		ID:               e.eventID,
		TargetID:         e.data.ID(),
		LastResultStatus: NewResultStatus(""),
		MetaInfo:         e.data.MetaInfo(),
	}
}

func (e *Target[T]) log() Log {
	return Log{
		TargetID:            e.data.ID(),
		EventID:             e.eventID,
		CurrentStateName:    e.state.Name,
		CurrentResultStatus: e.stateResult,
		CreatedAt:           time.Now(),
		UpdatedAt:           time.Now(),
	}
}

func (e *Target[T]) currentStateName() StateName {
	return e.state.Name
}

func (e *Target[T]) nextStateName(status ResultStatus) (StateName, error) {
	if state, err := e.state.getNext(status); err == nil {
		return state.Name, nil
	} else {
		return "", err
	}
}

type Result struct {
	Status ResultStatus
}
