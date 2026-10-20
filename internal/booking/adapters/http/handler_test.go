package http

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/example/coworking/internal/booking/application"
	"github.com/example/coworking/internal/booking/domain"
	"github.com/example/coworking/internal/booking/mocks"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

func TestCancelBooking_Success(t *testing.T) {
	bookingID := uuid.New()

	svc := mocks.NewMockBookingService(t)
	svc.EXPECT().
		CancelBooking(
			mock.Anything,
			application.CancelBookingInput{BookingID: bookingID},
		).
		Return(nil).
		Once()

	handler := &BookingHandler{svc: svc}

	req := httptest.NewRequest(
		http.MethodPost,
		"/bookings/"+bookingID.String()+"/cancel",
		nil,
	)
	req.SetPathValue("id", bookingID.String())

	rec := httptest.NewRecorder()

	handler.CancelBooking(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d", http.StatusNoContent, rec.Code)
	}
}

func TestCancelBooking_InvalidID(t *testing.T) {
	svc := mocks.NewMockBookingService(t)
	handler := &BookingHandler{svc: svc}

	req := httptest.NewRequest(
		http.MethodPost,
		"/bookings/invalid-id/cancel",
		nil,
	)
	req.SetPathValue("id", "invalid-id")

	rec := httptest.NewRecorder()

	handler.CancelBooking(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

func TestCancelBooking_ServiceError(t *testing.T) {
	bookingID := uuid.New()

	svc := mocks.NewMockBookingService(t)
	svc.EXPECT().
		CancelBooking(
			mock.Anything,
			application.CancelBookingInput{BookingID: bookingID},
		).
		Return(domain.ErrBookingNotFound).
		Once()

	handler := &BookingHandler{svc: svc}

	req := httptest.NewRequest(
		http.MethodPost,
		"/bookings/"+bookingID.String()+"/cancel",
		nil,
	)
	req.SetPathValue("id", bookingID.String())

	rec := httptest.NewRecorder()

	handler.CancelBooking(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, rec.Code)
	}
}

func TestCancelBooking_InternalServerError(t *testing.T) {
	bookingID := uuid.New()

	svc := mocks.NewMockBookingService(t)
	svc.EXPECT().
		CancelBooking(
			mock.Anything,
			application.CancelBookingInput{BookingID: bookingID},
		).
		Return(errors.New("unexpected database error")).
		Once()

	handler := &BookingHandler{svc: svc}

	req := httptest.NewRequest(
		http.MethodPost,
		"/bookings/"+bookingID.String()+"/cancel",
		nil,
	)
	req.SetPathValue("id", bookingID.String())

	rec := httptest.NewRecorder()

	handler.CancelBooking(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusInternalServerError,
			rec.Code,
		)
	}
}
