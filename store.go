package feedstore

import (
	"database/sql"
	"errors"
	"log"
	"strconv"

	"github.com/doug-martin/goqu/v9"
	"github.com/dromara/carbon/v2"
	"github.com/gouniverse/sb"
	"github.com/samber/lo"
)

var _ StoreInterface = (*storeImplementation)(nil) // verify it extends the interface

type storeImplementation struct {
	feedTableName      string
	linkTableName      string
	db                 *sql.DB
	dbDriverName       string
	automigrateEnabled bool
	debugEnabled       bool
}

// FeedCount returns the total number of feeds matching the query filters
func (storeImplementation *storeImplementation) FeedCount(query FeedQueryInterface) (int64, error) {
	if query == nil {
		query = FeedQuery()
	}

	// ensure count-only (disables limit/offset in ToSelectDataset)
	query = query.SetCountOnly(true)

	q, _, err := query.ToSelectDataset(storeImplementation)
	if err != nil {
		return 0, err
	}

	// Build SELECT COUNT(*) directly from the same dataset, removing
	// select list, ordering, limit and offset to avoid unnecessary columns
	countSQL, countParams, errSql := q.
		ClearSelect().
		ClearOrder().
		ClearLimit().
		ClearOffset().
		Prepared(true).
		Select(goqu.COUNT("*").As("count")).
		ToSQL()
	if errSql != nil {
		return 0, errSql
	}

	if storeImplementation.debugEnabled {
		log.Println(countSQL)
	}

	db := sb.NewDatabase(storeImplementation.db, storeImplementation.dbDriverName)
	rows, err := db.SelectToMapString(countSQL, countParams...)
	if err != nil {
		return 0, err
	}
	if len(rows) == 0 {
		return 0, nil
	}
	s := rows[0]["count"]
	if s == "" {
		return 0, nil
	}
	n, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0, err
	}
	return n, nil
}

// AutoMigrate auto migrate
func (storeImplementation *storeImplementation) AutoMigrate() error {
	sql := storeImplementation.sqlFeedTableCreate()

	if sql == "" {
		return errors.New("feed table create sql is empty")
	}

	_, err := storeImplementation.db.Exec(sql)

	if err != nil {
		return err
	}

	sql = storeImplementation.sqlLinkTableCreate()

	if sql == "" {
		return errors.New("link table create sql is empty")
	}

	_, err = storeImplementation.db.Exec(sql)

	if err != nil {
		return err
	}

	return nil
}

// EnableDebug - enables the debug option
func (st *storeImplementation) EnableDebug(debug bool) {
	st.debugEnabled = debug
}

func (storeImplementation *storeImplementation) GetDriverName() string {
	return storeImplementation.dbDriverName
}

func (storeImplementation *storeImplementation) GetFeedTableName() string {
	return storeImplementation.feedTableName
}

func (storeImplementation *storeImplementation) GetLinkTableName() string {
	return storeImplementation.linkTableName
}

func (storeImplementation *storeImplementation) FeedCreate(feed FeedInterface) error {
	feed.SetCreatedAt(carbon.Now(carbon.UTC).ToDateTimeString(carbon.UTC))
	feed.SetUpdatedAt(carbon.Now(carbon.UTC).ToDateTimeString(carbon.UTC))

	data := feed.Data()

	sqlStr, params, errSql := goqu.Dialect(storeImplementation.dbDriverName).
		Insert(storeImplementation.feedTableName).
		Prepared(true).
		Rows(data).
		ToSQL()

	if errSql != nil {
		return errSql
	}

	if storeImplementation.debugEnabled {
		log.Println(sqlStr)
	}

	_, err := storeImplementation.db.Exec(sqlStr, params...)

	if err != nil {
		return err
	}

	feed.MarkAsNotDirty()

	return nil
}

func (storeImplementation *storeImplementation) FeedDelete(feed FeedInterface) error {
	if feed == nil {
		return errors.New("feed is nil")
	}

	return storeImplementation.FeedDeleteByID(feed.ID())
}

