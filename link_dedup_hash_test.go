package feedstore

import (
	"context"
	"testing"
)

// TestStoreLinkDedupHash_RoundTrip verifies that a dedup_hash set on a link
// survives a create + find round-trip.
func TestStoreLinkDedupHash_RoundTrip(t *testing.T) {
	db := initDB(":memory:")
	defer db.Close()
	store := createTestStore(t, db, "feed_dedup_rt", "link_dedup_rt")
	ctx := context.Background()

	feed := NewFeed().SetName("Test Feed")
	if err := store.FeedCreate(ctx, feed); err != nil {
		t.Fatalf("FeedCreate failed: %v", err)
	}

	hash := "abc123def4567890abcdef1234567890abcdef1234567890abcdef1234567890"
	link := NewLink().
		SetFeedID(feed.GetID()).
		SetTitle("Dedup Test").
		SetURL("https://example.com/dedup").
		SetStatus(LINK_STATUS_ACTIVE).
		SetDedupHash(hash)

	if err := store.LinkCreate(ctx, link); err != nil {
		t.Fatalf("LinkCreate failed: %v", err)
	}

	found, err := store.LinkFindByID(ctx, link.GetID())
	if err != nil {
		t.Fatalf("LinkFindByID failed: %v", err)
	}
	if found == nil {
		t.Fatal("link not found after create")
	}
	if found.GetDedupHash() != hash {
		t.Errorf("expected dedup_hash '%s', got '%s'", hash, found.GetDedupHash())
	}
}

// TestStoreLinkDedupHash_Query verifies that LinkList filters by dedup_hash
// at the SQL level.
func TestStoreLinkDedupHash_Query(t *testing.T) {
	db := initDB(":memory:")
	defer db.Close()
	store := createTestStore(t, db, "feed_dedup_q", "link_dedup_q")
	ctx := context.Background()

	feed := NewFeed().SetName("Test Feed")
	if err := store.FeedCreate(ctx, feed); err != nil {
		t.Fatalf("FeedCreate failed: %v", err)
	}

	linkA := NewLink().
		SetFeedID(feed.GetID()).
		SetTitle("Link A").
		SetURL("https://example.com/a").
		SetStatus(LINK_STATUS_ACTIVE).
		SetDedupHash("hashA")
	if err := store.LinkCreate(ctx, linkA); err != nil {
		t.Fatalf("LinkCreate A failed: %v", err)
	}

	linkB := NewLink().
		SetFeedID(feed.GetID()).
		SetTitle("Link B").
		SetURL("https://example.com/b").
		SetStatus(LINK_STATUS_ACTIVE).
		SetDedupHash("hashB")
	if err := store.LinkCreate(ctx, linkB); err != nil {
		t.Fatalf("LinkCreate B failed: %v", err)
	}

	// Query by hashA — should return exactly 1 link
	links, err := store.LinkList(ctx, LinkQuery().SetDedupHash("hashA").SetLimit(10))
	if err != nil {
		t.Fatalf("LinkList by dedup_hash failed: %v", err)
	}
	if len(links) != 1 {
		t.Fatalf("expected 1 link with dedup_hash 'hashA', got %d", len(links))
	}
	if links[0].GetID() != linkA.GetID() {
		t.Errorf("expected linkA ID, got %s", links[0].GetID())
	}
}

