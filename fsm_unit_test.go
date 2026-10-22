package event_fsm

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"go.uber.org/zap"
)

type fakeStore struct {
	events []Event
	logs   []Log
}

func (s *fakeStore) saveEvent(_ context.Context, event Event) (string, error) {
	s.events = append(s.events, event)
	return event.ID, nil
}

func (s *fakeStore) updateEvent(_ context.Context, event Event) error {
	s.events = append(s.events, event)
	return nil
}

func (s *fakeStore) saveLog(_ context.Context, log Log) (string, error) {
	s.logs = append(s.logs, log)
	return "log-id", nil
}

func (s *fakeStore) createFullLog(_ context.Context, log Log) (string, error) {
	s.logs = append(s.logs, log)
	return "log-id", nil
}

func (s *fakeStore) updateLog(_ context.Context, log Log) error {
	s.logs = append(s.logs, log)
	return nil
}

type funcExecutor func(ctx context.Context, data *eventData) (ResultStatus, error)

func (f funcExecutor) Execute(ctx context.Context, data *eventData) (ResultStatus, error) {
	return f(ctx, data)
}

var (
	unitStateStart = NewStateName("unit_start")
	unitStateWait  = NewStateName("unit_wait")
)

// newUnitFSM builds an FSM with start -> (ok) -> wait, where start runs exec.
func newUnitFSM(exec funcExecutor) *FSM[*eventData] {
	sd := NewStateDetector[*eventData]()
	start := sd.NewState(unitStateStart, exec, StateTypeTransition)
	wait := sd.NewState(unitStateWait, funcExecutor(func(context.Context, *eventData) (ResultStatus, error) {
		return ResultStatusOk, nil
	}), StateTypeWaitEvent)
	start.SetNext(wait, ResultStatusOk)
	sd.SetMainState(unitStateStart)

	return &FSM[*eventData]{
		l:             zap.NewNop(),
		store:         &fakeStore{},
		stateDetector: sd,
	}
}

func okExecutor(context.Context, *eventData) (ResultStatus, error) {
	return ResultStatusOk, nil
}

func TestProcessEventTransitionsToWaitState(t *testing.T) {
	f := newUnitFSM(okExecutor)
	ed := newEventData(1)

	if _, err := f.ProcessEvent(context.Background(), NewTarget(&ed)); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if ed.StateName != unitStateWait {
		t.Fatalf("expected state %s, got %s", unitStateWait, ed.StateName)
	}
}

func TestProcessEventNilTarget(t *testing.T) {
	f := newUnitFSM(okExecutor)

	var typedNil *eventData
	targets := map[string]Target[*eventData]{
		"nil interface": NewTarget[*eventData](nil),
		"typed nil":     NewTarget[*eventData](typedNil),
		"zero target":   {},
	}

	for name, target := range targets {
		t.Run(name, func(t *testing.T) {
			if _, ok := target.Data(); ok {
				t.Fatal("expected Data() to report no data")
			}

			if _, err := f.ProcessEvent(context.Background(), target); !errors.Is(err, ErrTargetIsNil) {
				t.Fatalf("expected ErrTargetIsNil, got %v", err)
			}
		})
	}
}

func TestProcessEventPanicReturnsError(t *testing.T) {
	f := newUnitFSM(func(context.Context, *eventData) (ResultStatus, error) {
		panic("boom")
	})
	ed := newEventData(1)

	_, err := f.ProcessEvent(context.Background(), NewTarget(&ed))
	if !errors.Is(err, ErrPanic) {
		t.Fatalf("expected ErrPanic, got %v", err)
	}
}

func TestProcessEventExecutorErrorIsReturned(t *testing.T) {
	execErr := errors.New("executor failed")

	tests := []struct {
		name    string
		status  ResultStatus
		wantErr error
	}{
		{name: "fail status", status: ResultStatusFail, wantErr: ErrStateFailed},
		{name: "status without transition", status: ResultStatusEmpty, wantErr: ErrNoNextState},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newUnitFSM(func(context.Context, *eventData) (ResultStatus, error) {
				return tt.status, execErr
			})
			ed := newEventData(1)

			_, err := f.ProcessEvent(context.Background(), NewTarget(&ed))
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("expected %v, got %v", tt.wantErr, err)
			}
			if !errors.Is(err, execErr) {
				t.Fatalf("expected executor error in chain, got %v", err)
			}
		})
	}
}

func TestProcessEventStopsOnCanceledContext(t *testing.T) {
	called := false
	f := newUnitFSM(func(context.Context, *eventData) (ResultStatus, error) {
		called = true
		return ResultStatusOk, nil
	})
	ed := newEventData(1)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := f.ProcessEvent(ctx, NewTarget(&ed))
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", err)
	}
	if called {
		t.Fatal("executor must not run after context cancellation")
	}
}

func TestDTOColumnsMatchSchema(t *testing.T) {
	field, _ := reflect.TypeOf(eventDto{}).FieldByName("TargetID")
	if tag := field.Tag.Get("db"); tag != "target_id" {
		t.Fatalf("eventDto.TargetID db tag = %q, want %q", tag, "target_id")
	}
}
