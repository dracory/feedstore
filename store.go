package feedstore

import (
	"context"
	"database/sql"
	"encoding/json"
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

// ============================================================================
// == INTERFACE
// ============================================================================

type StoreInterface interface {
	// MigrateDown drops the feed and link tables
	MigrateDown(ctx context.Context, tx ...*sql.Tx) error

	// MigrateUp creates the feed and link tables
	MigrateUp(ctx context.Context, tx ...*sql.Tx) error
	EnableDebug(debug bool)

	GetDB() *sql.DB
	GetFeedTableName() string
	GetLinkTableName() string

	FeedCount(ctx context.Context, query FeedQueryInterface) (int64, error)
	FeedCreate(ctx context.Context, feed FeedInterface) error
	FeedDelete(ctx context.Context, feed FeedInterface) error
	FeedDeleteByID(ctx context.Context, id string) error
	FeedFindByID(ctx context.Context, id string) (FeedInterface, error)
	FeedList(ctx context.Context, query FeedQueryInterface) ([]FeedInterface, error)
	FeedSoftDelete(ctx context.Context, feed FeedInterface) error
	FeedSoftDeleteByID(ctx context.Context, id string) error
	FeedUpdate(ctx context.Context, feed FeedInterface) error

	LinkCount(ctx context.Context, query LinkQueryInterface) (int64, error)
	LinkCreate(ctx context.Context, link LinkInterface) error
	LinkDelete(ctx context.Context, link LinkInterface) error
	LinkDeleteByID(ctx context.Context, id string) error
	LinkFindByID(ctx context.Context, id string) (LinkInterface, error)
	LinkList(ctx context.Context, query LinkQueryInterface) ([]LinkInterface, error)
	LinkSoftDelete(ctx context.Context, link LinkInterface) error
	LinkSoftDeleteByID(ctx context.Context, id string) error
	LinkUpdate(ctx context.Context, link LinkInterface) error
}

// ============================================================================
// == TYPE
// ============================================================================

var _ StoreInterface = (*storeImplementation)(nil) // verify it extends the interface

type storeImplementation struct {
	feedTableName      string
	linkTableName      string
	db                 *neat.Database
	automigrateEnabled bool
	debugEnabled       bool
	logger             *slog.Logger
}

// ============================================================================
// == MIGRATE
// ============================================================================

// MigrateUp creates the feed and link tables
func (st *storeImplementation) MigrateUp(ctx context.Context, tx ...*sql.Tx) error {
	if st.db.Schema().HasTable(st.feedTableName) {
		if st.debugEnabled {
			st.logger.Info("MigrateUp: feed table already exists", "table", st.feedTableName)
		}

		// Upgrade path: add columns introduced after the table was created
		if !st.db.Schema().HasColumn(st.feedTableName, COLUMN_LANGUAGE) {
			err := st.db.Schema().Table(st.feedTableName, func(table contractsschema.Blueprint) {
				table.String(COLUMN_LANGUAGE, 10).Nullable()
			})
			if err != nil {
				return err
			}
		}
		if !st.db.Schema().HasColumn(st.feedTableName, COLUMN_METAS) {
			err := st.db.Schema().Table(st.feedTableName, func(table contractsschema.Blueprint) {
				table.LongText(COLUMN_METAS).Nullable()
			})
			if err != nil {
				return err
			}
		}
		// Widen ID column to 40 to support GUIDs/UUIDs
		if st.db.Schema().HasColumn(st.feedTableName, COLUMN_ID) {
			err := st.db.Schema().Table(st.feedTableName, func(table contractsschema.Blueprint) {
				table.String(COLUMN_ID, 40).Change()
			})
			if err != nil {
				return err
			}
		}
	} else {
		err := st.db.Schema().Create(st.feedTableName, func(table contractsschema.Blueprint) {
			table.String(COLUMN_ID, 40)
			table.Primary(COLUMN_ID)
			table.String(COLUMN_NAME, 255)
			table.String(COLUMN_DESCRIPTION, 1024).Nullable()
			table.String(COLUMN_URL, 1024)
			table.String(COLUMN_STATUS, 50)
			table.String(COLUMN_FETCH_INTERVAL, 50)
			table.DateTime(COLUMN_LAST_FETCHED_AT).Nullable()
			table.String(COLUMN_LANGUAGE, 10).Nullable()
			table.Text(COLUMN_MEMO).Nullable()
			table.LongText(COLUMN_METAS).Nullable()
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

		// Upgrade path: add columns introduced after the table was created
		if !st.db.Schema().HasColumn(st.linkTableName, COLUMN_LANGUAGE) {
			err := st.db.Schema().Table(st.linkTableName, func(table contractsschema.Blueprint) {
				table.String(COLUMN_LANGUAGE, 10).Nullable()
			})
			if err != nil {
				return err
			}
		}
		if !st.db.Schema().HasColumn(st.linkTableName, COLUMN_METAS) {
			err := st.db.Schema().Table(st.linkTableName, func(table contractsschema.Blueprint) {
				table.LongText(COLUMN_METAS).Nullable()
			})
			if err != nil {
				return err
			}
		}
		// Rename time → published_at for existing tables.
		// "time" is the old column name (hardcoded, not a constant — COLUMN_TIME was removed).
		if st.db.Schema().HasColumn(st.linkTableName, "time") {
			err := st.db.Schema().Table(st.linkTableName, func(table contractsschema.Blueprint) {
				table.RenameColumn("time", COLUMN_PUBLISHED_AT)
			})
			if err != nil {
				return err
			}
		}
		// Add dedup_hash column for existing tables
		if !st.db.Schema().HasColumn(st.linkTableName, COLUMN_DEDUP_HASH) {
			err := st.db.Schema().Table(st.linkTableName, func(table contractsschema.Blueprint) {
				table.String(COLUMN_DEDUP_HASH, 64).Nullable()
			})
			if err != nil {
				return err
			}
		}
		// Add dedup_hash index for existing tables (guard with HasIndex for idempotency).
		// neat generates index names as {table}_{column}_index, so we must check
		// the full index name, not just the column name.
		dedupHashIndexName := st.linkTableName + "_" + COLUMN_DEDUP_HASH + "_index"
		if !st.db.Schema().HasIndex(st.linkTableName, dedupHashIndexName) {
			err := st.db.Schema().Table(st.linkTableName, func(table contractsschema.Blueprint) {
				table.Index(COLUMN_DEDUP_HASH)
			})
			if err != nil {
				return err
			}
		}
		// Widen ID column to 40 to support GUIDs/UUIDs
		if st.db.Schema().HasColumn(st.linkTableName, COLUMN_ID) {
			err := st.db.Schema().Table(st.linkTableName, func(table contractsschema.Blueprint) {
				table.String(COLUMN_ID, 40).Change()
			})
			if err != nil {
				return err
			}
		}
		// Widen feed_id column to 40 to match the feed table's ID column
		if st.db.Schema().HasColumn(st.linkTableName, COLUMN_FEED_ID) {
			err := st.db.Schema().Table(st.linkTableName, func(table contractsschema.Blueprint) {
				table.String(COLUMN_FEED_ID, 40).Change()
			})
			if err != nil {
				return err
			}
		}
	} else {
		err := st.db.Schema().Create(st.linkTableName, func(table contractsschema.Blueprint) {
			table.String(COLUMN_ID, 40)
			table.Primary(COLUMN_ID)
			table.String(COLUMN_FEED_ID, 40)
			table.String(COLUMN_STATUS, 50)
			table.String(COLUMN_TITLE, 255)
			table.String(COLUMN_DESCRIPTION, 1024).Nullable()
			table.Text(COLUMN_CONTENT).Nullable()
			table.String(COLUMN_AUTHOR, 255).Nullable()
			table.String(COLUMN_PRIORITY, 1).Nullable()
			table.String(COLUMN_URL, 1024)
			table.String(COLUMN_VIEWS, 50)
			table.String(COLUMN_VOTES_UP, 50)
			table.String(COLUMN_VOTES_DOWN, 50)
			table.DateTime(COLUMN_REPORTED_AT).Nullable()
			table.Text(COLUMN_REPORT).Nullable()
			table.DateTime(COLUMN_CHECKED_AT).Nullable()
			table.DateTime(COLUMN_PUBLISHED_AT).Nullable()
			table.String(COLUMN_DEDUP_HASH, 64).Nullable()
			table.String(COLUMN_LANGUAGE, 10).Nullable()
			table.LongText(COLUMN_METAS).Nullable()
			table.DateTime(COLUMN_CREATED_AT)
			table.DateTime(COLUMN_UPDATED_AT)
			table.DateTime(COLUMN_SOFT_DELETED_AT)
			table.Index(COLUMN_FEED_ID)
			table.Index(COLUMN_DEDUP_HASH)
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

// ============================================================================
// == DEBUG
// ============================================================================

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

// ============================================================================
// == DB
// ============================================================================

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

// ============================================================================
// == FEED CRUD
// ============================================================================

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
	data, err := st.feedToMap(feed)
	if err != nil {
		return err
	}

	err = st.db.Query().Table(st.feedTableName).Create(data)
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

	return st.FeedDeleteByID(ctx, feed.GetID())
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

	data, err := st.feedToMap(feed)
	if err != nil {
		return err
	}
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
	data, err = st.feedToMap(feed)
	if err != nil {
		return err
	}
	delete(data, COLUMN_ID) // ID is not updateable

	_, err = st.db.Query().Table(st.feedTableName).Where("id = ?", feed.GetID()).Update(data)
	if err != nil {
		return err
	}

	feed.MarkAsNotDirty()
	return nil
}

// ============================================================================
// == LINK CRUD
// ============================================================================

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

	data, err := st.linkToMap(link)
	if err != nil {
		return err
	}

	err = st.db.Query().Table(st.linkTableName).Create(data)
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

	return st.LinkDeleteByID(ctx, link.GetID())
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

	data, err := st.linkToMap(link)
	if err != nil {
		return err
	}
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
	data, err = st.linkToMap(link)
	if err != nil {
		return err
	}
	delete(data, COLUMN_ID) // ID is not updateable

	_, err = st.db.Query().Table(st.linkTableName).Where("id = ?", link.GetID()).Update(data)
	if err != nil {
		return err
	}

	link.MarkAsNotDirty()
	return nil
}

// ============================================================================
// == QUERY BUILDER
// ============================================================================

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

	// Language filter
	if query.IsLanguageSet() {
		q = q.Where("language = ?", query.GetLanguage())
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
		q = q.Where("soft_deleted_at = ?", neat.MaxDateTime)
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

	// Language filter
	if query.IsLanguageSet() {
		q = q.Where("language = ?", query.GetLanguage())
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

	// PublishedAt (publication datetime) filters
	if query.IsPublishedAtGteSet() {
		q = q.Where("published_at >= ?", query.GetPublishedAtGte())
	}

	if query.IsPublishedAtLteSet() {
		q = q.Where("published_at <= ?", query.GetPublishedAtLte())
	}

	// Dedup hash filter
	if query.IsDedupHashSet() {
		q = q.Where("dedup_hash = ?", query.GetDedupHash())
	}

	// Priority filter
	if query.IsPrioritySet() {
		q = q.Where("priority = ?", query.GetPriority())
	}

	// Soft delete filters
	if query.IsOnlySoftDeletedSet() && query.GetOnlySoftDeleted() {
		q = q.Where("soft_deleted_at <= ?", carbon.Now(carbon.UTC).ToDateTimeString())
	} else if query.IsWithSoftDeletedSet() && query.GetWithSoftDeleted() {
		// Include soft deleted
		// No filter needed
	} else {
		// Exclude soft deleted by default
		q = q.Where("soft_deleted_at = ?", neat.MaxDateTime)
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

// ============================================================================
// == HELPERS
// ============================================================================

func (st *storeImplementation) feedToMap(feed FeedInterface) (map[string]any, error) {
	metas, err := feed.GetMetas()
	if err != nil {
		return nil, err
	}

	metasBytes, err := json.Marshal(metas)
	if err != nil {
		return nil, err
	}

	return map[string]any{
		COLUMN_ID:              feed.GetID(),
		COLUMN_NAME:            feed.GetName(),
		COLUMN_DESCRIPTION:     feed.GetDescription(),
		COLUMN_URL:             feed.GetURL(),
		COLUMN_STATUS:          feed.GetStatus(),
		COLUMN_FETCH_INTERVAL:  feed.GetFetchInterval(),
		COLUMN_LAST_FETCHED_AT: feed.GetLastFetchedAt(),
		COLUMN_LANGUAGE:        feed.GetLanguage(),
		COLUMN_MEMO:            feed.GetMemo(),
		COLUMN_METAS:           string(metasBytes),
		COLUMN_CREATED_AT:      feed.GetCreatedAt(),
		COLUMN_UPDATED_AT:      feed.GetUpdatedAt(),
		COLUMN_SOFT_DELETED_AT: feed.GetSoftDeletedAt(),
	}, nil
}

func (st *storeImplementation) linkToMap(link LinkInterface) (map[string]any, error) {
	metas, err := link.GetMetas()
	if err != nil {
		return nil, err
	}

	metasBytes, err := json.Marshal(metas)
	if err != nil {
		return nil, err
	}

	return map[string]any{
		COLUMN_ID:              link.GetID(),
		COLUMN_FEED_ID:         link.GetFeedID(),
		COLUMN_STATUS:          link.GetStatus(),
		COLUMN_TITLE:           link.GetTitle(),
		COLUMN_DESCRIPTION:     link.GetDescription(),
		COLUMN_CONTENT:         link.GetContent(),
		COLUMN_AUTHOR:          link.GetAuthor(),
		COLUMN_PRIORITY:        priorityToStr(link.GetPriority()),
		COLUMN_URL:             link.GetURL(),
		COLUMN_VIEWS:           link.GetViews(),
		COLUMN_VOTES_UP:        link.GetVotesUp(),
		COLUMN_VOTES_DOWN:      link.GetVotesDown(),
		COLUMN_REPORTED_AT:     link.GetReportedAt(),
		COLUMN_REPORT:          link.GetReport(),
		COLUMN_CHECKED_AT:      link.GetCheckedAt(),
		COLUMN_PUBLISHED_AT:    link.GetPublishedAt(),
		COLUMN_DEDUP_HASH:      link.GetDedupHash(),
		COLUMN_LANGUAGE:        link.GetLanguage(),
		COLUMN_METAS:           string(metasBytes),
		COLUMN_CREATED_AT:      link.GetCreatedAt(),
		COLUMN_UPDATED_AT:      link.GetUpdatedAt(),
		COLUMN_SOFT_DELETED_AT: link.GetSoftDeletedAt(),
	}, nil
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

func priorityToStr(b bool) string {
	if b {
		return "1"
	}
	return "0"
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
