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

func setupSeatMock(t *testing.T) (*SeatRepository, sqlmock.Sqlmock, func()) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)

	sqlxDB := sqlx.NewDb(db, "postgres")
	repo := NewSeatRepository(sqlxDB)

	return repo, mock, func() {
		db.Close()
	}
}

func TestSeatRepository_Create(t *testing.T) {
	repo, mock, cleanup := setupSeatMock(t)
	defer cleanup()

	seat := &model.Seat{
		ID:         uuid.New(),
		EventID:    uuid.New(),
		SeatNumber: "1",
		RowNumber:  "A",
		Status:     model.SeatStatusAvailable,
		Price:      100,
		Version:    1,
	}

	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO seats")).
		WithArgs(
			seat.ID,
			seat.EventID,
			seat.SeatNumber,
			seat.RowNumber,
			seat.Status,
			seat.Price,
			seat.Version,
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
		).WillReturnResult(sqlmock.NewResult(1, 1))

	err := repo.Create(context.Background(), seat)
	assert.NoError(t, err)
}

func TestSeatRepository_CreateBatch(t *testing.T) {
	repo, mock, cleanup := setupSeatMock(t)
	defer cleanup()

	eventID := uuid.New()
	seats := []*model.Seat{
		{ID: uuid.New(), EventID: eventID, SeatNumber: "1", RowNumber: "A", Status: "available", Price: 100},
		{ID: uuid.New(), EventID: eventID, SeatNumber: "2", RowNumber: "A", Status: "available", Price: 100},
	}

	mock.ExpectBegin()
	for range seats {
		mock.ExpectExec(regexp.QuoteMeta("INSERT INTO seats")).
			WillReturnResult(sqlmock.NewResult(1, 1))
	}
	mock.ExpectCommit()

	err := repo.CreateBatch(context.Background(), seats)
	assert.NoError(t, err)
}

func TestSeatRepository_GetByID(t *testing.T) {
	repo, mock, cleanup := setupSeatMock(t)
	defer cleanup()

	seatID := uuid.New()

	t.Run("Success", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "event_id", "seat_number", "row_number", "status", "price", "version", "created_at", "updated_at"}).
			AddRow(seatID, uuid.New(), "1", "A", "available", 100.0, 1, time.Now(), time.Now())

		mock.ExpectQuery(regexp.QuoteMeta("SELECT id, event_id, seat_number, row_number, status, price, version, created_at, updated_at FROM seats WHERE id = $1")).
			WithArgs(seatID).
			WillReturnRows(rows)

		seat, err := repo.GetByID(context.Background(), seatID)
		require.NoError(t, err)
		assert.Equal(t, seatID, seat.ID)
	})

	t.Run("Not Found", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta("SELECT id, event_id, seat_number, row_number, status, price, version, created_at, updated_at FROM seats WHERE id = $1")).
			WithArgs(seatID).
			WillReturnError(sql.ErrNoRows)

		seat, err := repo.GetByID(context.Background(), seatID)
		assert.ErrorIs(t, err, ErrSeatNotFound)
		assert.Nil(t, seat)
	})
}

func TestSeatRepository_Update(t *testing.T) {
	repo, mock, cleanup := setupSeatMock(t)
	defer cleanup()

	seat := &model.Seat{
		ID:      uuid.New(),
		Status:  model.SeatStatusBooked,
		Version: 1,
	}

	t.Run("Success", func(t *testing.T) {
		mock.ExpectExec(regexp.QuoteMeta("UPDATE seats SET status = $1, price = $2, version = $3 + 1, updated_at = $4 WHERE id = $5 AND version = $6")).
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := repo.Update(context.Background(), seat)
		assert.NoError(t, err)
	})

	t.Run("Optimistic Lock Fail", func(t *testing.T) {
		mock.ExpectExec(regexp.QuoteMeta("UPDATE seats SET status = $1, price = $2, version = $3 + 1, updated_at = $4 WHERE id = $5 AND version = $6")).
			WillReturnResult(sqlmock.NewResult(0, 0))

		err := repo.Update(context.Background(), seat)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "optimistic lock")
	})
}
