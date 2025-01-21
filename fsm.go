package event_fsm

import (
	"context"
	"fmt"

	"go.uber.org/zap"
)

type FSM struct {
	StateDetector *StateDetector

	l *zap.Logger
}

func NewFSM(st *StateDetector, l *zap.Logger) *FSM {
	return &FSM{
		StateDetector: st,
		l:             l,
	}
}

func (f *FSM) ProcessEvent(ctx context.Context, e Event) (bool, error) {
	var (
		err error
	)

	if e.state, err = f.StateDetector.GetStateByName(e.stateName); err != nil {
		return false, fmt.Errorf("%w: %w, state: %s", ErrStateNotFound, err, e.stateName)
	}

	return f.processEvent(ctx, e)
}

func (f *FSM) processEvent(ctx context.Context, e Event) (bool, error) {
	var (
		status   ResultStatus
		newState *State
		err      error
	)

	for {
		status, err = e.state.Executor(ctx, e)
		if err != nil {
			f.l.Error("error in usecase", zap.Error(err))
		}

		if status == Fail {
			return false, fmt.Errorf("usecase failed: %s", e.GetLog())
		}

		newState, err = f.StateDetector.getNextState(e.state, status)
		if err != nil || newState == nil {
			return true, fmt.Errorf(
				"f.StateDetector.getNextState: %w, event: %s, lastStatus: %d",
				err, e.GetLog(), status,
			)
		}

		e.prevState = e.state
		e.state = newState

		if newState.StateType == StateTypeWaitEvent {
			return true, nil
		}
	}
}
