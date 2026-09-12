package feedstore

import (
	"encoding/json"
	"time"

	"github.com/dracory/neat"
	"github.com/dracory/neat/database/orm"
	"github.com/dracory/neat/database/soft_delete"
	neatuid "github.com/dracory/neat/support/uid"
	"github.com/dromara/carbon/v2"
)

// ============================================================================
// == CLASS
// ============================================================================

type linkImplementation struct {
	orm.ShortID

	FeedIDField      string    `db:"feed_id"`
	StatusField      string    `db:"status"`
	TitleField       string    `db:"title"`
	DescriptionField string    `db:"description"`
	ContentField     string    `db:"content"`
	AuthorField      string    `db:"author"`
	PriorityField    string    `db:"priority"`
	URLField         string    `db:"url"`
	ViewsField       string    `db:"views"`
	VotesUpField     string    `db:"votes_up"`
	VotesDownField   string    `db:"votes_down"`
	ReportedAtField  time.Time `db:"reported_at"`
	ReportField      string    `db:"report"`
	CheckedAtField   time.Time `db:"checked_at"`
	PublishedAtField time.Time `db:"published_at"`
	DedupHashField   string    `db:"dedup_hash"`
	LanguageField    string    `db:"language"`
	MetasField       string    `db:"metas"`
	CreatedAtField   orm.CreatedAt
	UpdatedAtField   orm.UpdatedAt
	soft_delete.SoftDeletesMaxDate

	originalData map[string]string
}

// ============================================================================
// == INTERFACE
// ============================================================================

type LinkInterface interface {
	Data() map[string]string
	MarkAsNotDirty(...string)

	// Creation and update timestamps
	GetCreatedAt() string
	GetCreatedAtCarbon() *carbon.Carbon
	SetCreatedAt(createdAt string) LinkInterface
	GetDescription() string
	SetDescription(description string) LinkInterface
	GetContent() string
	SetContent(content string) LinkInterface
	GetAuthor() string
	SetAuthor(author string) LinkInterface
	GetPriority() bool
	SetPriority(priority bool) LinkInterface
	GetFeedID() string
	SetFeedID(feedID string) LinkInterface
	GetID() string
	SetID(id string) LinkInterface
	GetLanguage() string
	SetLanguage(language string) LinkInterface
	GetMeta(key string) (string, error)
	SetMeta(key string, value string) error
	DeleteMeta(key string) error
	GetMetas() (map[string]string, error)
	SetMetas(metas map[string]string) error
	GetStatus() string
	SetStatus(status string) LinkInterface
	GetTitle() string
	SetTitle(title string) LinkInterface
	GetPublishedAt() string
	GetPublishedAtCarbon() *carbon.Carbon
	SetPublishedAt(publishedAt time.Time) LinkInterface
	SetPublishedAtString(publishedAt string) LinkInterface
	GetDedupHash() string
	SetDedupHash(dedupHash string) LinkInterface
	GetSoftDeletedAt() string
	GetSoftDeletedAtCarbon() *carbon.Carbon
	SetSoftDeletedAt(softDeletedAt string) LinkInterface
	GetUpdatedAt() string
	GetUpdatedAtCarbon() *carbon.Carbon
	SetUpdatedAt(updatedAt string) LinkInterface
	GetURL() string
	SetURL(url string) LinkInterface
	GetViews() string
	SetViews(views string) LinkInterface
	GetVotesUp() string
	SetVotesUp(votesUp string) LinkInterface
	GetVotesDown() string
	SetVotesDown(votesDown string) LinkInterface
	GetReport() string
	SetReport(report string) LinkInterface
	GetReportedAt() string
	GetReportedAtCarbon() *carbon.Carbon
	SetReportedAt(reportedAt time.Time) LinkInterface
	SetReportedAtString(reportedAt string) LinkInterface
	GetCheckedAt() string
	GetCheckedAtCarbon() *carbon.Carbon
	SetCheckedAt(timeChecked time.Time) LinkInterface
	SetCheckedAtString(timeChecked string) LinkInterface
}

var _ LinkInterface = (*linkImplementation)(nil)

// ============================================================================
// == CONSTRUCTOR
// ============================================================================

