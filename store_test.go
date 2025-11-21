package feedstore

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/dracory/sb"
	"github.com/dromara/carbon/v2"

	_ "modernc.org/sqlite"
)

// Helper function to initialize an in-memory SQLite DB for testing
func initDB(filepath string) *sql.DB {
	os.Remove(filepath) // remove database if it exists
	dsn := filepath + "?_pragma=foreign_keys(1)&_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)&_pragma=synchronous(NORMAL)&parseTime=true"
	db, err := sql.Open("sqlite", dsn)

	if err != nil {
		panic(fmt.Sprintf("Failed to open database: %v", err))
	}

	// Basic check to ensure connection is alive
	err = db.Ping()
	if err != nil {
		panic(fmt.Sprintf("Failed to ping database: %v", err))
	}

	return db
}

// Helper function to create a store instance for testing
func createTestStore(t *testing.T, db *sql.DB, feedTable, linkTable string) StoreInterface {
	t.Helper() // Mark this as a test helper
	store, err := NewStore(NewStoreOptions{
		DB:                 db,
		FeedTableName:      feedTable,
		LinkTableName:      linkTable,
		AutomigrateEnabled: true, // Enable automigrate for tests
	})
	if err != nil {
		t.Fatalf("NewStore should not return an error, but got: %v", err)
	}
	if store == nil {
		t.Fatalf("Store should not be nil after NewStore")
	}
	return store
}

// Helper function to check if two slices contain the same elements, regardless of order.
func elementsMatch(t *testing.T, expected, actual []string) bool {
	t.Helper()
	if len(expected) != len(actual) {
		return false
	}
	expectedMap := make(map[string]int)
	actualMap := make(map[string]int)
	for _, item := range expected {
		expectedMap[item]++
	}
	for _, item := range actual {
		actualMap[item]++
	}
	return reflect.DeepEqual(expectedMap, actualMap)
}

func TestStoreAutoMigrate(t *testing.T) {
	db := initDB(":memory:")
	defer db.Close()

	feedTable := "feed_automigrate"
	linkTable := "link_automigrate"

	store, err := NewStore(NewStoreOptions{
		DB:                 db,
		FeedTableName:      feedTable,
		LinkTableName:      linkTable,
		AutomigrateEnabled: false, // Start with false
	})
	if err != nil {
		t.Fatalf("NewStore (with automigrate false) failed: %v", err)
	}
	if store == nil {
		t.Fatal("NewStore (with automigrate false) returned nil store")
	}

	// Run AutoMigrate
	err = store.AutoMigrate()
	if err != nil {
		t.Fatalf("AutoMigrate should not return an error, but got: %v", err)
	}

	// Verify tables exist
	var feedTableName string
	err = db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name=?;", feedTable).Scan(&feedTableName)
	if err != nil {
		t.Fatalf("Querying for feed table failed: %v", err)
	}
	if feedTableName != feedTable {
		t.Errorf("Expected feed table name '%s', but got '%s'", feedTable, feedTableName)
	}

	var linkTableName string
	err = db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name=?;", linkTable).Scan(&linkTableName)
	if err != nil {
		t.Fatalf("Querying for link table failed: %v", err)
	}
	if linkTableName != linkTable {
		t.Errorf("Expected link table name '%s', but got '%s'", linkTable, linkTableName)
	}
}

func TestStoreEnableDebug(t *testing.T) {
	db := initDB(":memory:")
	defer db.Close()
	store := createTestStore(t, db, "feed_debug", "link_debug")

	// Cast to implementation to check internal state (use cautiously)
	storeImpl, ok := store.(*storeImplementation)
	if !ok {
		t.Fatalf("Store should be of type *storeImplementation")
	}

	if storeImpl.debugEnabled {
		t.Error("Debug should be false initially")
	}
	store.EnableDebug(true)
	if !storeImpl.debugEnabled {
		t.Error("Debug should be true after enabling")
	}
	store.EnableDebug(false)
	if storeImpl.debugEnabled {
		t.Error("Debug should be false after disabling")
	}
}

func TestStoreGetters(t *testing.T) {
	db := initDB(":memory:")
	defer db.Close()
	feedTable := "feed_getters"
	linkTable := "link_getters"
	store := createTestStore(t, db, feedTable, linkTable)

	if store.GetDriverName() != "sqlite" {
		t.Errorf("GetDriverName: expected 'sqlite', got '%s'", store.GetDriverName())
	}
	if store.GetFeedTableName() != feedTable {
		t.Errorf("GetFeedTableName: expected '%s', got '%s'", feedTable, store.GetFeedTableName())
	}
	if store.GetLinkTableName() != linkTable {
		t.Errorf("GetLinkTableName: expected '%s', got '%s'", linkTable, store.GetLinkTableName())
	}
}

func TestStoreFeedDelete(t *testing.T) {
	db := initDB(":memory:")
	defer db.Close()
	store := createTestStore(t, db, "feed_delete", "link_delete")
	ctx := context.Background()

	// 1. Create a feed
	feed := NewFeed()
	feed.SetName("FeedToDelete")
	err := store.FeedCreate(ctx, feed)
	if err != nil {
		t.Fatalf("FeedCreate should succeed, but got error: %v", err)
	}

	// 2. Verify it exists
	foundFeed, err := store.FeedFindByID(ctx, feed.ID())
	if err != nil {
		t.Fatalf("FeedFindByID should succeed before delete, but got error: %v", err)
	}
	if foundFeed == nil {
		t.Fatal("Feed should be found before delete, but was nil")
	}

	// 3. Delete using FeedDelete
	err = store.FeedDelete(ctx, feed)
	if err != nil {
		t.Fatalf("FeedDelete should succeed, but got error: %v", err)
	}

	// 4. Verify it's gone
	foundFeed, err = store.FeedFindByID(ctx, feed.ID())
	if err != nil {
		t.Fatalf("FeedFindByID should succeed after delete, but got error: %v", err)
	}
	if foundFeed != nil {
		t.Error("Feed should not be found after delete, but was found")
	}

	// 5. Test deleting nil feed
	err = store.FeedDelete(ctx, nil)
	if err == nil {
		t.Error("FeedDelete should return error for nil feed, but got nil")
	}

	// 6. Test deleting non-existent feed (by ID)
	err = store.FeedDeleteByID(ctx, "non-existent-id")
	if err != nil {
		t.Errorf("FeedDeleteByID for non-existent ID should not error (idempotent), but got: %v", err)
	}
}

