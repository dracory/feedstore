package feedstore

import (
	"database/sql"
	"os"
	"testing"

	_ "modernc.org/sqlite"
)

func initDB(filepath string) *sql.DB {
	os.Remove(filepath) // remove database
	dsn := filepath + "?parseTime=true"
	db, err := sql.Open("sqlite", dsn)

	if err != nil {
		panic(err)
	}

	return db
}

func TestStoreFeedCreate(t *testing.T) {
	db := initDB(":memory:")

	store, err := NewStore(NewStoreOptions{
		DB:                 db,
		FeedTableName:      "feed_table_create",
		LinkTableName:      "link_table_create",
		AutomigrateEnabled: true,
	})

	if err != nil {
		t.Fatal("unexpected error:", err)
	}

	if store == nil {
		t.Fatal("unexpected nil store")
	}

	feed := NewFeed()

	feed.SetName("test feed")

	err = store.FeedCreate(feed)

	if err != nil {
		t.Fatal("unexpected error:", err)
	}
}

func TestStoreFeedFindByID(t *testing.T) {
	db := initDB(":memory:")

	store, err := NewStore(NewStoreOptions{
		DB:                 db,
		FeedTableName:      "feed_table_find_by_id",
		LinkTableName:      "link_table_find_by_id",
		AutomigrateEnabled: true,
	})

	if err != nil {
		t.Fatal("unexpected error:", err)
	}

	feed := NewFeed()
	feed.SetName("Feed 1")

	err = store.FeedCreate(feed)
	if err != nil {
		t.Error("unexpected error:", err)
	}

	feedFound, errFind := store.FeedFindByID(feed.ID())

	if errFind != nil {
		t.Fatal("unexpected error:", errFind)
	}

	if feedFound == nil {
		t.Fatal("Feed MUST NOT be nil")
	}

	if feedFound.ID() != feed.ID() {
		t.Fatal("IDs do not match")
	}

	if feedFound.Status() != feed.Status() {
		t.Fatal("Statuses do not match")
	}
}

// func TestStoreFeedSoftDelete(t *testing.T) {
// 	config.TestsConfigureAndInitialize()

// 	store, err := NewStore(NewStoreOptions{
// 		DB:                 config.Database.DB(),
// 		FeedTableName:      "exams_feed_find_by_id",
// 		AutomigrateEnabled: true,
// 	})

// 	if err != nil {
// 		t.Fatal("unexpected error:", err)
// 	}

// 	if store == nil {
// 		t.Fatal("unexpected nil store")
// 	}

// 	feed := NewFeed().
// 		SetStatus(CATEGORY_STATUS_ACTIVE).
// 		SetTitle("Feed 1")

// 	err = store.FeedCreate(feed)

// 	if err != nil {
// 		t.Fatal("unexpected error:", err)
// 	}

// 	err = store.FeedSoftDeleteByID(feed.ID())

// 	if err != nil {
// 		t.Fatal("unexpected error:", err)
// 	}

// 	if feed.DeletedAt() != sb.NULL_DATETIME {
// 		t.Fatal("Feed MUST NOT be soft deleted")
// 	}

// 	feedFound, errFind := store.FeedFindByID(feed.ID())

// 	if errFind != nil {
// 		t.Fatal("unexpected error:", errFind)
// 	}

// 	if feedFound != nil {
// 		t.Fatal("Feed MUST be nil")
// 	}

// 	feedFindWithDeleted, err := store.FeedList(FeedQueryOptions{
// 		ID:          feed.ID(),
// 		Limit:       1,
// 		WithDeleted: true,
// 	})

// 	if err != nil {
// 		t.Fatal("unexpected error:", err)
// 	}

// 	if len(feedFindWithDeleted) == 0 {
// 		t.Fatal("Exam MUST be soft deleted")
// 	}

// 	if strings.Contains(feedFindWithDeleted[0].DeletedAt(), sb.NULL_DATETIME) {
// 		t.Fatal("Exam MUST be soft deleted", feed.DeletedAt())
// 	}

// }

func TestStoreLinkCreate(t *testing.T) {
	db := initDB(":memory:")

	store, err := NewStore(NewStoreOptions{
		DB:                 db,
		FeedTableName:      "feed_table_create",
		LinkTableName:      "link_table_create",
		AutomigrateEnabled: true,
	})

	if err != nil {
		t.Fatal("unexpected error:", err)
	}

	if store == nil {
		t.Fatal("unexpected nil store")
	}

	link := NewLink()
	link.SetFeedID(`FeedID`)
	link.SetStatus(LINK_STATUS_ACTIVE)
	link.SetTitle(`Link 1`)
	link.SetURL(`https://example.com`)

	err = store.LinkCreate(link)

	if err != nil {
		t.Fatal("unexpected error:", err)
	}
}

func TestStoreLinkFindByID(t *testing.T) {
	db := initDB(":memory:")

	store, err := NewStore(NewStoreOptions{
		DB:                 db,
		FeedTableName:      "feed_table_find_by_id",
		LinkTableName:      "link_table_find_by_id",
		AutomigrateEnabled: true,
	})

	if err != nil {
		t.Fatal("unexpected error:", err)
	}

	link := NewLink()
	link.SetFeedID(`FeedID`)
	link.SetStatus(LINK_STATUS_ACTIVE)
	link.SetTitle(`Link 1`)
	link.SetURL(`https://example.com`)

	err = store.LinkCreate(link)
	if err != nil {
		t.Error("unexpected error:", err)
	}

	feedFound, errFind := store.LinkFindByID(link.ID())

	if errFind != nil {
		t.Fatal("unexpected error:", errFind)
	}

	if feedFound == nil {
		t.Fatal("Link MUST NOT be nil")
	}

	if feedFound.ID() != link.ID() {
		t.Fatal("IDs do not match")
	}

	if feedFound.Status() != link.Status() {
		t.Fatal("Statuses do not match")
	}
}
