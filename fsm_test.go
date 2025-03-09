package event_fsm

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"sync"
	"testing"

	"go.uber.org/zap"
)

type cache struct {
	mu   sync.Mutex
	data map[string]eventData
}

func newCache() *cache {
	return &cache{
		data: make(map[string]eventData),
	}
}

func (c *cache) Get(key string) (eventData, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	data, ok := c.data[key]
	return data, ok
}

func (c *cache) Set(key string, d eventData) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.data == nil {
		c.data = make(map[string]eventData)
	}

	c.data[key] = d
}

type eventData struct {
	InitNumber   int
	CurrentState string
	Number       int
	ManualAdd    int
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

func (e *eventData) SetStateName(state StateName) {
	e.CurrentState = state.String()
}

func (e *eventData) StateName() StateName {
	return GetStateName(e.CurrentState)
}

// ResultStatus is a type for the result status of a usecase
const (
	Ok ResultStatus = iota + 1
	NotEnough
	TooMuch
	WrongNumber
)

// State names
var (
	StateFirstCheck  = NewStateName("StateFirstCheck")
	StateAdd3        = NewStateName("StateAdd3")
	StateRemove2     = NewStateName("StateRemove2")
	StateManualAdd   = NewStateName("StateManualAdd")
	StateLastCheck   = NewStateName("StateLastCheck")
	StatePrintResult = NewStateName("StatePrintResult")
)

func TestProcess(t *testing.T) {
	cacheInstance := newCache()

	edSlice := []eventData{
		newEventData(1),
		newEventData(2),
		newEventData(3),
		newEventData(4),
		newEventData(5),
	}

	sd := NewStateDetector[*eventData]()

	firstCheck := sd.NewState(StateFirstCheck, &stateFirstCheck{}, StateTypeTransition)
	add3 := sd.NewState(StateAdd3, &stateAdd3{cache: cacheInstance}, StateTypeTransition)
	remove2 := sd.NewState(StateRemove2, &stateRemove2{cache: cacheInstance}, StateTypeTransition)
	manualAdd := sd.NewState(StateManualAdd, &stateManualAdd{cache: cacheInstance}, StateTypeWaitEvent)
	lastCheck := sd.NewState(StateLastCheck, &stateLastCheck{}, StateTypeTransition)
	printResult := sd.NewState(StatePrintResult, &statePrintResult{}, StateTypeTransition)

	firstCheck.SetNext(add3, NotEnough)
	firstCheck.SetNext(remove2, TooMuch)
	firstCheck.SetNext(manualAdd, Ok)
	add3.SetNext(manualAdd, Ok)

	add3.SetNext(firstCheck, Ok)
	remove2.SetNext(firstCheck, Ok)

	manualAdd.SetNext(lastCheck, Ok)

	lastCheck.SetNext(firstCheck, WrongNumber)
	lastCheck.SetNext(printResult, Ok)

	sd.SetMainState(firstCheck)

	fsm := NewFSM[*eventData](sd, zap.NewNop())

	wg := sync.WaitGroup{}
	for i, ed := range edSlice {
		wg.Add(1)
		go func(ed eventData) {
			defer wg.Done()

			cacheInstance.Set("test", ed)

			_, err := fsm.ProcessEvent(context.Background(), NewEvent(strconv.Itoa(i), &ed))
			if err != nil {
				t.Fatal("Error processing event:", err)
				return

			}

			ed.ManualAdd = -20

			_, err = fsm.ProcessEvent(context.Background(), NewEvent(strconv.Itoa(i), &ed))
			if err != nil {
				if !errors.Is(err, ErrNoNextState) {
					t.Fatal("Error processing event:", err)
					return
				}
			}

		}(ed)
	}

	wg.Wait()
}

type stateFirstCheck struct{}

func (s *stateFirstCheck) Execute(ctx context.Context, data *eventData) (ResultStatus, error) {
	if data.Number < 10 {
		return NotEnough, nil
	}

	if data.Number > 10 {
		return TooMuch, nil
	}

	return Ok, nil
}

type stateAdd3 struct {
	cache *cache
}

func (s *stateAdd3) Execute(ctx context.Context, data *eventData) (ResultStatus, error) {
	data.Number += 3

	s.cache.Set("test", *data)

	return Ok, nil
}

type stateRemove2 struct {
	cache *cache
}

func (s *stateRemove2) Execute(ctx context.Context, data *eventData) (ResultStatus, error) {
	data.Number -= 2

	s.cache.Set("test", *data)

	return Ok, nil
}

type stateManualAdd struct {
	cache *cache
}

func (s *stateManualAdd) Execute(ctx context.Context, data *eventData) (ResultStatus, error) {
	data.Number += data.ManualAdd

	s.cache.Set("test", *data)

	return Ok, nil
}

type stateLastCheck struct{}

func (s *stateLastCheck) Execute(ctx context.Context, data *eventData) (ResultStatus, error) {
	if data.Number == -10 {
		return Ok, nil
	}

	return WrongNumber, nil
}

type statePrintResult struct{}

func (s *statePrintResult) Execute(ctx context.Context, data *eventData) (ResultStatus, error) {
	fmt.Printf("init number: %d, current state: %s, number: %d, manual add: %d\n",
		data.InitNumber, data.CurrentState,
		data.Number, data.ManualAdd,
	)

	return Ok, nil
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
