package event_fsm

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

const (
	searchPath = "fsm"
)

type FSM[T comparable] struct {
	l *zap.Logger

	store *storage

	stateDetector *StateDetector[T]
}

func NewFSM[T comparable](cfg *Config[T]) (*FSM[T], error) {
	if err := cfg.check(); err != nil {
		return nil, fmt.Errorf("cfg.check() failed: %w", err)
	}

	dbConn, err := initDB(cfg)
	if err != nil {
		return nil, fmt.Errorf("initDB failed: %w", err)
	}

	db := newDBStore(dbConn)

	rdb, err := initRedis(cfg)
	if err != nil {
		return nil, fmt.Errorf("initRedis failed: %w", err)
	}

	return &FSM[T]{
		stateDetector: cfg.StateDetector,
		l:             cfg.Logger,

		store: newStorage(cfg.Logger, cfg.AppLabel, db, rdb),
	}, nil
}

func (f *FSM[T]) ProcessEvent(ctx context.Context, t Target[T]) (Target[T], error) {
	// check if the target is nil
	if t.data.IsNull() {
		return t, fmt.Errorf("target is nil")
	}

	// determine current state
	currentStateName := t.getStateName()

	if ok, err := checkStateName(currentStateName); !ok {
		if errors.Is(err, ErrStateNotFound) {
			return t, fmt.Errorf("invalid state name: %s, %w", currentStateName, ErrStateNotFound)
		}

		currentStateName, err = f.stateDetector.getMainState()
		if err != nil {
			return t, fmt.Errorf("f.stateDetector.getMainState: %w", err)
		}
	}

	var (
		err error
	)

	if t.state, err = f.stateDetector.stateByName(currentStateName); err != nil {
		return t, fmt.Errorf("%v: %w, state: %s", ErrStateNotFound, err, currentStateName)
	}

	t.eventID = uuid.NewString()
	t.eventID, err = f.store.saveEvent(ctx, t.event())
	if err != nil {
		return t, fmt.Errorf("f.store.saveEvent: %w", err)
	}

	return f.processEvent(ctx, t)
}

func (f *FSM[T]) processEvent(ctx context.Context, t Target[T]) (nt Target[T], err error) {
	defer func() {
		if err := recover(); err != nil {
			f.l.Error("panic in FSM", zap.Any("err", err))

			nt = t
		}
	}()

	var (
		ok bool
	)

	for {
		id, err := f.store.saveLog(ctx, t.log())
		if err != nil {
			return t, fmt.Errorf("f.store.createLog: %w", err)
		}

		t.stateResult, err = t.state.Executor.Execute(ctx, t.data.Data())
		if err != nil {
			f.l.Error("error executing state", zap.Error(err), zap.String("state", t.state.Name.String()))
		}

		log := t.log()
		log.ID = id
		if err = f.store.updateLog(ctx, log); err != nil {
			return t, fmt.Errorf("f.store.updateLog: %w", err)
		}

		if err = f.store.updateEvent(ctx, t.event()); err != nil {
			return t, fmt.Errorf("f.store.updateEvent: %w", err)
		}

		if t.stateResult == ResultStatusFail {
			return t, fmt.Errorf("state execution failed: %s", t.state.Name)
		}

		t.state, ok = f.stateDetector.getNextState(t.state, t.stateResult)
		if !ok {
			return t, fmt.Errorf("no next state for %s: %w", log.CurrentStateName, ErrNoNextState)
		}

		if err = t.setStateName(ctx, t.state); err != nil {
			return t, fmt.Errorf("t.setStateName: %w", err)
		}

		if t.state.StateType == StateTypeWaitEvent {
			// wait for the next event
			t.stateResult = resultStatusWaitNextEvent
			if _, err = f.store.createFullLog(ctx, t.log()); err != nil {
				return t, fmt.Errorf("f.store.createLog: %w", err)
			}

			return t, nil
		}
	}
}
