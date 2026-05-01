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

func TestBookingRepository_GetByUser(t *testing.T) {
	repo, mock, cleanup := setupBookingMock(t)
	defer cleanup()

	userID := uuid.New()

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, user_id, event_id, total_amount, discount, final_amount, status, expires_at, paid_at, created_at, updated_at FROM bookings WHERE user_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3")).
		WithArgs(userID, 10, 0).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uuid.New()))

	bookings, err := repo.GetByUser(context.Background(), userID, 10, 0)
	assert.NoError(t, err)
	assert.Len(t, bookings, 1)
}

func TestBookingRepository_GetByEvent(t *testing.T) {
	repo, mock, cleanup := setupBookingMock(t)
	defer cleanup()

	eventID := uuid.New()

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, user_id, event_id, total_amount, discount, final_amount, status, expires_at, paid_at, created_at, updated_at FROM bookings WHERE event_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3")).
		WithArgs(eventID, 10, 0).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uuid.New()))

	bookings, err := repo.GetByEvent(context.Background(), eventID, 10, 0)
	assert.NoError(t, err)
	assert.Len(t, bookings, 1)
}

func TestBookingRepository_GetSeats(t *testing.T) {
	repo, mock, cleanup := setupBookingMock(t)
	defer cleanup()

	bookingID := uuid.New()

	mock.ExpectQuery(regexp.QuoteMeta("SELECT booking_id, seat_id, price, created_at FROM booking_seats WHERE booking_id = $1")).
		WithArgs(bookingID).
		WillReturnRows(sqlmock.NewRows([]string{"booking_id", "seat_id", "price"}).AddRow(bookingID, uuid.New(), 50.0))

	seats, err := repo.GetSeats(context.Background(), bookingID)
	assert.NoError(t, err)
	assert.Len(t, seats, 1)
}

func TestBookingRepository_Update(t *testing.T) {
	repo, mock, cleanup := setupBookingMock(t)
	defer cleanup()

	booking := &model.Booking{
		ID:     uuid.New(),
		Status: model.BookingStatusConfirmed,
	}

	mock.ExpectExec(regexp.QuoteMeta("UPDATE bookings SET status = $1, total_amount = $2, discount = $3, final_amount = $4, expires_at = $5, paid_at = $6, updated_at = $7 WHERE id = $8")).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.Update(context.Background(), booking)
	assert.NoError(t, err)
}

func TestBookingRepository_CountByUser(t *testing.T) {
	repo, mock, cleanup := setupBookingMock(t)
	defer cleanup()

	userID := uuid.New()

	mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*) FROM bookings WHERE user_id = $1")).
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(5))

	count, err := repo.CountByUser(context.Background(), userID)
	assert.NoError(t, err)
	assert.Equal(t, 5, count)
}

func TestBookingRepository_GetExpiredPending(t *testing.T) {
	repo, mock, cleanup := setupBookingMock(t)
	defer cleanup()

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, user_id, event_id, total_amount, discount, final_amount, status, expires_at, paid_at, created_at, updated_at FROM bookings WHERE status = 'pending' AND expires_at < NOW()")).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uuid.New()))

	bookings, err := repo.GetExpiredPending(context.Background())
	assert.NoError(t, err)
	assert.Len(t, bookings, 1)
}

