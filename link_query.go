package feedstore

import (
	"errors"
	"strings"

	"github.com/doug-martin/goqu/v9"
	"github.com/dromara/carbon/v2"
	"github.com/gouniverse/sb"
)

// linkQuery implements the LinkQueryInterface
type linkQuery struct {
	params map[string]interface{}

	isCountOnlySet bool
	countOnly      bool

	isWithSoftDeletedSet bool
	withSoftDeleted      bool

	isOnlySoftDeletedSet bool
	onlySoftDeleted      bool

	isCreatedAtGteSet bool
	createdAtGte      string

	isCreatedAtLteSet bool
	createdAtLte      string

	isFeedIDSet bool
	feedID      string

	isIDSet bool
	id      string

	isIDInSet bool
	idIn      []string

	isLimitSet bool
	limit      int

	isOffsetSet bool
	offset      int

	isOwnerIDSet bool
	ownerID      string

	isOrderBySet bool
	orderBy      string

	isOrderDirectionSet bool
	orderDirection      string

	isStatusSet bool
	status      string

	isStatusInSet bool
	statusIn      []string

	isURLSet bool
	url      string

	isUpdatedAtGteSet bool
	updatedAtGte      string

	isUpdatedAtLteSet bool
	updatedAtLte      string
}

var _ LinkQueryInterface = (*linkQuery)(nil)

// LinkQuery creates a new link query
func LinkQuery() LinkQueryInterface {
	return &linkQuery{
		params: map[string]interface{}{},
	}
}

// Validate validates the query parameters
func (q *linkQuery) Validate() error {
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

	if q.IsURLSet() && q.GetURL() == "" {
		return errors.New("document query: url cannot be empty")
	}

	return nil
}

