package feedstore

import (
	"context"
	"testing"
)

// TestLinkQuery_PublishedAtGte_PublishedAtLte_GettersSetters verifies the query object
// correctly stores and retrieves PublishedAtGte/PublishedAtLte values.
func TestLinkQuery_PublishedAtGte_PublishedAtLte_GettersSetters(t *testing.T) {
	q := LinkQuery()

	// Initially not set
	if q.IsPublishedAtGteSet() {
		t.Error("IsPublishedAtGteSet should be false on new query")
	}
	if q.IsPublishedAtLteSet() {
		t.Error("IsPublishedAtLteSet should be false on new query")
	}
	if q.GetPublishedAtGte() != "" {
		t.Errorf("GetPublishedAtGte should return empty string when not set, got %q", q.GetPublishedAtGte())
	}
	if q.GetPublishedAtLte() != "" {
		t.Errorf("GetPublishedAtLte should return empty string when not set, got %q", q.GetPublishedAtLte())
	}

	// Set values
	q.SetPublishedAtGte("2026-01-01 00:00:00")
	q.SetPublishedAtLte("2026-12-31 23:59:59")

	if !q.IsPublishedAtGteSet() {
		t.Error("IsPublishedAtGteSet should be true after SetPublishedAtGte")
	}
	if !q.IsPublishedAtLteSet() {
		t.Error("IsPublishedAtLteSet should be true after SetPublishedAtLte")
	}
	if q.GetPublishedAtGte() != "2026-01-01 00:00:00" {
		t.Errorf("GetPublishedAtGte mismatch, got %q", q.GetPublishedAtGte())
	}
	if q.GetPublishedAtLte() != "2026-12-31 23:59:59" {
		t.Errorf("GetPublishedAtLte mismatch, got %q", q.GetPublishedAtLte())
	}
}

// TestLinkQuery_PublishedAtGte_PublishedAtLte_Validate verifies that Validate rejects
// empty strings when the flag is set.
func TestLinkQuery_PublishedAtGte_PublishedAtLte_Validate(t *testing.T) {
	q := LinkQuery()

	// No PublishedAt filters set — should be valid
	if err := q.Validate(); err != nil {
		t.Errorf("Validate should pass with no PublishedAt filters, got: %v", err)
	}

	// Set PublishedAtGte to empty — should fail
	q.SetPublishedAtGte("")
	if err := q.Validate(); err == nil {
		t.Error("Validate should fail when PublishedAtGte is set to empty string")
	}

	// Set PublishedAtLte to empty — should fail
	q2 := LinkQuery()
	q2.SetPublishedAtLte("")
	if err := q2.Validate(); err == nil {
		t.Error("Validate should fail when PublishedAtLte is set to empty string")
	}

	// Valid values — should pass
	q3 := LinkQuery()
	q3.SetPublishedAtGte("2026-01-01")
	q3.SetPublishedAtLte("2026-12-31")
	if err := q3.Validate(); err != nil {
		t.Errorf("Validate should pass with valid PublishedAtGte/PublishedAtLte, got: %v", err)
	}
}

// TestLinkQuery_PublishedAtGte_FluentChaining verifies that SetPublishedAtGte returns
// the query interface for fluent chaining.
func TestLinkQuery_PublishedAtGte_FluentChaining(t *testing.T) {
	q := LinkQuery()
	result := q.SetPublishedAtGte("2026-01-01")
	if result != q {
		t.Error("SetPublishedAtGte should return the same LinkQueryInterface for chaining")
	}
	result2 := q.SetPublishedAtLte("2026-12-31")
	if result2 != q {
		t.Error("SetPublishedAtLte should return the same LinkQueryInterface for chaining")
	}
}

// TestStoreLinkList_PublishedAtGteFilter verifies that the store correctly
// filters links by published_at >= ? at the SQL level.
func TestStoreLinkList_PublishedAtGteFilter(t *testing.T) {
	db := initDB(":memory:")
	defer db.Close()
	store := createTestStore(t, db, "feed_published_at_gte", "link_published_at_gte")
	ctx := context.Background()

	// Create links with specific publication times
	linkOld := NewLink().SetTitle("Old").SetFeedID("feedA").SetStatus(LINK_STATUS_ACTIVE).SetURL("url_old")
	linkOld.SetPublishedAtString("2020-01-01 10:00:00")
	if err := store.LinkCreate(ctx, linkOld); err != nil {
		t.Fatalf("Failed to create linkOld: %v", err)
	}

	linkRecent := NewLink().SetTitle("Recent").SetFeedID("feedA").SetStatus(LINK_STATUS_ACTIVE).SetURL("url_recent")
	linkRecent.SetPublishedAtString("2026-08-31 10:00:00")
	if err := store.LinkCreate(ctx, linkRecent); err != nil {
		t.Fatalf("Failed to create linkRecent: %v", err)
	}

	// Query with PublishedAtGte = 2026-01-01 — should only return linkRecent
	links, err := store.LinkList(ctx, LinkQuery().
		SetPublishedAtGte("2026-01-01 00:00:00").
		SetLimit(10))
	if err != nil {
		t.Fatalf("LinkList with PublishedAtGte failed: %v", err)
	}
	if len(links) != 1 {
		t.Fatalf("expected 1 link with PublishedAtGte filter, got %d", len(links))
	}
	if links[0].GetID() != linkRecent.GetID() {
		t.Errorf("expected linkRecent ID, got %s", links[0].GetID())
	}
}

