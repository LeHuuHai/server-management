package model

type ServerSortField string

const (
	SortByCreatedAt ServerSortField = "created_at"
	SortByName      ServerSortField = "server_name"
)

type ListServerFilter struct {
	From      int
	To        int
	SortField ServerSortField
	Desc      bool
}
