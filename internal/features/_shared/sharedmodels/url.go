package sharedmodels

import "context"

type urlKey struct{}

func WithURL(ctx context.Context, url string) context.Context {
	return context.WithValue(ctx, urlKey{}, url)
}

func GetURLFromContext(ctx context.Context) string {
	if v, ok := ctx.Value(urlKey{}).(string); ok {
		return v
	}
	return ""
}
