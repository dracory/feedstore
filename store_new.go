package feedstore

import (
	"database/sql"
	"errors"

	"github.com/gouniverse/sb"
)

// NewStoreOptions define the options for creating a new block store
type NewStoreOptions struct {
	FeedTableName      string
	LinkTableName      string
	DB                 *sql.DB
	DbDriverName       string
	AutomigrateEnabled bool
	DebugEnabled       bool
}

// NewStore creates a new block store
func NewStore(opts NewStoreOptions) (StoreInterface, error) {
	if opts.FeedTableName == "" {
		return nil, errors.New("feed store: FeedTableName is required")
	}

	if opts.LinkTableName == "" {
		return nil, errors.New("feed store: LinkTableName is required")
	}

	if opts.DB == nil {
		return nil, errors.New("feed store: DB is required")
	}

	if opts.DbDriverName == "" {
		opts.DbDriverName = sb.DatabaseDriverName(opts.DB)
	}

	store := &storeImplementation{
		feedTableName:      opts.FeedTableName,
		linkTableName:      opts.LinkTableName,
		automigrateEnabled: opts.AutomigrateEnabled,
		db:                 opts.DB,
		dbDriverName:       opts.DbDriverName,
		debugEnabled:       opts.DebugEnabled,
	}

	if store.automigrateEnabled {
		err := store.AutoMigrate()

		if err != nil {
			return nil, err
		}
	}

	return store, nil
}