func TestStoreFeedDeleteByID(t *testing.T) {
	db := initDB(":memory:")
	defer db.Close()
	store := createTestStore(t, db, "feed_delete_id", "link_delete_id")
	ctx := context.Background()

	// 1. Create a feed
	feed := NewFeed()
	feed.SetName("FeedToDeleteByID")
	err := store.FeedCreate(ctx, feed)
	if err != nil {
		t.Fatalf("FeedCreate should succeed, but got error: %v", err)
	}
	feedID := feed.ID()

	// 2. Verify it exists
	foundFeed, err := store.FeedFindByID(ctx, feedID)
	if err != nil {
		t.Fatalf("FeedFindByID should succeed before delete, but got error: %v", err)
	}
	if foundFeed == nil {
		t.Fatal("Feed should be found before delete, but was nil")
	}

	// 3. Delete using FeedDeleteByID
	err = store.FeedDeleteByID(ctx, feedID)
	if err != nil {
		t.Fatalf("FeedDeleteByID should succeed, but got error: %v", err)
	}

	// 4. Verify it's gone
	foundFeed, err = store.FeedFindByID(ctx, feedID)
	if err != nil {
		t.Fatalf("FeedFindByID should succeed after delete, but got error: %v", err)
	}
	if foundFeed != nil {
		t.Error("Feed should not be found after delete, but was found")
	}

	// 5. Test deleting with empty ID
	err = store.FeedDeleteByID(ctx, "")
	if err == nil {
		t.Error("FeedDeleteByID should return error for empty ID, but got nil")
	}
}

func TestStoreFeedList(t *testing.T) {
	db := initDB(":memory:")
	defer db.Close()
	store := createTestStore(t, db, "feed_list", "link_list")
	ctx := context.Background()

	// Create test feeds
	feed1 := NewFeed().SetName("Feed 1").SetStatus(FEED_STATUS_ACTIVE)
	feed2 := NewFeed().SetName("Feed 2").SetStatus(FEED_STATUS_INACTIVE)
	feed3 := NewFeed().SetName("Feed 3").SetStatus(FEED_STATUS_ACTIVE)
	feed4 := NewFeed().SetName("Feed 4").SetStatus(FEED_STATUS_ACTIVE) // To be soft deleted

	if err := store.FeedCreate(ctx, feed1); err != nil {
		t.Fatalf("Failed to create feed1: %v", err)
	}
	time.Sleep(1 * time.Second)
	if err := store.FeedCreate(ctx, feed2); err != nil {
		t.Fatalf("Failed to create feed2: %v", err)
	}
	time.Sleep(1 * time.Second)
	if err := store.FeedCreate(ctx, feed3); err != nil {
		t.Fatalf("Failed to create feed3: %v", err)
	}
	time.Sleep(1 * time.Second)
	if err := store.FeedCreate(ctx, feed4); err != nil {
		t.Fatalf("Failed to create feed4: %v", err)
	}
	if err := store.FeedSoftDelete(ctx, feed4); err != nil {
		t.Fatalf("Failed to soft delete feed4: %v", err)
	} // Soft delete feed4

	// Test cases
	testCases := []struct {
		name          string
		query         FeedQueryInterface
		expectedCount int
		expectedIDs   []string // Order might matter depending on query
		expectError   bool
	}{
		{
			name:          "List all (excluding soft deleted)",
			query:         FeedQuery().SetLimit(10),
			expectedCount: 3,
			expectedIDs:   []string{feed1.ID(), feed2.ID(), feed3.ID()},
		},
		{
			name:          "List with specific ID",
			query:         FeedQuery().SetID(feed2.ID()),
			expectedCount: 1,
			expectedIDs:   []string{feed2.ID()},
		},
		{
			name:          "List with specific Status",
			query:         FeedQuery().SetStatus(FEED_STATUS_ACTIVE).SetLimit(10),
			expectedCount: 2,
			expectedIDs:   []string{feed1.ID(), feed3.ID()},
		},
		{
			name:          "List with Status IN",
			query:         FeedQuery().SetStatusIn([]string{FEED_STATUS_INACTIVE}).SetLimit(10),
			expectedCount: 1,
			expectedIDs:   []string{feed2.ID()},
		},
		{
			name:          "List with Limit",
			query:         FeedQuery().SetLimit(2),
			expectedCount: 2,
		},
		{
			name:          "List with Offset",
			query:         FeedQuery().SetLimit(2).SetOffset(1),
			expectedCount: 2, // Should get feed2 and feed3 if ordered by creation
		},
		{
			name:          "List with OrderBy CreatedAt ASC",
			query:         FeedQuery().SetLimit(10).SetOrderBy(COLUMN_CREATED_AT).SetOrderDirection(sb.ASC),
			expectedCount: 3,
			expectedIDs:   []string{feed1.ID(), feed2.ID(), feed3.ID()},
		},
		{
			name:          "List with OrderBy CreatedAt DESC",
			query:         FeedQuery().SetLimit(10).SetOrderBy(COLUMN_CREATED_AT).SetOrderDirection(sb.DESC),
			expectedCount: 3,
			expectedIDs:   []string{feed3.ID(), feed2.ID(), feed1.ID()},
		},
		{
			name:          "List including soft deleted",
			query:         FeedQuery().SetLimit(10).SetWithSoftDeleted(true),
			expectedCount: 4,
			expectedIDs:   []string{feed1.ID(), feed2.ID(), feed3.ID(), feed4.ID()},
		},
		{
			name:          "List only soft deleted",
			query:         FeedQuery().SetLimit(10).SetOnlySoftDeleted(true),
			expectedCount: 1,
			expectedIDs:   []string{feed4.ID()},
		},
		{
			name:          "List non-existent ID",
			query:         FeedQuery().SetID("non-existent"),
			expectedCount: 0,
		},
		{
			name:        "List with invalid query (negative limit)",
			query:       FeedQuery().SetLimit(-1),
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			feeds, err := store.FeedList(ctx, tc.query)

			if tc.expectError {
				if err == nil {
					t.Errorf("Expected an error for invalid query '%s', but got nil", tc.name)
				}
				if len(feeds) != 0 {
					t.Errorf("Expected empty feed list on error for '%s', but got %d feeds", tc.name, len(feeds))
				}
				return // Skip further checks for error cases
			}

			if err != nil {
				t.Fatalf("FeedList for '%s' should not return an error, but got: %v", tc.name, err)
			}
			if len(feeds) != tc.expectedCount {
				t.Errorf("FeedList for '%s' returned wrong number of feeds: expected %d, got %d", tc.name, tc.expectedCount, len(feeds))
			}

			if len(tc.expectedIDs) > 0 {
				returnedIDs := make([]string, len(feeds))
				for i, f := range feeds {
					returnedIDs[i] = f.ID()
				}
				// Use reflect.DeepEqual for ordered comparison, elementsMatch helper for unordered
				if strings.Contains(tc.name, "OrderBy") {
					if !reflect.DeepEqual(tc.expectedIDs, returnedIDs) {
						t.Errorf("Returned feed IDs for '%s' do not match expected order. Expected %v, got %v", tc.name, tc.expectedIDs, returnedIDs)
					}
				} else {
					if !elementsMatch(t, tc.expectedIDs, returnedIDs) {
						t.Errorf("Returned feed IDs for '%s' do not match expected set. Expected %v (any order), got %v", tc.name, tc.expectedIDs, returnedIDs)
					}
				}
			}
		})
	}
}

