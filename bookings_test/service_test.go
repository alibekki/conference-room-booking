package bookings_test

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/jackc/pgx/v5/pgxpool"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"conference-room-booking/internal/bookings"
)

var db *pgxpool.Pool

func setupTestDB() {

	dsn := "postgres://test_user:test_password@localhost:5433/test_bookings?sslmode=disable" // Убедитесь, что переменная окружения содержит правильный DSN
	var err error
	db, err = pgxpool.New(context.Background(), dsn)
	if err != nil {
		log.Fatalf("Unable to connect to test database: %v\n", err)
	}

	_, err = db.Exec(context.Background(), `CREATE TABLE IF NOT EXISTS reservations (
                              id INTEGER PRIMARY KEY GENERATED  BY DEFAULT  AS IDENTITY ,
                              room_id Varchar(255) NOT NULL,
                              start_time timestamp NOT NULL,
                              end_time timestamp NOT NULL)`)
	if err != nil {
		log.Fatalf("Failed to create test table: %v\n", err)
	}
}

func teardownTestDB() {
	_, err := db.Exec(context.Background(), "DROP TABLE IF EXISTS reservations")
	if err != nil {
		log.Fatalf("Failed to drop test table: %v\n", err)
	}
	db.Close()
}

func TestCreateReservationSuccess(t *testing.T) {
	setupTestDB() // Инициализация тестовой базы данных
	defer teardownTestDB()

	service := bookings.NewService(db)

	reservation := bookings.Reservation{
		RoomId:    "room_1",
		StartTime: time.Date(2024, 9, 4, 9, 0, 0, 0, time.UTC),
		EndTime:   time.Date(2024, 9, 4, 10, 0, 0, 0, time.UTC),
	}

	body, _ := json.Marshal(reservation)

	req := httptest.NewRequest(http.MethodPost, "/reservations", bytes.NewReader(body))
	req = req.WithContext(context.Background())
	rec := httptest.NewRecorder()

	service.CreateReservation(rec, req)

	if status := rec.Result().StatusCode; status != http.StatusCreated {
		t.Errorf("expected status code %d, got %d", http.StatusCreated, status)
	}
}

func TestCreateReservationConflict(t *testing.T) {
	setupTestDB()
	defer teardownTestDB()

	service := bookings.NewService(db)

	reservation1 := bookings.Reservation{
		RoomId:    "room_1",
		StartTime: time.Date(2024, 9, 4, 9, 0, 0, 0, time.UTC),
		EndTime:   time.Date(2024, 9, 4, 10, 0, 0, 0, time.UTC),
	}

	body1, _ := json.Marshal(reservation1)
	req1 := httptest.NewRequest(http.MethodPost, "/reservations", bytes.NewReader(body1))
	req1 = req1.WithContext(context.Background())
	rec1 := httptest.NewRecorder()

	service.CreateReservation(rec1, req1)

	if status := rec1.Result().StatusCode; status != http.StatusCreated {
		t.Errorf("expected status code %d, got %d", http.StatusCreated, status)
	}

	reservation2 := bookings.Reservation{
		RoomId:    "room_1",
		StartTime: time.Date(2024, 9, 4, 9, 0, 0, 0, time.UTC),
		EndTime:   time.Date(2024, 9, 4, 10, 0, 0, 0, time.UTC),
	}

	body2, _ := json.Marshal(reservation2)
	req2 := httptest.NewRequest(http.MethodPost, "/reservations", bytes.NewReader(body2))
	req2 = req2.WithContext(context.Background())
	rec2 := httptest.NewRecorder()

	service.CreateReservation(rec2, req2)

	if status := rec2.Result().StatusCode; status != http.StatusConflict {
		t.Errorf("expected status code %d, got %d", http.StatusConflict, status)
	}
}

func TestCreateReservationBadRequest(t *testing.T) {
	setupTestDB()
	defer teardownTestDB()

	service := bookings.NewService(db)

	body := []byte(`{"room_id": "room_1", "start_time": "invalid time", "end_time": "invalid time"}`)
	req := httptest.NewRequest(http.MethodPost, "/reservations", bytes.NewReader(body))
	req = req.WithContext(context.Background())
	rec := httptest.NewRecorder()

	service.CreateReservation(rec, req)

	if status := rec.Result().StatusCode; status != http.StatusBadRequest {
		t.Errorf("expected status code %d, got %d", http.StatusBadRequest, status)
	}
}
