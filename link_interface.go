package feedstore

import "github.com/dromara/carbon/v2"

type LinkInterface interface {
	Data() map[string]string
	DataChanged() map[string]string
	MarkAsNotDirty()

	// Views() string
	// SetViews(views string) LinkInterface
	// VotesUp() string
	// SetVotesUp(votesUp string) LinkInterface
	// VotesDown() string
	// SetVotesDown(votesDown string) LinkInterface
	// Report() string
	// SetReport(report string) LinkInterface
	// ReportedAt() string
	// SetReportedAt(reportedAt string) LinkInterface
	// CheckedAt() string
	// SetCheckedAt(timeChecked string) LinkInterface

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
	SetTime(time string) LinkInterface
	SoftDeletedAt() string
	SoftDeletedAtCarbon() *carbon.Carbon
	SetSoftDeletedAt(softDeletedAt string) LinkInterface
	UpdatedAt() string
	UpdatedAtCarbon() *carbon.Carbon
	SetUpdatedAt(updatedAt string) LinkInterface
	URL() string
	SetURL(url string) LinkInterface
}
