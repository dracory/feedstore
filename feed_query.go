package feedstore

import (
	"errors"
	"strings"

	"github.com/doug-martin/goqu/v9"
	"github.com/dromara/carbon/v2"
	"github.com/gouniverse/sb"
)

// feedQuery implements the FeedQueryInterface
type feedQuery struct {
	isCountOnlySet bool
	countOnly      bool

	isCreatedAtGteSet bool
	createdAtGte      string

	isCreatedAtLteSet bool
	createdAtLte      string

	isIDSet bool
	id      string

	isIDInSet bool
	idIn      []string

	isLastFetchedAtLteSet bool
	lastFetchedAtLte      string

	isLastFetchedAtGteSet bool
	lastFetchedAtGte      string

	isLimitSet bool
	limit      int

	isOffsetSet bool
	offset      int

	isOnlySoftDeletedSet bool
	onlySoftDeleted      bool

	isOrderDirectionSet bool
	orderDirection      string

	isOrderBySet bool
	orderBy      string

	isOwnerIDSet bool
	ownerID      string

	isStatusSet bool
	status      string

	isStatusInSet bool
	statusIn      []string

	isWithSoftDeletedSet bool
	withSoftDeleted      bool

	isUpdatedAtGteSet bool
	updatedAtGte      string

	isUpdatedAtLteSet bool
	updatedAtLte      string
}

var _ FeedQueryInterface = (*feedQuery)(nil)

// FeedQuery creates a new feed query
func FeedQuery() FeedQueryInterface {
	return &feedQuery{}
}

// Validate validates the query parameters
func (q *feedQuery) Validate() error {
	if q.IsOwnerIDSet() && q.GetOwnerID() == "" {
		return errors.New("document query: owner_id cannot be empty")
	}

	if q.IsCreatedAtGteSet() && q.GetCreatedAtGte() == "" {
		return errors.New("document query: created_at_gte cannot be empty")
	}

	if q.IsCreatedAtLteSet() && q.GetCreatedAtLte() == "" {
		return errors.New("document query: created_at_lte cannot be empty")
	}

	if q.IsIDSet() && q.GetID() == "" {
		return errors.New("document query: id cannot be empty")
	}

	if q.IsIDInSet() && len(q.GetIDIn()) < 1 {
		return errors.New("document query: id_in cannot be empty array")
	}

	if q.IsLimitSet() && q.GetLimit() < 0 {
		return errors.New("document query: limit cannot be negative")
	}

	if q.IsOffsetSet() && q.GetOffset() < 0 {
		return errors.New("document query: offset cannot be negative")
	}

	if q.IsStatusSet() && q.GetStatus() == "" {
		return errors.New("document query: status cannot be empty")
	}

	if q.IsStatusInSet() && len(q.GetStatusIn()) < 1 {
		return errors.New("document query: status_in cannot be empty array")
	}

	return nil
}

func (q *feedQuery) ToSelectDataset(st StoreInterface) (selectDataset *goqu.SelectDataset, columns []any, err error) {
	if st == nil {
		return nil, []any{}, errors.New("store cannot be nil")
	}

	if err := q.Validate(); err != nil {
		return nil, []any{}, err
	}

	sql := goqu.Dialect(st.GetDriverName()).From(st.GetFeedTableName())

	// Created At filter
	if q.IsCreatedAtGteSet() {
		sql = sql.Where(goqu.C(COLUMN_CREATED_AT).Gte(q.GetCreatedAtGte()))
	}

	if q.IsCreatedAtLteSet() {
		sql = sql.Where(goqu.C(COLUMN_CREATED_AT).Lte(q.GetCreatedAtLte()))
	}

	// ID filter
	if q.IsIDSet() {
		sql = sql.Where(goqu.C(COLUMN_ID).Eq(q.GetID()))
	}

	// ID IN filter
	if q.IsIDInSet() {
		sql = sql.Where(goqu.C(COLUMN_ID).In(q.GetIDIn()))
	}

	// Status filter
	if q.IsStatusSet() {
		sql = sql.Where(goqu.C(COLUMN_STATUS).Eq(q.GetStatus()))
	}

	// Status IN filter
	if q.IsStatusInSet() {
		sql = sql.Where(goqu.C(COLUMN_STATUS).In(q.GetStatusIn()))
	}

	// Updated At filter
	if q.IsUpdatedAtGteSet() {
		sql = sql.Where(goqu.C(COLUMN_UPDATED_AT).Gte(q.GetUpdatedAtGte()))
	}

	if q.IsUpdatedAtLteSet() {
		sql = sql.Where(goqu.C(COLUMN_UPDATED_AT).Lte(q.GetUpdatedAtLte()))
	}

	if !q.IsCountOnlySet() {
		if q.IsLimitSet() {
			sql = sql.Limit(uint(q.GetLimit()))
		}

		if q.IsOffsetSet() {
			sql = sql.Offset(uint(q.GetOffset()))
		}
	}

	sortOrder := sb.DESC
	if q.IsOrderDirectionSet() {
		sortOrder = q.GetOrderDirection()
	}

	if q.IsOrderBySet() {
		if strings.EqualFold(sortOrder, sb.ASC) {
			sql = sql.Order(goqu.I(q.GetOrderBy()).Asc())
		} else {
			sql = sql.Order(goqu.I(q.GetOrderBy()).Desc())
		}
	}

	// Limit (if count only is not set)
	if !q.IsCountOnlySet() || !q.GetCountOnly() {
		if q.IsLimitSet() {
			sql = sql.Limit(uint(q.GetLimit()))
		}

		if q.IsOffsetSet() {
			sql = sql.Offset(uint(q.GetOffset()))
		}
	}

	// Sort order
	if q.IsOrderBySet() {
		sortOrder := q.GetOrderDirection()

		if strings.EqualFold(sortOrder, sb.ASC) {
			sql = sql.Order(goqu.I(q.GetOrderBy()).Asc())
		} else {
			sql = sql.Order(goqu.I(q.GetOrderBy()).Desc())
		}
	}

	// Soft delete filters

	// Only soft deleted
	if q.IsOnlySoftDeletedSet() && q.GetOnlySoftDeleted() {
		sql = sql.Where(goqu.C(COLUMN_SOFT_DELETED_AT).Lte(carbon.Now(carbon.UTC).ToDateTimeString()))
		return sql, []any{}, nil
	}

	// Include soft deleted
	if q.IsWithSoftDeletedSet() && q.GetWithSoftDeleted() {
		return sql, []any{}, nil
	}

	// Exclude soft deleted, not in the past (default)
	softDeleted := goqu.C(COLUMN_SOFT_DELETED_AT).
		Gt(carbon.Now(carbon.UTC).ToDateTimeString())

	sql = sql.Where(softDeleted)

	return sql, []any{}, nil
}

