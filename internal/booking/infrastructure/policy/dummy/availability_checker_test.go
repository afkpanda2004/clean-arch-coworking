package dummy

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/example/coworking/internal/booking/domain"
	"github.com/example/coworking/internal/booking/mocks"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

func TestAvailabilityChecker(t *testing.T) {
	ctx := context.Background()
	roomID := uuid.New()
	slot := newTestDateRange(t)

	repo := mocks.NewMockBookingRepo(t)
	repo.EXPECT().
		FindOverlapping(mock.Anything, roomID, slot).
		Return([]*domain.Booking{}, nil).
		Once()

	checker := NewAvailabilityChecker(repo)

	err := checker.CheckAvailability(ctx, roomID, slot)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
}

func TestAvailabilityCheckerRepositoryError(t *testing.T) {
	ctx := context.Background()
	roomID := uuid.New()
	slot := newTestDateRange(t)
	repoErr := errors.New("failed")

	repo := mocks.NewMockBookingRepo(t)
	repo.EXPECT().
		FindOverlapping(mock.Anything, roomID, slot).
		Return(nil, repoErr).
		Once()

	checker := NewAvailabilityChecker(repo)
	err := checker.CheckAvailability(ctx, roomID, slot)

	if !errors.Is(err, repoErr) {
		t.Fatalf("expected repository error %v, got: %v", repoErr, err)
	}
}

func TestAvailabilityCheckerCancelledBooking(t *testing.T) {
	ctx := context.Background()
	roomID := uuid.New()
	slot := newTestDateRange(t)

	booking, err := domain.NewBooking(
		roomID,
		uuid.New(),
		slot,
		domain.NewMoney(100, "USD"),
	)
	if err != nil {
		t.Fatalf("failed to create booking: %v", err)
	}

	if err := booking.Cancel(); err != nil {
		t.Fatalf("failed to cancel booking: %v", err)
	}

	repo := mocks.NewMockBookingRepo(t)
	repo.EXPECT().
		FindOverlapping(mock.Anything, roomID, slot).
		Return([]*domain.Booking{booking}, nil).
		Once()

	checker := NewAvailabilityChecker(repo)
	err = checker.CheckAvailability(ctx, roomID, slot)

	if err != nil {
		t.Fatalf("expected no error for cancelled booking, got: %v", err)
	}
}

func TestAvailabilityCheckerNotCancelBooking(t *testing.T) {
	ctx := context.Background()
	roomID := uuid.New()
	slot := newTestDateRange(t)

	booking, err := domain.NewBooking(
		roomID,
		uuid.New(),
		slot,
		domain.NewMoney(100, "USD"),
	)
	if err != nil {
		t.Fatalf("failed to create booking: %v", err)
	}

	repo := mocks.NewMockBookingRepo(t)
	repo.EXPECT().
		FindOverlapping(mock.Anything, roomID, slot).
		Return([]*domain.Booking{booking}, nil).
		Once()

	checker := NewAvailabilityChecker(repo)
	err = checker.CheckAvailability(ctx, roomID, slot)

	if !errors.Is(err, domain.ErrBookingConflict) {
		t.Fatalf("expected ErrBookingConflict, got: %v", err)
	}
}

func TestAvailabilityCheckerCancelledAndActiveBookings(t *testing.T) {
	ctx := context.Background()
	roomID := uuid.New()
	slot := newTestDateRange(t)

	cancelledBooking, err := domain.NewBooking(
		roomID,
		uuid.New(),
		slot,
		domain.NewMoney(100, "USD"),
	)
	if err != nil {
		t.Fatalf("failed to create cancelled booking: %v", err)
	}

	if err := cancelledBooking.Cancel(); err != nil {
		t.Fatalf("failed to cancel booking: %v", err)
	}

	activeBooking, err := domain.NewBooking(
		roomID,
		uuid.New(),
		slot,
		domain.NewMoney(200, "USD"),
	)
	if err != nil {
		t.Fatalf("failed to create active booking: %v", err)
	}

	t.Run("only cancelled booking", func(t *testing.T) {
		repo := mocks.NewMockBookingRepo(t)
		repo.EXPECT().
			FindOverlapping(mock.Anything, roomID, slot).
			Return([]*domain.Booking{cancelledBooking}, nil).
			Once()

		checker := NewAvailabilityChecker(repo)
		err := checker.CheckAvailability(ctx, roomID, slot)

		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
	})

	t.Run("cancelled and active bookings", func(t *testing.T) {
		repo := mocks.NewMockBookingRepo(t)
		repo.EXPECT().
			FindOverlapping(mock.Anything, roomID, slot).
			Return([]*domain.Booking{cancelledBooking, activeBooking}, nil).
			Once()

		checker := NewAvailabilityChecker(repo)
		err := checker.CheckAvailability(ctx, roomID, slot)

		if !errors.Is(err, domain.ErrBookingConflict) {
			t.Fatalf("expected ErrBookingConflict, got: %v", err)
		}
	})
}

func newTestDateRange(t *testing.T) domain.DateRange {
	t.Helper()

	from := time.Now()
	to := from.Add(time.Hour)

	slot, err := domain.NewDateRange(from, to)
	if err != nil {
		t.Fatalf("failed to create date range: %v", err)
	}

	return slot
}
