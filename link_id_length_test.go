package feedstore

import (
	"context"
	"strings"
	"testing"
)

// TestStoreLinkID_40Chars verifies that a 40-char ID survives a round-trip
// through create + find. This confirms the widened VARCHAR(40) column.
func TestStoreLinkID_40Chars(t *testing.T) {
	db := initDB(":memory:")
	defer db.Close()
	store := createTestStore(t, db, "feed_id40", "link_id40")
	ctx := context.Background()

	feed := NewFeed().SetName("Test Feed")
	if err := store.FeedCreate(ctx, feed); err != nil {
		t.Fatalf("FeedCreate failed: %v", err)
	}

	longID := strings.Repeat("a", 40)
	link := NewLink().
		SetID(longID).
		SetFeedID(feed.GetID()).
		SetTitle("Long ID Link").
		SetURL("https://example.com/long-id").
		SetStatus(LINK_STATUS_ACTIVE)

	if err := store.LinkCreate(ctx, link); err != nil {
		t.Fatalf("LinkCreate with 40-char ID failed: %v", err)
	}

	found, err := store.LinkFindByID(ctx, longID)
	if err != nil {
		t.Fatalf("LinkFindByID failed: %v", err)
	}
	if found == nil {
		t.Fatal("link with 40-char ID not found after create")
	}
	if found.GetID() != longID {
		t.Errorf("expected ID '%s', got '%s'", longID, found.GetID())
	}
}

// TestStoreLinkID_UUID verifies that a UUID (36 chars with hyphens) survives
// a round-trip through create + find.
func TestStoreLinkID_UUID(t *testing.T) {
	db := initDB(":memory:")
	defer db.Close()
	store := createTestStore(t, db, "feed_id_uuid", "link_id_uuid")
	ctx := context.Background()

	feed := NewFeed().SetName("Test Feed")
	if err := store.FeedCreate(ctx, feed); err != nil {
		t.Fatalf("FeedCreate failed: %v", err)
	}

	uuidID := "550e8400-e29b-41d4-a716-446655440000"
	link := NewLink().
		SetID(uuidID).
		SetFeedID(feed.GetID()).
		SetTitle("UUID Link").
		SetURL("https://example.com/uuid").
		SetStatus(LINK_STATUS_ACTIVE)

	if err := store.LinkCreate(ctx, link); err != nil {
		t.Fatalf("LinkCreate with UUID failed: %v", err)
	}

	found, err := store.LinkFindByID(ctx, uuidID)
	if err != nil {
		t.Fatalf("LinkFindByID failed: %v", err)
	}
	if found == nil {
		t.Fatal("link with UUID not found after create")
	}
	if found.GetID() != uuidID {
		t.Errorf("expected ID '%s', got '%s'", uuidID, found.GetID())
	}
}

// TestStoreLinkID_Legacy9Chars verifies that a legacy 9-char ID still works
// after the column widening (backward compatibility).
func TestStoreLinkID_Legacy9Chars(t *testing.T) {
	db := initDB(":memory:")
	defer db.Close()
	store := createTestStore(t, db, "feed_id_legacy", "link_id_legacy")
	ctx := context.Background()

	feed := NewFeed().SetName("Test Feed")
	if err := store.FeedCreate(ctx, feed); err != nil {
		t.Fatalf("FeedCreate failed: %v", err)
	}

	legacyID := "abc123def" // 9 chars — the old column width
	link := NewLink().
		SetID(legacyID).
		SetFeedID(feed.GetID()).
		SetTitle("Legacy ID Link").
		SetURL("https://example.com/legacy").
		SetStatus(LINK_STATUS_ACTIVE)

	if err := store.LinkCreate(ctx, link); err != nil {
		t.Fatalf("LinkCreate with legacy 9-char ID failed: %v", err)
	}

	found, err := store.LinkFindByID(ctx, legacyID)
	if err != nil {
		t.Fatalf("LinkFindByID failed: %v", err)
	}
	if found == nil {
		t.Fatal("link with legacy 9-char ID not found after create")
	}
	if found.GetID() != legacyID {
		t.Errorf("expected ID '%s', got '%s'", legacyID, found.GetID())
	}
}

// TestStoreFeedID_40Chars verifies that a 40-char feed ID survives a round-trip.
func TestStoreFeedID_40Chars(t *testing.T) {
	db := initDB(":memory:")
	defer db.Close()
	store := createTestStore(t, db, "feed_id40_feed", "link_id40_feed")
	ctx := context.Background()

	longID := strings.Repeat("f", 40)
	feed := NewFeed().
		SetID(longID).
		SetName("Long ID Feed").
		SetURL("https://example.com/feed").
		SetStatus(FEED_STATUS_ACTIVE)

	if err := store.FeedCreate(ctx, feed); err != nil {
		t.Fatalf("FeedCreate with 40-char ID failed: %v", err)
	}

	found, err := store.FeedFindByID(ctx, longID)
	if err != nil {
		t.Fatalf("FeedFindByID failed: %v", err)
	}
	if found == nil {
		t.Fatal("feed with 40-char ID not found after create")
	}
	if found.GetID() != longID {
		t.Errorf("expected ID '%s', got '%s'", longID, found.GetID())
	}
}