func TestStoreFeedSoftDelete(t *testing.T) {
	db := initDB(":memory:")
	defer db.Close()
	store := createTestStore(t, db, "feed_soft_delete", "link_soft_delete")
	ctx := context.Background()

	// 1. Create a feed
	feed := NewFeed().SetName("FeedToSoftDelete")
	err := store.FeedCreate(ctx, feed)
	if err != nil {
		t.Fatalf("FeedCreate failed: %v", err)
	}
	initialUpdatedAt := feed.UpdatedAt()

	// 2. Verify it exists and is not soft deleted
	foundFeed, err := store.FeedFindByID(ctx, feed.ID())
	if err != nil {
		t.Fatalf("FeedFindByID before soft delete failed: %v", err)
	}
	if foundFeed == nil {
		t.Fatal("Feed not found before soft delete")
	}
	// Check if SoftDeletedAt is in the future (or MAX_DATETIME)
	if !foundFeed.SoftDeletedAtCarbon().Gt(carbon.Now()) {
		t.Errorf("SoftDeletedAt should be in the future initially, but was %s", foundFeed.SoftDeletedAt())
	}

	// 3. Soft delete using FeedSoftDelete
	time.Sleep(1 * time.Second) // Ensure UpdatedAt changes
	err = store.FeedSoftDelete(ctx, feed)
	if err != nil {
		t.Fatalf("FeedSoftDelete failed: %v", err)
	}

	// 4. Verify it's marked as deleted in the object (check internal state)
	if feed.SoftDeletedAtCarbon().Gt(carbon.Now()) {
		t.Errorf("SoftDeletedAt should be in the past after soft delete in object, but was %s", feed.SoftDeletedAt())
	}
	if initialUpdatedAt == feed.UpdatedAt() {
		t.Error("UpdatedAt should have changed after soft delete")
	}

	// 5. Verify it's not found by default FindByID
	foundFeed, err = store.FeedFindByID(ctx, feed.ID())
	if err != nil {
		t.Fatalf("FeedFindByID after soft delete failed: %v", err)
	}
	if foundFeed != nil {
		t.Error("Feed should not be found by default FindByID after soft delete")
	}

	// 6. Verify it IS found when including soft deleted
	list, err := store.FeedList(ctx, FeedQuery().SetID(feed.ID()).SetWithSoftDeleted(true))
	if err != nil {
		t.Fatalf("FeedList with soft deleted failed: %v", err)
	}
	if len(list) != 1 {
		t.Fatalf("Feed should be found when including soft deleted, expected 1 got %d", len(list))
	}
	if list[0].SoftDeletedAtCarbon().Gt(carbon.Now()) {
		t.Errorf("Found feed's SoftDeletedAt should be in the past, but was %s", list[0].SoftDeletedAt())
	}

	// 7. Test soft deleting nil feed
	err = store.FeedSoftDelete(ctx, nil)
	if err == nil {
		t.Error("FeedSoftDelete should return error for nil feed")
	}
}

