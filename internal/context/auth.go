package context

import "context"

type userIDContextKey struct{}

func WithUserID(ctx context.Context, userID int32) context.Context {
	return context.WithValue(ctx, userIDContextKey{}, userID)
}

func UserIDFromContext(ctx context.Context) (int32, bool) {
	userID, ok := ctx.Value(userIDContextKey{}).(int32)

	return userID, ok
}
