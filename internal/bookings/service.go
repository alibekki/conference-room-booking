package bookings

import (
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"net/http"
	"time"
)

type Service struct {
	db *pgxpool.Pool
}

func NewService(db *pgxpool.Pool) *Service {
	return &Service{db: db}
}

type Reservation struct {
	RoomId    string    `json:"room_id"`
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
}

func (s *Service) CreateReservation(w http.ResponseWriter, r *http.Request) {
	var req Reservation
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	if req.StartTime.After(req.EndTime) || req.StartTime.Equal(req.EndTime) {
		http.Error(w, "Invalid time range", http.StatusBadRequest)
		return
	}

	var count int
	err := s.db.QueryRow(r.Context(), `
        SELECT COUNT(1) 
        FROM reservations 
        WHERE room_id=$1 AND 
              (start_time < $3 AND end_time > $2)`,
		req.RoomId, req.StartTime, req.EndTime).Scan(&count)

	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	if count > 0 {
		http.Error(w, "Reservation conflict", http.StatusConflict)
		return
	}

	_, err = s.db.Exec(r.Context(), `
        INSERT INTO reservations (room_id, start_time, end_time) 
        VALUES ($1, $2, $3)`,
		req.RoomId, req.StartTime, req.EndTime)

	if err != nil {
		http.Error(w, "Failed to create reservation", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (s *Service) GetReservations(w http.ResponseWriter, r *http.Request) {
	roomID := chi.URLParam(r, "room_id")

	rows, err := s.db.Query(r.Context(), "SELECT  room_id, start_time, end_time FROM reservations WHERE room_id=$1", roomID)
	if err != nil {
		http.Error(w, "Failed to query reservations", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var reservations []Reservation
	for rows.Next() {
		var res Reservation
		if err := rows.Scan(&res.RoomId, &res.StartTime, &res.EndTime); err != nil {
			http.Error(w, "Failed to scan reservation", http.StatusInternalServerError)
			return
		}
		reservations = append(reservations, res)
	}

	if err := rows.Err(); err != nil {
		http.Error(w, "Error reading reservations", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(reservations)
}
