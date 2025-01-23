package event_fsm

import "errors"

var (
	ErrStateNotFound          = errors.New("state not found")
	ErrStateNameAlreadyExists = errors.New("state name already exists")
	ErrStateNameNotFound      = errors.New("state name not found")
)
