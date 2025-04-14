package feedstore

import "github.com/dromara/carbon/v2"

type FeedInterface interface {
	Data() map[string]string
	DataChanged() map[string]string
	MarkAsNotDirty()

	CreatedAt() string
	CreatedAtCarbon() *carbon.Carbon
	SetCreatedAt(createdAt string) FeedInterface
	DeletedAt() string
	DeletedAtCarbon() *carbon.Carbon
	SetDeletedAt(deletedAt string) FeedInterface
	Description() string
	SetDescription(description string) FeedInterface
	FetchInterval() string
	SetFetchInterval(fetchInterval string) FeedInterface
	ID() string
	SetID(id string) FeedInterface
	LastFetchedAt() string
	SetLastFetchedAt(lastFetchedAt string) FeedInterface
	Memo() string
	SetMemo(memo string) FeedInterface
	Name() string
	SetName(name string) FeedInterface
	Status() string
	SetStatus(status string) FeedInterface
	UpdatedAt() string
	UpdatedAtCarbon() *carbon.Carbon
	SetUpdatedAt(updatedAt string) FeedInterface
	URL() string
	SetURL(url string) FeedInterface
}
