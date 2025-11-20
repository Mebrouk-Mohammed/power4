package auth

import "context"

// User represents a minimal user record used by the auth layer.
type User struct {
    ID        int
    Username  string
    Email     string
    AvatarURL string
}

// Repository is the persistence abstraction for users.
type Repository interface {
    CreateUser(ctx context.Context, username, email, password string) (*User, error)
    GetByUsername(ctx context.Context, username string) (*User, error)
    GetByID(ctx context.Context, id int) (*User, error)
    Authenticate(ctx context.Context, username, password string) (*User, error)
    // DeleteUser removes a user by ID.
    DeleteUser(ctx context.Context, id int) error
    // UpdateAvatar sets the avatar URL for a user by ID.
    UpdateAvatar(ctx context.Context, id int, avatarURL string) error
    Close() error
}
