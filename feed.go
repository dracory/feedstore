package feedstore

import (
	"time"

	"github.com/dracory/neat/database/orm"
	"github.com/dracory/neat/database/soft_delete"
	neatuid "github.com/dracory/neat/support/uid"
	"github.com/dromara/carbon/v2"
	"github.com/spf13/cast"
)

// ============================================================================
// == CLASS
// ============================================================================

type feedImplementation struct {
	orm.ShortID

	NameField          string    `db:"name"`
	DescriptionField   string    `db:"description"`
	URLField           string    `db:"url"`
	StatusField        string    `db:"status"`
	FetchIntervalField string    `db:"fetch_interval"`
	LastFetchedAtField time.Time `db:"last_fetched_at"`
	MemoField          string    `db:"memo"`
	CreatedAtField     orm.CreatedAt
	UpdatedAtField     orm.UpdatedAt
	soft_delete.SoftDeletesMaxDate

	originalData map[string]string
}

// ============================================================================
// == INTERFACE
// ============================================================================

type FeedInterface interface {
	Data() map[string]string
	MarkAsNotDirty(...string)

	CreatedAt() string
	CreatedAtCarbon() *carbon.Carbon
	SetCreatedAt(createdAt string) FeedInterface
	Description() string
	SetDescription(description string) FeedInterface
	FetchInterval() string
	SetFetchInterval(fetchInterval string) FeedInterface
	ID() string
	SetID(id string) FeedInterface
	LastFetchedAt() string
	LastFetchedAtCarbon() *carbon.Carbon
	SetLastFetchedAt(lastFetchedAt time.Time) FeedInterface
	SetLastFetchedAtString(lastFetchedAt string) FeedInterface
	Memo() string
	SetMemo(memo string) FeedInterface
	Name() string
	SetName(name string) FeedInterface
	GetSoftDeletedAt() string
	GetSoftDeletedAtCarbon() *carbon.Carbon
	SetSoftDeletedAt(softDeletedAt string) FeedInterface
	Status() string
	SetStatus(status string) FeedInterface
	UpdatedAt() string
	UpdatedAtCarbon() *carbon.Carbon
	SetUpdatedAt(updatedAt string) FeedInterface
	URL() string
	SetURL(url string) FeedInterface
}

var _ FeedInterface = (*feedImplementation)(nil)

// ============================================================================
// == CONSTRUCTOR
// ============================================================================

func NewFeed() *feedImplementation {
	feed := &feedImplementation{}
	feed.SetID(neatuid.GenerateShortID())
	feed.SetStatus(FEED_STATUS_INACTIVE)
	feed.SetDescription("")
	feed.SetURL("")
	feed.SetFetchInterval("600")
	feed.SetLastFetchedAt(time.Time{})
	feed.SetMemo("")
	feed.SetCreatedAt(carbon.Now(carbon.UTC).ToDateTimeString())
	feed.SetUpdatedAt(carbon.Now(carbon.UTC).ToDateTimeString())
	feed.SetSoftDeletedAt(MAX_DATETIME)
	feed.MarkAsNotDirty()

	return feed
}

func NewFeedFromExistingData(data map[string]string) *feedImplementation {
	feed := &feedImplementation{}

	feed.SetID(data[COLUMN_ID])
	feed.SetName(data[COLUMN_NAME])
	feed.SetDescription(data[COLUMN_DESCRIPTION])
	feed.SetURL(data[COLUMN_URL])
	feed.SetStatus(data[COLUMN_STATUS])
	feed.SetFetchInterval(data[COLUMN_FETCH_INTERVAL])
	if v, ok := data[COLUMN_LAST_FETCHED_AT]; ok {
		feed.SetLastFetchedAtString(v)
	}
	feed.SetMemo(data[COLUMN_MEMO])
	if v, ok := data[COLUMN_CREATED_AT]; ok {
		feed.SetCreatedAt(v)
	}
	if v, ok := data[COLUMN_UPDATED_AT]; ok {
		feed.SetUpdatedAt(v)
	}
	if v, ok := data[COLUMN_SOFT_DELETED_AT]; ok {
		feed.SetSoftDeletedAt(v)
	}
	feed.MarkAsNotDirty()

	return feed
}

// == SETTERS AND GETTERS =====================================================

func (feed *feedImplementation) CreatedAt() string {
	if feed.CreatedAtField.CreatedAt.IsZero() {
		return ""
	}
	return carbon.CreateFromStdTime(feed.CreatedAtField.CreatedAt).ToDateTimeString()
}
func (feed *feedImplementation) CreatedAtCarbon() *carbon.Carbon {
	return carbon.CreateFromStdTime(feed.CreatedAtField.CreatedAt)
}
func (feed *feedImplementation) SetCreatedAt(createdAt string) FeedInterface {
	if createdAt == "" {
		return feed
	}
	feed.CreatedAtField.CreatedAt = carbon.Parse(createdAt, carbon.UTC).StdTime()
	return feed
}

