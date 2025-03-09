package event_fsm

import (
	"context"
	"fmt"

	"go.uber.org/zap"
)

type FSM[T comparable] struct {
	stateDetector *StateDetector[T]

	l *zap.Logger
}

func NewFSM[T comparable](sd *StateDetector[T], l *zap.Logger) *FSM[T] {
	return &FSM[T]{
		stateDetector: sd,
		l:             l,
	}
}

func (f *FSM[T]) ProcessEvent(ctx context.Context, e Event[T]) (Event[T], error) {
	var (
		err error
	)

	sn := e.data.StateName()
	if sn.String() == "" {

		ms, err := f.stateDetector.getMainState()
		if err != nil {
			return e, fmt.Errorf("f.stateDetector.getMainState: %v: %w", ErrStateNotFound, err)
		}

		fmt.Println("Main state:", ms.Name)
		sn = ms.Name
	}

	if e.state, err = f.stateDetector.GetStateByName(sn); err != nil {
		return e, fmt.Errorf("%v: %w, state: %s", ErrStateNotFound, err, sn)
	}

	return f.processEvent(ctx, e)
}

func (f *FSM[T]) processEvent(ctx context.Context, e Event[T]) (Event[T], error) {
	var (
		status ResultStatus
		pErr   error
		ok     bool
	)

	for {
		if e.data.IsNull() {
			return e, fmt.Errorf("eventData is null")
		}

		status, pErr = e.state.Executor.Execute(ctx, e.data.Data())
		if pErr != nil {
			f.l.Error("error in usecase", zap.Error(pErr))
		}

		if status == Fail {
			return e, fmt.Errorf("usecase failed: %s", e.GetLog())
		}

		e.prevState = e.state

		fmt.Printf("State: %s, status: %d \n", e.state.Name, status)

		e.state, ok = f.stateDetector.getNextState(e.state, status)
		if !ok {
			return e, fmt.Errorf("no next state for %s: %w", e.prevState.Name, ErrNoNextState)
		}

		e.data.SetStateName(e.state.Name)

		if e.state.StateType == StateTypeWaitEvent {
			return e, nil
		}
	}
}
