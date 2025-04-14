package feedstore

import "github.com/doug-martin/goqu/v9"

// LinkQueryInterface defines the interface for querying links
type LinkQueryInterface interface {
	// Validation method
	Validate() error

	// Count related methods
	IsCountOnlySet() bool
	GetCountOnly() bool
	SetCountOnly(countOnly bool) LinkQueryInterface

	// Soft delete related query methods
	IsWithSoftDeletedSet() bool
	GetWithSoftDeleted() bool
	SetWithSoftDeleted(withSoftDeleted bool) LinkQueryInterface

	IsOnlySoftDeletedSet() bool
	GetOnlySoftDeleted() bool
	SetOnlySoftDeleted(onlySoftDeleted bool) LinkQueryInterface

	// Dataset conversion methods
	ToSelectDataset(store StoreInterface) (selectDataset *goqu.SelectDataset, columns []any, err error)

	// Field query methods

	IsCreatedAtGteSet() bool
	GetCreatedAtGte() string
	SetCreatedAtGte(createdAt string) LinkQueryInterface

	IsCreatedAtLteSet() bool
	GetCreatedAtLte() string
	SetCreatedAtLte(createdAt string) LinkQueryInterface

	IsFeedIDSet() bool
	GetFeedID() string
	SetFeedID(feedID string) LinkQueryInterface

	IsIDSet() bool
	GetID() string
	SetID(id string) LinkQueryInterface

	IsIDInSet() bool
	GetIDIn() []string
	SetIDIn(ids []string) LinkQueryInterface

	IsLimitSet() bool
	GetLimit() int
	SetLimit(limit int) LinkQueryInterface

	IsOffsetSet() bool
	GetOffset() int
	SetOffset(offset int) LinkQueryInterface

	IsOrderBySet() bool
	GetOrderBy() string
	SetOrderBy(orderBy string) LinkQueryInterface

	IsOrderDirectionSet() bool
	GetOrderDirection() string
	SetOrderDirection(orderDirection string) LinkQueryInterface

	IsStatusSet() bool
	GetStatus() string
	SetStatus(status string) LinkQueryInterface
	SetStatusIn(statuses []string) LinkQueryInterface

	IsURLSet() bool
	GetURL() string
	SetURL(url string) LinkQueryInterface

	IsUpdatedAtGteSet() bool
	GetUpdatedAtGte() string
	SetUpdatedAtGte(updatedAt string) LinkQueryInterface

	IsUpdatedAtLteSet() bool
	GetUpdatedAtLte() string
	SetUpdatedAtLte(updatedAt string) LinkQueryInterface
}