func (feed *feedImplementation) Description() string {
	return feed.DescriptionField
}

func (feed *feedImplementation) SetDescription(description string) FeedInterface {
	feed.DescriptionField = description
	return feed
}

func (feed *feedImplementation) FetchInterval() string {
	return feed.FetchIntervalField
}

func (feed *feedImplementation) FetchIntervalInt64() (int64, error) {
	return cast.ToInt64E(feed.FetchInterval())
}

func (feed *feedImplementation) SetFetchInterval(fetchInterval string) FeedInterface {
	feed.FetchIntervalField = fetchInterval
	return feed
}

func (feed *feedImplementation) ID() string {
	return feed.ShortID.ID
}

func (feed *feedImplementation) SetID(id string) FeedInterface {
	feed.ShortID.ID = id
	return feed
}

func (feed *feedImplementation) LastFetchedAt() string {
	if feed.LastFetchedAtField.IsZero() {
		return ""
	}
	return carbon.CreateFromStdTime(feed.LastFetchedAtField).ToDateTimeString()
}

func (feed *feedImplementation) LastFetchedAtCarbon() *carbon.Carbon {
	return carbon.CreateFromStdTime(feed.LastFetchedAtField)
}

func (feed *feedImplementation) SetLastFetchedAt(lastFetchedAt time.Time) FeedInterface {
	feed.LastFetchedAtField = lastFetchedAt
	return feed
}

func (feed *feedImplementation) SetLastFetchedAtString(lastFetchedAt string) FeedInterface {
	if lastFetchedAt == "" {
		feed.LastFetchedAtField = time.Time{}
		return feed
	}
	feed.LastFetchedAtField = carbon.Parse(lastFetchedAt, carbon.UTC).StdTime()
	return feed
}

func (feed *feedImplementation) Memo() string {
	return feed.MemoField
}
func (feed *feedImplementation) SetMemo(memo string) FeedInterface {
	feed.MemoField = memo
	return feed
}

func (feed *feedImplementation) Name() string {
	return feed.NameField
}
func (feed *feedImplementation) SetName(name string) FeedInterface {
	feed.NameField = name
	return feed
}

func (feed *feedImplementation) GetSoftDeletedAt() string {
	if feed.SoftDeletedAt.IsZero() {
		return ""
	}
	return carbon.CreateFromStdTime(feed.SoftDeletedAt).ToDateTimeString()
}

func (feed *feedImplementation) GetSoftDeletedAtCarbon() *carbon.Carbon {
	return carbon.CreateFromStdTime(feed.SoftDeletedAt)
}

func (feed *feedImplementation) SetSoftDeletedAt(softDeletedAt string) FeedInterface {
	if softDeletedAt == "" {
		feed.SoftDeletedAt = time.Time{}
		return feed
	}
	feed.SoftDeletedAt = carbon.Parse(softDeletedAt, carbon.UTC).StdTime()
	return feed
}

func (feed *feedImplementation) Status() string {
	return feed.StatusField
}
func (feed *feedImplementation) SetStatus(status string) FeedInterface {
	feed.StatusField = status
	return feed
}

func (feed *feedImplementation) UpdatedAt() string {
	if feed.UpdatedAtField.UpdatedAt.IsZero() {
		return ""
	}
	return carbon.CreateFromStdTime(feed.UpdatedAtField.UpdatedAt).ToDateTimeString()
}
func (feed *feedImplementation) UpdatedAtCarbon() *carbon.Carbon {
	return carbon.CreateFromStdTime(feed.UpdatedAtField.UpdatedAt)
}
func (feed *feedImplementation) SetUpdatedAt(updatedAt string) FeedInterface {
	if updatedAt == "" {
		return feed
	}
	feed.UpdatedAtField.UpdatedAt = carbon.Parse(updatedAt, carbon.UTC).StdTime()
	return feed
}

func (feed *feedImplementation) URL() string {
	return feed.URLField
}

func (feed *feedImplementation) SetURL(url string) FeedInterface {
	feed.URLField = url
	return feed
}

func (feed *feedImplementation) Data() map[string]string {
	data := map[string]string{}
	data[COLUMN_ID] = feed.ID()
	data[COLUMN_NAME] = feed.Name()
	data[COLUMN_DESCRIPTION] = feed.Description()
	data[COLUMN_URL] = feed.URL()
	data[COLUMN_STATUS] = feed.Status()
	data[COLUMN_FETCH_INTERVAL] = feed.FetchInterval()
	data[COLUMN_LAST_FETCHED_AT] = feed.LastFetchedAt()
	data[COLUMN_MEMO] = feed.Memo()
	data[COLUMN_CREATED_AT] = feed.CreatedAt()
	data[COLUMN_UPDATED_AT] = feed.UpdatedAt()
	data[COLUMN_SOFT_DELETED_AT] = feed.GetSoftDeletedAt()
	return data
}

func (feed *feedImplementation) MarkAsNotDirty(columns ...string) {
	feed.originalData = feed.Data()
}
