package feedstore

import (
	"context"
	"testing"
)

func TestStoreLinkLanguageAndMetas(t *testing.T) {
	db := initDB(":memory:")
	defer db.Close()
	store := createTestStore(t, db, "feed_link_lang_metas", "link_link_lang_metas")
	ctx := context.Background()

	feed := NewFeed().SetName("Test Feed")
	if err := store.FeedCreate(ctx, feed); err != nil {
		t.Fatalf("FeedCreate failed: %v", err)
	}

	// 1. Create a link with language and metas
	link := NewLink().
		SetFeedID(feed.GetID()).
		SetTitle("German Article").
		SetURL("https://example.com/de/article").
		SetLanguage("de")

	if err := link.SetMeta("original_title", "Deutscher Artikel"); err != nil {
		t.Fatalf("SetMeta failed: %v", err)
	}
	if err := link.SetMeta("original_language", "de"); err != nil {
		t.Fatalf("SetMeta original_language failed: %v", err)
	}

	if err := store.LinkCreate(ctx, link); err != nil {
		t.Fatalf("LinkCreate failed: %v", err)
	}

	// 2. Read back and verify round-trip
	found, err := store.LinkFindByID(ctx, link.GetID())
	if err != nil {
		t.Fatalf("LinkFindByID failed: %v", err)
	}
	if found == nil {
		t.Fatal("link not found after create")
	}

	if found.GetLanguage() != "de" {
		t.Errorf("expected language 'de', got '%s'", found.GetLanguage())
	}

	origTitle, err := found.GetMeta("original_title")
	if err != nil {
		t.Fatalf("GetMeta failed: %v", err)
	}
	if origTitle != "Deutscher Artikel" {
		t.Errorf("expected meta 'Deutscher Artikel', got '%s'", origTitle)
	}

	metas, err := found.GetMetas()
	if err != nil {
		t.Fatalf("GetMetas failed: %v", err)
	}
	if len(metas) != 2 {
		t.Errorf("expected 2 metas, got %d", len(metas))
	}

	// 3. Update meta and delete one
	if err := found.SetMeta("original_language", "en"); err != nil {
		t.Fatalf("SetMeta update failed: %v", err)
	}
	if err := found.DeleteMeta("original_title"); err != nil {
		t.Fatalf("DeleteMeta failed: %v", err)
	}
	if err := store.LinkUpdate(ctx, found); err != nil {
		t.Fatalf("LinkUpdate failed: %v", err)
	}

	reloaded, err := store.LinkFindByID(ctx, link.GetID())
	if err != nil {
		t.Fatalf("LinkFindByID after update failed: %v", err)
	}
	lang, err := reloaded.GetMeta("original_language")
	if err != nil {
		t.Fatalf("GetMeta after update failed: %v", err)
	}
	if lang != "en" {
		t.Errorf("expected meta original_language 'en', got '%s'", lang)
	}
	if v, _ := reloaded.GetMeta("original_title"); v != "" {
		t.Errorf("expected original_title meta to be deleted, got '%s'", v)
	}
}

func TestStoreLinkLanguageQuery(t *testing.T) {
	db := initDB(":memory:")
	defer db.Close()
	store := createTestStore(t, db, "feed_link_lang_query", "link_link_lang_query")
	ctx := context.Background()

	feed := NewFeed().SetName("Test Feed")
	if err := store.FeedCreate(ctx, feed); err != nil {
		t.Fatalf("FeedCreate failed: %v", err)
	}

	linkEn := NewLink().SetFeedID(feed.GetID()).SetTitle("English Link").SetURL("https://example.com/en").SetLanguage("en").SetStatus(LINK_STATUS_ACTIVE)
	linkDe := NewLink().SetFeedID(feed.GetID()).SetTitle("German Link").SetURL("https://example.com/de").SetLanguage("de").SetStatus(LINK_STATUS_ACTIVE)

	if err := store.LinkCreate(ctx, linkEn); err != nil {
		t.Fatalf("LinkCreate en failed: %v", err)
	}
	if err := store.LinkCreate(ctx, linkDe); err != nil {
		t.Fatalf("LinkCreate de failed: %v", err)
	}

	// Filter by language
	links, err := store.LinkList(ctx, LinkQuery().SetLanguage("de"))
	if err != nil {
		t.Fatalf("LinkList by language failed: %v", err)
	}
	if len(links) != 1 {
		t.Fatalf("expected 1 German link, got %d", len(links))
	}
	if links[0].GetID() != linkDe.GetID() {
		t.Errorf("expected German link ID '%s', got '%s'", linkDe.GetID(), links[0].GetID())
	}

	// Empty language fails validation
	_, err = store.LinkList(ctx, LinkQuery().SetLanguage(""))
	if err == nil {
		t.Error("expected validation error for empty language, got nil")
	}
}