func TestStoreFeedSoftDeleteByID(t *testing.T) {
	db := initDB(":memory:")
	defer db.Close()
	store := createTestStore(t, db, "feed_soft_delete_id", "link_soft_delete_id")
	ctx := context.Background()

	// 1. Create a feed
	feed := NewFeed().SetName("FeedToSoftDeleteByID")
	err := store.FeedCreate(ctx, feed)
	if err != nil {
		t.Fatalf("FeedCreate failed: %v", err)
	}
	feedID := feed.ID()

	// 2. Verify it exists and is not soft deleted
	foundFeed, err := store.FeedFindByID(ctx, feedID)
	if err != nil {
		t.Fatalf("FeedFindByID before soft delete failed: %v", err)
	}
	if foundFeed == nil {
		t.Fatal("Feed not found before soft delete")
	}

	// 3. Soft delete using FeedSoftDeleteByID
	err = store.FeedSoftDeleteByID(ctx, feedID)
	if err != nil {
		t.Fatalf("FeedSoftDeleteByID failed: %v", err)
	}

	// 4. Verify it's not found by default FindByID
	foundFeed, err = store.FeedFindByID(ctx, feedID)
	if err != nil {
		t.Fatalf("FeedFindByID after soft delete failed: %v", err)
	}
	if foundFeed != nil {
		t.Error("Feed should not be found by default FindByID after soft delete")
	}

	// 5. Verify it IS found when including soft deleted
	list, err := store.FeedList(ctx, FeedQuery().SetID(feedID).SetWithSoftDeleted(true))
	if err != nil {
		t.Fatalf("FeedList with soft deleted failed: %v", err)
	}
	if len(list) != 1 {
		t.Fatalf("Feed should be found when including soft deleted, expected 1 got %d", len(list))
	}
	if list[0].SoftDeletedAtCarbon().Gt(carbon.Now()) {
		t.Errorf("Found feed's SoftDeletedAt should be in the past, but was %s", list[0].SoftDeletedAt())
	}

	// 6. Test soft deleting non-existent ID
	err = store.FeedSoftDeleteByID(ctx, "non-existent-id")
	// This case is tricky. FeedSoftDeleteByID calls FindByID first.
	// If FindByID returns (nil, nil) for non-existent, then FeedSoftDelete(nil) is called, which errors.
	// If FindByID returns (nil, error), then that error is returned.
	// Let's assume FindByID returns (nil, nil) for not found.
	// The current implementation will then call FeedSoftDelete(nil) which returns an error.
	// So, we expect an error here based on the current code structure.
	if err == nil {
		t.Error("FeedSoftDeleteByID for non-existent ID should error (due to FeedSoftDelete(nil)), but got nil")
	}

	// 7. Test soft deleting with empty ID
	err = store.FeedSoftDeleteByID(ctx, "")
	if err == nil {
		t.Error("FeedSoftDeleteByID should return error for empty ID")
	}
}

func TestStoreFeedUpdate(t *testing.T) {
	db := initDB(":memory:")
	defer db.Close()
	store := createTestStore(t, db, "feed_update", "link_update")
	ctx := context.Background()

	// 1. Create a feed
	feed := NewFeed().SetName("Original Name").SetStatus(FEED_STATUS_INACTIVE)
	err := store.FeedCreate(ctx, feed)
	if err != nil {
		t.Fatalf("FeedCreate failed: %v", err)
	}
	feedID := feed.ID()
	initialUpdatedAt := feed.UpdatedAt()

	// 2. Modify the feed object
	newName := "Updated Name"
	newStatus := FEED_STATUS_ACTIVE
	newMemo := "Test Memo"
	feed.SetName(newName)
	feed.SetStatus(newStatus)
	feed.SetMemo(newMemo)

	// 3. Update the feed in the store
	time.Sleep(1 * time.Second) // Ensure UpdatedAt changes
	err = store.FeedUpdate(ctx, feed)
	if err != nil {
		t.Fatalf("FeedUpdate failed: %v", err)
	}

	// 4. Verify the object is marked as not dirty
	if len(feed.DataChanged()) != 0 {
		t.Errorf("DataChanged should be empty after successful update, but got %v", feed.DataChanged())
	}
	if initialUpdatedAt == feed.UpdatedAt() {
		t.Error("UpdatedAt should have changed after update")
	}
	updatedAtAfterUpdate := feed.UpdatedAt() // Store for next check

	// 5. Retrieve the feed and verify changes
	updatedFeed, err := store.FeedFindByID(ctx, feedID)
	if err != nil {
		t.Fatalf("FeedFindByID after update failed: %v", err)
	}
	if updatedFeed == nil {
		t.Fatal("Updated feed not found")
	}
	if updatedFeed.Name() != newName {
		t.Errorf("Name update failed: expected '%s', got '%s'", newName, updatedFeed.Name())
	}
	if updatedFeed.Status() != newStatus {
		t.Errorf("Status update failed: expected '%s', got '%s'", newStatus, updatedFeed.Status())
	}
	if updatedFeed.Memo() != newMemo {
		t.Errorf("Memo update failed: expected '%s', got '%s'", newMemo, updatedFeed.Memo())
	}
	if strings.ReplaceAll(updatedFeed.UpdatedAt(), " +0000 UTC", "") != updatedAtAfterUpdate {
		t.Errorf("UpdatedAt mismatch: expected '%s', got '%s'", updatedAtAfterUpdate, strings.ReplaceAll(updatedFeed.UpdatedAt(), " +0000 UTC", ""))
	}

	// 6. Test updating with no changes
	// updatedFeed was retrieved, MarkAsNotDirty was called implicitly by NewFeedFromExistingData
	time.Sleep(1 * time.Second)
	err = store.FeedUpdate(ctx, updatedFeed) // No fields changed since last MarkAsNotDirty
	if err != nil {
		t.Fatalf("Update with no changes should not error, but got: %v", err)
	}
	// Retrieve again to check if UpdatedAt changed in DB (it shouldn't if no query ran)
	finalFeed, err := store.FeedFindByID(ctx, feedID)
	if err != nil {
		t.Fatalf("FeedFindByID after no-change update failed: %v", err)
	}
	if finalFeed == nil {
		t.Fatal("Feed not found after no-change update")
	}
	if strings.ReplaceAll(finalFeed.UpdatedAt(), " +0000 UTC", "") != updatedAtAfterUpdate {
		t.Errorf("UpdatedAt should not change if no fields were modified, expected '%s', got '%s'", updatedAtAfterUpdate, strings.ReplaceAll(finalFeed.UpdatedAt(), " +0000 UTC", ""))
	}

	// 7. Test updating nil feed
	err = store.FeedUpdate(ctx, nil)
	if err == nil {
		t.Error("FeedUpdate should return error for nil feed")
	}
}

