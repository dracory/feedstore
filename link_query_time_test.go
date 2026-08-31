package feedstore

import (
	"context"
	"testing"
)

// TestLinkQuery_TimeGte_TimeLte_GettersSetters verifies the query object
// correctly stores and retrieves TimeGte/TimeLte values.
func TestLinkQuery_TimeGte_TimeLte_GettersSetters(t *testing.T) {
	q := LinkQuery()

	// Initially not set
	if q.IsTimeGteSet() {
		t.Error("IsTimeGteSet should be false on new query")
	}
	if q.IsTimeLteSet() {
		t.Error("IsTimeLteSet should be false on new query")
	}
	if q.GetTimeGte() != "" {
		t.Errorf("GetTimeGte should return empty string when not set, got %q", q.GetTimeGte())
	}
	if q.GetTimeLte() != "" {
		t.Errorf("GetTimeLte should return empty string when not set, got %q", q.GetTimeLte())
	}

	// Set values
	q.SetTimeGte("2026-01-01 00:00:00")
	q.SetTimeLte("2026-12-31 23:59:59")

	if !q.IsTimeGteSet() {
		t.Error("IsTimeGteSet should be true after SetTimeGte")
	}
	if !q.IsTimeLteSet() {
		t.Error("IsTimeLteSet should be true after SetTimeLte")
	}
	if q.GetTimeGte() != "2026-01-01 00:00:00" {
		t.Errorf("GetTimeGte mismatch, got %q", q.GetTimeGte())
	}
	if q.GetTimeLte() != "2026-12-31 23:59:59" {
		t.Errorf("GetTimeLte mismatch, got %q", q.GetTimeLte())
	}
}

// TestLinkQuery_TimeGte_TimeLte_Validate verifies that Validate rejects
// empty strings when the flag is set.
func TestLinkQuery_TimeGte_TimeLte_Validate(t *testing.T) {
	q := LinkQuery()

	// No Time filters set — should be valid
	if err := q.Validate(); err != nil {
		t.Errorf("Validate should pass with no Time filters, got: %v", err)
	}

	// Set TimeGte to empty — should fail
	q.SetTimeGte("")
	if err := q.Validate(); err == nil {
		t.Error("Validate should fail when TimeGte is set to empty string")
	}

	// Set TimeLte to empty — should fail
	q2 := LinkQuery()
	q2.SetTimeLte("")
	if err := q2.Validate(); err == nil {
		t.Error("Validate should fail when TimeLte is set to empty string")
	}

	// Valid values — should pass
	q3 := LinkQuery()
	q3.SetTimeGte("2026-01-01")
	q3.SetTimeLte("2026-12-31")
	if err := q3.Validate(); err != nil {
		t.Errorf("Validate should pass with valid TimeGte/TimeLte, got: %v", err)
	}
}

// TestLinkQuery_TimeGte_FluentChaining verifies that SetTimeGte returns
// the query interface for fluent chaining.
func TestLinkQuery_TimeGte_FluentChaining(t *testing.T) {
	q := LinkQuery()
	result := q.SetTimeGte("2026-01-01")
	if result != q {
		t.Error("SetTimeGte should return the same LinkQueryInterface for chaining")
	}
	result2 := q.SetTimeLte("2026-12-31")
	if result2 != q {
		t.Error("SetTimeLte should return the same LinkQueryInterface for chaining")
	}
}

// TestStoreLinkList_TimeGteFilter verifies that the store correctly
// filters links by time >= ? at the SQL level.
func TestStoreLinkList_TimeGteFilter(t *testing.T) {
	db := initDB(":memory:")
	defer db.Close()
	store := createTestStore(t, db, "feed_time_gte", "link_time_gte")
	ctx := context.Background()

	// Create links with specific publication times
	linkOld := NewLink().SetTitle("Old").SetFeedID("feedA").SetStatus(LINK_STATUS_ACTIVE).SetURL("url_old")
	linkOld.SetTimeString("2020-01-01 10:00:00")
	if err := store.LinkCreate(ctx, linkOld); err != nil {
		t.Fatalf("Failed to create linkOld: %v", err)
	}

	linkRecent := NewLink().SetTitle("Recent").SetFeedID("feedA").SetStatus(LINK_STATUS_ACTIVE).SetURL("url_recent")
	linkRecent.SetTimeString("2026-08-31 10:00:00")
	if err := store.LinkCreate(ctx, linkRecent); err != nil {
		t.Fatalf("Failed to create linkRecent: %v", err)
	}

	// Query with TimeGte = 2026-01-01 — should only return linkRecent
	links, err := store.LinkList(ctx, LinkQuery().
		SetTimeGte("2026-01-01 00:00:00").
		SetLimit(10))
	if err != nil {
		t.Fatalf("LinkList with TimeGte failed: %v", err)
	}
	if len(links) != 1 {
		t.Fatalf("expected 1 link with TimeGte filter, got %d", len(links))
	}
	if links[0].ID() != linkRecent.ID() {
		t.Errorf("expected linkRecent ID, got %s", links[0].ID())
	}
}

