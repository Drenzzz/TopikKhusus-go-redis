package tracker

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	rds "github.com/redis/go-redis/v9"

	"topikkhusus-methodtracker/internal/middleware"
)

const methodTrackKey = "method:tracks"

type MethodTracker struct {
	client  *rds.Client
	timeout time.Duration
}

func NewMethodTracker(client *rds.Client, timeout time.Duration) *MethodTracker {
	return &MethodTracker{client: client, timeout: timeout}
}

func (t *MethodTracker) Track(ctx context.Context, payload middleware.TrackPayload) error {
	operationCtx, cancel := context.WithTimeout(ctx, t.timeout)
	defer cancel()

	entry, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal track payload failed: %w", err)
	}

	if err := t.client.LPush(operationCtx, methodTrackKey, entry).Err(); err != nil {
		return fmt.Errorf("lpush track log failed: %w", err)
	}

	if err := t.client.LTrim(operationCtx, methodTrackKey, 0, 999).Err(); err != nil {
		return fmt.Errorf("ltrim track log failed: %w", err)
	}

	return nil
}