// --- Link Tests ---

func TestStoreLinkCount(t *testing.T) {
	db := initDB(":memory:")
	defer db.Close()
	store := createTestStore(t, db, "feed_link_count", "link_link_count")
	ctx := context.Background()

	// 0. Empty store should return 0
	total, err := store.LinkCount(ctx, LinkQuery().SetStatus(LINK_STATUS_ACTIVE))
	if err != nil {
		t.Fatalf("LinkCount on empty store failed: %v", err)
	}
	if total != 0 {
		t.Errorf("Expected 0, got %d", total)
	}

	// 1. Create test data: 5 active, 2 inactive across two feeds
	feeds := []string{"feedA", "feedB"}
	mk := func(title, status, feedID, url string) string {
		l := NewLink().SetTitle(title).SetStatus(status).SetFeedID(feedID).SetURL(url)
		if err := store.LinkCreate(ctx, l); err != nil {
			t.Fatalf("LinkCreate failed: %v", err)
		}
		return l.ID()
	}

	ids := []string{}
	ids = append(ids, mk("A1", LINK_STATUS_ACTIVE, feeds[0], "http://a1"))
	ids = append(ids, mk("A2", LINK_STATUS_ACTIVE, feeds[0], "http://a2"))
	ids = append(ids, mk("A3", LINK_STATUS_ACTIVE, feeds[1], "http://a3"))
	ids = append(ids, mk("A4", LINK_STATUS_ACTIVE, feeds[1], "http://a4"))
	ids = append(ids, mk("A5", LINK_STATUS_ACTIVE, feeds[1], "http://a5"))
	ids = append(ids, mk("I1", LINK_STATUS_INACTIVE, feeds[0], "http://i1"))
	ids = append(ids, mk("I2", LINK_STATUS_INACTIVE, feeds[1], "http://i2"))

	// 2. Count by status
	totalActive, err := store.LinkCount(ctx, LinkQuery().SetStatus(LINK_STATUS_ACTIVE))
	if err != nil {
		t.Fatalf("LinkCount active failed: %v", err)
	}
	if totalActive != 5 {
		t.Errorf("Expected 5 active, got %d", totalActive)
	}

	totalInactive, err := store.LinkCount(ctx, LinkQuery().SetStatus(LINK_STATUS_INACTIVE))
	if err != nil {
		t.Fatalf("LinkCount inactive failed: %v", err)
	}
	if totalInactive != 2 {
		t.Errorf("Expected 2 inactive, got %d", totalInactive)
	}

	// 3. Soft delete one active and verify default excludes it
	if err := store.LinkSoftDeleteByID(ctx, ids[0]); err != nil {
		t.Fatalf("LinkSoftDeleteByID failed: %v", err)
	}
	totalActiveAfterSD, err := store.LinkCount(ctx, LinkQuery().SetStatus(LINK_STATUS_ACTIVE))
	if err != nil {
		t.Fatalf("LinkCount after soft delete failed: %v", err)
	}
	if totalActiveAfterSD != 4 {
		t.Errorf("Expected 4 active after soft-delete, got %d", totalActiveAfterSD)
	}

	// 4. Including soft-deleted should bring it back
	totalActiveWithSD, err := store.LinkCount(ctx, LinkQuery().SetStatus(LINK_STATUS_ACTIVE).SetWithSoftDeleted(true))
	if err != nil {
		t.Fatalf("LinkCount with soft deleted failed: %v", err)
	}
	if totalActiveWithSD != 5 {
		t.Errorf("Expected 5 active including soft-deleted, got %d", totalActiveWithSD)
	}

	// 5. Filter by FeedID
	totalFeedA, err := store.LinkCount(ctx, LinkQuery().SetStatus(LINK_STATUS_ACTIVE).SetFeedID(feeds[0]))
	if err != nil {
		t.Fatalf("LinkCount by feed failed: %v", err)
	}
	if totalFeedA != 1 { // one remaining active in feedA after soft delete
		t.Errorf("Expected 1 active for feedA, got %d", totalFeedA)
	}
}
func TestStoreLinkDelete(t *testing.T) {
	db := initDB(":memory:")
	defer db.Close()
	store := createTestStore(t, db, "feed_link_delete", "link_link_delete")
	ctx := context.Background()

	// 1. Create a link
	link := NewLink().SetTitle("LinkToDelete").SetFeedID("feed1").SetURL("http://delete.me").SetStatus(LINK_STATUS_ACTIVE)
	err := store.LinkCreate(ctx, link)
	if err != nil {
		t.Fatalf("LinkCreate failed: %v", err)
	}

	// 2. Verify it exists
	foundLink, err := store.LinkFindByID(ctx, link.ID())
	if err != nil {
		t.Fatalf("LinkFindByID before delete failed: %v", err)
	}
	if foundLink == nil {
		t.Fatal("Link not found before delete")
	}

	// 3. Delete using LinkDelete
	err = store.LinkDelete(ctx, link)
	if err != nil {
		t.Fatalf("LinkDelete failed: %v", err)
	}

	// 4. Verify it's gone
	foundLink, err = store.LinkFindByID(ctx, link.ID())
	if err != nil {
		t.Fatalf("LinkFindByID after delete failed: %v", err)
	}
	if foundLink != nil {
		t.Error("Link should not be found after delete")
	}

	// 5. Test deleting nil link
	err = store.LinkDelete(ctx, nil)
	if err == nil {
		t.Error("LinkDelete should return error for nil link")
	}

	// 6. Test deleting non-existent link (by ID)
	err = store.LinkDeleteByID(ctx, "non-existent-id")
	if err != nil {
		t.Errorf("LinkDeleteByID for non-existent ID should not error (idempotent), but got: %v", err)
	}
}

