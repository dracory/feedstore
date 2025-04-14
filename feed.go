package feedstore

import (
	"github.com/dromara/carbon/v2"
	"github.com/gouniverse/dataobject"
	"github.com/gouniverse/sb"
	"github.com/gouniverse/uid"
	"github.com/gouniverse/utils"
)

// ============================================================================
// == CLASS
// ============================================================================

type feedImplementation struct {
	dataobject.DataObject
}

// ============================================================================
// == INTERFACE
// ============================================================================

var _ FeedInterface = (*feedImplementation)(nil) // verify it extends the interface

// ============================================================================
// == CONSTRUCTOR
// ============================================================================

func NewFeed() *feedImplementation {
	feed := &feedImplementation{}
	feed.SetID(uid.NanoUid())
	feed.SetStatus(FEED_STATUS_INACTIVE)
	// feed.SetName("")
	feed.SetDescription("")
	feed.SetURL("")
	feed.SetFetchInterval("600")
	feed.SetLastFetchedAt(sb.NULL_DATETIME)
	feed.SetMemo("")
	feed.SetCreatedAt(carbon.Now(carbon.UTC).ToDateTimeString())
	feed.SetUpdatedAt(carbon.Now(carbon.UTC).ToDateTimeString())
	feed.SetSoftDeletedAt(sb.MAX_DATETIME)

	return feed
}

func NewFeedFromExistingData(data map[string]string) *feedImplementation {
	feed := &feedImplementation{}

	for k, v := range data {
		feed.Set(k, v)
	}

	feed.MarkAsNotDirty()

	return feed
}

// == SETTERS AND GETTERS =====================================================

func (feed *feedImplementation) CreatedAt() string {
	return feed.Get(COLUMN_CREATED_AT)
}
func (feed *feedImplementation) CreatedAtCarbon() *carbon.Carbon {
	return carbon.Parse(feed.CreatedAt())
}
func (feed *feedImplementation) SetCreatedAt(createdAt string) FeedInterface {
	feed.Set(COLUMN_CREATED_AT, createdAt)
	return feed
}

func (feed *feedImplementation) Description() string {
	return feed.Get(COLUMN_DESCRIPTION)
}

func (feed *feedImplementation) SetDescription(description string) FeedInterface {
	feed.Set(COLUMN_DESCRIPTION, description)
	return feed
}

func (feed *feedImplementation) FetchInterval() string {
	return feed.Get(COLUMN_FETCH_INTERVAL)
}

func (feed *feedImplementation) FetchIntervalInt64() (int64, error) {
	return utils.ToInt(feed.FetchInterval())
}

func (feed *feedImplementation) SetFetchInterval(fetchInterval string) FeedInterface {
	feed.Set(COLUMN_FETCH_INTERVAL, fetchInterval)
	return feed
}

func (feed *feedImplementation) ID() string {
	return feed.Get(COLUMN_ID)
}

func (feed *feedImplementation) SetID(id string) FeedInterface {
	feed.Set(COLUMN_ID, id)
	return feed
}

func (feed *feedImplementation) LastFetchedAt() string {
	return feed.Get(COLUMN_LAST_FETCHED_AT)
}
func (feed *feedImplementation) SetLastFetchedAt(lastFetchedAt string) FeedInterface {
	feed.Set(COLUMN_LAST_FETCHED_AT, lastFetchedAt)
	return feed
}

func (feed *feedImplementation) Memo() string {
	return feed.Get(COLUMN_MEMO)
}
func (feed *feedImplementation) SetMemo(memo string) FeedInterface {
	feed.Set(COLUMN_MEMO, memo)
	return feed
}

func (feed *feedImplementation) Name() string {
	return feed.Get(COLUMN_NAME)
}
func (feed *feedImplementation) SetName(name string) FeedInterface {
	feed.Set(COLUMN_NAME, name)
	return feed
}

func (feed *feedImplementation) SoftDeletedAt() string {
	return feed.Get(COLUMN_SOFT_DELETED_AT)
}

func (feed *feedImplementation) SoftDeletedAtCarbon() *carbon.Carbon {
	return carbon.Parse(feed.SoftDeletedAt())
}

func (feed *feedImplementation) SetSoftDeletedAt(softDeletedAt string) FeedInterface {
	feed.Set(COLUMN_SOFT_DELETED_AT, softDeletedAt)
	return feed
}

func (feed *feedImplementation) Status() string {
	return feed.Get(COLUMN_STATUS)
}
func (feed *feedImplementation) SetStatus(status string) FeedInterface {
	feed.Set(COLUMN_STATUS, status)
	return feed
}

func (feed *feedImplementation) UpdatedAt() string {
	return feed.Get(COLUMN_UPDATED_AT)
}
func (feed *feedImplementation) UpdatedAtCarbon() *carbon.Carbon {
	return carbon.Parse(feed.UpdatedAt())
}
func (feed *feedImplementation) SetUpdatedAt(updatedAt string) FeedInterface {
	feed.Set(COLUMN_UPDATED_AT, updatedAt)
	return feed
}

func (feed *feedImplementation) URL() string {
	return feed.Get(COLUMN_URL)
}

func (feed *feedImplementation) SetURL(url string) FeedInterface {
	feed.Set(COLUMN_URL, url)
	return feed
}
