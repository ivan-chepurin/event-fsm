package event_fsm

import (
	"encoding/json"
	"time"
)

type Event struct {
	ID               string
	TargetID         string
	LastResultStatus ResultStatus
	MetaInfo         json.RawMessage
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

type eventDto struct {
	ID               string          `db:"id" json:"id"`
	TargetID         string          `db:"entity_id" json:"entity_id"`
	LastResultStatus ResultStatus    `db:"last_result_status" json:"last_result_status"`
	MetaInfo         json.RawMessage `db:"meta_info" json:"meta_info"`
	CreatedAt        time.Time       `db:"created_at" json:"created_at"`
	UpdatedAt        time.Time       `db:"updated_at" json:"updated_at"`
}

func (e *eventDto) toEvent() Event {
	return Event{
		ID:               e.ID,
		TargetID:         e.TargetID,
		LastResultStatus: e.LastResultStatus,
		MetaInfo:         e.MetaInfo,
		CreatedAt:        e.CreatedAt,
		UpdatedAt:        e.UpdatedAt,
	}
}

func eventToDTO(e Event) eventDto {
	return eventDto{
		ID:               e.ID,
		TargetID:         e.TargetID,
		LastResultStatus: e.LastResultStatus,
		MetaInfo:         e.MetaInfo,
		CreatedAt:        e.CreatedAt,
		UpdatedAt:        e.UpdatedAt,
	}
}
