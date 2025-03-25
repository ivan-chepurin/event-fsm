package event_fsm

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"sync"
	"testing"
	"time"

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

// ResultStatus is a type for the result status of a usecase
var (
	NotEnough   = NewResultStatus("not_enough")
	TooMuch     = NewResultStatus("too_much")
	WrongNumber = NewResultStatus("wrong_number")
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
	sd.SetMainState(StateFirstCheck)

	add3 := sd.NewState(StateAdd3, &stateAdd3{cache: cacheInstance}, StateTypeTransition)
	remove2 := sd.NewState(StateRemove2, &stateRemove2{cache: cacheInstance}, StateTypeTransition)
	manualAdd := sd.NewState(StateManualAdd, &stateManualAdd{cache: cacheInstance}, StateTypeWaitEvent)
	lastCheck := sd.NewState(StateLastCheck, &stateLastCheck{}, StateTypeTransition)
	printResult := sd.NewState(StatePrintResult, &statePrintResult{t}, StateTypeTransition)

	firstCheck.SetNext(add3, NotEnough)
	firstCheck.SetNext(remove2, TooMuch)
	firstCheck.SetNext(manualAdd, ResultStatusOk)
	add3.SetNext(manualAdd, ResultStatusOk)

	add3.SetNext(firstCheck, ResultStatusOk)
	remove2.SetNext(firstCheck, ResultStatusOk)

	manualAdd.SetNext(lastCheck, ResultStatusOk)

	lastCheck.SetNext(firstCheck, WrongNumber)
	lastCheck.SetNext(printResult, ResultStatusOk)

	zap.NewNop().Error("State detector initialized")
	cfg := &Config[*eventData]{
		Logger:        zap.NewNop(),
		StateDetector: sd,
		DBConf:        "host=localhost port=5432 user=user password=qwerty dbname=fsm_test_db sslmode=disable",
		RedisConf: &Redis{
			URL:      "localhost:6379",
			Password: "qwerty",
		},
		AppLabel:              "fsm_test",
		MaxOpenConnections:    5,
		MaxIdleConnections:    5,
		ConnectionMaxLifetime: 5 * time.Second,
	}

	fsm, err := NewFSM[*eventData](cfg)
	if err != nil {
		t.Fatal("Error creating FSM:", err)
		return
	}

	wg := sync.WaitGroup{}
	for _, ed := range edSlice {
		wg.Add(1)
		go func(ed eventData) {
			defer wg.Done()

			cacheInstance.Set("test", ed)

			target, err := fsm.ProcessEvent(context.Background(), NewTarget(&ed))
			if err != nil {
				if errors.Is(err, ErrNoNextState) {
					t.Log("For test need to clean data base and redis")

					return
				}
				t.Fatal("Error processing event:", err.Error())
				return

			}

			if _, ok := target.Data(); !ok {
				t.Fatal("Data is nil")
				return
			}

			ed.ManualAdd = -20

			_, err = fsm.ProcessEvent(context.Background(), NewTarget(&ed))
			if err != nil {
				if !errors.Is(err, ErrNoNextState) {
					t.Fatal("Error processing event:", err.Error())
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

	return ResultStatusOk, nil
}

type stateAdd3 struct {
	cache *cache
}

func (s *stateAdd3) Execute(ctx context.Context, data *eventData) (ResultStatus, error) {
	data.Number += 3

	s.cache.Set("test", *data)

	return ResultStatusOk, nil
}

type stateRemove2 struct {
	cache *cache
}

func (s *stateRemove2) Execute(ctx context.Context, data *eventData) (ResultStatus, error) {
	data.Number -= 2

	s.cache.Set("test", *data)

	return ResultStatusOk, nil
}

type stateManualAdd struct {
	cache *cache
}

func (s *stateManualAdd) Execute(ctx context.Context, data *eventData) (ResultStatus, error) {
	data.Number += data.ManualAdd

	s.cache.Set("test", *data)

	return ResultStatusOk, nil
}

type stateLastCheck struct{}

func (s *stateLastCheck) Execute(ctx context.Context, data *eventData) (ResultStatus, error) {
	if data.Number == -10 {
		return ResultStatusOk, nil
	}

	return WrongNumber, nil
}

type statePrintResult struct {
	log *testing.T
}

func (s *statePrintResult) Execute(ctx context.Context, data *eventData) (ResultStatus, error) {
	s.log.Log(fmt.Sprintf("init number: %d, number: %d, manual add: %d\n",
		data.InitNumber,
		data.Number, data.ManualAdd,
	))

	return ResultStatusOk, nil
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