func TestStoreLinkDeleteByID(t *testing.T) {
	db := initDB(":memory:")
	defer db.Close()
	store := createTestStore(t, db, "feed_link_delete_id", "link_link_delete_id")
	ctx := context.Background()

	// 1. Create a link
	link := NewLink().SetTitle("LinkToDeleteByID").SetFeedID("feed1").SetURL("http://delete.id").SetStatus(LINK_STATUS_ACTIVE)
	err := store.LinkCreate(ctx, link)
	if err != nil {
		t.Fatalf("LinkCreate failed: %v", err)
	}
	linkID := link.ID()

	// 2. Verify it exists
	foundLink, err := store.LinkFindByID(ctx, linkID)
	if err != nil {
		t.Fatalf("LinkFindByID before delete failed: %v", err)
	}
	if foundLink == nil {
		t.Fatal("Link not found before delete")
	}

	// 3. Delete using LinkDeleteByID
	err = store.LinkDeleteByID(ctx, linkID)
	if err != nil {
		t.Fatalf("LinkDeleteByID failed: %v", err)
	}

	// 4. Verify it's gone
	foundLink, err = store.LinkFindByID(ctx, linkID)
	if err != nil {
		t.Fatalf("LinkFindByID after delete failed: %v", err)
	}
	if foundLink != nil {
		t.Error("Link should not be found after delete")
	}

	// 5. Test deleting with empty ID
	err = store.LinkDeleteByID(ctx, "")
	if err == nil {
		t.Error("LinkDeleteByID should return error for empty ID")
	}
}

// --- LinkQuery needs SetFeedID ---
// Add the following to link_query_interface.go:
/*
	IsFeedIDSet() bool
	GetFeedID() string
	SetFeedID(feedID string) LinkQueryInterface
*/

// Add the following to link_query.go:
/*
func (q *linkQuery) IsFeedIDSet() bool {
	return q.hasProperty(COLUMN_FEED_ID)
}

func (q *linkQuery) GetFeedID() string {
	if q.IsFeedIDSet() {
		return q.params[COLUMN_FEED_ID].(string)
	}
	return ""
}

func (q *linkQuery) SetFeedID(feedID string) LinkQueryInterface {
	q.params[COLUMN_FEED_ID] = feedID
	return q
}

// Inside ToSelectDataset in link_query.go, add:
	// FeedID filter
	if q.IsFeedIDSet() {
		sql = sql.Where(goqu.C(COLUMN_FEED_ID).Eq(q.GetFeedID()))
	}
*/

func TestStoreLinkList(t *testing.T) {
	db := initDB(":memory:")
	defer db.Close()
	store := createTestStore(t, db, "feed_link_list", "link_link_list")
	ctx := context.Background()

	// Create test links
	link1 := NewLink().SetTitle("Link 1").SetFeedID("feedA").SetStatus(LINK_STATUS_ACTIVE).SetURL("url1")
	link2 := NewLink().SetTitle("Link 2").SetFeedID("feedB").SetStatus(LINK_STATUS_INACTIVE).SetURL("url2")
	link3 := NewLink().SetTitle("Link 3").SetFeedID("feedA").SetStatus(LINK_STATUS_ACTIVE).SetURL("url3")
	link4 := NewLink().SetTitle("Link 4").SetFeedID("feedC").SetStatus(LINK_STATUS_ACTIVE).SetURL("url4") // To be soft deleted

	if err := store.LinkCreate(ctx, link1); err != nil {
		t.Fatalf("Failed to create link1: %v", err)
	}
	time.Sleep(1 * time.Second)
	if err := store.LinkCreate(ctx, link2); err != nil {
		t.Fatalf("Failed to create link2: %v", err)
	}
	time.Sleep(1 * time.Second)
	if err := store.LinkCreate(ctx, link3); err != nil {
		t.Fatalf("Failed to create link3: %v", err)
	}
	time.Sleep(1 * time.Second)
	if err := store.LinkCreate(ctx, link4); err != nil {
		t.Fatalf("Failed to create link4: %v", err)
	}
	if err := store.LinkSoftDelete(ctx, link4); err != nil {
		t.Fatalf("Failed to soft delete link4: %v", err)
	} // Soft delete link4

	// Test cases
	testCases := []struct {
		name          string
		query         LinkQueryInterface
		expectedCount int
		expectedIDs   []string
		expectError   bool
	}{
		{
			name:          "List all (excluding soft deleted)",
			query:         LinkQuery().SetLimit(10),
			expectedCount: 3,
			expectedIDs:   []string{link1.ID(), link2.ID(), link3.ID()},
		},
		{
			name:          "List with specific ID",
			query:         LinkQuery().SetID(link2.ID()),
			expectedCount: 1,
			expectedIDs:   []string{link2.ID()},
		},
		{
			name:          "List by FeedID",
			query:         LinkQuery().SetFeedID("feedA").SetLimit(10), // Assumes SetFeedID is implemented
			expectedCount: 2,
			expectedIDs:   []string{link1.ID(), link3.ID()},
		},
		{
			name:          "List with specific Status",
			query:         LinkQuery().SetStatus(LINK_STATUS_ACTIVE).SetLimit(10),
			expectedCount: 2, // link1, link3
			expectedIDs:   []string{link1.ID(), link3.ID()},
		},
		{
			name:          "List including soft deleted",
			query:         LinkQuery().SetLimit(10).SetWithSoftDeleted(true),
			expectedCount: 4,
			expectedIDs:   []string{link1.ID(), link2.ID(), link3.ID(), link4.ID()},
		},
		{
			name:          "List only soft deleted",
			query:         LinkQuery().SetLimit(10).SetOnlySoftDeleted(true),
			expectedCount: 1,
			expectedIDs:   []string{link4.ID()},
		},
		{
			name:          "List non-existent ID",
			query:         LinkQuery().SetID("non-existent"),
			expectedCount: 0,
		},
		{
			name:        "List with invalid query (negative limit)",
			query:       LinkQuery().SetLimit(-1),
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Check if SetFeedID is implemented before running that specific test
			if strings.Contains(tc.name, "FeedID") {
				if _, ok := tc.query.(interface {
					SetFeedID(string) LinkQueryInterface
				}); !ok {
					t.Skip("Skipping FeedID test: SetFeedID not implemented on LinkQueryInterface")
				}
			}

			links, err := store.LinkList(ctx, tc.query)

			if tc.expectError {
				if err == nil {
					t.Errorf("Expected an error for invalid query '%s', but got nil", tc.name)
				}
				if len(links) != 0 {
					t.Errorf("Expected empty link list on error for '%s', but got %d links", tc.name, len(links))
				}
				return
			}

			if err != nil {
				t.Fatalf("LinkList for '%s' should not return an error, but got: %v", tc.name, err)
			}
			if len(links) != tc.expectedCount {
				t.Errorf("LinkList for '%s' returned wrong number of links: expected %d, got %d", tc.name, tc.expectedCount, len(links))
			}

			if len(tc.expectedIDs) > 0 {
				returnedIDs := make([]string, len(links))
				for i, l := range links {
					returnedIDs[i] = l.ID()
				}
				// Use elementsMatch helper for unordered comparison
				if !elementsMatch(t, tc.expectedIDs, returnedIDs) {
					t.Errorf("Returned link IDs for '%s' do not match expected set. Expected %v (any order), got %v", tc.name, tc.expectedIDs, returnedIDs)
				}
			}
		})
	}
}