func NewLink() *linkImplementation {
	link := &linkImplementation{}
	link.SetID(neatuid.GenerateShortID())
	link.SetDescription("")
	link.SetContent("")
	link.SetAuthor("")
	link.SetPriority(false)
	link.SetViews("0")
	link.SetVotesUp("0")
	link.SetVotesDown("0")
	link.SetReportedAtString(neat.NullDateTime)
	link.SetReport("")
	link.SetCheckedAtString(neat.NullDateTime)
	link.SetPublishedAtString(neat.NullDateTime)
	link.SetDedupHash("")
	link.SetLanguage("")
	_ = link.SetMetas(map[string]string{})
	link.SetCreatedAt(carbon.Now(carbon.UTC).ToDateTimeString())
	link.SetUpdatedAt(carbon.Now(carbon.UTC).ToDateTimeString())
	link.SetSoftDeletedAt(neat.MaxDateTime)
	link.MarkAsNotDirty()
	return link
}

func NewLinkFromExistingData(data map[string]string) *linkImplementation {
	link := &linkImplementation{}

	link.SetID(data[COLUMN_ID])
	link.SetFeedID(data[COLUMN_FEED_ID])
	link.SetStatus(data[COLUMN_STATUS])
	link.SetTitle(data[COLUMN_TITLE])
	link.SetDescription(data[COLUMN_DESCRIPTION])
	if v, ok := data[COLUMN_CONTENT]; ok {
		link.SetContent(v)
	}
	if v, ok := data[COLUMN_AUTHOR]; ok {
		link.SetAuthor(v)
	}
	if v, ok := data[COLUMN_PRIORITY]; ok {
		link.PriorityField = v
	}
	link.SetURL(data[COLUMN_URL])
	link.SetViews(data[COLUMN_VIEWS])
	link.SetVotesUp(data[COLUMN_VOTES_UP])
	link.SetVotesDown(data[COLUMN_VOTES_DOWN])
	if v, ok := data[COLUMN_REPORTED_AT]; ok {
		link.SetReportedAtString(v)
	}
	link.SetReport(data[COLUMN_REPORT])
	if v, ok := data[COLUMN_CHECKED_AT]; ok {
		link.SetCheckedAtString(v)
	}
	if v, ok := data[COLUMN_PUBLISHED_AT]; ok {
		link.SetPublishedAtString(v)
	}
	if v, ok := data[COLUMN_DEDUP_HASH]; ok {
		link.SetDedupHash(v)
	}
	if v, ok := data[COLUMN_LANGUAGE]; ok {
		link.SetLanguage(v)
	}
	if v, ok := data[COLUMN_METAS]; ok {
		link.MetasField = v
	}
	if v, ok := data[COLUMN_CREATED_AT]; ok {
		link.SetCreatedAt(v)
	}
	if v, ok := data[COLUMN_UPDATED_AT]; ok {
		link.SetUpdatedAt(v)
	}
	if v, ok := data[COLUMN_SOFT_DELETED_AT]; ok {
		link.SetSoftDeletedAt(v)
	}
	link.MarkAsNotDirty()

	return link
}

// == SETTERS AND GETTERS =====================================================

func (link *linkImplementation) GetCheckedAt() string {
	if link.CheckedAtField.IsZero() {
		return neat.NullDateTime
	}
	return carbon.CreateFromStdTime(link.CheckedAtField).ToDateTimeString()
}

func (link *linkImplementation) GetCheckedAtCarbon() *carbon.Carbon {
	return carbon.CreateFromStdTime(link.CheckedAtField)
}

func (link *linkImplementation) SetCheckedAt(timeChecked time.Time) LinkInterface {
	link.CheckedAtField = timeChecked
	return link
}

func (link *linkImplementation) SetCheckedAtString(timeChecked string) LinkInterface {
	if timeChecked == "" || timeChecked == neat.NullDateTime {
		link.CheckedAtField = time.Time{}
		return link
	}
	link.CheckedAtField = carbon.Parse(timeChecked, carbon.UTC).StdTime()
	return link
}

func (link *linkImplementation) GetCreatedAt() string {
	if link.CreatedAtField.CreatedAt.IsZero() {
		return neat.NullDateTime
	}
	return carbon.CreateFromStdTime(link.CreatedAtField.CreatedAt).ToDateTimeString()
}

func (link *linkImplementation) GetCreatedAtCarbon() *carbon.Carbon {
	return carbon.CreateFromStdTime(link.CreatedAtField.CreatedAt)
}

func (link *linkImplementation) SetCreatedAt(createdAt string) LinkInterface {
	if createdAt == "" || createdAt == neat.NullDateTime {
		link.CreatedAtField.CreatedAt = time.Time{}
		return link
	}
	link.CreatedAtField.CreatedAt = carbon.Parse(createdAt, carbon.UTC).StdTime()
	return link
}

func (link *linkImplementation) GetDescription() string {
	return link.DescriptionField
}

func (link *linkImplementation) SetDescription(description string) LinkInterface {
	link.DescriptionField = description
	return link
}

