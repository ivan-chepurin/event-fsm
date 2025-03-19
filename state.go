package event_fsm

import (
	"context"
)

type StateType int

const (
	StateTypeTransition StateType = iota + 1
	StateTypeWaitEvent
)

type Executor[T comparable] interface {
	Execute(ctx context.Context, e T) (ResultStatus, error)
}

type State[T comparable] struct {
	Name StateName

	StateType StateType

	Next map[string]*State[T]

	Executor Executor[T]
}

func (s *State[T]) SetNext(nextState *State[T], response ResultStatus) {
	s.Next[response.String()] = nextState
}

func (s *State[T]) getNext(response ResultStatus) (*State[T], error) {
	if state, ok := s.Next[response.String()]; ok {
		return state, nil
	}
	return nil, ErrStateNotFound
}
