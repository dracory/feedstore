package feedstore

import "context"

type StoreInterface interface {
	AutoMigrate() error
	EnableDebug(debug bool)

	GetDriverName() string
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