// TestStoreLinkDedupHash_Empty verifies that a link with an empty dedup_hash
// saves and reads back as empty string (nullable column).
func TestStoreLinkDedupHash_Empty(t *testing.T) {
	db := initDB(":memory:")
	defer db.Close()
	store := createTestStore(t, db, "feed_dedup_empty", "link_dedup_empty")
	ctx := context.Background()

	feed := NewFeed().SetName("Test Feed")
	if err := store.FeedCreate(ctx, feed); err != nil {
		t.Fatalf("FeedCreate failed: %v", err)
	}

	link := NewLink().
		SetFeedID(feed.GetID()).
		SetTitle("No Dedup").
		SetURL("https://example.com/no-dedup").
		SetStatus(LINK_STATUS_ACTIVE).
		SetDedupHash("")

	if err := store.LinkCreate(ctx, link); err != nil {
		t.Fatalf("LinkCreate failed: %v", err)
	}

	found, err := store.LinkFindByID(ctx, link.GetID())
	if err != nil {
		t.Fatalf("LinkFindByID failed: %v", err)
	}
	if found == nil {
		t.Fatal("link not found after create")
	}
	if found.GetDedupHash() != "" {
		t.Errorf("expected empty dedup_hash, got '%s'", found.GetDedupHash())
	}
}

// TestLinkQuery_DedupHash_Validate verifies that Validate rejects an empty
// dedup_hash when the flag is set.
func TestLinkQuery_DedupHash_Validate(t *testing.T) {
	q := LinkQuery()

	// No dedup_hash set — should be valid
	if err := q.Validate(); err != nil {
		t.Errorf("Validate should pass with no dedup_hash filter, got: %v", err)
	}

	// Set dedup_hash to empty — should fail
	q.SetDedupHash("")
	if err := q.Validate(); err == nil {
		t.Error("Validate should fail when dedup_hash is set to empty string")
	}

	// Valid value — should pass
	q2 := LinkQuery()
	q2.SetDedupHash("somehash")
	if err := q2.Validate(); err != nil {
		t.Errorf("Validate should pass with valid dedup_hash, got: %v", err)
	}
}

// TestLinkQuery_DedupHash_FluentChaining verifies that SetDedupHash returns
// the same LinkQueryInterface for chaining.
func TestLinkQuery_DedupHash_FluentChaining(t *testing.T) {
	q := LinkQuery()
	result := q.SetDedupHash("hashX")
	if result != q {
		t.Error("SetDedupHash should return the same LinkQueryInterface for chaining")
	}
}

// TestStoreLinkCount_DedupHash verifies that LinkCount respects the dedup_hash filter.
func TestStoreLinkCount_DedupHash(t *testing.T) {
	db := initDB(":memory:")
	defer db.Close()
	store := createTestStore(t, db, "feed_dedup_count", "link_dedup_count")
	ctx := context.Background()

	feed := NewFeed().SetName("Test Feed")
	if err := store.FeedCreate(ctx, feed); err != nil {
		t.Fatalf("FeedCreate failed: %v", err)
	}

	linkA := NewLink().
		SetFeedID(feed.GetID()).
		SetTitle("Link A").
		SetURL("https://example.com/a").
		SetStatus(LINK_STATUS_ACTIVE).
		SetDedupHash("hashA")
	if err := store.LinkCreate(ctx, linkA); err != nil {
		t.Fatalf("LinkCreate A failed: %v", err)
	}

	linkB := NewLink().
		SetFeedID(feed.GetID()).
		SetTitle("Link B").
		SetURL("https://example.com/b").
		SetStatus(LINK_STATUS_ACTIVE).
		SetDedupHash("hashB")
	if err := store.LinkCreate(ctx, linkB); err != nil {
		t.Fatalf("LinkCreate B failed: %v", err)
	}

	// Count with dedup_hash filter — should be 1
	count, err := store.LinkCount(ctx, LinkQuery().SetDedupHash("hashA"))
	if err != nil {
		t.Fatalf("LinkCount with dedup_hash failed: %v", err)
	}
	if count != 1 {
		t.Errorf("expected count=1 with dedup_hash filter, got %d", count)
	}

	// Count without filter — should be 2
	count, err = store.LinkCount(ctx, LinkQuery())
	if err != nil {
		t.Fatalf("LinkCount without filter failed: %v", err)
	}
	if count != 2 {
		t.Errorf("expected count=2 without filter, got %d", count)
	}
}
