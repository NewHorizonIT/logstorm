package user

import "go.uber.org/fx"

// Module provides user domain services.
// Repository implementation is bound at the bootstrap layer to avoid import cycles.
var Module = fx.Module(
	"user",
	fx.Provide(NewUserService),
)