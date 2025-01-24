package event_fsm

import (
	"context"
	"fmt"

	"go.uber.org/zap"
)

type FSM[T comparable] struct {
	StateDetector *StateDetector[T]

	l *zap.Logger
}

func NewFSM[T comparable](st *StateDetector[T], l *zap.Logger) *FSM[T] {
	return &FSM[T]{
		StateDetector: st,
		l:             l,
	}
}

func (f *FSM[T]) ProcessEvent(ctx context.Context, e Event[T]) (Event[T], bool, error) {
	var (
		err error
	)

	if e.state, err = f.StateDetector.GetStateByName(e.stateName); err != nil {
		return e, false, fmt.Errorf("%w: %w, state: %s", ErrStateNotFound, err, e.stateName)
	}

	return f.processEvent(ctx, e)
}

func (f *FSM[T]) processEvent(ctx context.Context, e Event[T]) (Event[T], bool, error) {
	var (
		status   ResultStatus
		newState *State[T]
		err      error
		errs     []error
	)

	for {
		status, errs = e.state.Executor.Execute(ctx, e)
		if len(errs) > 0 {
			for _, sErr := range errs {
				f.l.Error("error in usecase", zap.Error(sErr))
			}
		}

		if status == Fail {
			return e, false, fmt.Errorf("usecase failed: %s", e.GetLog())
		}

		e.prevState = e.state

		e.state, err = f.StateDetector.getNextState(e.state, status)
		if err != nil {

			return e, true, fmt.Errorf(
				"f.StateDetector.getNextState: %w, event: %s, lastStatus: %d",
				err, e.GetLog(), status,
			)
		}

		if newState.StateType == StateTypeWaitEvent {
			return e, true, nil
		}
	}
}
