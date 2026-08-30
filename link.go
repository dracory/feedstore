package feedstore

import (
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
	URLField         string    `db:"url"`
	ViewsField       string    `db:"views"`
	VotesUpField     string    `db:"votes_up"`
	VotesDownField   string    `db:"votes_down"`
	ReportedAtField  time.Time `db:"reported_at"`
	ReportField      string    `db:"report"`
	CheckedAtField   time.Time `db:"checked_at"`
	TimeField        time.Time `db:"time"`
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

	CreatedAt() string
	CreatedAtCarbon() *carbon.Carbon
	SetCreatedAt(createdAt string) LinkInterface
	Description() string
	SetDescription(description string) LinkInterface
	FeedID() string
	SetFeedID(feedID string) LinkInterface
	ID() string
	SetID(id string) LinkInterface
	Status() string
	SetStatus(status string) LinkInterface
	Title() string
	SetTitle(title string) LinkInterface
	Time() string
	TimeCarbon() *carbon.Carbon
	SetTime(time time.Time) LinkInterface
	SetTimeString(time string) LinkInterface
	GetSoftDeletedAt() string
	GetSoftDeletedAtCarbon() *carbon.Carbon
	SetSoftDeletedAt(softDeletedAt string) LinkInterface
	UpdatedAt() string
	UpdatedAtCarbon() *carbon.Carbon
	SetUpdatedAt(updatedAt string) LinkInterface
	URL() string
	SetURL(url string) LinkInterface
	Views() string
	SetViews(views string) LinkInterface
	VotesUp() string
	SetVotesUp(votesUp string) LinkInterface
	VotesDown() string
	SetVotesDown(votesDown string) LinkInterface
	Report() string
	SetReport(report string) LinkInterface
	ReportedAt() string
	ReportedAtCarbon() *carbon.Carbon
	SetReportedAt(reportedAt time.Time) LinkInterface
	SetReportedAtString(reportedAt string) LinkInterface
	CheckedAt() string
	CheckedAtCarbon() *carbon.Carbon
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
	link.SetViews("0")
	link.SetVotesUp("0")
	link.SetVotesDown("0")
	link.SetReportedAt(time.Time{})
	link.SetReport("")
	link.SetCheckedAt(time.Time{})
	link.SetTime(time.Time{})
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
	if v, ok := data[COLUMN_TIME]; ok {
		link.SetTimeString(v)
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

func (link *linkImplementation) CheckedAt() string {
	if link.CheckedAtField.IsZero() {
		return ""
	}
	return carbon.CreateFromStdTime(link.CheckedAtField).ToDateTimeString()
}

func (link *linkImplementation) CheckedAtCarbon() *carbon.Carbon {
	return carbon.CreateFromStdTime(link.CheckedAtField)
}

func (link *linkImplementation) SetCheckedAt(timeChecked time.Time) LinkInterface {
	link.CheckedAtField = timeChecked
	return link
}

func (link *linkImplementation) SetCheckedAtString(timeChecked string) LinkInterface {
	if timeChecked == "" {
		link.CheckedAtField = time.Time{}
		return link
	}
	link.CheckedAtField = carbon.Parse(timeChecked, carbon.UTC).StdTime()
	return link
}

func (link *linkImplementation) CreatedAt() string {
	if link.CreatedAtField.CreatedAt.IsZero() {
		return ""
	}
	return carbon.CreateFromStdTime(link.CreatedAtField.CreatedAt).ToDateTimeString()
}

func (link *linkImplementation) CreatedAtCarbon() *carbon.Carbon {
	return carbon.CreateFromStdTime(link.CreatedAtField.CreatedAt)
}

func (link *linkImplementation) SetCreatedAt(createdAt string) LinkInterface {
	if createdAt == "" {
		return link
	}
	link.CreatedAtField.CreatedAt = carbon.Parse(createdAt, carbon.UTC).StdTime()
	return link
}

func (link *linkImplementation) Description() string {
	return link.DescriptionField
}

func (link *linkImplementation) SetDescription(description string) LinkInterface {
	link.DescriptionField = description
	return link
}

func (link *linkImplementation) FeedID() string {
	return link.FeedIDField
}

func (link *linkImplementation) SetFeedID(feedID string) LinkInterface {
	link.FeedIDField = feedID
	return link
}

func (link *linkImplementation) ID() string {
	return link.ShortID.ID
}

func (link *linkImplementation) SetID(id string) LinkInterface {
	link.ShortID.ID = id
	return link
}

func (link *linkImplementation) Status() string {
	return link.StatusField
}

func (link *linkImplementation) SetStatus(status string) LinkInterface {
	link.StatusField = status
	return link
}

func (link *linkImplementation) Title() string {
	return link.TitleField
}

func (link *linkImplementation) SetTitle(title string) LinkInterface {
	link.TitleField = title
	return link
}

func (link *linkImplementation) URL() string {
	return link.URLField
}

func (link *linkImplementation) SetURL(url string) LinkInterface {
	link.URLField = url
	return link
}

func (link *linkImplementation) VotesDown() string {
	return link.VotesDownField
}

func (link *linkImplementation) SetVotesDown(votesDown string) LinkInterface {
	link.VotesDownField = votesDown
	return link
}

func (link *linkImplementation) VotesUp() string {
	return link.VotesUpField
}

func (link *linkImplementation) SetVotesUp(votesUp string) LinkInterface {
	link.VotesUpField = votesUp
	return link
}

func (link *linkImplementation) Views() string {
	return link.ViewsField
}

func (link *linkImplementation) SetViews(views string) LinkInterface {
	link.ViewsField = views
	return link
}

func (link *linkImplementation) Report() string {
	return link.ReportField
}

func (link *linkImplementation) SetReport(report string) LinkInterface {
	link.ReportField = report
	return link
}

func (link *linkImplementation) ReportedAt() string {
	if link.ReportedAtField.IsZero() {
		return ""
	}
	return carbon.CreateFromStdTime(link.ReportedAtField).ToDateTimeString()
}

func (link *linkImplementation) ReportedAtCarbon() *carbon.Carbon {
	return carbon.CreateFromStdTime(link.ReportedAtField)
}

func (link *linkImplementation) SetReportedAt(reportedAt time.Time) LinkInterface {
	link.ReportedAtField = reportedAt
	return link
}

func (link *linkImplementation) SetReportedAtString(reportedAt string) LinkInterface {
	if reportedAt == "" {
		link.ReportedAtField = time.Time{}
		return link
	}
	link.ReportedAtField = carbon.Parse(reportedAt, carbon.UTC).StdTime()
	return link
}

func (link *linkImplementation) Time() string {
	if link.TimeField.IsZero() {
		return ""
	}
	return carbon.CreateFromStdTime(link.TimeField).ToDateTimeString()
}

func (link *linkImplementation) TimeCarbon() *carbon.Carbon {
	return carbon.CreateFromStdTime(link.TimeField)
}

func (link *linkImplementation) SetTime(time time.Time) LinkInterface {
	link.TimeField = time
	return link
}

func (link *linkImplementation) SetTimeString(timeStr string) LinkInterface {
	if timeStr == "" {
		link.TimeField = time.Time{}
		return link
	}
	link.TimeField = carbon.Parse(timeStr, carbon.UTC).StdTime()
	return link
}

func (link *linkImplementation) GetSoftDeletedAt() string {
	if link.SoftDeletedAt.IsZero() {
		return ""
	}
	return carbon.CreateFromStdTime(link.SoftDeletedAt).ToDateTimeString()
}
func (link *linkImplementation) GetSoftDeletedAtCarbon() *carbon.Carbon {
	return carbon.CreateFromStdTime(link.SoftDeletedAt)
}

func (link *linkImplementation) SetSoftDeletedAt(softDeletedAt string) LinkInterface {
	if softDeletedAt == "" {
		link.SoftDeletedAt = time.Time{}
		return link
	}
	link.SoftDeletedAt = carbon.Parse(softDeletedAt, carbon.UTC).StdTime()
	return link
}

func (link *linkImplementation) UpdatedAt() string {
	if link.UpdatedAtField.UpdatedAt.IsZero() {
		return ""
	}
	return carbon.CreateFromStdTime(link.UpdatedAtField.UpdatedAt).ToDateTimeString()
}

func (link *linkImplementation) UpdatedAtCarbon() *carbon.Carbon {
	return carbon.CreateFromStdTime(link.UpdatedAtField.UpdatedAt)
}

func (link *linkImplementation) SetUpdatedAt(updatedAt string) LinkInterface {
	if updatedAt == "" {
		return link
	}
	link.UpdatedAtField.UpdatedAt = carbon.Parse(updatedAt, carbon.UTC).StdTime()
	return link
}

func (link *linkImplementation) Data() map[string]string {
	data := map[string]string{}
	data[COLUMN_ID] = link.ID()
	data[COLUMN_FEED_ID] = link.FeedID()
	data[COLUMN_STATUS] = link.Status()
	data[COLUMN_TITLE] = link.Title()
	data[COLUMN_DESCRIPTION] = link.Description()
	data[COLUMN_URL] = link.URL()
	data[COLUMN_VIEWS] = link.Views()
	data[COLUMN_VOTES_UP] = link.VotesUp()
	data[COLUMN_VOTES_DOWN] = link.VotesDown()
	data[COLUMN_REPORTED_AT] = link.ReportedAt()
	data[COLUMN_REPORT] = link.Report()
	data[COLUMN_CHECKED_AT] = link.CheckedAt()
	data[COLUMN_TIME] = link.Time()
	data[COLUMN_CREATED_AT] = link.CreatedAt()
	data[COLUMN_UPDATED_AT] = link.UpdatedAt()
	data[COLUMN_SOFT_DELETED_AT] = link.GetSoftDeletedAt()
	return data
}

func (link *linkImplementation) MarkAsNotDirty(columns ...string) {
	link.originalData = link.Data()
}
