package feedstore

import (
	"context"
	"testing"
)

func TestStoreFeedLanguageAndMetas(t *testing.T) {
	db := initDB(":memory:")
	defer db.Close()
	store := createTestStore(t, db, "feed_lang_metas", "link_lang_metas")
	ctx := context.Background()

	// 1. Create a feed with language and metas
	feed := NewFeed().
		SetName("German Feed").
		SetLanguage("de")

	if err := feed.SetMeta("type", "rss"); err != nil {
		t.Fatalf("SetMeta failed: %v", err)
	}
	if err := feed.SetMeta("etag", "W/\"abc123\""); err != nil {
		t.Fatalf("SetMeta etag failed: %v", err)
	}

	if err := store.FeedCreate(ctx, feed); err != nil {
		t.Fatalf("FeedCreate failed: %v", err)
	}

	// 2. Read back and verify round-trip
	found, err := store.FeedFindByID(ctx, feed.GetID())
	if err != nil {
		t.Fatalf("FeedFindByID failed: %v", err)
	}
	if found == nil {
		t.Fatal("feed not found after create")
	}

	if found.GetLanguage() != "de" {
		t.Errorf("expected language 'de', got '%s'", found.GetLanguage())
	}

	etag, err := found.GetMeta("etag")
	if err != nil {
		t.Fatalf("Meta failed: %v", err)
	}
	if etag != "W/\"abc123\"" {
		t.Errorf("expected etag meta 'W/\"abc123\"', got '%s'", etag)
	}

	metas, err := found.GetMetas()
	if err != nil {
		t.Fatalf("Metas failed: %v", err)
	}
	if len(metas) != 2 {
		t.Errorf("expected 2 metas, got %d", len(metas))
	}

	// 3. Update meta and delete one
	if err := found.SetMeta("type", "api"); err != nil {
		t.Fatalf("SetMeta update failed: %v", err)
	}
	if err := found.DeleteMeta("etag"); err != nil {
		t.Fatalf("DeleteMeta failed: %v", err)
	}
	if err := store.FeedUpdate(ctx, found); err != nil {
		t.Fatalf("FeedUpdate failed: %v", err)
	}

	reloaded, err := store.FeedFindByID(ctx, feed.GetID())
	if err != nil {
		t.Fatalf("FeedFindByID after update failed: %v", err)
	}
	typ, err := reloaded.GetMeta("type")
	if err != nil {
		t.Fatalf("Meta after update failed: %v", err)
	}
	if typ != "api" {
		t.Errorf("expected meta type 'api', got '%s'", typ)
	}
	if _, err := reloaded.GetMeta("etag"); err != nil {
		t.Fatalf("Meta after delete failed: %v", err)
	}
	if v, _ := reloaded.GetMeta("etag"); v != "" {
		t.Errorf("expected etag meta to be deleted, got '%s'", v)
	}
}

func TestStoreFeedLanguageQuery(t *testing.T) {
	db := initDB(":memory:")
	defer db.Close()
	store := createTestStore(t, db, "feed_lang_query", "link_lang_query")
	ctx := context.Background()

	feedEn := NewFeed().SetName("English Feed").SetLanguage("en").SetStatus(FEED_STATUS_ACTIVE)
	feedDe := NewFeed().SetName("German Feed").SetLanguage("de").SetStatus(FEED_STATUS_ACTIVE)

	if err := store.FeedCreate(ctx, feedEn); err != nil {
		t.Fatalf("FeedCreate en failed: %v", err)
	}
	if err := store.FeedCreate(ctx, feedDe); err != nil {
		t.Fatalf("FeedCreate de failed: %v", err)
	}

	// Filter by language
	feeds, err := store.FeedList(ctx, FeedQuery().SetLanguage("de"))
	if err != nil {
		t.Fatalf("FeedList by language failed: %v", err)
	}
	if len(feeds) != 1 {
		t.Fatalf("expected 1 German feed, got %d", len(feeds))
	}
	if feeds[0].GetID() != feedDe.GetID() {
		t.Errorf("expected German feed ID '%s', got '%s'", feedDe.GetID(), feeds[0].GetID())
	}

	// Empty language fails validation
	_, err = store.FeedList(ctx, FeedQuery().SetLanguage(""))
	if err == nil {
		t.Error("expected validation error for empty language, got nil")
	}
}
