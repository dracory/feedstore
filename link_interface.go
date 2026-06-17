package feedstore

import (
	"time"

	"github.com/dromara/carbon/v2"
)

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
