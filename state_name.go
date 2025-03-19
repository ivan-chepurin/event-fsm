package event_fsm

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

type StateName string

var _stateNames = map[string]StateName{}

func checkStateName(sn StateName) (bool, error) {
	if sn.String() == "" {
		return false, ErrEmptyStateName
	}
	_, ok := _stateNames[sn.String()]
	if !ok {
		return false, ErrStateNameNotFound
	}

	return true, nil
}

// State result statuses
func NewStateName(name string) StateName {
	if sn, ok := _stateNames[name]; ok {
		return sn
	}

	sn := StateName(name)
	_stateNames[name] = sn

	return sn
}

func (sn *StateName) String() string {
	return string(*sn)
}

// MarshalJSON check if state name is valid or not before marshalling with Alias
func (sn *StateName) MarshalJSON() ([]byte, error) {
	if _, ok := _stateNames[sn.String()]; !ok {
		return nil, fmt.Errorf("state name %s not found", sn.String())
	}

	return json.Marshal(sn.String())
}

// UnmarshalJSON unmarshal eventData to string, then check if it is a valid state name or not
func (sn *StateName) UnmarshalJSON(data []byte) error {
	var name string
	if err := json.Unmarshal(data, &name); err != nil {
		return err
	}

	if _, ok := _stateNames[name]; !ok {
		return fmt.Errorf("state name %s not found", name)
	}

	*sn = StateName(name)

	return nil
}

func (sn *StateName) Scan(value interface{}) error {
	if str, ok := value.(string); ok {
		*sn = StateName(str)
		return nil
	}

	return fmt.Errorf("cannot convert %v to string", value)
}

func (sn *StateName) Value() (driver.Value, error) {
	if sn.String() == "" {
		return nil, fmt.Errorf("state name is empty")
	}

	return sn.String(), nil
}
