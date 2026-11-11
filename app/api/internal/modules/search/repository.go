package search

import "context"

type LogRepository interface {
	Search(ctx context.Context, query *SearchQuery) (*SearchResult, error)
}
