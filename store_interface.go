package feedstore

type StoreInterface interface {
	AutoMigrate() error
	EnableDebug(debug bool)

	GetDriverName() string
	GetFeedTableName() string
	GetLinkTableName() string

	FeedCount(query FeedQueryInterface) (int64, error)
	FeedCreate(feed FeedInterface) error
	FeedDelete(feed FeedInterface) error
	FeedDeleteByID(id string) error
	FeedFindByID(id string) (FeedInterface, error)
	FeedList(query FeedQueryInterface) ([]FeedInterface, error)
	FeedSoftDelete(feed FeedInterface) error
	FeedSoftDeleteByID(id string) error
	FeedUpdate(feed FeedInterface) error

	LinkCount(query LinkQueryInterface) (int64, error)
	LinkCreate(link LinkInterface) error
	LinkDelete(link LinkInterface) error
	LinkDeleteByID(id string) error
	LinkFindByID(id string) (LinkInterface, error)
	LinkList(query LinkQueryInterface) ([]LinkInterface, error)
	LinkSoftDelete(link LinkInterface) error
	LinkSoftDeleteByID(id string) error
	LinkUpdate(link LinkInterface) error
}
