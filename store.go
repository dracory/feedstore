package feedstore

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/dracory/neat"
	contractsorm "github.com/dracory/neat/contracts/database/orm"
	contractsschema "github.com/dracory/neat/contracts/database/schema"
	"github.com/dromara/carbon/v2"
	"github.com/samber/lo"
)

var _ StoreInterface = (*storeImplementation)(nil) // verify it extends the interface

type storeImplementation struct {
	feedTableName      string
	linkTableName      string
	db                 *neat.Database
	automigrateEnabled bool
	debugEnabled       bool
	logger             *slog.Logger
}

// MigrateUp creates the feed and link tables
func (st *storeImplementation) MigrateUp(ctx context.Context, tx ...*sql.Tx) error {
	if st.db.Schema().HasTable(st.feedTableName) {
		if st.debugEnabled {
			st.logger.Info("MigrateUp: feed table already exists", "table", st.feedTableName)
		}
	} else {
		err := st.db.Schema().Create(st.feedTableName, func(table contractsschema.Blueprint) {
			table.String(COLUMN_ID, 9)
			table.Primary(COLUMN_ID)
			table.String(COLUMN_NAME, 255)
			table.String(COLUMN_DESCRIPTION, 1024).Nullable()
			table.String(COLUMN_URL, 1024)
			table.String(COLUMN_STATUS, 50)
			table.String(COLUMN_FETCH_INTERVAL, 50)
			table.DateTime(COLUMN_LAST_FETCHED_AT).Nullable()
			table.Text(COLUMN_MEMO).Nullable()
			table.DateTime(COLUMN_CREATED_AT)
			table.DateTime(COLUMN_UPDATED_AT)
			table.DateTime(COLUMN_SOFT_DELETED_AT)
		})

		if err != nil {
			if st.debugEnabled {
				st.logger.Error("MigrateUp failed for feed table", "error", err)
			}
			return err
		}
	}

	if st.db.Schema().HasTable(st.linkTableName) {
		if st.debugEnabled {
			st.logger.Info("MigrateUp: link table already exists", "table", st.linkTableName)
		}
	} else {
		err := st.db.Schema().Create(st.linkTableName, func(table contractsschema.Blueprint) {
			table.String(COLUMN_ID, 9)
			table.Primary(COLUMN_ID)
			table.String(COLUMN_FEED_ID, 9)
			table.String(COLUMN_STATUS, 50)
			table.String(COLUMN_TITLE, 255)
			table.String(COLUMN_DESCRIPTION, 1024).Nullable()
			table.String(COLUMN_URL, 1024)
			table.String(COLUMN_VIEWS, 50)
			table.String(COLUMN_VOTES_UP, 50)
			table.String(COLUMN_VOTES_DOWN, 50)
			table.DateTime(COLUMN_REPORTED_AT).Nullable()
			table.Text(COLUMN_REPORT).Nullable()
			table.DateTime(COLUMN_CHECKED_AT).Nullable()
			table.DateTime(COLUMN_TIME).Nullable()
			table.DateTime(COLUMN_CREATED_AT)
			table.DateTime(COLUMN_UPDATED_AT)
			table.DateTime(COLUMN_SOFT_DELETED_AT)
			table.Index(COLUMN_FEED_ID)
		})

		if err != nil {
			if st.debugEnabled {
				st.logger.Error("MigrateUp failed for link table", "error", err)
			}
			return err
		}
	}

	return nil
}

// MigrateDown drops the feed and link tables
func (st *storeImplementation) MigrateDown(ctx context.Context, tx ...*sql.Tx) error {
	if st.db.Schema().HasTable(st.linkTableName) {
		err := st.db.Schema().Drop(st.linkTableName)
		if err != nil {
			if st.debugEnabled {
				st.logger.Error("MigrateDown failed for link table", "error", err)
			}
			return err
		}
	}

	if st.db.Schema().HasTable(st.feedTableName) {
		err := st.db.Schema().Drop(st.feedTableName)
		if err != nil {
			if st.debugEnabled {
				st.logger.Error("MigrateDown failed for feed table", "error", err)
			}
			return err
		}
	}

	return nil
}

