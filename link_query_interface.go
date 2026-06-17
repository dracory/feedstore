package feedstore

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

	IsStatusInSet() bool
	GetStatusIn() []string
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

	// Owner ID (legacy, kept for compatibility)
	IsOwnerIDSet() bool
	GetOwnerID() string
	SetOwnerID(ownerID string) LinkQueryInterface
}