func (q *linkQuery) ToSelectDataset(st StoreInterface) (selectDataset *goqu.SelectDataset, columns []any, err error) {
	if st == nil {
		return nil, []any{}, errors.New("store cannot be nil")
	}

	if err := q.Validate(); err != nil {
		return nil, []any{}, err
	}

	sql := goqu.Dialect(st.GetDriverName()).From(st.GetLinkTableName())

	// Created At filter
	if q.IsCreatedAtGteSet() {
		sql = sql.Where(goqu.C(COLUMN_CREATED_AT).Gte(q.GetCreatedAtGte()))
	}

	if q.IsCreatedAtLteSet() {
		sql = sql.Where(goqu.C(COLUMN_CREATED_AT).Lte(q.GetCreatedAtLte()))
	}

	// Feed ID filter
	if q.IsFeedIDSet() {
		sql = sql.Where(goqu.C(COLUMN_FEED_ID).Eq(q.GetFeedID()))
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

	// URL filter
	if q.IsURLSet() {
		sql = sql.Where(goqu.C(COLUMN_URL).Eq(q.GetURL()))
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
		} else {
			sql = sql.Limit(1000)
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
		} else {
			sql = sql.Limit(1000)
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

func (q *linkQuery) IsOwnerIDSet() bool {
	return q.isOwnerIDSet
}

func (q *linkQuery) GetOwnerID() string {
	if q.IsOwnerIDSet() {
		return q.ownerID
	}

	return ""
}

func (q *linkQuery) SetOwnerID(ownerID string) LinkQueryInterface {
	q.isOwnerIDSet = true
	q.ownerID = ownerID
	return q
}

func (q *linkQuery) IsCountOnlySet() bool {
	return q.isCountOnlySet
}

func (q *linkQuery) GetCountOnly() bool {
	if q.IsCountOnlySet() {
		return q.countOnly
	}
	return false
}

func (q *linkQuery) SetCountOnly(countOnly bool) LinkQueryInterface {
	q.isCountOnlySet = true
	q.countOnly = countOnly
	return q
}

func (q *linkQuery) IsCreatedAtGteSet() bool {
	return q.isCreatedAtGteSet
}

func (q *linkQuery) GetCreatedAtGte() string {
	if q.IsCreatedAtGteSet() {
		return q.createdAtGte
	}

	return ""
}

func (q *linkQuery) SetCreatedAtGte(createdAtGte string) LinkQueryInterface {
	q.isCreatedAtGteSet = true
	q.createdAtGte = createdAtGte
	return q
}

func (q *linkQuery) IsCreatedAtLteSet() bool {
	return q.isCreatedAtLteSet
}

func (q *linkQuery) GetCreatedAtLte() string {
	if q.IsCreatedAtLteSet() {
		return q.createdAtLte
	}

	return ""
}

func (q *linkQuery) SetCreatedAtLte(createdAtLte string) LinkQueryInterface {
	q.isCreatedAtLteSet = true
	q.createdAtLte = createdAtLte
	return q
}

func (q *linkQuery) IsFeedIDSet() bool {
	return q.isFeedIDSet
}
func (q *linkQuery) GetFeedID() string {
	if q.IsFeedIDSet() {
		return q.feedID
	}

	return ""
}
func (q *linkQuery) SetFeedID(feedID string) LinkQueryInterface {
	q.isFeedIDSet = true
	q.feedID = feedID
	return q
}

func (q *linkQuery) IsIDSet() bool {
	return q.isIDSet
}
func (q *linkQuery) GetID() string {
	if q.IsIDSet() {
		return q.id
	}

	return ""
}

func (q *linkQuery) SetID(id string) LinkQueryInterface {
	q.isIDSet = true
	q.id = id
	return q
}

func (q *linkQuery) IsIDInSet() bool {
	return q.isIDInSet
}

func (q *linkQuery) GetIDIn() []string {
	if q.IsIDInSet() {
		return q.idIn
	}

	return []string{}
}

func (q *linkQuery) SetIDIn(idIn []string) LinkQueryInterface {
	q.isIDInSet = true
	q.idIn = idIn
	return q
}

func (q *linkQuery) IsLimitSet() bool {
	return q.isLimitSet
}

func (q *linkQuery) GetLimit() int {
	if q.IsLimitSet() {
		return q.limit
	}

	return 0
}

func (q *linkQuery) SetLimit(limit int) LinkQueryInterface {
	q.isLimitSet = true
	q.limit = limit
	return q
}

func (q *linkQuery) IsOffsetSet() bool {
	return q.isOffsetSet
}

func (q *linkQuery) GetOffset() int {
	if q.IsOffsetSet() {
		return q.offset
	}

	return 0
}

func (q *linkQuery) SetOffset(offset int) LinkQueryInterface {
	q.isOffsetSet = true
	q.offset = offset
	return q
}

func (q *linkQuery) IsOnlySoftDeletedSet() bool {
	return q.isOnlySoftDeletedSet
}

func (q *linkQuery) GetOnlySoftDeleted() bool {
	if q.IsOnlySoftDeletedSet() {
		return q.onlySoftDeleted
	}

	return false
}

func (q *linkQuery) SetOnlySoftDeleted(onlySoftDeleted bool) LinkQueryInterface {
	q.isOnlySoftDeletedSet = true
	q.onlySoftDeleted = onlySoftDeleted
	return q
}

func (q *linkQuery) IsOrderDirectionSet() bool {
	return q.isOrderDirectionSet
}

func (q *linkQuery) GetOrderDirection() string {
	if q.IsOrderDirectionSet() {
		return q.orderDirection
	}

	return ""
}

func (q *linkQuery) SetOrderDirection(orderDirection string) LinkQueryInterface {
	q.isOrderDirectionSet = true
	q.orderDirection = orderDirection
	return q
}

func (q *linkQuery) IsOrderBySet() bool {
	return q.isOrderBySet
}

func (q *linkQuery) GetOrderBy() string {
	if q.IsOrderBySet() {
		return q.orderBy
	}

	return ""
}

func (q *linkQuery) SetOrderBy(orderBy string) LinkQueryInterface {
	q.isOrderBySet = true
	q.orderBy = orderBy
	return q
}

func (q *linkQuery) IsStatusSet() bool {
	return q.isStatusSet
}

func (q *linkQuery) GetStatus() string {
	if q.IsStatusSet() {
		return q.status
	}

	return ""
}

func (q *linkQuery) SetStatus(status string) LinkQueryInterface {
	q.isStatusSet = true
	q.status = status
	return q
}

func (q *linkQuery) IsStatusInSet() bool {
	return q.isStatusInSet
}

func (q *linkQuery) GetStatusIn() []string {
	if q.IsStatusInSet() {
		return q.statusIn
	}

	return []string{}
}

func (q *linkQuery) SetStatusIn(statusIn []string) LinkQueryInterface {
	q.isStatusInSet = true
	q.statusIn = statusIn
	return q
}

func (q *linkQuery) IsUpdatedAtGteSet() bool {
	return q.isUpdatedAtGteSet
}
func (q *linkQuery) GetUpdatedAtGte() string {
	if q.IsUpdatedAtGteSet() {
		return q.updatedAtGte
	}

	return ""
}

func (q *linkQuery) SetUpdatedAtGte(updatedAt string) LinkQueryInterface {
	q.isUpdatedAtGteSet = true
	q.updatedAtGte = updatedAt
	return q
}

func (q *linkQuery) IsUpdatedAtLteSet() bool {
	return q.isUpdatedAtLteSet
}
func (q *linkQuery) GetUpdatedAtLte() string {
	if q.IsUpdatedAtLteSet() {
		return q.updatedAtLte
	}

	return ""
}

func (q *linkQuery) SetUpdatedAtLte(updatedAt string) LinkQueryInterface {
	q.isUpdatedAtLteSet = true
	q.updatedAtLte = updatedAt
	return q
}

func (q *linkQuery) IsURLSet() bool {
	return q.isURLSet
}

func (q *linkQuery) GetURL() string {
	if q.IsURLSet() {
		return q.url
	}

	return ""
}

func (q *linkQuery) SetURL(url string) LinkQueryInterface {
	q.isURLSet = true
	q.url = url
	return q
}

func (q *linkQuery) IsWithSoftDeletedSet() bool {
	return q.isWithSoftDeletedSet
}

func (q *linkQuery) GetWithSoftDeleted() bool {
	if q.IsWithSoftDeletedSet() {
		return q.withSoftDeleted
	}

	return false
}

func (q *linkQuery) SetWithSoftDeleted(withSoftDeleted bool) LinkQueryInterface {
	q.isWithSoftDeletedSet = true
	q.withSoftDeleted = withSoftDeleted
	return q
}