// EnableDebug - enables the debug option
func (st *storeImplementation) EnableDebug(debug bool) {
	st.debugEnabled = debug
	if debug {
		st.db.EnableDebug()
		st.logger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	} else {
		st.db.DisableDebug()
		st.logger = slog.New(slog.NewTextHandler(os.Stdout, nil))
	}
}

func (st *storeImplementation) GetFeedTableName() string {
	return st.feedTableName
}

func (st *storeImplementation) GetLinkTableName() string {
	return st.linkTableName
}

func (st *storeImplementation) GetDB() *sql.DB {
	db, _ := st.db.DB()
	return db
}

// FeedCount returns the total number of feeds matching the query filters
func (st *storeImplementation) FeedCount(ctx context.Context, query FeedQueryInterface) (int64, error) {
	if query == nil {
		query = FeedQuery()
	}

	if err := query.Validate(); err != nil {
		return 0, err
	}

	q := st.buildFeedQuery(query)

	var count int64
	err := q.Count(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (st *storeImplementation) FeedCreate(ctx context.Context, feed FeedInterface) error {
	feed.SetCreatedAt(carbon.Now(carbon.UTC).ToDateTimeString())
	feed.SetUpdatedAt(carbon.Now(carbon.UTC).ToDateTimeString())

	// Convert feed implementation to map for neat
	data := st.feedToMap(feed)

	err := st.db.Query().Table(st.feedTableName).Create(data)
	if err != nil {
		return err
	}

	feed.MarkAsNotDirty()
	return nil
}

func (st *storeImplementation) FeedDelete(ctx context.Context, feed FeedInterface) error {
	if feed == nil {
		return errors.New("feed is nil")
	}

	return st.FeedDeleteByID(ctx, feed.ID())
}

func (st *storeImplementation) FeedDeleteByID(ctx context.Context, id string) error {
	if id == "" {
		return errors.New("feed id is empty")
	}

	_, err := st.db.Query().Table(st.feedTableName).Where("id = ?", id).Delete()
	return err
}

func (st *storeImplementation) FeedFindByID(ctx context.Context, id string) (FeedInterface, error) {
	if id == "" {
		return nil, errors.New("feed id is empty")
	}

	list, err := st.FeedList(ctx, FeedQuery().
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

func (st *storeImplementation) FeedList(ctx context.Context, query FeedQueryInterface) ([]FeedInterface, error) {
	if err := query.Validate(); err != nil {
		return []FeedInterface{}, err
	}

	q := st.buildFeedQuery(query)

	var results []map[string]any
	err := q.Get(&results)
	if err != nil {
		return []FeedInterface{}, err
	}

	list := []FeedInterface{}
	lo.ForEach(results, func(result map[string]any, index int) {
		model := st.mapToFeed(result)
		list = append(list, model)
	})

	return list, nil
}

func (st *storeImplementation) FeedSoftDelete(ctx context.Context, feed FeedInterface) error {
	if feed == nil {
		return errors.New("feed is nil")
	}

	feed.SetSoftDeletedAt(carbon.Now(carbon.UTC).ToDateTimeString())

	return st.FeedUpdate(ctx, feed)
}

func (st *storeImplementation) FeedSoftDeleteByID(ctx context.Context, id string) error {
	feed, err := st.FeedFindByID(ctx, id)

	if err != nil {
		return err
	}

	return st.FeedSoftDelete(ctx, feed)
}

func (st *storeImplementation) FeedUpdate(ctx context.Context, feed FeedInterface) error {
	if feed == nil {
		return errors.New("feed is nil")
	}

	data := st.feedToMap(feed)
	delete(data, COLUMN_ID) // ID is not updateable

	// Check if any meaningful field has changed
	feedImpl, ok := feed.(*feedImplementation)
	if ok && feedImpl.originalData != nil {
		hasChanges := false
		for k, v := range data {
			if k == COLUMN_ID || k == COLUMN_CREATED_AT || k == COLUMN_UPDATED_AT {
				continue
			}
			if feedImpl.originalData[k] != v {
				hasChanges = true
				break
			}
		}
		if !hasChanges {
			return nil
		}
	}

	feed.SetUpdatedAt(carbon.Now(carbon.UTC).ToDateTimeString())
	data = st.feedToMap(feed)
	delete(data, COLUMN_ID) // ID is not updateable

	_, err := st.db.Query().Table(st.feedTableName).Where("id = ?", feed.ID()).Update(data)
	if err != nil {
		return err
	}

	feed.MarkAsNotDirty()
	return nil
}

// LinkCount returns the total number of links matching the query filters
func (st *storeImplementation) LinkCount(ctx context.Context, query LinkQueryInterface) (int64, error) {
	if query == nil {
		query = LinkQuery()
	}

	if err := query.Validate(); err != nil {
		return 0, err
	}

	q := st.buildLinkQuery(query)

	var count int64
	err := q.Count(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (st *storeImplementation) LinkCreate(ctx context.Context, link LinkInterface) error {
	link.SetCreatedAt(carbon.Now(carbon.UTC).ToDateTimeString())
	link.SetUpdatedAt(carbon.Now(carbon.UTC).ToDateTimeString())

	data := st.linkToMap(link)

	err := st.db.Query().Table(st.linkTableName).Create(data)
	if err != nil {
		return err
	}

	link.MarkAsNotDirty()
	return nil
}

func (st *storeImplementation) LinkDelete(ctx context.Context, link LinkInterface) error {
	if link == nil {
		return errors.New("link is nil")
	}

	return st.LinkDeleteByID(ctx, link.ID())
}

func (st *storeImplementation) LinkDeleteByID(ctx context.Context, id string) error {
	if id == "" {
		return errors.New("link id is empty")
	}

	_, err := st.db.Query().Table(st.linkTableName).Where("id = ?", id).Delete()
	return err
}

func (st *storeImplementation) LinkFindByID(ctx context.Context, id string) (LinkInterface, error) {
	if id == "" {
		return nil, errors.New("link id is empty")
	}

	list, err := st.LinkList(ctx, LinkQuery().
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

func (st *storeImplementation) LinkList(ctx context.Context, query LinkQueryInterface) ([]LinkInterface, error) {
	if err := query.Validate(); err != nil {
		return []LinkInterface{}, err
	}

	q := st.buildLinkQuery(query)

	var results []map[string]any
	err := q.Get(&results)
	if err != nil {
		return []LinkInterface{}, err
	}

	list := []LinkInterface{}
	lo.ForEach(results, func(result map[string]any, index int) {
		model := st.mapToLink(result)
		list = append(list, model)
	})

	return list, nil
}

func (st *storeImplementation) LinkSoftDelete(ctx context.Context, link LinkInterface) error {
	if link == nil {
		return errors.New("link is nil")
	}

	link.SetSoftDeletedAt(carbon.Now(carbon.UTC).ToDateTimeString())

	return st.LinkUpdate(ctx, link)
}

func (st *storeImplementation) LinkSoftDeleteByID(ctx context.Context, id string) error {
	link, err := st.LinkFindByID(ctx, id)

	if err != nil {
		return err
	}

	return st.LinkSoftDelete(ctx, link)
}

func (st *storeImplementation) LinkUpdate(ctx context.Context, link LinkInterface) error {
	if link == nil {
		return errors.New("link is nil")
	}

	data := st.linkToMap(link)
	delete(data, COLUMN_ID) // ID is not updateable

	// Check if any meaningful field has changed
	linkImpl, ok := link.(*linkImplementation)
	if ok && linkImpl.originalData != nil {
		hasChanges := false
		for k, v := range data {
			if k == COLUMN_ID || k == COLUMN_CREATED_AT || k == COLUMN_UPDATED_AT {
				continue
			}
			if linkImpl.originalData[k] != v {
				hasChanges = true
				break
			}
		}
		if !hasChanges {
			return nil
		}
	}

	link.SetUpdatedAt(carbon.Now(carbon.UTC).ToDateTimeString())
	data = st.linkToMap(link)
	delete(data, COLUMN_ID) // ID is not updateable

	_, err := st.db.Query().Table(st.linkTableName).Where("id = ?", link.ID()).Update(data)
	if err != nil {
		return err
	}

	link.MarkAsNotDirty()
	return nil
}

// Helper methods for building queries and converting data

func (st *storeImplementation) buildFeedQuery(query FeedQueryInterface) contractsorm.Query {
	q := st.db.Query().Table(st.feedTableName)

	// ID filter
	if query.IsIDSet() {
		q = q.Where("id = ?", query.GetID())
	}

	// ID IN filter
	if query.IsIDInSet() {
		q = q.WhereIn("id", lo.ToAnySlice(query.GetIDIn()))
	}

	// Status filter
	if query.IsStatusSet() {
		q = q.Where("status = ?", query.GetStatus())
	}

	// Status IN filter
	if query.IsStatusInSet() {
		q = q.WhereIn("status", lo.ToAnySlice(query.GetStatusIn()))
	}

	// Created At filters
	if query.IsCreatedAtGteSet() {
		q = q.Where("created_at >= ?", query.GetCreatedAtGte())
	}

	if query.IsCreatedAtLteSet() {
		q = q.Where("created_at <= ?", query.GetCreatedAtLte())
	}

	// Updated At filters
	if query.IsUpdatedAtGteSet() {
		q = q.Where("updated_at >= ?", query.GetUpdatedAtGte())
	}

	if query.IsUpdatedAtLteSet() {
		q = q.Where("updated_at <= ?", query.GetUpdatedAtLte())
	}

	// Last Fetched At filters
	if query.IsLastFetchedAtGteSet() {
		q = q.Where("last_fetched_at >= ?", query.GetLastFetchedAtGte())
	}

	if query.IsLastFetchedAtLteSet() {
		q = q.Where("last_fetched_at <= ?", query.GetLastFetchedAtLte())
	}

	// Soft delete filters
	if query.IsOnlySoftDeletedSet() && query.GetOnlySoftDeleted() {
		q = q.Where("soft_deleted_at <= ?", carbon.Now(carbon.UTC).ToDateTimeString())
	} else if query.IsWithSoftDeletedSet() && query.GetWithSoftDeleted() {
		// Include soft deleted
		// No filter needed
	} else {
		// Exclude soft deleted by default
		q = q.Where("soft_deleted_at = ?", MAX_DATETIME)
	}

	// Ordering
	if query.IsOrderBySet() {
		orderDirection := "desc"
		if query.IsOrderDirectionSet() {
			orderDirection = query.GetOrderDirection()
		}
		q = q.OrderBy(query.GetOrderBy(), orderDirection)
	}

	// Limit and Offset
	if query.IsLimitSet() {
		q = q.Limit(query.GetLimit())
	}

	if query.IsOffsetSet() {
		q = q.Offset(query.GetOffset())
	}

	return q
}

func (st *storeImplementation) buildLinkQuery(query LinkQueryInterface) contractsorm.Query {
	q := st.db.Query().Table(st.linkTableName)

	// ID filter
	if query.IsIDSet() {
		q = q.Where("id = ?", query.GetID())
	}

	// ID IN filter
	if query.IsIDInSet() {
		q = q.WhereIn("id", lo.ToAnySlice(query.GetIDIn()))
	}

	// Feed ID filter
	if query.IsFeedIDSet() {
		q = q.Where("feed_id = ?", query.GetFeedID())
	}

	// Status filter
	if query.IsStatusSet() {
		q = q.Where("status = ?", query.GetStatus())
	}

	// Status IN filter
	if query.IsStatusInSet() {
		q = q.WhereIn("status", lo.ToAnySlice(query.GetStatusIn()))
	}

	// URL filter
	if query.IsURLSet() {
		q = q.Where("url = ?", query.GetURL())
	}

	// Created At filters
	if query.IsCreatedAtGteSet() {
		q = q.Where("created_at >= ?", query.GetCreatedAtGte())
	}

	if query.IsCreatedAtLteSet() {
		q = q.Where("created_at <= ?", query.GetCreatedAtLte())
	}

	// Updated At filters
	if query.IsUpdatedAtGteSet() {
		q = q.Where("updated_at >= ?", query.GetUpdatedAtGte())
	}

	if query.IsUpdatedAtLteSet() {
		q = q.Where("updated_at <= ?", query.GetUpdatedAtLte())
	}

	// Soft delete filters
	if query.IsOnlySoftDeletedSet() && query.GetOnlySoftDeleted() {
		q = q.Where("soft_deleted_at <= ?", carbon.Now(carbon.UTC).ToDateTimeString())
	} else if query.IsWithSoftDeletedSet() && query.GetWithSoftDeleted() {
		// Include soft deleted
		// No filter needed
	} else {
		// Exclude soft deleted by default
		q = q.Where("soft_deleted_at = ?", MAX_DATETIME)
	}

	// Ordering
	if query.IsOrderBySet() {
		orderDirection := "desc"
		if query.IsOrderDirectionSet() {
			orderDirection = query.GetOrderDirection()
		}
		q = q.OrderBy(query.GetOrderBy(), orderDirection)
	}

	// Limit and Offset
	if query.IsLimitSet() {
		q = q.Limit(query.GetLimit())
	}

	if query.IsOffsetSet() {
		q = q.Offset(query.GetOffset())
	}

	return q
}

func (st *storeImplementation) feedToMap(feed FeedInterface) map[string]any {
	return map[string]any{
		COLUMN_ID:              feed.ID(),
		COLUMN_NAME:            feed.Name(),
		COLUMN_DESCRIPTION:     feed.Description(),
		COLUMN_URL:             feed.URL(),
		COLUMN_STATUS:          feed.Status(),
		COLUMN_FETCH_INTERVAL:  feed.FetchInterval(),
		COLUMN_LAST_FETCHED_AT: feed.LastFetchedAt(),
		COLUMN_MEMO:            feed.Memo(),
		COLUMN_CREATED_AT:      feed.CreatedAt(),
		COLUMN_UPDATED_AT:      feed.UpdatedAt(),
		COLUMN_SOFT_DELETED_AT: feed.GetSoftDeletedAt(),
	}
}

func (st *storeImplementation) linkToMap(link LinkInterface) map[string]any {
	return map[string]any{
		COLUMN_ID:              link.ID(),
		COLUMN_FEED_ID:         link.FeedID(),
		COLUMN_STATUS:          link.Status(),
		COLUMN_TITLE:           link.Title(),
		COLUMN_DESCRIPTION:     link.Description(),
		COLUMN_URL:             link.URL(),
		COLUMN_VIEWS:           link.Views(),
		COLUMN_VOTES_UP:        link.VotesUp(),
		COLUMN_VOTES_DOWN:      link.VotesDown(),
		COLUMN_REPORTED_AT:     link.ReportedAt(),
		COLUMN_REPORT:          link.Report(),
		COLUMN_CHECKED_AT:      link.CheckedAt(),
		COLUMN_TIME:            link.Time(),
		COLUMN_CREATED_AT:      link.CreatedAt(),
		COLUMN_UPDATED_AT:      link.UpdatedAt(),
		COLUMN_SOFT_DELETED_AT: link.GetSoftDeletedAt(),
	}
}

func (st *storeImplementation) mapToFeed(data map[string]any) FeedInterface {
	stringData := make(map[string]string)
	for k, v := range data {
		if v != nil {
			stringData[k] = toString(v)
		} else {
			stringData[k] = ""
		}
	}
	return NewFeedFromExistingData(stringData)
}

func (st *storeImplementation) mapToLink(data map[string]any) LinkInterface {
	stringData := make(map[string]string)
	for k, v := range data {
		if v != nil {
			stringData[k] = toString(v)
		} else {
			stringData[k] = ""
		}
	}
	return NewLinkFromExistingData(stringData)
}

func toString(v any) string {
	switch val := v.(type) {
	case string:
		return val
	case time.Time:
		if val.IsZero() {
			return ""
		}
		return carbon.CreateFromStdTime(val).ToDateTimeString()
	case []byte:
		return string(val)
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		return fmt.Sprintf("%d", val)
	case float32, float64:
		return fmt.Sprintf("%f", val)
	case bool:
		return fmt.Sprintf("%t", val)
	default:
		return fmt.Sprintf("%v", val)
	}
}
