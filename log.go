package event_fsm

import (
	"time"
)

type Log struct {
	ID                  string
	TargetID            string
	EventID             string
	CurrentStateName    StateName
	CurrentResultStatus ResultStatus
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

type logDto struct {
	ID            string       `db:"id" json:"id"`
	TargetID      string       `db:"target_id" json:"target_id"`
	EventID       string       `db:"event_id" json:"event_id"`
	CurrentState  StateName    `db:"current_state" json:"current_state"`
	CurrentResult ResultStatus `db:"current_result_status" json:"current_result_status"`
	CreatedAt     time.Time    `db:"created_at" json:"created_at"`
	UpdatedAt     time.Time    `db:"updated_at" json:"updated_at"`
}

func logToDTO(log Log) logDto {
	return logDto{
		ID:            log.ID,
		TargetID:      log.TargetID,
		EventID:       log.EventID,
		CurrentState:  log.CurrentStateName,
		CurrentResult: log.CurrentResultStatus,
		CreatedAt:     log.CreatedAt,
		UpdatedAt:     log.UpdatedAt,
	}
}

func (l *logDto) toLog() Log {
	return Log{
		ID:                  l.ID,
		TargetID:            l.TargetID,
		EventID:             l.EventID,
		CurrentStateName:    l.CurrentState,
		CurrentResultStatus: l.CurrentResult,
		CreatedAt:           l.CreatedAt,
		UpdatedAt:           l.UpdatedAt,
	}
}
