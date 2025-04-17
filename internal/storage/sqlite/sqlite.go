package sqlite

import (
	"database/sql"
	"fmt"

	"github.com/mattn/go-sqlite3"
	"github.com/sokratgruzit/goCatan/internal/models"
	"github.com/sokratgruzit/goCatan/internal/storage"
)

type Storage struct {
	db *sql.DB
}

func New(storagePath string) (*Storage, error) {
	const op = "storage.sqlite.New"

	db, err := sql.Open("sqlite3", storagePath)

	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			email TEXT NOT NULL UNIQUE,
			password TEXT NOT NULL,
			username TEXT,
			balance INTEGER NOT NULL DEFAULT 0,
			demo_balance INTEGER NOT NULL DEFAULT 0
		);

		CREATE INDEX IF NOT EXISTS idx_email ON users(email);
	`)

	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &Storage{db: db}, nil
}

func (s *Storage) RegisterUser(email string, password string, username string) (int64, error) {
	const (
		op             = "storage.sqlite.RegisterUser"
		initialBalance = 0
		initialDemo    = 5000
	)

	stmt, err := s.db.Prepare(`
		INSERT INTO users(email, password, username, balance, demo_balance)
		VALUES(?, ?, ?, ?, ?)
	`)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	res, err := stmt.Exec(email, password, username, initialBalance, initialDemo)
	if err != nil {
		if sqliteErr, ok := err.(sqlite3.Error); ok && sqliteErr.ExtendedCode == sqlite3.ErrConstraintUnique {
			return 0, fmt.Errorf("%s: %w", op, storage.ErrUserExists)
		}
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("%s: failed to insert id: %w", op, err)
	}

	return id, nil
}

func (s *Storage) User(email string) (*models.User, error) {
	const op = "storage.sqlite.GetUser"

	row := s.db.QueryRow(`
		SELECT id, email, password, username, balance, demo_balance
		FROM users
		WHERE email = ?
	`, email)

	var (
		id          int64
		storedEmail string
		password    string
		username    string
		balance     int64
		demoBalance int64
	)

	err := row.Scan(&id, &storedEmail, &password, &username, &balance, &demoBalance)

	if err != nil {
		if err == sql.ErrNoRows {
			// handle not found
			return nil, fmt.Errorf("%s: %w", op, err)
		}

		return nil, fmt.Errorf("%s: %w", op, err)
	}

	// you can return a struct if you have one
	return &models.User{
		ID:          id,
		Email:       storedEmail,
		Password:    password,
		Username:    username,
		Balance:     balance,
		DemoBalance: demoBalance,
	}, nil
}

func (s *Storage) Users() ([]*models.User, error) {
	const op = "storage.sqlite.Users"

	rows, err := s.db.Query(`
		SELECT id, email, password, username, balance, demo_balance
		FROM users
	`)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	var users []*models.User

	for rows.Next() {
		var user models.User

		err := rows.Scan(
			&user.ID,
			&user.Email,
			&user.Password,
			&user.Username,
			&user.Balance,
			&user.DemoBalance,
		)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}

		users = append(users, &user)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: rows iteration error: %w", op, err)
	}

	return users, nil
}