func TestStoreLinkSoftDelete(t *testing.T) {
	db := initDB(":memory:")
	defer db.Close()
	store := createTestStore(t, db, "feed_link_soft_delete", "link_link_soft_delete")
	ctx := context.Background()

	// 1. Create a link
	link := NewLink().SetTitle("LinkToSoftDelete").SetFeedID("feed1").SetURL("http://softdel.me").SetStatus(LINK_STATUS_ACTIVE)
	err := store.LinkCreate(ctx, link)
	if err != nil {
		t.Fatalf("LinkCreate failed: %v", err)
	}

	// 2. Verify it exists and is not soft deleted
	foundLink, err := store.LinkFindByID(ctx, link.ID())
	if err != nil {
		t.Fatalf("LinkFindByID before soft delete failed: %v", err)
	}
	if foundLink == nil {
		t.Fatal("Link not found before soft delete")
	}
	if !foundLink.SoftDeletedAtCarbon().Gt(carbon.Now()) {
		t.Errorf("SoftDeletedAt should be in the future initially, but was %s", foundLink.SoftDeletedAt())
	}

	// 3. Soft delete using LinkSoftDelete
	err = store.LinkSoftDelete(ctx, link)
	if err != nil {
		t.Fatalf("LinkSoftDelete failed: %v", err)
	}

	// 4. Verify it's marked as deleted in the object
	if link.SoftDeletedAtCarbon().Gt(carbon.Now()) {
		t.Errorf("SoftDeletedAt should be in the past after soft delete in object, but was %s", link.SoftDeletedAt())
	}

	// 5. Verify it's not found by default FindByID
	foundLink, err = store.LinkFindByID(ctx, link.ID())
	if err != nil {
		t.Fatalf("LinkFindByID after soft delete failed: %v", err)
	}
	if foundLink != nil {
		t.Error("Link should not be found by default FindByID after soft delete")
	}

	// 6. Verify it IS found when including soft deleted
	list, err := store.LinkList(ctx, LinkQuery().SetID(link.ID()).SetWithSoftDeleted(true))
	if err != nil {
		t.Fatalf("LinkList with soft deleted failed: %v", err)
	}
	if len(list) != 1 {
		t.Fatalf("Link should be found when including soft deleted, expected 1 got %d", len(list))
	}
	if list[0].SoftDeletedAtCarbon().Gt(carbon.Now()) {
		t.Errorf("Found link's SoftDeletedAt should be in the past, but was %s", list[0].SoftDeletedAt())
	}

	// 7. Test soft deleting nil link
	err = store.LinkSoftDelete(ctx, nil)
	if err == nil {
		t.Error("LinkSoftDelete should return error for nil link")
	}
}

