package auth

import (
	"context"

	"github.com/logstorm/api/internal/modules/user"
)

// UserProvider is the narrow interface auth needs from the user domain.
// Satisfied by *user.UserService at bootstrap.
type UserProvider interface {
	CreateUser(ctx context.Context, input user.CreateUserInput) (*user.User, error)
	GetByEmail(ctx context.Context, email string) (*user.User, error)
}
