package context

import "context"

type ctxKey string

const (
	userIDKey ctxKey = "userID"
)

type Auth struct {
	context.Context
}

func NewAuth(ctx context.Context) *Auth {
	if ctx == nil {
		ctx = context.Background()
	}

	return &Auth{Context: ctx}
}

func (a *Auth) WithUserID(userID int32) *Auth {
	return &Auth{
		Context: context.WithValue(a.Context, userIDKey, userID),
	}
}

func (a *Auth) UserID() int32 {
	userID, ok := a.Context.Value(userIDKey).(int32)
	if !ok {
		return 0
	}

	return userID
}
