package application_test

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/example/coworking/internal/booking/application"
	"github.com/example/coworking/internal/booking/domain"
	"github.com/example/coworking/internal/booking/mocks"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

func TestCancelBooking_Success(t *testing.T) {
	userID := uuid.New()
	roomID := uuid.New()

	from := time.Now()
	to := from.Add(time.Hour)

	slot, err := domain.NewDateRange(from, to)
	if err != nil {
		t.Fatalf("failed to create DateRange: %v", err)
	}

	booking, err := domain.NewBooking(
		roomID,
		userID,
		slot,
		domain.NewMoney(100, "USD"),
	)
	if err != nil {
		t.Fatalf("failed to create booking: %v", err)
	}

	bookingID := booking.ID()

	repo := mocks.NewMockBookingRepo(t)
	store := mocks.NewMockEventStore(t)
	uow := mocks.NewMockUnitOfWork(t)

	repo.EXPECT().
		FindByID(mock.Anything, bookingID).
		Return(booking, nil).
		Once()

	repo.EXPECT().
		Save(mock.Anything, booking).
		Return(nil).
		Once()

	uow.EXPECT().
		Execute(mock.Anything, mock.Anything).
		RunAndReturn(func(
			ctx context.Context,
			fn func(application.BookingRepo, application.EventStore) error,
		) error {
			return fn(repo, store)
		}).
		Once()

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	service := application.NewService(
		repo,
		nil,
		nil,
		nil,
		uow,
		logger,
	)

	err = service.CancelBooking(
		context.Background(),
		application.CancelBookingInput{BookingID: bookingID},
	)
	if err != nil {
		t.Fatalf("CancelBooking() error = %v", err)
	}

	if booking.Status() != domain.Cancelled {
		t.Errorf(
			"expected status %v, got %v",
			domain.Cancelled,
			booking.Status(),
		)
	}
}

func TestCancelBooking_InvalidInput(t *testing.T) {
	repo := mocks.NewMockBookingRepo(t)
	uow := mocks.NewMockUnitOfWork(t)

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	service := application.NewService(
		repo,
		nil,
		nil,
		nil,
		uow,
		logger,
	)

	err := service.CancelBooking(
		context.Background(),
		application.CancelBookingInput{BookingID: uuid.Nil},
	)

	if err == nil {
		t.Fatal("expected error for empty BookingID, got nil")
	}
}

func TestCancelBooking_FindBookingError(t *testing.T) {
	repo := mocks.NewMockBookingRepo(t)
	store := mocks.NewMockEventStore(t)
	uow := mocks.NewMockUnitOfWork(t)

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	repo.EXPECT().
		FindByID(mock.Anything, mock.Anything).
		Return(nil, errors.New("booking not found")).
		Once()

	uow.EXPECT().
		Execute(mock.Anything, mock.Anything).
		RunAndReturn(func(
			ctx context.Context,
			fn func(application.BookingRepo, application.EventStore) error,
		) error {
			return fn(repo, store)
		}).
		Once()

	bookingID := uuid.New()

	service := application.NewService(
		repo,
		nil,
		nil,
		nil,
		uow,
		logger,
	)

	err := service.CancelBooking(
		context.Background(),
		application.CancelBookingInput{BookingID: bookingID},
	)

	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestCancelBooking_CancelError(t *testing.T) {
	repo := mocks.NewMockBookingRepo(t)
	store := mocks.NewMockEventStore(t)
	uow := mocks.NewMockUnitOfWork(t)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	userID := uuid.New()
	roomID := uuid.New()
	from := time.Now()
	to := from.Add(time.Hour)

	slot, err := domain.NewDateRange(from, to)
	if err != nil {
		t.Fatalf("failed to create DateRange: %v", err)
	}

	booking, err := domain.NewBooking(
		roomID,
		userID,
		slot,
		domain.NewMoney(100, "USD"),
	)
	if err != nil {
		t.Fatalf("failed to create booking: %v", err)
	}

	if err := booking.Cancel(); err != nil {
		t.Fatalf("failed to cancel booking for test setup: %v", err)
	}

	bookingID := booking.ID()

	repo.EXPECT().
		FindByID(mock.Anything, bookingID).
		Return(booking, nil).
		Once()

	uow.EXPECT().
		Execute(mock.Anything, mock.Anything).
		RunAndReturn(func(
			ctx context.Context,
			fn func(application.BookingRepo, application.EventStore) error,
		) error {
			return fn(repo, store)
		}).
		Once()

	service := application.NewService(repo, nil, nil, nil, uow, logger)

	err = service.CancelBooking(
		context.Background(),
		application.CancelBookingInput{BookingID: bookingID},
	)

	if err == nil {
		t.Fatal("expected error when cancelling an already cancelled booking")
	}
}

func TestCancelBooking_SaveError(t *testing.T) {
	repo := mocks.NewMockBookingRepo(t)
	store := mocks.NewMockEventStore(t)
	uow := mocks.NewMockUnitOfWork(t)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	userID := uuid.New()
	roomID := uuid.New()
	from := time.Now()
	to := from.Add(time.Hour)

	slot, err := domain.NewDateRange(from, to)
	if err != nil {
		t.Fatalf("failed to create DateRange: %v", err)
	}

	booking, err := domain.NewBooking(
		roomID,
		userID,
		slot,
		domain.NewMoney(100, "USD"),
	)
	if err != nil {
		t.Fatalf("failed to create booking: %v", err)
	}

	bookingID := booking.ID()

	repo.EXPECT().
		FindByID(mock.Anything, bookingID).
		Return(booking, nil).
		Once()

	repo.EXPECT().
		Save(mock.Anything, booking).
		Return(errors.New("database error")).
		Once()

	uow.EXPECT().
		Execute(mock.Anything, mock.Anything).
		RunAndReturn(func(
			ctx context.Context,
			fn func(application.BookingRepo, application.EventStore) error,
		) error {
			return fn(repo, store)
		}).
		Once()

	service := application.NewService(repo, nil, nil, nil, uow, logger)

	err = service.CancelBooking(
		context.Background(),
		application.CancelBookingInput{BookingID: bookingID},
	)

	if err == nil {
		t.Fatal("expected error when saving booking, got nil")
	}
}
