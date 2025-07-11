package event_fsm

import (
	"context"
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

	// GetState returns the state of the object, whose state is being processed
	GetState() StateName

	// SetState sets the state of the object, whose state is being processed
	SetState(state StateName)

	// Save saves the current state of the object, whose state is being processed
	Save(ctx context.Context) error

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
		id:          data.ID(),
		data:        data,
		stateResult: ResultStatusEmpty,
	}
}

func (e *Target[T]) Data() (T, bool) {
	return e.data.Data(), !e.data.IsNull()
}

func (e *Target[T]) ID() string {
	return e.id
}

func (e *Target[T]) event() Event {
	return Event{
		ID:               e.eventID,
		TargetID:         e.data.ID(),
		LastResultStatus: e.stateResult,
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

func (e *Target[T]) setStateName(state *State[T]) {
	e.data.SetState(state.Name)
}

func (e *Target[T]) save(ctx context.Context) error {
	if err := e.data.Save(ctx); err != nil {
		return err
	}

	return nil
}

func (e *Target[T]) getStateName() StateName {
	return e.data.GetState()
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
