package feedstore

type StoreInterface interface {
	AutoMigrate() error
	FeedCreate(feed FeedInterface) error
	FeedDelete(feed FeedInterface) error
	FeedDeleteByID(id string) error
	FeedFindByID(id string) (FeedInterface, error)
	FeedList(options FeedQueryOptions) ([]FeedInterface, error)
	FeedSoftDelete(feed FeedInterface) error
	FeedSoftDeleteByID(id string) error
	FeedUpdate(feed FeedInterface) error
	LinkCreate(link LinkInterface) error
	LinkDelete(link LinkInterface) error
	LinkDeleteByID(id string) error
	LinkFindByID(id string) (LinkInterface, error)
	LinkList(options LinkQueryOptions) ([]LinkInterface, error)
	LinkSoftDelete(link LinkInterface) error
	LinkSoftDeleteByID(id string) error
	LinkUpdate(link LinkInterface) error
}
