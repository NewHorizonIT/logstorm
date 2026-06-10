package redis

func SessionKey(sessionID string) string {
	return "session:" + ":" + sessionID
}

func CacheKey(prefix, identifier string) string {
	return prefix + ":" + identifier
}
