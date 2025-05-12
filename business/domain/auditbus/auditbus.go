// Package auditbus provides a business logic layer for handling audit events.
package auditbus

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/ardanlabs/service/business/sdk/order"
	"github.com/ardanlabs/service/business/sdk/page"
	"github.com/ardanlabs/service/foundation/logger"
	"github.com/ardanlabs/service/foundation/otel"
	"github.com/google/uuid"
)

// Storer interface declares the behavior this package needs to persist and
// retrieve data.
type Storer interface {
	Create(ctx context.Context, audit Audit) error
	Query(ctx context.Context, filter QueryFilter, orderBy order.By, page page.Page) ([]Audit, error)
	Count(ctx context.Context, filter QueryFilter) (int, error)
}

// Business manages the set of APIs for audit access.
type Business struct {
	log    *logger.Logger
	storer Storer
}

// NewBusiness constructs a audit business API for use.
func NewBusiness(log *logger.Logger, storer Storer) *Business {
	return &Business{
		log:    log,
		storer: storer,
	}
}

// Create adds a new audit record to the system.
func (b *Business) Create(ctx context.Context, na NewAudit) (Audit, error) {
	ctx, span := otel.AddSpan(ctx, "business.auditbus.create")
	defer span.End()

	jsonData, err := json.Marshal(na.Data)
	if err != nil {
		return Audit{}, fmt.Errorf("marshal object: %w", err)
	}

	audit := Audit{
		ID:        uuid.New(),
		ObjID:     na.ObjID,
		ObjDomain: na.ObjDomain,
		ObjName:   na.ObjName,
		ActorID:   na.ActorID,
		Action:    na.Action,
		Data:      jsonData,
		Message:   na.Message,
		Timestamp: time.Now(),
	}

	if err := b.storer.Create(ctx, audit); err != nil {
		return Audit{}, fmt.Errorf("create audit: %w", err)
	}

	return audit, nil
}

// Query retrieves a list of existing audit records.
func (b *Business) Query(ctx context.Context, filter QueryFilter, orderBy order.By, page page.Page) ([]Audit, error) {
	ctx, span := otel.AddSpan(ctx, "business.auditbus.query")
	defer span.End()

	audits, err := b.storer.Query(ctx, filter, orderBy, page)
	if err != nil {
		return nil, fmt.Errorf("query audits: %w", err)
	}

	return audits, nil
}

// Count returns the total number of users.
func (b *Business) Count(ctx context.Context, filter QueryFilter) (int, error) {
	ctx, span := otel.AddSpan(ctx, "business.auditbus.count")
	defer span.End()

	return b.storer.Count(ctx, filter)
}
