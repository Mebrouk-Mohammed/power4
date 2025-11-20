package auth

import (
    "context"
    "database/sql"
    "errors"
    "fmt"
    "os"
    "strings"
    "time"

    _ "github.com/go-sql-driver/mysql"
    "golang.org/x/crypto/bcrypt"
)

type mysqlRepo struct {
    db *sql.DB
}

// NewMySQLFromEnv constructs a MySQL repository from env vars:
// DB_USER, DB_PASS, DB_HOST, DB_PORT, DB_NAME
func NewMySQLFromEnv() (Repository, error) {
    user := os.Getenv("DB_USER")
    pass := os.Getenv("DB_PASS")
    host := os.Getenv("DB_HOST")
    port := os.Getenv("DB_PORT")
    name := os.Getenv("DB_NAME")
    if user == "" || host == "" || name == "" {
        return nil, errors.New("DB_USER/DB_HOST/DB_NAME required")
    }
    if port == "" {
        port = "3306"
    }
    return NewMySQLFromConfig(user, pass, host, port, name)
}

// NewMySQLFromConfig creates a MySQL repo from explicit parameters.
func NewMySQLFromConfig(user, pass, host, port, name string) (Repository, error) {
    if port == "" { port = "3306" }
    dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true", user, pass, host, port, name)
    db, err := sql.Open("mysql", dsn)
    if err != nil {
        return nil, err
    }
    db.SetConnMaxLifetime(5 * time.Minute)
    db.SetMaxOpenConns(10)
    db.SetMaxIdleConns(5)
    if err := db.Ping(); err != nil {
        _ = db.Close()
        return nil, err
    }
    return &mysqlRepo{db: db}, nil
}

// NewMySQLFromDefaults tries to connect using sensible defaults (matching the PHP `db.php`).
func NewMySQLFromDefaults() (Repository, error) {
    return NewMySQLFromConfig("root", "", "127.0.0.1", "3306", "power4")
}

func (m *mysqlRepo) Close() error { return m.db.Close() }

func (m *mysqlRepo) CreateUser(ctx context.Context, username, email, password string) (*User, error) {
    username = strings.TrimSpace(username)
    if username == "" {
        return nil, errors.New("username required")
    }
    // check existing username to return a friendly error instead of SQL duplicate-key
    if existing, err := m.GetByUsername(ctx, username); err == nil && existing != nil {
        return nil, errors.New("username exists")
    } else if err != nil {
        return nil, err
    }
    hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
    if err != nil {
        return nil, err
    }
    res, err := m.db.ExecContext(ctx, "INSERT INTO users (username, email, password_hash, created_at) VALUES (?, ?, ?, NOW())", username, email, string(hash))
    if err != nil {
        return nil, err
    }
    id64, _ := res.LastInsertId()
    return &User{ID: int(id64), Username: username, Email: email}, nil
}

func (m *mysqlRepo) GetByUsername(ctx context.Context, username string) (*User, error) {
    row := m.db.QueryRowContext(ctx, "SELECT id, username, email, avatar_url FROM users WHERE username=?", username)
    var u User
    if err := row.Scan(&u.ID, &u.Username, &u.Email, &u.AvatarURL); err != nil {
        if err == sql.ErrNoRows { return nil, nil }
        return nil, err
    }
    return &u, nil
}

func (m *mysqlRepo) GetByID(ctx context.Context, id int) (*User, error) {
    row := m.db.QueryRowContext(ctx, "SELECT id, username, email, avatar_url FROM users WHERE id=?", id)
    var u User
    if err := row.Scan(&u.ID, &u.Username, &u.Email, &u.AvatarURL); err != nil {
        if err == sql.ErrNoRows { return nil, nil }
        return nil, err
    }
    return &u, nil
}

func (m *mysqlRepo) Authenticate(ctx context.Context, username, password string) (*User, error) {
    username = strings.TrimSpace(username)
    row := m.db.QueryRowContext(ctx, "SELECT id, password_hash, username, email, avatar_url FROM users WHERE username=?", username)
    var id int
    var hash string
    var u User
    if err := row.Scan(&id, &hash, &u.Username, &u.Email, &u.AvatarURL); err != nil {
        if err == sql.ErrNoRows { return nil, errors.New("not found") }
        return nil, err
    }
    // Normalize PHP-style bcrypt prefix $2y$ to $2a$ which Go's bcrypt accepts
    if strings.HasPrefix(hash, "$2y$") {
        hash = strings.Replace(hash, "$2y$", "$2a$", 1)
        fmt.Printf("auth: converted bcrypt prefix $2y$ -> $2a$ for user %s\n", username)
    }
    if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)); err != nil {
        return nil, errors.New("invalid password")
    }
    u.ID = id
    return &u, nil
}

// DeleteUser removes a user row from the database.
func (m *mysqlRepo) DeleteUser(ctx context.Context, id int) error {
    _, err := m.db.ExecContext(ctx, "DELETE FROM users WHERE id = ?", id)
    return err
}

// UpdateAvatar sets the avatar_url for a user.
func (m *mysqlRepo) UpdateAvatar(ctx context.Context, id int, avatarURL string) error {
    _, err := m.db.ExecContext(ctx, "UPDATE users SET avatar_url = ? WHERE id = ?", avatarURL, id)
    return err
}
