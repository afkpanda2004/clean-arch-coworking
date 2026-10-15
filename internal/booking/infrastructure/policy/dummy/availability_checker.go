package dummy

import (
	"context"
	"fmt"

	"github.com/example/coworking/internal/booking/application"
	"github.com/google/uuid"

	"github.com/example/coworking/internal/booking/domain"
)

// TODO: replace with real availability check against booking repository.
// Current implementation always returns nil, allowing double-bookings.
type AvailabilityChecker struct {
	repo application.BookingRepo
}

func NewAvailabilityChecker(repo application.BookingRepo) *AvailabilityChecker {
	return &AvailabilityChecker{repo: repo}
}

func (a *AvailabilityChecker) CheckAvailability(ctx context.Context, roomID uuid.UUID, slot domain.DateRange) error {

	bookings, err := a.repo.FindOverlapping(ctx, roomID, slot)
	if err != nil {
		return fmt.Errorf("find overlapping bookings: %w", err)
	}
	for _, booking := range bookings {
		if booking.Status() == domain.Cancelled {
			continue
		}
		return fmt.Errorf("booking conflict: %w", domain.ErrBookingConflict)
	}

	return nil

}
