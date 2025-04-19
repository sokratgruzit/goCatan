package sqlite

import (
	"database/sql"
	"fmt"

	"github.com/mattn/go-sqlite3"
	"github.com/sokratgruzit/goCatan/internal/models"
	"github.com/sokratgruzit/goCatan/internal/storage"
	"golang.org/x/crypto/bcrypt"
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
			demo_balance INTEGER NOT NULL DEFAULT 0,
			address TEXT DEFAULT '',
			access_token TEXT DEFAULT '',
			roles TEXT DEFAULT '',
			avatar TEXT DEFAULT 'avatar.jpg',
			game_started BOOLEAN NOT NULL DEFAULT 0,
			switch_account BOOLEAN NOT NULL DEFAULT 0
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

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return 0, fmt.Errorf("%s: failed to hash password: %w", op, err)
	}

	stmt, err := s.db.Prepare(`
		INSERT INTO users(email, password, username, balance, demo_balance, address, access_token, roles, avatar, game_started, switch_account)
		VALUES(?, ?, ?, ?, ?, '', '', '', 'avatar.jpg', 0, 0)
	`)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	res, err := stmt.Exec(email, string(hashedPassword), username, initialBalance, initialDemo)
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
		SELECT id, email, password, username, balance, demo_balance, address, access_token, roles, avatar, game_started, switch_account
		FROM users WHERE email = ?
	`, email)

	var (
		id            int64
		storedEmail   string
		password      string
		username      string
		balance       int64
		demoBalance   int64
		address       string
		accessToken   string
		roles         string
		avatar        string
		gameStarted   bool
		switchAccount bool
	)

	err := row.Scan(&id, &storedEmail, &password, &username, &balance, &demoBalance, &address, &accessToken, &roles, &avatar, &gameStarted, &switchAccount)

	if err != nil {
		if err == sql.ErrNoRows {
			// handle not found
			return nil, fmt.Errorf("%s: %w", op, err)
		}

		return nil, fmt.Errorf("%s: %w", op, err)
	}

	// you can return a struct if you have one
	return &models.User{
		ID:            id,
		Email:         storedEmail,
		Password:      "",
		Username:      username,
		Balance:       balance,
		DemoBalance:   demoBalance,
		Address:       address,
		AccessToken:   accessToken,
		Roles:         roles,
		Avatar:        avatar,
		GameStarted:   gameStarted,
		SwitchAccount: switchAccount,
	}, nil
}

func (s *Storage) Login(email string, password string) (*models.User, error) {
	const op = "storage.sqlite.Login"

	row := s.db.QueryRow(`
		SELECT id, email, password, username, balance, demo_balance, address, access_token, roles, avatar, game_started, switch_account
		FROM users WHERE email = ?
	`, email)

	var (
		id             int64
		storedEmail    string
		hashedPassword string
		username       string
		balance        int64
		demoBalance    int64
		address        string
		accessToken    string
		roles          string
		avatar         string
		gameStarted    bool
		switchAccount  bool
	)

	err := row.Scan(&id, &storedEmail, &hashedPassword, &username, &balance, &demoBalance, &address, &accessToken, &roles, &avatar, &gameStarted, &switchAccount)

	if err != nil {
		if err == sql.ErrNoRows {
			// handle not found
			return nil, fmt.Errorf("%s: %w", op, err)
		}

		return nil, fmt.Errorf("%s: %w", op, err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password)); err != nil {
		return nil, fmt.Errorf("%s: invalid password", op)
	}

	// you can return a struct if you have one
	return &models.User{
		ID:            id,
		Email:         storedEmail,
		Password:      "",
		Username:      username,
		Balance:       balance,
		DemoBalance:   demoBalance,
		Address:       address,
		AccessToken:   accessToken,
		Roles:         roles,
		Avatar:        avatar,
		GameStarted:   gameStarted,
		SwitchAccount: switchAccount,
	}, nil
}

func (s *Storage) Users() ([]*models.User, error) {
	const op = "storage.sqlite.Users"

	rows, err := s.db.Query(`
		SELECT id, email, password, username, balance, demo_balance, address, access_token, roles, avatar, game_started, switch_account
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
			&user.Address,
			&user.AccessToken,
			&user.Roles,
			&user.Avatar,
			&user.GameStarted,
			&user.SwitchAccount,
		)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}

		user.Password = ""

		users = append(users, &user)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: rows iteration error: %w", op, err)
	}

	return users, nil
}
