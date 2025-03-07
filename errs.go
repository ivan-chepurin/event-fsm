package event_fsm

import "errors"

var (
	ErrStateNotFound          = errors.New("state not found")
	ErrStateNameAlreadyExists = errors.New("state name already exists")
	ErrStateNameNotFound      = errors.New("state name not found")
	ErrStateNameEmpty         = errors.New("state name is empty")
	ErrMainStateNotFound      = errors.New("main state not found")
	ErrNoNextState            = errors.New("no next state found")
)
