package event_fsm

import "fmt"

// EventType is a type of event that can be processed by usecase
type EventType string

// Event is a struct that represents an event that can be processed by usecase
type Event struct {
	id string

	stateName string
	state     *State
	prevState *State

	data map[string]interface{}

	et EventType
}

func NewEvent(id, stateName string, data map[string]interface{}, eventType EventType) Event {
	return Event{
		id:        id,
		stateName: stateName,
		data:      data,
		et:        eventType,
	}
}

func (e *Event) Data(key string) (interface{}, bool) {
	val, ok := e.data[key]
	return val, ok
}

func (e *Event) ID() string {
	return e.id
}

func (e *Event) Type() EventType {
	return e.et
}

func (e *Event) GetLog() string {
	log := fmt.Sprintf("id: %s, stateName: %s, type: %s", e.id, e.state.Name, e.et)
	return log
}

func (e *Event) CurrentStateName() string {
	return e.state.Name
}

func (e *Event) NextStateName(status ResultStatus) (string, error) {
	if state, err := e.state.getNext(status); err != nil {
		return "", err
	} else {
		return state.Name, nil
	}
}

type Result struct {
	Status ResultStatus
}