func (storeImplementation *storeImplementation) FeedDeleteByID(id string) error {
	if id == "" {
		return errors.New("feed id is empty")
	}

	sqlStr, params, errSql := goqu.Dialect(storeImplementation.dbDriverName).
		Delete(storeImplementation.feedTableName).
		Prepared(true).
		Where(goqu.C(COLUMN_ID).Eq(id)).
		ToSQL()

	if errSql != nil {
		return errSql
	}

	if storeImplementation.debugEnabled {
		log.Println(sqlStr)
	}

	_, err := storeImplementation.db.Exec(sqlStr, params...)

	return err
}

func (storeImplementation *storeImplementation) FeedFindByID(id string) (FeedInterface, error) {
	if id == "" {
		return nil, errors.New("feed id is empty")
	}

	list, err := storeImplementation.FeedList(FeedQuery().
		SetID(id).
		SetLimit(1))

	if err != nil {
		return nil, err
	}

	if len(list) > 0 {
		return list[0], nil
	}

	return nil, nil
}

func (storeImplementation *storeImplementation) FeedList(query FeedQueryInterface) ([]FeedInterface, error) {
	q, columns, err := query.ToSelectDataset(storeImplementation)

	if err != nil {
		return []FeedInterface{}, err
	}

	sqlStr, sqlParams, errSql := q.Prepared(true).Select(columns...).ToSQL()

	if errSql != nil {
		return []FeedInterface{}, errSql
	}

	if storeImplementation.debugEnabled {
		log.Println(sqlStr)
	}

	db := sb.NewDatabase(storeImplementation.db, storeImplementation.dbDriverName)
	modelMaps, err := db.SelectToMapString(sqlStr, sqlParams...)
	if err != nil {
		return []FeedInterface{}, err
	}

	list := []FeedInterface{}

	lo.ForEach(modelMaps, func(modelMap map[string]string, index int) {
		model := NewFeedFromExistingData(modelMap)
		list = append(list, model)
	})

	return list, nil
}

// LinkCount returns the total number of links matching the query filters
func (storeImplementation *storeImplementation) LinkCount(query LinkQueryInterface) (int64, error) {
	if query == nil {
		query = LinkQuery()
	}

	// ensure count-only (disables limit/offset in ToSelectDataset)
	query = query.SetCountOnly(true)

	q, _, err := query.ToSelectDataset(storeImplementation)
	if err != nil {
		return 0, err
	}

	// Build SELECT COUNT(*) directly from the same dataset, removing
	// select list, ordering, limit and offset to avoid unnecessary columns
	countSQL, countParams, errSql := q.
		ClearSelect().
		ClearOrder().
		ClearLimit().
		ClearOffset().
		Prepared(true).
		Select(goqu.COUNT("*").As("count")).
		ToSQL()
	if errSql != nil {
		return 0, errSql
	}

	if storeImplementation.debugEnabled {
		log.Println(countSQL)
	}

	db := sb.NewDatabase(storeImplementation.db, storeImplementation.dbDriverName)
	rows, err := db.SelectToMapString(countSQL, countParams...)
	if err != nil {
		return 0, err
	}
	if len(rows) == 0 {
		return 0, nil
	}
	s := rows[0]["count"]
	if s == "" {
		return 0, nil
	}
	n, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0, err
	}
	return n, nil
}

func (storeImplementation *storeImplementation) FeedSoftDelete(feed FeedInterface) error {
	if feed == nil {
		return errors.New("feed is nil")
	}

	feed.SetSoftDeletedAt(carbon.Now(carbon.UTC).ToDateTimeString(carbon.UTC))

	return storeImplementation.FeedUpdate(feed)
}

func (storeImplementation *storeImplementation) FeedSoftDeleteByID(id string) error {
	feed, err := storeImplementation.FeedFindByID(id)

	if err != nil {
		return err
	}

	return storeImplementation.FeedSoftDelete(feed)
}

