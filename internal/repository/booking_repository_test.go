package repository

import (
	"context"
	"database/sql"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/bekzat-kamen/booking_system_api/internal/model"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupBookingMock(t *testing.T) (*BookingRepository, sqlmock.Sqlmock, func()) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)

	sqlxDB := sqlx.NewDb(db, "postgres")
	repo := NewBookingRepository(sqlxDB)

	return repo, mock, func() {
		_ = db.Close()
	}
}

func TestBookingRepository_Create(t *testing.T) {
	repo, mock, cleanup := setupBookingMock(t)
	defer cleanup()

	booking := &model.Booking{
		ID:          uuid.New(),
		UserID:      uuid.New(),
		EventID:     uuid.New(),
		TotalAmount: 100,
		FinalAmount: 100,
		Status:      model.BookingStatusPending,
		ExpiresAt:   time.Now().Add(15 * time.Minute),
	}

	seats := []*model.BookingSeat{
		{BookingID: booking.ID, SeatID: uuid.New(), Price: 50},
		{BookingID: booking.ID, SeatID: uuid.New(), Price: 50},
	}

	t.Run("Success", func(t *testing.T) {
		mock.ExpectBegin()

		mock.ExpectExec(regexp.QuoteMeta("INSERT INTO bookings")).
			WillReturnResult(sqlmock.NewResult(1, 1))

		mock.ExpectExec(regexp.QuoteMeta("INSERT INTO booking_seats")).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectExec(regexp.QuoteMeta("INSERT INTO booking_seats")).
			WillReturnResult(sqlmock.NewResult(1, 1))

		mock.ExpectCommit()

		err := repo.Create(context.Background(), booking, seats)
		assert.NoError(t, err)
	})

	t.Run("Transaction Error", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta("INSERT INTO bookings")).
			WillReturnError(sql.ErrConnDone)
		mock.ExpectRollback()

		err := repo.Create(context.Background(), booking, seats)
		assert.Error(t, err)
	})
}

func TestBookingRepository_GetByID(t *testing.T) {
	repo, mock, cleanup := setupBookingMock(t)
	defer cleanup()

	bookingID := uuid.New()

	t.Run("Success", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "user_id", "event_id", "total_amount", "discount", "final_amount", "status", "expires_at", "paid_at", "created_at", "updated_at"}).
			AddRow(bookingID, uuid.New(), uuid.New(), 100, 0, 100, "pending", time.Now(), nil, time.Now(), time.Now())

		mock.ExpectQuery(regexp.QuoteMeta("SELECT id, user_id, event_id, total_amount, discount, final_amount, status, expires_at, paid_at, created_at, updated_at FROM bookings WHERE id = $1")).
			WithArgs(bookingID).
			WillReturnRows(rows)

		booking, err := repo.GetByID(context.Background(), bookingID)
		require.NoError(t, err)
		assert.Equal(t, bookingID, booking.ID)
	})

	t.Run("Not Found", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta("SELECT id, user_id, event_id, total_amount, discount, final_amount, status, expires_at, paid_at, created_at, updated_at FROM bookings WHERE id = $1")).
			WithArgs(bookingID).
			WillReturnError(sql.ErrNoRows)

		booking, err := repo.GetByID(context.Background(), bookingID)
		assert.ErrorIs(t, err, ErrBookingNotFound)
		assert.Nil(t, booking)
	})
}

func TestBookingRepository_UpdateStatus(t *testing.T) {
	repo, mock, cleanup := setupBookingMock(t)
	defer cleanup()

	bookingID := uuid.New()

	mock.ExpectExec(regexp.QuoteMeta("UPDATE bookings SET status = $1")).
		WithArgs(model.BookingStatusConfirmed, sqlmock.AnyArg(), bookingID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.UpdateStatus(context.Background(), bookingID, model.BookingStatusConfirmed)
	assert.NoError(t, err)
}
