package feedstore

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"os"

	"github.com/dracory/neat"
)

// NewStoreOptions define the options for creating a new feed store
type NewStoreOptions struct {
	FeedTableName      string
	LinkTableName      string
	DB                 *sql.DB
	AutomigrateEnabled bool
	DebugEnabled       bool
}

// NewStore creates a new feed store
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

	neatDB, err := neat.NewFromSQLDB(opts.DB)
	if err != nil {
		return nil, err
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	store := &storeImplementation{
		feedTableName:      opts.FeedTableName,
		linkTableName:      opts.LinkTableName,
		automigrateEnabled: opts.AutomigrateEnabled,
		db:                 neatDB,
		debugEnabled:       opts.DebugEnabled,
		logger:             logger,
	}

	if store.automigrateEnabled {
		err := store.MigrateUp(context.Background())

		if err != nil {
			return nil, err
		}
	}

	return store, nil
}
