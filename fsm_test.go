package event_fsm

import (
	"context"
	"sync"
	"testing"
	"time"

	"go.uber.org/zap"
)

// ResultStatus is a type for the result status of a usecase
const (
	Ok ResultStatus = iota + 1
	Done
	TooBig
	TooSmall
	Next
	NeedInsert
)

// State names
const (
	StateNameCheckWhenPlus       StateName = "checkWhenPlus"
	StateNameCheckWhenMinus      StateName = "checkWhenMinus"
	StateNamePlus2               StateName = "statePlus2"
	StateNameMinus3              StateName = "stateMinus3"
	StateNameNeedInsertWhenPlus  StateName = "needInsertWhenPlus"
	StateNameNeedInsertWhenMinus StateName = "needInsertWhenMinus"
	StateNameDone                StateName = "stateDone"
)

// EventTypes
const (
	InsertNumEvent EventType = "insert"
)

// event data key
const (
	cachedNum = "num"
	insertNum = "insertNum"
)

type cache struct {
	m   map[string]int
	rwm sync.RWMutex
}

func (c *cache) Set(key string, value int) {
	c.rwm.Lock()
	defer c.rwm.Unlock()

	c.m[key] = value
}

func (c *cache) Get(key string) int {
	c.rwm.RLock()
	defer c.rwm.RUnlock()

	if val, ok := c.m[key]; ok {
		return val
	}

	return 0
}

type process struct {
	id           string
	status       ResultStatus
	currentState StateName
	nextState    StateName
}

func (p *process) Data() process {
	return *p
}

func (p *process) IsNull() bool {
	result := false
	if p.id == "" {
		result = true
	}

	return result
}

var currentNum = 0

func getInsertNum() int {
	currentNum++
	return currentNum
}

