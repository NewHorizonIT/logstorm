package search

type SearchRepository interface {
	Search(query *SearchQuery) (*SearchResult, error)
}
