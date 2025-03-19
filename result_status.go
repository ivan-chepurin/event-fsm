package event_fsm

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

type ResultStatus string

var _resultStatuses = map[string]ResultStatus{}

// NewResultStatus creates a new ResultStatus instance
func NewResultStatus(status string) ResultStatus {
	if rs, ok := _resultStatuses[status]; ok {
		return rs
	}

	rs := ResultStatus(status)
	_resultStatuses[status] = rs

	return rs
}

func (r *ResultStatus) String() string {
	return string(*r)
}

func (r *ResultStatus) MarshalJSON() ([]byte, error) {
	if _, ok := _resultStatuses[r.String()]; !ok {
		return nil, fmt.Errorf("result status %s not found", r.String())
	}

	return json.Marshal(r.String())
}

func (r *ResultStatus) UnmarshalJSON(data []byte) error {
	var status string
	if err := json.Unmarshal(data, &status); err != nil {
		return err
	}

	if status == "" {
		return fmt.Errorf("result status is empty")
	}

	if _, ok := _resultStatuses[status]; !ok {
		return fmt.Errorf("result status %s not found", status)
	}

	*r = _resultStatuses[status]

	return nil
}

func (r *ResultStatus) Scan(value interface{}) error {
	if str, ok := value.(string); ok {
		*r, ok = _resultStatuses[str]
		if !ok {
			return fmt.Errorf("result status %s not found", str)
		}

		return nil
	}

	return fmt.Errorf("cannot convert %v to string", value)
}

func (r *ResultStatus) Value() (driver.Value, error) {
	if r.String() == "" {
		return nil, fmt.Errorf("result status is empty")
	}

	return r.String(), nil
}

var (
	ResultStatusEmpty         = NewResultStatus("")
	ResultStatusFail          = NewResultStatus("fail")
	ResultStatusOk            = NewResultStatus("ok")
	resultStatusWaitNextEvent = NewResultStatus("wait_next_event")
)
