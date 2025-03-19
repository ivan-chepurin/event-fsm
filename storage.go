package event_fsm

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

const (
	eventKeyPrefix = "fsm:event:"
	logKeyPrefix   = "fsm:log:"
	cacheTTL       = time.Minute * 15
)

type storage struct {
	l *zap.Logger

	appLabel string

	db    *stateRepo
	cache *rClient
}

func newStorage(l *zap.Logger, appLabel string, db *dbStore, cache *rClient) *storage {
	return &storage{
		l:        l,
		appLabel: appLabel,

		db:    newRepo(db),
		cache: cache,
	}
}

func (s *storage) makeKey(keyPrefix, id string) string {
	b := strings.Builder{}
	b.Grow(len(keyPrefix) + len(s.appLabel) + len(id) + 2)
	b.WriteString(keyPrefix)
	b.WriteString(s.appLabel)
	b.WriteByte(':')
	b.WriteString(id)
	return b.String()
}

func (s *storage) saveLog(ctx context.Context, log Log) (string, error) {
	// Save the log to the database
	id, err := s.db.createLog(ctx, log)
	if err != nil {
		return "", fmt.Errorf("db.createLog: %w", err)
	}

	// Save the log to cache
	if err := s.cache.Set(ctx, s.makeKey(logKeyPrefix, log.TargetID), logToDTO(log), cacheTTL); err != nil {
		s.l.Error(
			"createLog.cache.Set", zap.String("key", s.makeKey(logKeyPrefix, log.TargetID)), zap.Error(err),
		)
	}

	return id, nil
}

func (s *storage) createFullLog(ctx context.Context, log Log) (string, error) {
	// Save the log to the database
	id, err := s.db.createFullLog(ctx, log)
	if err != nil {
		return "", fmt.Errorf("db.createFullLog: %w", err)
	}

	// Save the log to cache
	if err := s.cache.Set(ctx, s.makeKey(logKeyPrefix, log.TargetID), logToDTO(log), cacheTTL); err != nil {
		s.l.Error(
			"createFullLog.cache.Set", zap.String("key", s.makeKey(logKeyPrefix, log.TargetID)), zap.Error(err),
		)
	}

	return id, nil
}

func (s *storage) updateLog(ctx context.Context, log Log) error {
	// Update the log in the database
	if err := s.db.updateLog(ctx, log); err != nil {
		return fmt.Errorf("db.updateLog: %w", err)
	}

	// Update the log in cache
	if err := s.cache.Set(ctx, s.makeKey(logKeyPrefix, log.TargetID), logToDTO(log), cacheTTL); err != nil {
		s.l.Error(
			"updateLog.cache.Set", zap.String("key", s.makeKey(logKeyPrefix, log.TargetID)), zap.Error(err),
		)
	}

	return nil
}

func (s *storage) getEvent(ctx context.Context, id string) (Event, error) {
	// Check the cache first
	var eventDTO eventDto
	if err := s.cache.Get(ctx, s.makeKey(eventKeyPrefix, id), &eventDTO); err != nil {
		if !errors.Is(err, redis.Nil) {
			s.l.Error(
				"getEvent.s.cache.Get", zap.String("key", s.makeKey(eventKeyPrefix, id)), zap.Error(err),
			)
		}
	} else {
		return eventDTO.toEvent(), nil
	}

	event, err := s.db.getEventByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Event{}, ErrLastLogNotFound
		}

		return Event{}, fmt.Errorf("db.getEventByID: %w", err)
	}

	// Save the event to cache
	if err := s.cache.Set(ctx, s.makeKey(eventKeyPrefix, id), eventToDTO(event), cacheTTL); err != nil {
		s.l.Error("getEvent.cache.Set", zap.String("key", s.makeKey(eventKeyPrefix, id)), zap.Error(err))
	}

	return event, nil
}

func (s *storage) saveEvent(ctx context.Context, event Event) (string, error) {
	// Save the event to the database
	id, err := s.db.createEvent(ctx, event)
	if err != nil {
		return "", fmt.Errorf("db.createEvent: %w", err)
	}

	// Save the event to cache
	if err = s.cache.Set(ctx, s.makeKey(eventKeyPrefix, event.ID), eventToDTO(event), cacheTTL); err != nil {
		s.l.Error(
			"saveEvent.cache.Set", zap.String("key", s.makeKey(eventKeyPrefix, event.ID)), zap.Error(err),
		)
	}

	return id, nil
}

func (s *storage) updateEvent(ctx context.Context, event Event) error {
	// Update the event in the database
	if err := s.db.updateEvent(ctx, event); err != nil {
		return fmt.Errorf("db.updateEvent: %w", err)
	}

	// Update the event in cache
	if err := s.cache.Set(ctx, s.makeKey(eventKeyPrefix, event.ID), eventToDTO(event), cacheTTL); err != nil {
		s.l.Error(
			"updateEvent.cache.Set", zap.String("key", s.makeKey(eventKeyPrefix, event.ID)), zap.Error(err),
		)
	}

	return nil
}
