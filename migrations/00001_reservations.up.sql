-- migrations/00001_reservations.up.sql
CREATE TABLE IF NOT EXISTS reservations (
                              id INTEGER PRIMARY KEY GENERATED  BY DEFAULT  AS IDENTITY ,
                              room_id Varchar(255) NOT NULL,
                              start_time DATE NOT NULL,
                              end_time DATE NOT NULL
);