// TestStoreLinkList_PublishedAtLteFilter verifies that the store correctly
// filters links by published_at <= ? at the SQL level.
func TestStoreLinkList_PublishedAtLteFilter(t *testing.T) {
	db := initDB(":memory:")
	defer db.Close()
	store := createTestStore(t, db, "feed_published_at_lte", "link_published_at_lte")
	ctx := context.Background()

	linkOld := NewLink().SetTitle("Old").SetFeedID("feedA").SetStatus(LINK_STATUS_ACTIVE).SetURL("url_old")
	linkOld.SetPublishedAtString("2020-01-01 10:00:00")
	if err := store.LinkCreate(ctx, linkOld); err != nil {
		t.Fatalf("Failed to create linkOld: %v", err)
	}

	linkRecent := NewLink().SetTitle("Recent").SetFeedID("feedA").SetStatus(LINK_STATUS_ACTIVE).SetURL("url_recent")
	linkRecent.SetPublishedAtString("2026-08-31 10:00:00")
	if err := store.LinkCreate(ctx, linkRecent); err != nil {
		t.Fatalf("Failed to create linkRecent: %v", err)
	}

	// Query with PublishedAtLte = 2021-01-01 — should only return linkOld
	links, err := store.LinkList(ctx, LinkQuery().
		SetPublishedAtLte("2021-01-01 00:00:00").
		SetLimit(10))
	if err != nil {
		t.Fatalf("LinkList with PublishedAtLte failed: %v", err)
	}
	if len(links) != 1 {
		t.Fatalf("expected 1 link with PublishedAtLte filter, got %d", len(links))
	}
	if links[0].GetID() != linkOld.GetID() {
		t.Errorf("expected linkOld ID, got %s", links[0].GetID())
	}
}

// TestStoreLinkList_PublishedAtRangeFilter verifies that the store correctly
// filters links by both published_at >= ? AND published_at <= ? at the SQL level.
func TestStoreLinkList_PublishedAtRangeFilter(t *testing.T) {
	db := initDB(":memory:")
	defer db.Close()
	store := createTestStore(t, db, "feed_published_at_range", "link_published_at_range")
	ctx := context.Background()

	// Create three links: before, in, and after the target range
	linkBefore := NewLink().SetTitle("Before").SetFeedID("feedA").SetStatus(LINK_STATUS_ACTIVE).SetURL("url_before")
	linkBefore.SetPublishedAtString("2020-01-01 10:00:00")
	if err := store.LinkCreate(ctx, linkBefore); err != nil {
		t.Fatalf("Failed to create linkBefore: %v", err)
	}

	linkInRange := NewLink().SetTitle("InRange").SetFeedID("feedA").SetStatus(LINK_STATUS_ACTIVE).SetURL("url_in_range")
	linkInRange.SetPublishedAtString("2025-06-15 12:00:00")
	if err := store.LinkCreate(ctx, linkInRange); err != nil {
		t.Fatalf("Failed to create linkInRange: %v", err)
	}

	linkAfter := NewLink().SetTitle("After").SetFeedID("feedA").SetStatus(LINK_STATUS_ACTIVE).SetURL("url_after")
	linkAfter.SetPublishedAtString("2026-08-31 10:00:00")
	if err := store.LinkCreate(ctx, linkAfter); err != nil {
		t.Fatalf("Failed to create linkAfter: %v", err)
	}

	// Query with PublishedAtGte=2025-01-01 AND PublishedAtLte=2025-12-31 — should only return linkInRange
	links, err := store.LinkList(ctx, LinkQuery().
		SetPublishedAtGte("2025-01-01 00:00:00").
		SetPublishedAtLte("2025-12-31 23:59:59").
		SetLimit(10))
	if err != nil {
		t.Fatalf("LinkList with PublishedAtRange failed: %v", err)
	}
	if len(links) != 1 {
		t.Fatalf("expected 1 link in published_at range, got %d", len(links))
	}
	if links[0].GetID() != linkInRange.GetID() {
		t.Errorf("expected linkInRange ID, got %s", links[0].GetID())
	}
}

