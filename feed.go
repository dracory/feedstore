package feedstore

import (
	"encoding/json"
	"time"

	"github.com/dracory/neat"
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
	LanguageField      string    `db:"language"`
	MemoField          string    `db:"memo"`
	MetasField         string    `db:"metas"`
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

	// ============================================================================
	// == GETTERS AND SETTERS
	// ============================================================================

	// GetCreatedAt returns the created at timestamp as a string
	GetCreatedAt() string

	// GetCreatedAtCarbon returns the created at timestamp as a carbon instance
	GetCreatedAtCarbon() *carbon.Carbon

	// SetCreatedAt sets the created at timestamp
	SetCreatedAt(createdAt string) FeedInterface

	// GetDescription returns the description
	GetDescription() string

	// SetDescription sets the description
	SetDescription(description string) FeedInterface

	// GetFetchInterval returns the fetch interval
	GetFetchInterval() string

	// SetFetchInterval sets the fetch interval
	SetFetchInterval(fetchInterval string) FeedInterface

	// GetID returns the ID
	GetID() string

	// SetID sets the ID
	SetID(id string) FeedInterface

	// GetLanguage returns the language
	GetLanguage() string

	// SetLanguage sets the language
	SetLanguage(language string) FeedInterface

	// GetLastFetchedAt returns the last fetched at timestamp as a string
	GetLastFetchedAt() string

	// GetLastFetchedAtCarbon returns the last fetched at timestamp as a carbon instance
	GetLastFetchedAtCarbon() *carbon.Carbon

	// SetLastFetchedAt sets the last fetched at timestamp
	SetLastFetchedAt(lastFetchedAt time.Time) FeedInterface

	// SetLastFetchedAtString sets the last fetched at timestamp from a string
	SetLastFetchedAtString(lastFetchedAt string) FeedInterface

	// GetMemo returns the memo
	GetMemo() string

	// SetMemo sets the memo
	SetMemo(memo string) FeedInterface

	// GetMeta returns a specific meta value by key
	GetMeta(key string) (string, error)

	// SetMeta sets a specific meta value by key
	SetMeta(key string, value string) error

	// DeleteMeta removes a specific meta value by key
	DeleteMeta(key string) error

	// GetMetas returns all meta values as a map
	GetMetas() (map[string]string, error)

	// SetMetas sets all meta values from a map
	SetMetas(metas map[string]string) error

	// GetName returns the name
	GetName() string

	// SetName sets the name
	SetName(name string) FeedInterface

	// GetSoftDeletedAt returns the soft deleted at timestamp as a string
	GetSoftDeletedAt() string

	// GetSoftDeletedAtCarbon returns the soft deleted at timestamp as a carbon instance
	GetSoftDeletedAtCarbon() *carbon.Carbon

	// SetSoftDeletedAt sets the soft deleted at timestamp
	SetSoftDeletedAt(softDeletedAt string) FeedInterface

	// GetStatus returns the status
	GetStatus() string

	// SetStatus sets the status
	SetStatus(status string) FeedInterface

	// GetUpdatedAt returns the updated at timestamp as a string
	GetUpdatedAt() string

	// GetUpdatedAtCarbon returns the updated at timestamp as a carbon instance
	GetUpdatedAtCarbon() *carbon.Carbon

	// SetUpdatedAt sets the updated at timestamp
	SetUpdatedAt(updatedAt string) FeedInterface

	// GetURL returns the URL
	GetURL() string

	// SetURL sets the URL
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
	feed.SetLastFetchedAtString(neat.NullDateTime)
	feed.SetLanguage("")
	feed.SetMemo("")
	_ = feed.SetMetas(map[string]string{})
	feed.SetCreatedAt(carbon.Now(carbon.UTC).ToDateTimeString())
	feed.SetUpdatedAt(carbon.Now(carbon.UTC).ToDateTimeString())
	feed.SetSoftDeletedAt(neat.MaxDateTime)
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
	if v, ok := data[COLUMN_LANGUAGE]; ok {
		feed.SetLanguage(v)
	}
	feed.SetMemo(data[COLUMN_MEMO])
	if v, ok := data[COLUMN_METAS]; ok {
		feed.MetasField = v
	}
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

func (feed *feedImplementation) GetCreatedAt() string {
	if feed.CreatedAtField.CreatedAt.IsZero() {
		return neat.NullDateTime
	}
	return carbon.CreateFromStdTime(feed.CreatedAtField.CreatedAt).ToDateTimeString()
}
func (feed *feedImplementation) GetCreatedAtCarbon() *carbon.Carbon {
	return carbon.CreateFromStdTime(feed.CreatedAtField.CreatedAt)
}
func (feed *feedImplementation) SetCreatedAt(createdAt string) FeedInterface {
	if createdAt == "" || createdAt == neat.NullDateTime {
		feed.CreatedAtField.CreatedAt = time.Time{}
		return feed
	}
	feed.CreatedAtField.CreatedAt = carbon.Parse(createdAt, carbon.UTC).StdTime()
	return feed
}

func (feed *feedImplementation) GetDescription() string {
	return feed.DescriptionField
}

func (feed *feedImplementation) SetDescription(description string) FeedInterface {
	feed.DescriptionField = description
	return feed
}

func (feed *feedImplementation) GetFetchInterval() string {
	return feed.FetchIntervalField
}

func (feed *feedImplementation) GetFetchIntervalInt64() (int64, error) {
	return cast.ToInt64E(feed.GetFetchInterval())
}

func (feed *feedImplementation) SetFetchInterval(fetchInterval string) FeedInterface {
	feed.FetchIntervalField = fetchInterval
	return feed
}

func (feed *feedImplementation) GetID() string {
	return feed.ShortID.ID
}

func (feed *feedImplementation) SetID(id string) FeedInterface {
	feed.ShortID.ID = id
	return feed
}

func (feed *feedImplementation) GetLastFetchedAt() string {
	if feed.LastFetchedAtField.IsZero() {
		return neat.NullDateTime
	}
	return carbon.CreateFromStdTime(feed.LastFetchedAtField).ToDateTimeString()
}

func (feed *feedImplementation) GetLastFetchedAtCarbon() *carbon.Carbon {
	return carbon.CreateFromStdTime(feed.LastFetchedAtField)
}

func (feed *feedImplementation) SetLastFetchedAt(lastFetchedAt time.Time) FeedInterface {
	feed.LastFetchedAtField = lastFetchedAt
	return feed
}

func (feed *feedImplementation) SetLastFetchedAtString(lastFetchedAt string) FeedInterface {
	if lastFetchedAt == "" || lastFetchedAt == neat.NullDateTime {
		feed.LastFetchedAtField = time.Time{}
		return feed
	}
	feed.LastFetchedAtField = carbon.Parse(lastFetchedAt, carbon.UTC).StdTime()
	return feed
}

func (feed *feedImplementation) GetLanguage() string {
	return feed.LanguageField
}

func (feed *feedImplementation) SetLanguage(language string) FeedInterface {
	feed.LanguageField = language
	return feed
}

func (feed *feedImplementation) GetMemo() string {
	return feed.MemoField
}
func (feed *feedImplementation) SetMemo(memo string) FeedInterface {
	feed.MemoField = memo
	return feed
}

func (feed *feedImplementation) GetMeta(key string) (string, error) {
	metas, err := feed.GetMetas()
	if err != nil {
		return "", err
	}
	value, ok := metas[key]
	if !ok {
		return "", nil
	}
	return value, nil
}

func (feed *feedImplementation) SetMeta(key string, value string) error {
	metas, err := feed.GetMetas()
	if err != nil {
		return err
	}
	metas[key] = value
	return feed.SetMetas(metas)
}

func (feed *feedImplementation) DeleteMeta(key string) error {
	metas, err := feed.GetMetas()
	if err != nil {
		return err
	}
	delete(metas, key)
	return feed.SetMetas(metas)
}

func (feed *feedImplementation) GetMetas() (map[string]string, error) {
	if feed.MetasField == "" {
		return map[string]string{}, nil
	}
	var metas map[string]string
	err := json.Unmarshal([]byte(feed.MetasField), &metas)
	if err != nil {
		return map[string]string{}, err
	}
	return metas, nil
}

func (feed *feedImplementation) SetMetas(metas map[string]string) error {
	metasBytes, err := json.Marshal(metas)
	if err != nil {
		return err
	}
	feed.MetasField = string(metasBytes)
	return nil
}

func (feed *feedImplementation) GetName() string {
	return feed.NameField
}
func (feed *feedImplementation) SetName(name string) FeedInterface {
	feed.NameField = name
	return feed
}

func (feed *feedImplementation) GetSoftDeletedAt() string {
	if feed.SoftDeletedAt.IsZero() {
		return neat.MaxDateTime
	}
	return carbon.CreateFromStdTime(feed.SoftDeletedAt).ToDateTimeString()
}

func (feed *feedImplementation) GetSoftDeletedAtCarbon() *carbon.Carbon {
	return carbon.CreateFromStdTime(feed.SoftDeletedAt)
}

func (feed *feedImplementation) SetSoftDeletedAt(softDeletedAt string) FeedInterface {
	if softDeletedAt == "" || softDeletedAt == neat.NullDateTime {
		feed.SoftDeletedAt = time.Time{}
		return feed
	}
	feed.SoftDeletedAt = carbon.Parse(softDeletedAt, carbon.UTC).StdTime()
	return feed
}

func (feed *feedImplementation) GetStatus() string {
	return feed.StatusField
}
func (feed *feedImplementation) SetStatus(status string) FeedInterface {
	feed.StatusField = status
	return feed
}

func (feed *feedImplementation) GetUpdatedAt() string {
	if feed.UpdatedAtField.UpdatedAt.IsZero() {
		return neat.NullDateTime
	}
	return carbon.CreateFromStdTime(feed.UpdatedAtField.UpdatedAt).ToDateTimeString()
}
func (feed *feedImplementation) GetUpdatedAtCarbon() *carbon.Carbon {
	return carbon.CreateFromStdTime(feed.UpdatedAtField.UpdatedAt)
}
func (feed *feedImplementation) SetUpdatedAt(updatedAt string) FeedInterface {
	if updatedAt == "" || updatedAt == neat.NullDateTime {
		feed.UpdatedAtField.UpdatedAt = time.Time{}
		return feed
	}
	feed.UpdatedAtField.UpdatedAt = carbon.Parse(updatedAt, carbon.UTC).StdTime()
	return feed
}

func (feed *feedImplementation) GetURL() string {
	return feed.URLField
}

func (feed *feedImplementation) SetURL(url string) FeedInterface {
	feed.URLField = url
	return feed
}

func (feed *feedImplementation) Data() map[string]string {
	data := map[string]string{}
	data[COLUMN_ID] = feed.GetID()
	data[COLUMN_NAME] = feed.GetName()
	data[COLUMN_DESCRIPTION] = feed.GetDescription()
	data[COLUMN_URL] = feed.GetURL()
	data[COLUMN_STATUS] = feed.GetStatus()
	data[COLUMN_FETCH_INTERVAL] = feed.GetFetchInterval()
	data[COLUMN_LAST_FETCHED_AT] = feed.GetLastFetchedAt()
	data[COLUMN_LANGUAGE] = feed.GetLanguage()
	data[COLUMN_MEMO] = feed.GetMemo()
	data[COLUMN_METAS] = feed.MetasField
	data[COLUMN_CREATED_AT] = feed.GetCreatedAt()
	data[COLUMN_UPDATED_AT] = feed.GetUpdatedAt()
	data[COLUMN_SOFT_DELETED_AT] = feed.GetSoftDeletedAt()
	return data
}

func (feed *feedImplementation) MarkAsNotDirty(columns ...string) {
	feed.originalData = feed.Data()
}