func (storeImplementation *storeImplementation) FeedUpdate(feed FeedInterface) error {
	if feed == nil {
		return errors.New("feed is nil")
	}

	feed.SetUpdatedAt(carbon.Now(carbon.UTC).ToDateTimeString())

	dataChanged := feed.DataChanged()

	delete(dataChanged, COLUMN_ID) // ID is not updateable

	if len(dataChanged) <= 1 {
		return nil // only the updated_at field is changed, no need to update
	}

	sqlStr, params, errSql := goqu.Dialect(storeImplementation.dbDriverName).
		Update(storeImplementation.feedTableName).
		Prepared(true).
		Set(dataChanged).
		Where(goqu.C(COLUMN_ID).Eq(feed.ID())).
		ToSQL()

	if errSql != nil {
		return errSql
	}

	if storeImplementation.debugEnabled {
		log.Println(sqlStr)
	}

	_, err := storeImplementation.db.Exec(sqlStr, params...)

	feed.MarkAsNotDirty()

	return err
}

func (storeImplementation *storeImplementation) LinkCreate(link LinkInterface) error {
	link.SetCreatedAt(carbon.Now(carbon.UTC).ToDateTimeString(carbon.UTC))
	link.SetUpdatedAt(carbon.Now(carbon.UTC).ToDateTimeString(carbon.UTC))

	data := link.Data()

	sqlStr, params, errSql := goqu.Dialect(storeImplementation.dbDriverName).
		Insert(storeImplementation.linkTableName).
		Prepared(true).
		Rows(data).
		ToSQL()

	if errSql != nil {
		return errSql
	}

	if storeImplementation.debugEnabled {
		log.Println(sqlStr)
	}

	_, err := storeImplementation.db.Exec(sqlStr, params...)

	if err != nil {
		return err
	}

	link.MarkAsNotDirty()

	return nil
}

func (storeImplementation *storeImplementation) LinkDelete(link LinkInterface) error {
	if link == nil {
		return errors.New("link is nil")
	}

	return storeImplementation.LinkDeleteByID(link.ID())
}

func (storeImplementation *storeImplementation) LinkDeleteByID(id string) error {
	if id == "" {
		return errors.New("link id is empty")
	}

	sqlStr, params, errSql := goqu.Dialect(storeImplementation.dbDriverName).
		Delete(storeImplementation.linkTableName).
		Prepared(true).
		Where(goqu.C("id").Eq(id)).
		ToSQL()

	if errSql != nil {
		return errSql
	}

	if storeImplementation.debugEnabled {
		log.Println(sqlStr)
	}

	_, err := storeImplementation.db.Exec(sqlStr, params...)

	return err
}

func (storeImplementation *storeImplementation) LinkFindByID(id string) (LinkInterface, error) {
	if id == "" {
		return nil, errors.New("link id is empty")
	}

	list, err := storeImplementation.LinkList(LinkQuery().
		SetID(id).
		SetLimit(1))

	if err != nil {
		return nil, err
	}

	if len(list) > 0 {
		return list[0], nil
	}

	return nil, nil
}

func (storeImplementation *storeImplementation) LinkList(query LinkQueryInterface) ([]LinkInterface, error) {
	q, columns, err := query.ToSelectDataset(storeImplementation)

	if err != nil {
		return []LinkInterface{}, err
	}

	sqlStr, sqlParams, errSql := q.Prepared(true).Select(columns...).ToSQL()

	if errSql != nil {
		return []LinkInterface{}, nil
	}

	if storeImplementation.debugEnabled {
		log.Println(sqlStr)
	}

	db := sb.NewDatabase(storeImplementation.db, storeImplementation.dbDriverName)
	modelMaps, err := db.SelectToMapString(sqlStr, sqlParams...)
	if err != nil {
		return []LinkInterface{}, err
	}

	list := []LinkInterface{}

	lo.ForEach(modelMaps, func(modelMap map[string]string, index int) {
		model := NewLinkFromExistingData(modelMap)
		list = append(list, model)
	})

	return list, nil
}