// TestStoreLinkList_TimeLteFilter verifies that the store correctly
// filters links by time <= ? at the SQL level.
func TestStoreLinkList_TimeLteFilter(t *testing.T) {
	db := initDB(":memory:")
	defer db.Close()
	store := createTestStore(t, db, "feed_time_lte", "link_time_lte")
	ctx := context.Background()

	linkOld := NewLink().SetTitle("Old").SetFeedID("feedA").SetStatus(LINK_STATUS_ACTIVE).SetURL("url_old")
	linkOld.SetTimeString("2020-01-01 10:00:00")
	if err := store.LinkCreate(ctx, linkOld); err != nil {
		t.Fatalf("Failed to create linkOld: %v", err)
	}

	linkRecent := NewLink().SetTitle("Recent").SetFeedID("feedA").SetStatus(LINK_STATUS_ACTIVE).SetURL("url_recent")
	linkRecent.SetTimeString("2026-08-31 10:00:00")
	if err := store.LinkCreate(ctx, linkRecent); err != nil {
		t.Fatalf("Failed to create linkRecent: %v", err)
	}

	// Query with TimeLte = 2021-01-01 — should only return linkOld
	links, err := store.LinkList(ctx, LinkQuery().
		SetTimeLte("2021-01-01 00:00:00").
		SetLimit(10))
	if err != nil {
		t.Fatalf("LinkList with TimeLte failed: %v", err)
	}
	if len(links) != 1 {
		t.Fatalf("expected 1 link with TimeLte filter, got %d", len(links))
	}
	if links[0].ID() != linkOld.ID() {
		t.Errorf("expected linkOld ID, got %s", links[0].ID())
	}
}

// TestStoreLinkList_TimeRangeFilter verifies that the store correctly
// filters links by both time >= ? AND time <= ? at the SQL level.
func TestStoreLinkList_TimeRangeFilter(t *testing.T) {
	db := initDB(":memory:")
	defer db.Close()
	store := createTestStore(t, db, "feed_time_range", "link_time_range")
	ctx := context.Background()

	// Create three links: before, in, and after the target range
	linkBefore := NewLink().SetTitle("Before").SetFeedID("feedA").SetStatus(LINK_STATUS_ACTIVE).SetURL("url_before")
	linkBefore.SetTimeString("2020-01-01 10:00:00")
	if err := store.LinkCreate(ctx, linkBefore); err != nil {
		t.Fatalf("Failed to create linkBefore: %v", err)
	}

	linkInRange := NewLink().SetTitle("InRange").SetFeedID("feedA").SetStatus(LINK_STATUS_ACTIVE).SetURL("url_in_range")
	linkInRange.SetTimeString("2025-06-15 12:00:00")
	if err := store.LinkCreate(ctx, linkInRange); err != nil {
		t.Fatalf("Failed to create linkInRange: %v", err)
	}

	linkAfter := NewLink().SetTitle("After").SetFeedID("feedA").SetStatus(LINK_STATUS_ACTIVE).SetURL("url_after")
	linkAfter.SetTimeString("2026-08-31 10:00:00")
	if err := store.LinkCreate(ctx, linkAfter); err != nil {
		t.Fatalf("Failed to create linkAfter: %v", err)
	}

	// Query with TimeGte=2025-01-01 AND TimeLte=2025-12-31 — should only return linkInRange
	links, err := store.LinkList(ctx, LinkQuery().
		SetTimeGte("2025-01-01 00:00:00").
		SetTimeLte("2025-12-31 23:59:59").
		SetLimit(10))
	if err != nil {
		t.Fatalf("LinkList with TimeRange failed: %v", err)
	}
	if len(links) != 1 {
		t.Fatalf("expected 1 link in time range, got %d", len(links))
	}
	if links[0].ID() != linkInRange.ID() {
		t.Errorf("expected linkInRange ID, got %s", links[0].ID())
	}
}

