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
	return q.hasProperty("owner_id")
}

func (q *linkQuery) GetOwnerID() string {
	if q.IsOwnerIDSet() {
		return q.params["owner_id"].(string)
	}

	return ""
}

func (q *linkQuery) SetOwnerID(ownerID string) LinkQueryInterface {
	q.params["owner_id"] = ownerID
	return q
}

func (q *linkQuery) IsCountOnlySet() bool {
	return q.hasProperty("count_only")
}
func (q *linkQuery) GetCountOnly() bool {
	if q.IsCountOnlySet() {
		return q.params["count_only"].(bool)
	}
	return false
}

func (q *linkQuery) SetCountOnly(countOnly bool) LinkQueryInterface {
	q.params["count_only"] = countOnly
	return q
}

func (q *linkQuery) IsCreatedAtGteSet() bool {
	return q.hasProperty("created_at_gte")
}

func (q *linkQuery) GetCreatedAtGte() string {
	if q.IsCreatedAtGteSet() {
		return q.params["created_at_gte"].(string)
	}

	return ""
}

func (q *linkQuery) SetCreatedAtGte(createdAtGte string) LinkQueryInterface {
	q.params["created_at_gte"] = createdAtGte
	return q
}

func (q *linkQuery) IsCreatedAtLteSet() bool {
	return q.hasProperty("created_at_lte")
}
func (q *linkQuery) GetCreatedAtLte() string {
	if q.IsCreatedAtLteSet() {
		return q.params["created_at_lte"].(string)
	}

	return ""
}

func (q *linkQuery) SetCreatedAtLte(createdAtLte string) LinkQueryInterface {
	q.params["created_at_lte"] = createdAtLte
	return q
}

func (q *linkQuery) IsIDSet() bool {
	return q.hasProperty("id")
}
func (q *linkQuery) GetID() string {
	if q.IsIDSet() {
		return q.params["id"].(string)
	}

	return ""
}

func (q *linkQuery) SetID(id string) LinkQueryInterface {
	q.params["id"] = id
	return q
}

func (q *linkQuery) IsIDInSet() bool {
	return q.hasProperty("id_in")
}

func (q *linkQuery) GetIDIn() []string {
	if q.IsIDInSet() {
		return q.params["id_in"].([]string)
	}

	return []string{}
}

func (q *linkQuery) SetIDIn(idIn []string) LinkQueryInterface {
	q.params["id_in"] = idIn
	return q
}

func (q *linkQuery) IsLimitSet() bool {
	return q.hasProperty("limit")
}

func (q *linkQuery) GetLimit() int {
	if q.IsLimitSet() {
		return q.params["limit"].(int)
	}

	return 0
}

func (q *linkQuery) SetLimit(limit int) LinkQueryInterface {
	q.params["limit"] = limit
	return q
}

func (q *linkQuery) IsOffsetSet() bool {
	return q.hasProperty("offset")
}

func (q *linkQuery) GetOffset() int {
	if q.IsOffsetSet() {
		return q.params["offset"].(int)
	}

	return 0
}

func (q *linkQuery) SetOffset(offset int) LinkQueryInterface {
	q.params["offset"] = offset
	return q
}

func (q *linkQuery) IsOnlySoftDeletedSet() bool {
	return q.hasProperty("only_soft_deleted")
}

func (q *linkQuery) GetOnlySoftDeleted() bool {
	if q.IsOnlySoftDeletedSet() {
		return q.params["only_soft_deleted"].(bool)
	}

	return false
}

func (q *linkQuery) SetOnlySoftDeleted(onlySoftDeleted bool) LinkQueryInterface {
	q.params["only_soft_deleted"] = onlySoftDeleted
	return q
}

func (q *linkQuery) IsOrderDirectionSet() bool {
	return q.hasProperty("order_direction")
}

func (q *linkQuery) GetOrderDirection() string {
	if q.IsOrderDirectionSet() {
		return q.params["order_direction"].(string)
	}

	return ""
}

func (q *linkQuery) SetOrderDirection(orderDirection string) LinkQueryInterface {
	q.params["order_direction"] = orderDirection
	return q
}

func (q *linkQuery) IsOrderBySet() bool {
	return q.hasProperty("order_by")
}

func (q *linkQuery) GetOrderBy() string {
	if q.IsOrderBySet() {
		return q.params["order_by"].(string)
	}

	return ""
}

func (q *linkQuery) SetOrderBy(orderBy string) LinkQueryInterface {
	q.params["order_by"] = orderBy
	return q
}

func (q *linkQuery) IsStatusSet() bool {
	return q.hasProperty("status")
}

func (q *linkQuery) GetStatus() string {
	if q.IsStatusSet() {
		return q.params["status"].(string)
	}

	return ""
}

func (q *linkQuery) SetStatus(status string) LinkQueryInterface {
	q.params["status"] = status
	return q
}

func (q *linkQuery) IsStatusInSet() bool {
	return q.hasProperty("status_in")
}

func (q *linkQuery) GetStatusIn() []string {
	if q.IsStatusInSet() {
		return q.params["status_in"].([]string)
	}

	return []string{}
}

func (q *linkQuery) SetStatusIn(statusIn []string) LinkQueryInterface {
	q.params["status_in"] = statusIn
	return q
}

func (q *linkQuery) IsUpdatedAtGteSet() bool {
	return q.hasProperty("updated_at_gte")
}
func (q *linkQuery) GetUpdatedAtGte() string {
	if q.IsUpdatedAtGteSet() {
		return q.params["updated_at_gte"].(string)
	}

	return ""
}

func (q *linkQuery) SetUpdatedAtGte(updatedAt string) LinkQueryInterface {
	q.params["updated_at_gte"] = updatedAt
	return q
}

func (q *linkQuery) IsUpdatedAtLteSet() bool {
	return q.hasProperty("updated_at_lte")
}
func (q *linkQuery) GetUpdatedAtLte() string {
	if q.IsUpdatedAtLteSet() {
		return q.params["updated_at_lte"].(string)
	}

	return ""
}

func (q *linkQuery) SetUpdatedAtLte(updatedAt string) LinkQueryInterface {
	q.params["updated_at_lte"] = updatedAt
	return q
}

func (q *linkQuery) IsWithSoftDeletedSet() bool {
	return q.hasProperty("with_soft_deleted")
}

func (q *linkQuery) GetWithSoftDeleted() bool {
	if q.IsWithSoftDeletedSet() {
		return q.params["with_soft_deleted"].(bool)
	}

	return false
}

func (q *linkQuery) SetWithSoftDeleted(withSoftDeleted bool) LinkQueryInterface {
	q.params["with_soft_deleted"] = withSoftDeleted
	return q
}

func (q *linkQuery) hasProperty(key string) bool {
	return q.params[key] != nil
}