func (link *linkImplementation) GetContent() string {
	return link.ContentField
}

func (link *linkImplementation) SetContent(content string) LinkInterface {
	link.ContentField = content
	return link
}

func (link *linkImplementation) GetAuthor() string {
	return link.AuthorField
}

func (link *linkImplementation) SetAuthor(author string) LinkInterface {
	link.AuthorField = author
	return link
}

func (link *linkImplementation) GetPriority() bool {
	return link.PriorityField == "1" || link.PriorityField == "true"
}

func (link *linkImplementation) SetPriority(priority bool) LinkInterface {
	if priority {
		link.PriorityField = "1"
	} else {
		link.PriorityField = "0"
	}
	return link
}

func (link *linkImplementation) GetFeedID() string {
	return link.FeedIDField
}

func (link *linkImplementation) SetFeedID(feedID string) LinkInterface {
	link.FeedIDField = feedID
	return link
}

func (link *linkImplementation) GetID() string {
	return link.ShortID.ID
}

func (link *linkImplementation) SetID(id string) LinkInterface {
	link.ShortID.ID = id
	return link
}

func (link *linkImplementation) GetLanguage() string {
	return link.LanguageField
}

func (link *linkImplementation) SetLanguage(language string) LinkInterface {
	link.LanguageField = language
	return link
}

func (link *linkImplementation) GetMeta(key string) (string, error) {
	metas, err := link.GetMetas()
	if err != nil {
		return "", err
	}
	value, ok := metas[key]
	if !ok {
		return "", nil
	}
	return value, nil
}

func (link *linkImplementation) SetMeta(key string, value string) error {
	metas, err := link.GetMetas()
	if err != nil {
		return err
	}
	metas[key] = value
	return link.SetMetas(metas)
}

func (link *linkImplementation) DeleteMeta(key string) error {
	metas, err := link.GetMetas()
	if err != nil {
		return err
	}
	delete(metas, key)
	return link.SetMetas(metas)
}

func (link *linkImplementation) GetMetas() (map[string]string, error) {
	if link.MetasField == "" {
		return map[string]string{}, nil
	}
	var metas map[string]string
	err := json.Unmarshal([]byte(link.MetasField), &metas)
	if err != nil {
		return map[string]string{}, err
	}
	return metas, nil
}

func (link *linkImplementation) SetMetas(metas map[string]string) error {
	metasBytes, err := json.Marshal(metas)
	if err != nil {
		return err
	}
	link.MetasField = string(metasBytes)
	return nil
}

func (link *linkImplementation) GetStatus() string {
	return link.StatusField
}

func (link *linkImplementation) SetStatus(status string) LinkInterface {
	link.StatusField = status
	return link
}

func (link *linkImplementation) GetTitle() string {
	return link.TitleField
}

func (link *linkImplementation) SetTitle(title string) LinkInterface {
	link.TitleField = title
	return link
}

func (link *linkImplementation) GetURL() string {
	return link.URLField
}

func (link *linkImplementation) SetURL(url string) LinkInterface {
	link.URLField = url
	return link
}

func (link *linkImplementation) GetVotesDown() string {
	return link.VotesDownField
}

func (link *linkImplementation) SetVotesDown(votesDown string) LinkInterface {
	link.VotesDownField = votesDown
	return link
}

func (link *linkImplementation) GetVotesUp() string {
	return link.VotesUpField
}

func (link *linkImplementation) SetVotesUp(votesUp string) LinkInterface {
	link.VotesUpField = votesUp
	return link
}

func (link *linkImplementation) GetViews() string {
	return link.ViewsField
}

func (link *linkImplementation) SetViews(views string) LinkInterface {
	link.ViewsField = views
	return link
}

func (link *linkImplementation) GetReport() string {
	return link.ReportField
}

func (link *linkImplementation) SetReport(report string) LinkInterface {
	link.ReportField = report
	return link
}

func (link *linkImplementation) GetReportedAt() string {
	if link.ReportedAtField.IsZero() {
		return neat.NullDateTime
	}
	return carbon.CreateFromStdTime(link.ReportedAtField).ToDateTimeString()
}

func (link *linkImplementation) GetReportedAtCarbon() *carbon.Carbon {
	return carbon.CreateFromStdTime(link.ReportedAtField)
}

func (link *linkImplementation) SetReportedAt(reportedAt time.Time) LinkInterface {
	link.ReportedAtField = reportedAt
	return link
}

func (link *linkImplementation) SetReportedAtString(reportedAt string) LinkInterface {
	if reportedAt == "" || reportedAt == neat.NullDateTime {
		link.ReportedAtField = time.Time{}
		return link
	}
	link.ReportedAtField = carbon.Parse(reportedAt, carbon.UTC).StdTime()
	return link
}

