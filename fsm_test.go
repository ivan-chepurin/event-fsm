package event_fsm

import (
	"context"
	"encoding/json"
	"strconv"
	"testing"
)

type eventData struct {
	InitNumber int
	Number     int
	ManualAdd  int
	StateName  StateName
}

func newEventData(initData int) eventData {
	return eventData{
		InitNumber: initData,
		Number:     initData,
		ManualAdd:  0,
	}
}

func (e *eventData) Data() *eventData {
	return e
}

func (e *eventData) IsNull() bool {
	return e == nil
}

func (e *eventData) GetState() StateName {
	return e.StateName
}

func (e *eventData) SetState(state StateName) {
	e.StateName = state
}

func (e *eventData) Save(ctx context.Context) error {
	// Simulate saving to a database or cache
	// In a real implementation, this would involve actual database operations
	return nil
}

func (e *eventData) ID() string {
	return strconv.Itoa(e.InitNumber)
}

func (e *eventData) MetaInfo() json.RawMessage {
	m := map[string]string{
		"InitNumber": strconv.Itoa(e.InitNumber),
		"ManualAdd":  strconv.Itoa(e.ManualAdd),
	}

	mi, _ := json.Marshal(m)

	return mi
}

func TestStateName(t *testing.T) {
	sn := NewStateName("test")

	msn, err := sn.MarshalJSON()
	if err != nil {
		t.Fatalf("failed to marshal state name: %v", err)
	}

	t.Logf("state name: %s", string(msn))

	var nsn StateName
	err = nsn.UnmarshalJSON(msn)
	if err != nil {
		t.Fatalf("failed to unmarshal state name: %v", err)
	}

	if sn.String() != nsn.String() {
		t.Fatalf("expected %s, got %s", sn.String(), nsn.String())
	}
}