// ============================================================================
// == Getters and Setters
// ============================================================================

func (q *feedQuery) IsOwnerIDSet() bool {
	return q.isOwnerIDSet
}

func (q *feedQuery) GetOwnerID() string {
	if q.IsOwnerIDSet() {
		return q.ownerID
	}

	return ""
}

func (q *feedQuery) SetOwnerID(ownerID string) FeedQueryInterface {
	q.isOwnerIDSet = true
	q.ownerID = ownerID
	return q
}

func (q *feedQuery) IsCountOnlySet() bool {
	return q.isCountOnlySet
}
func (q *feedQuery) GetCountOnly() bool {
	if q.IsCountOnlySet() {
		return q.countOnly
	}
	return false
}

func (q *feedQuery) SetCountOnly(countOnly bool) FeedQueryInterface {
	q.isCountOnlySet = true
	q.countOnly = countOnly
	return q
}

func (q *feedQuery) IsCreatedAtGteSet() bool {
	return q.isCreatedAtGteSet
}

func (q *feedQuery) GetCreatedAtGte() string {
	if q.IsCreatedAtGteSet() {
		return q.createdAtGte
	}

	return ""
}

func (q *feedQuery) SetCreatedAtGte(createdAtGte string) FeedQueryInterface {
	q.isCreatedAtGteSet = true
	q.createdAtGte = createdAtGte
	return q
}

func (q *feedQuery) IsCreatedAtLteSet() bool {
	return q.isCreatedAtLteSet
}

func (q *feedQuery) GetCreatedAtLte() string {
	if q.IsCreatedAtLteSet() {
		return q.createdAtLte
	}

	return ""
}

func (q *feedQuery) SetCreatedAtLte(createdAtLte string) FeedQueryInterface {
	q.isCreatedAtLteSet = true
	q.createdAtLte = createdAtLte
	return q
}

func (q *feedQuery) IsIDSet() bool {
	return q.isIDSet
}

func (q *feedQuery) GetID() string {
	if q.IsIDSet() {
		return q.id
	}

	return ""
}

func (q *feedQuery) SetID(id string) FeedQueryInterface {
	q.isIDSet = true
	q.id = id
	return q
}

func (q *feedQuery) IsIDInSet() bool {
	return q.isIDInSet
}

func (q *feedQuery) GetIDIn() []string {
	if q.IsIDInSet() {
		return q.idIn
	}

	return []string{}
}

func (q *feedQuery) SetIDIn(idIn []string) FeedQueryInterface {
	q.isIDInSet = true
	q.idIn = idIn
	return q
}

func (q *feedQuery) IsLastFetchedAtLteSet() bool {
	return q.isLastFetchedAtLteSet
}

func (q *feedQuery) GetLastFetchedAtLte() string {
	if q.IsLastFetchedAtLteSet() {
		return q.lastFetchedAtLte
	}

	return ""
}

func (q *feedQuery) SetLastFetchedAtLte(lastFetchedAtLte string) FeedQueryInterface {
	q.isLastFetchedAtLteSet = true
	q.lastFetchedAtLte = lastFetchedAtLte
	return q
}

