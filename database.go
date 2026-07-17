package main

import (
	"database/sql"
	"log"
	"os"

	_ "modernc.org/sqlite"
)

var db *sql.DB

// initDB inicializa la base de datos SQLite y crea la tabla de mensajes si no existe.
func initDB() error {
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "tureparto.db"
	}

	var err error
	db, err = sql.Open("sqlite", dbPath)
	if err != nil {
		return err
	}

	// Verificar conexión
	if err = db.Ping(); err != nil {
		return err
	}

	// Crear tabla de mensajes
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS messages (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			from_number TEXT NOT NULL,
			message_body TEXT NOT NULL,
			received_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		return err
	}

	log.Printf("💾 Base de datos inicializada: %s", dbPath)
	return nil
}

// saveMessage guarda un mensaje de WhatsApp en la base de datos.
func saveMessage(from, body string) error {
	_, err := db.Exec(
		"INSERT INTO messages (from_number, message_body) VALUES (?, ?)",
		from, body,
	)
	if err != nil {
		return err
	}
	log.Printf("💾 Mensaje guardado en BD ✅")
	return nil
}
