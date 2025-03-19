package event_fsm

import "errors"

var (
	ErrEmptyStateName    = errors.New("empty state")
	ErrStateNotFound     = errors.New("state not found")
	ErrStateNameNotFound = errors.New("state name not found")
	ErrMainStateNotFound = errors.New("main state not found")
	ErrNoNextState       = errors.New("no next state found")
	ErrLastLogNotFound   = errors.New("last log not found")
)
