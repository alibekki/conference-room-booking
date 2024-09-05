package main

import (
	"conference-room-booking/internal/db"
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"log"
	"net/http"
	"os"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx/v5/stdlib"

	"conference-room-booking/internal/bookings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {
	wPath, err := os.Getwd()
	if err != nil {
		fmt.Println(err.Error())
	}

	fmt.Println(wPath)
	dbUrl := os.Getenv("DATABASE_URL")

	// Создаем соединение через database/sql
	dbConn, err := db.NewConnection(dbUrl)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}
	defer dbConn.Close()

	// Запуск миграций
	if err := runMigrations(dbConn); err != nil {
		log.Fatalf("Migration failed: %v\n", err)
	}

	// Создаем соединение через pgxpool для дальнейшего использования в приложении
	pgConn, err := pgxpool.New(context.Background(), dbUrl)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}
	defer pgConn.Close()

	bookingService := bookings.NewService(pgConn)

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Post("/reservations", bookingService.CreateReservation)
	r.Get("/reservations/{room_id}", bookingService.GetReservations)

	log.Println("Starting server on :8080")
	http.ListenAndServe(":8080", r)
}

func runMigrations(pool *pgxpool.Pool) error {
	// Преобразуем pgxpool.Pool в *sql.DB
	newDb := stdlib.OpenDB(*pool.Config().ConnConfig)

	// Проверка соединения с базой данных
	if err := newDb.Ping(); err != nil {
		return fmt.Errorf("could not connect to the database: %w", err)
	}
	log.Println("Successfully connected to the database")

	// Настройка драйвера для PostgreSQL
	driver, err := postgres.WithInstance(newDb, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("could not create postgres driver: %w", err)
	}

	// Создание объекта migrate
	m, err := migrate.NewWithDatabaseInstance(
		"file:///app/migrations", // путь к миграциям
		"bookings", driver)
	if err != nil {
		return fmt.Errorf("could not create migrate instance: %w", err)
	}
	log.Println("Migration object successfully initialized")

	// Выполнение тестовой миграции
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("could not apply migrations: %w", err)
	}
	log.Println("Migrations applied successfully")

	return nil
}
