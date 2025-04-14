package feedstore

import "github.com/doug-martin/goqu/v9"

// FeedQueryInterface defines the interface for querying feeds
type FeedQueryInterface interface {
	// Validation method
	Validate() error

	// Count related methods
	IsCountOnlySet() bool
	GetCountOnly() bool
	SetCountOnly(countOnly bool) FeedQueryInterface

	// Soft delete related query methods
	IsWithSoftDeletedSet() bool
	GetWithSoftDeleted() bool
	SetWithSoftDeleted(withSoftDeleted bool) FeedQueryInterface

	IsOnlySoftDeletedSet() bool
	GetOnlySoftDeleted() bool
	SetOnlySoftDeleted(onlySoftDeleted bool) FeedQueryInterface

	// Dataset conversion methods
	ToSelectDataset(store StoreInterface) (selectDataset *goqu.SelectDataset, columns []any, err error)

	// Field query methods

	IsCreatedAtGteSet() bool
	GetCreatedAtGte() string
	SetCreatedAtGte(createdAt string) FeedQueryInterface

	IsCreatedAtLteSet() bool
	GetCreatedAtLte() string
	SetCreatedAtLte(createdAt string) FeedQueryInterface

	IsIDSet() bool
	GetID() string
	SetID(id string) FeedQueryInterface

	IsIDInSet() bool
	GetIDIn() []string
	SetIDIn(ids []string) FeedQueryInterface

	IsLimitSet() bool
	GetLimit() int
	SetLimit(limit int) FeedQueryInterface

	IsOffsetSet() bool
	GetOffset() int
	SetOffset(offset int) FeedQueryInterface

	IsOrderBySet() bool
	GetOrderBy() string
	SetOrderBy(orderBy string) FeedQueryInterface

	IsOrderDirectionSet() bool
	GetOrderDirection() string
	SetOrderDirection(orderDirection string) FeedQueryInterface

	IsStatusSet() bool
	GetStatus() string
	SetStatus(status string) FeedQueryInterface
	SetStatusIn(statuses []string) FeedQueryInterface

	IsUpdatedAtGteSet() bool
	GetUpdatedAtGte() string
	SetUpdatedAtGte(updatedAt string) FeedQueryInterface

	IsUpdatedAtLteSet() bool
	GetUpdatedAtLte() string
	SetUpdatedAtLte(updatedAt string) FeedQueryInterface
}