func (q *feedQuery) IsLastFetchedAtGteSet() bool {
	return q.isLastFetchedAtGteSet
}

func (q *feedQuery) GetLastFetchedAtGte() string {
	if q.IsLastFetchedAtGteSet() {
		return q.lastFetchedAtGte
	}

	return ""
}

func (q *feedQuery) SetLastFetchedAtGte(lastFetchedAtGte string) FeedQueryInterface {
	q.isLastFetchedAtGteSet = true
	q.lastFetchedAtGte = lastFetchedAtGte
	return q
}

func (q *feedQuery) IsLimitSet() bool {
	return q.isLimitSet
}

func (q *feedQuery) GetLimit() int {
	if q.IsLimitSet() {
		return q.limit
	}

	return 0
}

func (q *feedQuery) SetLimit(limit int) FeedQueryInterface {
	q.isLimitSet = true
	q.limit = limit
	return q
}

func (q *feedQuery) IsOffsetSet() bool {
	return q.isOffsetSet
}

func (q *feedQuery) GetOffset() int {
	if q.IsOffsetSet() {
		return q.offset
	}

	return 0
}

func (q *feedQuery) SetOffset(offset int) FeedQueryInterface {
	q.isOffsetSet = true
	q.offset = offset
	return q
}

func (q *feedQuery) IsOnlySoftDeletedSet() bool {
	return q.isOnlySoftDeletedSet
}

func (q *feedQuery) GetOnlySoftDeleted() bool {
	if q.IsOnlySoftDeletedSet() {
		return q.onlySoftDeleted
	}

	return false
}

func (q *feedQuery) SetOnlySoftDeleted(onlySoftDeleted bool) FeedQueryInterface {
	q.isOnlySoftDeletedSet = true
	q.onlySoftDeleted = onlySoftDeleted
	return q
}

func (q *feedQuery) IsOrderDirectionSet() bool {
	return q.isOrderDirectionSet
}

func (q *feedQuery) GetOrderDirection() string {
	if q.IsOrderDirectionSet() {
		return q.orderDirection
	}

	return ""
}

func (q *feedQuery) SetOrderDirection(orderDirection string) FeedQueryInterface {
	q.isOrderDirectionSet = true
	q.orderDirection = orderDirection
	return q
}

func (q *feedQuery) IsOrderBySet() bool {
	return q.isOrderBySet
}

func (q *feedQuery) GetOrderBy() string {
	if q.IsOrderBySet() {
		return q.orderBy
	}

	return ""
}

func (q *feedQuery) SetOrderBy(orderBy string) FeedQueryInterface {
	q.isOrderBySet = true
	q.orderBy = orderBy
	return q
}

func (q *feedQuery) IsStatusSet() bool {
	return q.isStatusSet
}

func (q *feedQuery) GetStatus() string {
	if q.IsStatusSet() {
		return q.status
	}

	return ""
}

func (q *feedQuery) SetStatus(status string) FeedQueryInterface {
	q.isStatusSet = true
	q.status = status
	return q
}

func (q *feedQuery) IsStatusInSet() bool {
	return q.isStatusInSet
}

func (q *feedQuery) GetStatusIn() []string {
	if q.IsStatusInSet() {
		return q.statusIn
	}

	return []string{}
}

func (q *feedQuery) SetStatusIn(statusIn []string) FeedQueryInterface {
	q.isStatusInSet = true
	q.statusIn = statusIn
	return q
}

func (q *feedQuery) IsUpdatedAtGteSet() bool {
	return q.isUpdatedAtGteSet
}

func (q *feedQuery) GetUpdatedAtGte() string {
	if q.IsUpdatedAtGteSet() {
		return q.updatedAtGte
	}

	return ""
}

func (q *feedQuery) SetUpdatedAtGte(updatedAt string) FeedQueryInterface {
	q.isUpdatedAtGteSet = true
	q.updatedAtGte = updatedAt
	return q
}

func (q *feedQuery) IsUpdatedAtLteSet() bool {
	return q.isUpdatedAtLteSet
}

func (q *feedQuery) GetUpdatedAtLte() string {
	if q.IsUpdatedAtLteSet() {
		return q.updatedAtLte
	}

	return ""
}

func (q *feedQuery) SetUpdatedAtLte(updatedAt string) FeedQueryInterface {
	q.isUpdatedAtLteSet = true
	q.updatedAtLte = updatedAt
	return q
}

func (q *feedQuery) IsWithSoftDeletedSet() bool {
	return q.isWithSoftDeletedSet
}

func (q *feedQuery) GetWithSoftDeleted() bool {
	if q.IsWithSoftDeletedSet() {
		return q.withSoftDeleted
	}

	return false
}

func (q *feedQuery) SetWithSoftDeleted(withSoftDeleted bool) FeedQueryInterface {
	q.isWithSoftDeletedSet = true
	q.withSoftDeleted = withSoftDeleted
	return q
}