func (link *linkImplementation) GetPublishedAt() string {
	if link.PublishedAtField.IsZero() {
		return neat.NullDateTime
	}
	return carbon.CreateFromStdTime(link.PublishedAtField).ToDateTimeString()
}

func (link *linkImplementation) GetPublishedAtCarbon() *carbon.Carbon {
	return carbon.CreateFromStdTime(link.PublishedAtField)
}

func (link *linkImplementation) SetPublishedAt(publishedAt time.Time) LinkInterface {
	link.PublishedAtField = publishedAt
	return link
}

func (link *linkImplementation) SetPublishedAtString(publishedAt string) LinkInterface {
	if publishedAt == "" || publishedAt == neat.NullDateTime {
		link.PublishedAtField = time.Time{}
		return link
	}
	link.PublishedAtField = carbon.Parse(publishedAt, carbon.UTC).StdTime()
	return link
}

func (link *linkImplementation) GetDedupHash() string {
	return link.DedupHashField
}

func (link *linkImplementation) SetDedupHash(dedupHash string) LinkInterface {
	link.DedupHashField = dedupHash
	return link
}

func (link *linkImplementation) GetSoftDeletedAt() string {
	if link.SoftDeletedAt.IsZero() {
		return neat.MaxDateTime
	}
	return carbon.CreateFromStdTime(link.SoftDeletedAt).ToDateTimeString()
}
func (link *linkImplementation) GetSoftDeletedAtCarbon() *carbon.Carbon {
	return carbon.CreateFromStdTime(link.SoftDeletedAt)
}

func (link *linkImplementation) SetSoftDeletedAt(softDeletedAt string) LinkInterface {
	if softDeletedAt == "" || softDeletedAt == neat.NullDateTime {
		link.SoftDeletedAt = time.Time{}
		return link
	}
	link.SoftDeletedAt = carbon.Parse(softDeletedAt, carbon.UTC).StdTime()
	return link
}

func (link *linkImplementation) GetUpdatedAt() string {
	if link.UpdatedAtField.UpdatedAt.IsZero() {
		return neat.NullDateTime
	}
	return carbon.CreateFromStdTime(link.UpdatedAtField.UpdatedAt).ToDateTimeString()
}

func (link *linkImplementation) GetUpdatedAtCarbon() *carbon.Carbon {
	return carbon.CreateFromStdTime(link.UpdatedAtField.UpdatedAt)
}

func (link *linkImplementation) SetUpdatedAt(updatedAt string) LinkInterface {
	if updatedAt == "" || updatedAt == neat.NullDateTime {
		link.UpdatedAtField.UpdatedAt = time.Time{}
		return link
	}
	link.UpdatedAtField.UpdatedAt = carbon.Parse(updatedAt, carbon.UTC).StdTime()
	return link
}

func (link *linkImplementation) Data() map[string]string {
	data := map[string]string{}
	data[COLUMN_ID] = link.GetID()
	data[COLUMN_FEED_ID] = link.GetFeedID()
	data[COLUMN_STATUS] = link.GetStatus()
	data[COLUMN_TITLE] = link.GetTitle()
	data[COLUMN_DESCRIPTION] = link.GetDescription()
	data[COLUMN_CONTENT] = link.GetContent()
	data[COLUMN_AUTHOR] = link.GetAuthor()
	data[COLUMN_PRIORITY] = link.PriorityField
	data[COLUMN_URL] = link.GetURL()
	data[COLUMN_VIEWS] = link.GetViews()
	data[COLUMN_VOTES_UP] = link.GetVotesUp()
	data[COLUMN_VOTES_DOWN] = link.GetVotesDown()
	data[COLUMN_REPORTED_AT] = link.GetReportedAt()
	data[COLUMN_REPORT] = link.GetReport()
	data[COLUMN_CHECKED_AT] = link.GetCheckedAt()
	data[COLUMN_PUBLISHED_AT] = link.GetPublishedAt()
	data[COLUMN_DEDUP_HASH] = link.GetDedupHash()
	data[COLUMN_LANGUAGE] = link.GetLanguage()
	data[COLUMN_METAS] = link.MetasField
	data[COLUMN_CREATED_AT] = link.GetCreatedAt()
	data[COLUMN_UPDATED_AT] = link.GetUpdatedAt()
	data[COLUMN_SOFT_DELETED_AT] = link.GetSoftDeletedAt()
	return data
}

func (link *linkImplementation) MarkAsNotDirty(columns ...string) {
	link.originalData = link.Data()
}