// TestStoreLinkList_TimeGteNoMatch verifies that TimeGte filter returns
// empty when no links match the time range.
func TestStoreLinkList_TimeGteNoMatch(t *testing.T) {
	db := initDB(":memory:")
	defer db.Close()
	store := createTestStore(t, db, "feed_time_gte_nomatch", "link_time_gte_nomatch")
	ctx := context.Background()

	link := NewLink().SetTitle("Old").SetFeedID("feedA").SetStatus(LINK_STATUS_ACTIVE).SetURL("url1")
	link.SetTimeString("2020-01-01 10:00:00")
	if err := store.LinkCreate(ctx, link); err != nil {
		t.Fatalf("Failed to create link: %v", err)
	}

	// Query with TimeGte in the future — should return 0 links
	links, err := store.LinkList(ctx, LinkQuery().
		SetTimeGte("2030-01-01 00:00:00").
		SetLimit(10))
	if err != nil {
		t.Fatalf("LinkList with future TimeGte failed: %v", err)
	}
	if len(links) != 0 {
		t.Errorf("expected 0 links with future TimeGte, got %d", len(links))
	}
}

// TestStoreLinkList_TimeFilterWithOtherFilters verifies that Time filters
// work correctly in combination with other filters (FeedID, Status).
func TestStoreLinkList_TimeFilterWithOtherFilters(t *testing.T) {
	db := initDB(":memory:")
	defer db.Close()
	store := createTestStore(t, db, "feed_time_combo", "link_time_combo")
	ctx := context.Background()

	// feedA, active, old time
	link1 := NewLink().SetTitle("A-Old").SetFeedID("feedA").SetStatus(LINK_STATUS_ACTIVE).SetURL("url1")
	link1.SetTimeString("2020-01-01 10:00:00")
	if err := store.LinkCreate(ctx, link1); err != nil {
		t.Fatalf("Failed to create link1: %v", err)
	}

	// feedA, active, recent time
	link2 := NewLink().SetTitle("A-Recent").SetFeedID("feedA").SetStatus(LINK_STATUS_ACTIVE).SetURL("url2")
	link2.SetTimeString("2026-08-31 10:00:00")
	if err := store.LinkCreate(ctx, link2); err != nil {
		t.Fatalf("Failed to create link2: %v", err)
	}

	// feedB, active, recent time
	link3 := NewLink().SetTitle("B-Recent").SetFeedID("feedB").SetStatus(LINK_STATUS_ACTIVE).SetURL("url3")
	link3.SetTimeString("2026-08-31 10:00:00")
	if err := store.LinkCreate(ctx, link3); err != nil {
		t.Fatalf("Failed to create link3: %v", err)
	}

	// Query: feedA + TimeGte=2026-01-01 — should only return link2
	links, err := store.LinkList(ctx, LinkQuery().
		SetFeedID("feedA").
		SetStatus(LINK_STATUS_ACTIVE).
		SetTimeGte("2026-01-01 00:00:00").
		SetLimit(10))
	if err != nil {
		t.Fatalf("LinkList with combo filters failed: %v", err)
	}
	if len(links) != 1 {
		t.Fatalf("expected 1 link with combo filters, got %d", len(links))
	}
	if links[0].ID() != link2.ID() {
		t.Errorf("expected link2 ID, got %s", links[0].ID())
	}
}

// TestStoreLinkCount_TimeFilter verifies that LinkCount also respects
// the TimeGte/TimeLte filters.
func TestStoreLinkCount_TimeFilter(t *testing.T) {
	db := initDB(":memory:")
	defer db.Close()
	store := createTestStore(t, db, "feed_time_count", "link_time_count")
	ctx := context.Background()

	linkOld := NewLink().SetTitle("Old").SetFeedID("feedA").SetStatus(LINK_STATUS_ACTIVE).SetURL("url_old")
	linkOld.SetTimeString("2020-01-01 10:00:00")
	if err := store.LinkCreate(ctx, linkOld); err != nil {
		t.Fatalf("Failed to create linkOld: %v", err)
	}

	linkRecent := NewLink().SetTitle("Recent").SetFeedID("feedA").SetStatus(LINK_STATUS_ACTIVE).SetURL("url_recent")
	linkRecent.SetTimeString("2026-08-31 10:00:00")
	if err := store.LinkCreate(ctx, linkRecent); err != nil {
		t.Fatalf("Failed to create linkRecent: %v", err)
	}

	// Count with TimeGte filter — should be 1
	count, err := store.LinkCount(ctx, LinkQuery().SetTimeGte("2026-01-01 00:00:00"))
	if err != nil {
		t.Fatalf("LinkCount with TimeGte failed: %v", err)
	}
	if count != 1 {
		t.Errorf("expected count=1 with TimeGte filter, got %d", count)
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