func TestFSM_ProcessEvent(t *testing.T) {

	c := cache{m: make(map[string]int)}

	pMap := map[string]*process{
		"p1": {"p1", Ok, "", StateNameCheckWhenPlus},
		"p2": {"p2", Ok, "", StateNameCheckWhenPlus},
		"p3": {"p3", Ok, "", StateNameCheckWhenMinus},
	}
	ch := make(chan process)

	check := func(ctx context.Context, e Event[process]) (status ResultStatus, err error) {
		defer func() {
			if nextSate, err := e.NextStateName(status); err == nil {
				ch <- process{
					id:           e.ID(),
					status:       status,
					nextState:    nextSate,
					currentState: e.CurrentStateName(),
				}

				return
			}

			ch <- process{
				id:           e.ID(),
				status:       status,
				nextState:    "",
				currentState: e.CurrentStateName(),
			}
		}()

		// get number from cache
		number := c.Get(e.id + cachedNum)
		if number == 0 {
			return Ok, nil
		}

		if number > 10 {
			return TooBig, nil
		}

		if number < -10 {
			return TooSmall, nil
		}

		if number == -5 || number == 5 {
			return NeedInsert, nil
		}

		return Next, nil
	}

	plus2 := func(ctx context.Context, e Event[process]) (status ResultStatus, err error) {
		defer func() {
			if nextSate, err := e.NextStateName(status); err == nil {
				ch <- process{
					id:           e.ID(),
					status:       status,
					nextState:    nextSate,
					currentState: e.CurrentStateName(),
				}

				return
			}

			ch <- process{
				id:           e.ID(),
				status:       status,
				nextState:    "",
				currentState: e.CurrentStateName(),
			}
		}()

		// get number from cache
		number := c.Get(e.id + cachedNum)

		// add 2 to the number
		newNumber := number + 2
		c.Set(e.id+cachedNum, newNumber)

		return Ok, nil
	}

	minus3 := func(ctx context.Context, e Event[process]) (status ResultStatus, err error) {
		defer func() {
			if nextSate, err := e.NextStateName(status); err == nil {
				ch <- process{
					id:           e.ID(),
					status:       status,
					nextState:    nextSate,
					currentState: e.CurrentStateName(),
				}

				return
			}

			ch <- process{
				id:           e.ID(),
				status:       status,
				nextState:    "",
				currentState: e.CurrentStateName(),
			}
		}()

		// get number from cache
		number := c.Get(e.id + cachedNum)

		// subtract 3 from the number
		newNumber := number - 3
		c.Set(e.id+cachedNum, newNumber)

		return Ok, nil
	}

	plusInsert := func(ctx context.Context, e Event[process]) (status ResultStatus, err error) {
		defer func() {
			if nextSate, err := e.NextStateName(status); err == nil {
				ch <- process{
					id:           e.ID(),
					status:       status,
					nextState:    nextSate,
					currentState: e.CurrentStateName(),
				}

				return
			}

			ch <- process{
				id:           e.ID(),
				status:       status,
				nextState:    "",
				currentState: e.CurrentStateName(),
			}
		}()

		// get number from cache
		insertNumber := c.Get(e.id + insertNum)
		cachedNumber := c.Get(e.id + cachedNum)

		// insert number to cache
		newNumber := cachedNumber + insertNumber
		c.Set(e.id+cachedNum, newNumber)

		return Ok, nil
	}

	done := func(ctx context.Context, e Event[process]) (status ResultStatus, err error) {
		defer func() {
			if nextSate, err := e.NextStateName(status); err == nil {
				ch <- process{
					id:           e.ID(),
					status:       status,
					nextState:    nextSate,
					currentState: e.CurrentStateName(),
				}

				return
			}

			ch <- process{
				id:           e.ID(),
				status:       status,
				nextState:    "",
				currentState: e.CurrentStateName(),
			}
		}()

		newNumber := 0
		c.Set(e.id+cachedNum, newNumber)

		return Done, nil
	}

	sd := NewStateDetector[process]()

	// create states
	checkWhenPlus := sd.NewState(StateNameCheckWhenPlus, check, StateTypeTransition)
	checkWhenMinus := sd.NewState(StateNameCheckWhenMinus, check, StateTypeTransition)

	statePlus2 := sd.NewState(StateNamePlus2, plus2, StateTypeTransition)
	stateMinus3 := sd.NewState(StateNameMinus3, minus3, StateTypeTransition)

	needInsertWhenPlus := sd.NewState(StateNameNeedInsertWhenPlus, plusInsert, StateTypeWaitEvent)
	needInsertWhenMinus := sd.NewState(StateNameNeedInsertWhenMinus, plusInsert, StateTypeWaitEvent)

	stateDone := sd.NewState(StateNameDone, done, StateTypeTransition)

	// set next states
	// for statePlus2
	statePlus2.SetNext(checkWhenPlus, Ok)

	// for stateMinus3
	stateMinus3.SetNext(checkWhenMinus, Ok)

	// for checkWhenPlus
	checkWhenPlus.SetNext(statePlus2, Next)
	checkWhenPlus.SetNext(statePlus2, TooSmall)
	checkWhenPlus.SetNext(needInsertWhenPlus, NeedInsert)
	checkWhenPlus.SetNext(stateDone, Ok)
	checkWhenPlus.SetNext(stateMinus3, TooBig)

	// for checkWhenMinus
	checkWhenMinus.SetNext(stateMinus3, Next)
	checkWhenMinus.SetNext(stateMinus3, TooBig)
	checkWhenMinus.SetNext(needInsertWhenMinus, NeedInsert)
	checkWhenMinus.SetNext(stateDone, Ok)
	checkWhenMinus.SetNext(statePlus2, TooSmall)

	// for needInsertWhenPlus
	needInsertWhenPlus.SetNext(checkWhenPlus, Ok)

	// for needInsertWhenMinus
	needInsertWhenMinus.SetNext(checkWhenMinus, Ok)

	l := zap.NewNop()
	fsm := NewFSM(sd, l)

	delCounter := 0

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func(wg *sync.WaitGroup) {
		defer wg.Done()

		for {
			select {
			case p := <-ch:
				t.Logf(
					"process: %s, status: %d, current state: %s, next_state: %s, current_number: %d",
					p.id, p.status, p.currentState, p.nextState, c.Get(p.id+cachedNum),
				)

				if p.status == Done {
					delCounter++
					delete(pMap, p.id)
				}

				if p.status == NeedInsert {
					pMap[p.id] = &process{
						p.id,
						p.status,
						p.currentState,
						p.nextState,
					}

					// create event for insert
					e := NewEvent(p.id, p.nextState, pMap[p.id],
						InsertNumEvent,
					)

					c.Set(p.id+insertNum, getInsertNum())

					wg.Add(1)
					go func(wg *sync.WaitGroup) {
						defer wg.Done()

						// process event
						if ok, err := fsm.ProcessEvent(context.Background(), e); !ok {
							if err != nil {
								t.Errorf("select: case p := <-ch: error: fsm.ProcessEvent: %v", err)
							}
						} else {
							t.Logf("select: case p := <-ch: fsm.ProcessEvent: ok, id: %s", p.id)
						}
					}(wg)
				}

				if delCounter == 3 {
					return
				}

			case <-time.After(10 * time.Second):
				t.Errorf("select: case <-time.After(10 * time.Second): timeout")
				return
			}
		}
	}(wg)

	for id, p := range pMap {
		e := NewEvent(id, StateNameCheckWhenPlus, pMap[id],
			InsertNumEvent,
		)

		c.Set(e.ID()+cachedNum, getInsertNum())

		wg.Add(1)
		go func(id string, i *process, wg *sync.WaitGroup) {
			defer wg.Done()

			if ok, err := fsm.ProcessEvent(context.Background(), e); !ok {
				if err != nil {
					t.Errorf("fsm.ProcessEvent: %v", err)
				}
			}
		}(id, p, wg)
	}

	wg.Wait()

	if delCounter != 3 {
		t.Errorf("delCounter != 3, delCounter: %d", delCounter)
	}
}
