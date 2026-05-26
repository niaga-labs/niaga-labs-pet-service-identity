package events

import (
	"context"
	"fmt"

	commonKafka "github.com/Kilat-Pet-Delivery/lib-common/kafka"
	protoEvents "github.com/Kilat-Pet-Delivery/lib-proto/events"
	"github.com/Kilat-Pet-Delivery/service-identity/internal/application"
	segmentioKafka "github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

// ShopEventConsumer applies role grants from shop.events.
type ShopEventConsumer struct {
	service *application.AuthService
	logger  *zap.Logger
}

// NewShopEventConsumer creates a shop event consumer.
func NewShopEventConsumer(service *application.AuthService, logger *zap.Logger) *ShopEventConsumer {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &ShopEventConsumer{service: service, logger: logger}
}

// HandleKafkaMessage parses and handles a Kafka CloudEvent envelope.
func (c *ShopEventConsumer) HandleKafkaMessage(ctx context.Context, msg segmentioKafka.Message) error {
	event, err := commonKafka.ParseCloudEvent(msg.Value)
	if err != nil {
		return err
	}
	return c.HandleCloudEvent(ctx, event)
}

// HandleCloudEvent dispatches shop events to scoped role mutations.
func (c *ShopEventConsumer) HandleCloudEvent(ctx context.Context, event commonKafka.CloudEvent) error {
	switch event.Type {
	case protoEvents.ShopCreated:
		var payload protoEvents.ShopCreatedEvent
		if err := event.ParseData(&payload); err != nil {
			return fmt.Errorf("parse shop.created: %w", err)
		}
		return c.service.GrantShopRole(ctx, payload.OwnerUserID, payload.ShopID, "shop_owner")
	case protoEvents.ShopStaffAccepted:
		var payload protoEvents.ShopStaffAcceptedEvent
		if err := event.ParseData(&payload); err != nil {
			return fmt.Errorf("parse shop.staff_accepted: %w", err)
		}
		return c.service.GrantShopRole(ctx, payload.StaffUserID, payload.ShopID, payload.Role)
	case protoEvents.ShopStaffRemoved:
		var payload protoEvents.ShopStaffRemovedEvent
		if err := event.ParseData(&payload); err != nil {
			return fmt.Errorf("parse shop.staff_removed: %w", err)
		}
		for _, role := range []string{"shop_owner", "shop_manager", "shop_staff"} {
			if err := c.service.RevokeShopRole(ctx, payload.StaffUserID, payload.ShopID, role); err != nil {
				return err
			}
		}
		return nil
	default:
		return nil
	}
}