func TestStoreLinkSoftDeleteByID(t *testing.T) {
	db := initDB(":memory:")
	defer db.Close()
	store := createTestStore(t, db, "feed_link_soft_delete_id", "link_link_soft_delete_id")
	ctx := context.Background()

	// 1. Create a link
	link := NewLink().SetTitle("LinkToSoftDeleteByID").SetFeedID("feed1").SetURL("http://softdel.id").SetStatus(LINK_STATUS_ACTIVE)
	err := store.LinkCreate(ctx, link)
	if err != nil {
		t.Fatalf("LinkCreate failed: %v", err)
	}
	linkID := link.ID()

	// 2. Verify it exists
	foundLink, err := store.LinkFindByID(ctx, linkID)
	if err != nil {
		t.Fatalf("LinkFindByID before soft delete failed: %v", err)
	}
	if foundLink == nil {
		t.Fatal("Link not found before soft delete")
	}

	// 3. Soft delete using LinkSoftDeleteByID
	err = store.LinkSoftDeleteByID(ctx, linkID)
	if err != nil {
		t.Fatalf("LinkSoftDeleteByID failed: %v", err)
	}

	// 4. Verify it's not found by default FindByID
	foundLink, err = store.LinkFindByID(ctx, linkID)
	if err != nil {
		t.Fatalf("LinkFindByID after soft delete failed: %v", err)
	}
	if foundLink != nil {
		t.Error("Link should not be found by default FindByID after soft delete")
	}

	// 5. Verify it IS found when including soft deleted
	list, err := store.LinkList(ctx, LinkQuery().SetID(linkID).SetWithSoftDeleted(true))
	if err != nil {
		t.Fatalf("LinkList with soft deleted failed: %v", err)
	}
	if len(list) != 1 {
		t.Fatalf("Link should be found when including soft deleted, expected 1 got %d", len(list))
	}
	if list[0].SoftDeletedAtCarbon().Gt(carbon.Now()) {
		t.Errorf("Found link's SoftDeletedAt should be in the past, but was %s", list[0].SoftDeletedAt())
	}

	// 6. Test soft deleting non-existent ID
	// Similar to FeedSoftDeleteByID, this will likely error because LinkSoftDelete(nil) is called.
	err = store.LinkSoftDeleteByID(ctx, "non-existent-id")
	if err == nil {
		t.Error("LinkSoftDeleteByID for non-existent ID should error (due to LinkSoftDelete(nil)), but got nil")
	}

	// 7. Test soft deleting with empty ID
	err = store.LinkSoftDeleteByID(ctx, "")
	if err == nil {
		t.Error("LinkSoftDeleteByID should return error for empty ID")
	}
}

func TestStoreLinkUpdate(t *testing.T) {
	db := initDB(":memory:")
	defer db.Close()
	store := createTestStore(t, db, "feed_link_update", "link_link_update")
	ctx := context.Background()

	// 1. Create a link
	link := NewLink().SetTitle("Original Title").SetStatus(LINK_STATUS_INACTIVE).SetFeedID("feedX").SetURL("http://original.url")
	err := store.LinkCreate(ctx, link)
	if err != nil {
		t.Fatalf("LinkCreate failed: %v", err)
	}
	linkID := link.ID()
	initialUpdatedAt := link.UpdatedAt()

	// 2. Modify the link object
	newTitle := "Updated Title"
	newStatus := LINK_STATUS_ACTIVE
	newURL := "http://updated.url"
	link.SetTitle(newTitle)
	link.SetStatus(newStatus)
	link.SetURL(newURL)

	// 3. Update the link in the store
	time.Sleep(1 * time.Second)
	err = store.LinkUpdate(ctx, link)
	if err != nil {
		t.Fatalf("LinkUpdate failed: %v", err)
	}

	// 4. Verify the object is marked as not dirty
	if len(link.DataChanged()) != 0 {
		t.Errorf("DataChanged should be empty after successful update, but got %v", link.DataChanged())
	}
	if initialUpdatedAt == link.UpdatedAt() {
		t.Error("UpdatedAt should have changed after update")
	}
	updatedAtAfterUpdate := link.UpdatedAt() // Store for next check

	// 5. Retrieve the link and verify changes
	updatedLink, err := store.LinkFindByID(ctx, linkID)
	if err != nil {
		t.Fatalf("LinkFindByID after update failed: %v", err)
	}
	if updatedLink == nil {
		t.Fatal("Updated link not found")
	}
	if updatedLink.Title() != newTitle {
		t.Errorf("Title update failed: expected '%s', got '%s'", newTitle, updatedLink.Title())
	}
	if updatedLink.Status() != newStatus {
		t.Errorf("Status update failed: expected '%s', got '%s'", newStatus, updatedLink.Status())
	}
	if updatedLink.URL() != newURL {
		t.Errorf("URL update failed: expected '%s', got '%s'", newURL, updatedLink.URL())
	}
	if strings.ReplaceAll(updatedLink.UpdatedAt(), " +0000 UTC", "") != updatedAtAfterUpdate {
		t.Errorf("UpdatedAt mismatch: expected '%s', got '%s'", updatedAtAfterUpdate, strings.ReplaceAll(updatedLink.UpdatedAt(), " +0000 UTC", ""))
	}

	// 6. Test updating with no changes
	time.Sleep(1 * time.Second)
	err = store.LinkUpdate(ctx, updatedLink) // No fields changed since last MarkAsNotDirty
	if err != nil {
		t.Fatalf("Update with no changes should not error, but got: %v", err)
	}
	// Retrieve again to check DB
	finalLink, err := store.LinkFindByID(ctx, linkID)
	if err != nil {
		t.Fatalf("LinkFindByID after no-change update failed: %v", err)
	}
	if finalLink == nil {
		t.Fatal("Link not found after no-change update")
	}
	if strings.ReplaceAll(finalLink.UpdatedAt(), " +0000 UTC", "") != updatedAtAfterUpdate {
		t.Errorf("UpdatedAt should not change if no fields were modified, expected '%s', got '%s'", updatedAtAfterUpdate, strings.ReplaceAll(finalLink.UpdatedAt(), " +0000 UTC", ""))
	}

	// 7. Test updating nil link
	err = store.LinkUpdate(ctx, nil)
	if err == nil {
		t.Error("LinkUpdate should return error for nil link")
	}
}