func (storeImplementation *storeImplementation) LinkSoftDelete(link LinkInterface) error {
	if link == nil {
		return errors.New("link is nil")
	}

	link.SetSoftDeletedAt(carbon.Now(carbon.UTC).ToDateTimeString(carbon.UTC))

	return storeImplementation.LinkUpdate(link)
}

func (storeImplementation *storeImplementation) LinkSoftDeleteByID(id string) error {
	link, err := storeImplementation.LinkFindByID(id)

	if err != nil {
		return err
	}

	return storeImplementation.LinkSoftDelete(link)
}

func (storeImplementation *storeImplementation) LinkUpdate(link LinkInterface) error {
	if link == nil {
		return errors.New("link is nil")
	}

	link.SetUpdatedAt(carbon.Now(carbon.UTC).ToDateTimeString())

	dataChanged := link.DataChanged()

	delete(dataChanged, COLUMN_ID) // ID is not updateable

	if len(dataChanged) <= 1 {
		return nil // only the updated_at field is changed, no need to update
	}

	sqlStr, params, errSql := goqu.Dialect(storeImplementation.dbDriverName).
		Update(storeImplementation.linkTableName).
		Prepared(true).
		Set(dataChanged).
		Where(goqu.C(COLUMN_ID).Eq(link.ID())).
		ToSQL()

	if errSql != nil {
		return errSql
	}

	if storeImplementation.debugEnabled {
		log.Println(sqlStr)
	}

	_, err := storeImplementation.db.Exec(sqlStr, params...)

	link.MarkAsNotDirty()

	return err
}

// func (storeImplementation *storeImplementation) linkQuery(options LinkQueryOptions) *goqu.SelectDataset {
// 	q := goqu.Dialect(storeImplementation.dbDriverName).From(storeImplementation.linkTableName)

// 	if options.ID != "" {
// 		q = q.Where(goqu.C(COLUMN_ID).Eq(options.ID))
// 	}

// 	if len(options.IDIn) > 0 {
// 		q = q.Where(goqu.C(COLUMN_ID).In(options.IDIn))
// 	}

// 	if options.FeedID != "" {
// 		q = q.Where(goqu.C(COLUMN_FEED_ID).Eq(options.FeedID))
// 	}

// 	if options.Status != "" {
// 		q = q.Where(goqu.C(COLUMN_STATUS).Eq(options.Status))
// 	}

// 	if len(options.StatusIn) > 0 {
// 		q = q.Where(goqu.C(COLUMN_STATUS).In(options.StatusIn))
// 	}

// 	if options.URL != "" {
// 		q = q.Where(goqu.C(COLUMN_URL).Eq(options.URL))
// 	}

// 	if !options.CountOnly {
// 		if options.Limit > 0 {
// 			q = q.Limit(uint(options.Limit))
// 		}

// 		if options.Offset > 0 {
// 			q = q.Offset(uint(options.Offset))
// 		}
// 	}

// 	sortOrder := sb.DESC
// 	if options.SortOrder != "" {
// 		sortOrder = options.SortOrder
// 	}

// 	if options.OrderBy != "" {
// 		if strings.EqualFold(sortOrder, sb.ASC) {
// 			q = q.Order(goqu.I(options.OrderBy).Asc())
// 		} else {
// 			q = q.Order(goqu.I(options.OrderBy).Desc())
// 		}
// 	}

// 	if !options.WithDeleted {
// 		q = q.Where(goqu.C(COLUMN_SOFT_DELETED_AT).Eq(sb.MAX_DATETIME))
// 	}

// 	return q
// }

// type LinkQueryOptions struct {
// 	ID          string
// 	IDIn        []string
// 	FeedID      string
// 	Status      string
// 	StatusIn    []string
// 	URL         string
// 	Offset      int
// 	Limit       int
// 	SortOrder   string
// 	OrderBy     string
// 	CountOnly   bool
// 	WithDeleted bool
// }
