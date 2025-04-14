package feedstore

import (
	"github.com/dromara/carbon/v2"
	"github.com/gouniverse/dataobject"
	"github.com/gouniverse/sb"
	"github.com/gouniverse/uid"
)

// ============================================================================
// == CLASS
// ============================================================================

type linkImplementation struct {
	dataobject.DataObject
}

// ============================================================================
// == INTERFACE
// ============================================================================

var _ LinkInterface = (*linkImplementation)(nil) // verify it extends the interface

// ============================================================================
// == CONSTRUCTOR
// ============================================================================

func NewLink() *linkImplementation {
	link := &linkImplementation{}
	link.SetID(uid.NanoUid())
	// link.SetStatus(LINK_STATUS_INACTIVE)
	// link.SetTitle("")
	link.SetDescription("")
	// link.SetURL("")
	// link.SetFeedID("") // required
	link.SetViews("0")
	link.SetVotesUp("0")
	link.SetVotesDown("0")
	link.SetReportedAt(sb.NULL_DATETIME)
	link.SetReport("")
	link.SetCheckedAt(sb.NULL_DATETIME)
	link.SetTime(sb.NULL_DATETIME)
	link.SetCreatedAt(carbon.Now(carbon.UTC).ToDateTimeString())
	link.SetUpdatedAt(carbon.Now(carbon.UTC).ToDateTimeString())
	link.SetDeletedAt(sb.NULL_DATETIME)

	return link
}

func NewLinkFromExistingData(data map[string]string) *linkImplementation {
	link := &linkImplementation{}

	for k, v := range data {
		link.Set(k, v)
	}

	link.MarkAsNotDirty()

	return link
}

// == SETTERS AND GETTERS =====================================================

func (link *linkImplementation) CreatedAt() string {
	return link.Get(COLUMN_CREATED_AT)
}

func (link *linkImplementation) CreatedAtCarbon() *carbon.Carbon {
	return carbon.Parse(link.CreatedAt())
}

func (link *linkImplementation) SetCreatedAt(createdAt string) LinkInterface {
	link.Set(COLUMN_CREATED_AT, createdAt)
	return link
}

func (link *linkImplementation) DeletedAt() string {
	return link.Get(COLUMN_DELETED_AT)
}

func (link *linkImplementation) DeletedAtCarbon() *carbon.Carbon {
	return carbon.Parse(link.DeletedAt())
}

func (link *linkImplementation) SetDeletedAt(deletedAt string) LinkInterface {
	link.Set(COLUMN_DELETED_AT, deletedAt)
	return link
}

func (link *linkImplementation) Description() string {
	return link.Get(COLUMN_DESCRIPTION)
}

func (link *linkImplementation) SetDescription(description string) LinkInterface {
	link.Set(COLUMN_DESCRIPTION, description)
	return link
}

func (link *linkImplementation) FeedID() string {
	return link.Get(COLUMN_FEED_ID)
}

func (link *linkImplementation) SetFeedID(feedID string) LinkInterface {
	link.Set(COLUMN_FEED_ID, feedID)
	return link
}

func (link *linkImplementation) ID() string {
	return link.Get(COLUMN_ID)
}

func (link *linkImplementation) SetID(id string) LinkInterface {
	link.Set(COLUMN_ID, id)
	return link
}

func (link *linkImplementation) Status() string {
	return link.Get(COLUMN_STATUS)
}

func (link *linkImplementation) SetStatus(status string) LinkInterface {
	link.Set(COLUMN_STATUS, status)
	return link
}

func (link *linkImplementation) Title() string {
	return link.Get(COLUMN_TITLE)
}

func (link *linkImplementation) SetTitle(title string) LinkInterface {
	link.Set(COLUMN_TITLE, title)
	return link
}

func (link *linkImplementation) URL() string {
	return link.Get(COLUMN_URL)
}

func (link *linkImplementation) SetURL(url string) LinkInterface {
	link.Set(COLUMN_URL, url)
	return link
}

func (link *linkImplementation) VotesDown() string {
	return link.Get(COLUMN_VOTES_DOWN)
}

func (link *linkImplementation) SetVotesDown(votesDown string) LinkInterface {
	link.Set(COLUMN_VOTES_DOWN, votesDown)
	return link
}

func (link *linkImplementation) VotesUp() string {
	return link.Get(COLUMN_VOTES_UP)
}

func (link *linkImplementation) SetVotesUp(votesUp string) LinkInterface {
	link.Set(COLUMN_VOTES_UP, votesUp)
	return link
}

func (link *linkImplementation) SetViews(views string) LinkInterface {
	link.Set(COLUMN_VIEWS, views)
	return link
}
func (link *linkImplementation) Report() string {
	return link.Get(COLUMN_REPORT)
}

func (link *linkImplementation) SetReport(report string) LinkInterface {
	link.Set(COLUMN_REPORT, report)
	return link
}

func (link *linkImplementation) ReportedAt() string {
	return link.Get(COLUMN_REPORTED_AT)
}

func (link *linkImplementation) ReportedAtCarbon() *carbon.Carbon {
	return carbon.Parse(link.ReportedAt())
}

func (link *linkImplementation) SetReportedAt(reportedAt string) LinkInterface {
	link.Set(COLUMN_REPORTED_AT, reportedAt)
	return link
}

func (link *linkImplementation) Time() string {
	return link.Get(COLUMN_TIME)
}

func (link *linkImplementation) TimeCarbon() *carbon.Carbon {
	return carbon.Parse(link.Time())
}

func (link *linkImplementation) SetTime(time string) LinkInterface {
	link.Set(COLUMN_TIME, time)
	return link
}

func (link *linkImplementation) CheckedAt() string {
	return link.Get(COLUMN_CHECKED_AT)
}

func (link *linkImplementation) SetCheckedAt(timeChecked string) LinkInterface {
	link.Set(COLUMN_CHECKED_AT, timeChecked)
	return link
}

func (link *linkImplementation) UpdatedAt() string {
	return link.Get(COLUMN_UPDATED_AT)
}

func (link *linkImplementation) UpdatedAtCarbon() *carbon.Carbon {
	return carbon.Parse(link.UpdatedAt())
}

func (link *linkImplementation) SetUpdatedAt(updatedAt string) LinkInterface {
	link.Set(COLUMN_UPDATED_AT, updatedAt)
	return link
}