// TestStoreLinkList_PublishedAtGteNoMatch verifies that PublishedAtGte filter returns
// empty when no links match the published_at range.
func TestStoreLinkList_PublishedAtGteNoMatch(t *testing.T) {
	db := initDB(":memory:")
	defer db.Close()
	store := createTestStore(t, db, "feed_published_at_gte_nomatch", "link_published_at_gte_nomatch")
	ctx := context.Background()

	link := NewLink().SetTitle("Old").SetFeedID("feedA").SetStatus(LINK_STATUS_ACTIVE).SetURL("url1")
	link.SetPublishedAtString("2020-01-01 10:00:00")
	if err := store.LinkCreate(ctx, link); err != nil {
		t.Fatalf("Failed to create link: %v", err)
	}

	// Query with PublishedAtGte in the future — should return 0 links
	links, err := store.LinkList(ctx, LinkQuery().
		SetPublishedAtGte("2030-01-01 00:00:00").
		SetLimit(10))
	if err != nil {
		t.Fatalf("LinkList with future PublishedAtGte failed: %v", err)
	}
	if len(links) != 0 {
		t.Errorf("expected 0 links with future PublishedAtGte, got %d", len(links))
	}
}

// TestStoreLinkList_PublishedAtFilterWithOtherFilters verifies that PublishedAt filters
// work correctly in combination with other filters (FeedID, Status).
func TestStoreLinkList_PublishedAtFilterWithOtherFilters(t *testing.T) {
	db := initDB(":memory:")
	defer db.Close()
	store := createTestStore(t, db, "feed_published_at_combo", "link_published_at_combo")
	ctx := context.Background()

	// feedA, active, old published_at
	link1 := NewLink().SetTitle("A-Old").SetFeedID("feedA").SetStatus(LINK_STATUS_ACTIVE).SetURL("url1")
	link1.SetPublishedAtString("2020-01-01 10:00:00")
	if err := store.LinkCreate(ctx, link1); err != nil {
		t.Fatalf("Failed to create link1: %v", err)
	}

	// feedA, active, recent published_at
	link2 := NewLink().SetTitle("A-Recent").SetFeedID("feedA").SetStatus(LINK_STATUS_ACTIVE).SetURL("url2")
	link2.SetPublishedAtString("2026-08-31 10:00:00")
	if err := store.LinkCreate(ctx, link2); err != nil {
		t.Fatalf("Failed to create link2: %v", err)
	}

	// feedB, active, recent published_at
	link3 := NewLink().SetTitle("B-Recent").SetFeedID("feedB").SetStatus(LINK_STATUS_ACTIVE).SetURL("url3")
	link3.SetPublishedAtString("2026-08-31 10:00:00")
	if err := store.LinkCreate(ctx, link3); err != nil {
		t.Fatalf("Failed to create link3: %v", err)
	}

	// Query: feedA + PublishedAtGte=2026-01-01 — should only return link2
	links, err := store.LinkList(ctx, LinkQuery().
		SetFeedID("feedA").
		SetStatus(LINK_STATUS_ACTIVE).
		SetPublishedAtGte("2026-01-01 00:00:00").
		SetLimit(10))
	if err != nil {
		t.Fatalf("LinkList with combo filters failed: %v", err)
	}
	if len(links) != 1 {
		t.Fatalf("expected 1 link with combo filters, got %d", len(links))
	}
	if links[0].GetID() != link2.GetID() {
		t.Errorf("expected link2 ID, got %s", links[0].GetID())
	}
}

// TestStoreLinkCount_PublishedAtFilter verifies that LinkCount also respects
// the PublishedAtGte/PublishedAtLte filters.
func TestStoreLinkCount_PublishedAtFilter(t *testing.T) {
	db := initDB(":memory:")
	defer db.Close()
	store := createTestStore(t, db, "feed_published_at_count", "link_published_at_count")
	ctx := context.Background()

	linkOld := NewLink().SetTitle("Old").SetFeedID("feedA").SetStatus(LINK_STATUS_ACTIVE).SetURL("url_old")
	linkOld.SetPublishedAtString("2020-01-01 10:00:00")
	if err := store.LinkCreate(ctx, linkOld); err != nil {
		t.Fatalf("Failed to create linkOld: %v", err)
	}

	linkRecent := NewLink().SetTitle("Recent").SetFeedID("feedA").SetStatus(LINK_STATUS_ACTIVE).SetURL("url_recent")
	linkRecent.SetPublishedAtString("2026-08-31 10:00:00")
	if err := store.LinkCreate(ctx, linkRecent); err != nil {
		t.Fatalf("Failed to create linkRecent: %v", err)
	}

	// Count with PublishedAtGte filter — should be 1
	count, err := store.LinkCount(ctx, LinkQuery().SetPublishedAtGte("2026-01-01 00:00:00"))
	if err != nil {
		t.Fatalf("LinkCount with PublishedAtGte failed: %v", err)
	}
	if count != 1 {
		t.Errorf("expected count=1 with PublishedAtGte filter, got %d", count)
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
