package feedstore

import (
	"context"
	"testing"
)

// TestLink_ContentAuthor_GettersSetters verifies the Content and Author
// getters and setters work correctly on the link model.
func TestLink_ContentAuthor_GettersSetters(t *testing.T) {
	link := NewLink()

	// Initially empty
	if link.Content() != "" {
		t.Errorf("expected empty Content on new link, got %q", link.Content())
	}
	if link.Author() != "" {
		t.Errorf("expected empty Author on new link, got %q", link.Author())
	}

	// Set values
	link.SetContent("This is the article content")
	link.SetAuthor("John Doe")

	if link.Content() != "This is the article content" {
		t.Errorf("Content mismatch, got %q", link.Content())
	}
	if link.Author() != "John Doe" {
		t.Errorf("Author mismatch, got %q", link.Author())
	}
}

// TestLink_ContentAuthor_FluentChaining verifies that SetContent and SetAuthor
// return the link interface for fluent chaining.
func TestLink_ContentAuthor_FluentChaining(t *testing.T) {
	link := NewLink()
	result := link.SetContent("test content")
	if result != link {
		t.Error("SetContent should return the same LinkInterface for chaining")
	}
	result2 := link.SetAuthor("test author")
	if result2 != link {
		t.Error("SetAuthor should return the same LinkInterface for chaining")
	}
}

// TestLink_Data_ContainsContentAuthor verifies that the Data() map
// includes content and author fields.
func TestLink_Data_ContainsContentAuthor(t *testing.T) {
	link := NewLink()
	link.SetContent("some content")
	link.SetAuthor("some author")

	data := link.Data()

	if data[COLUMN_CONTENT] != "some content" {
		t.Errorf("Data()[COLUMN_CONTENT] mismatch, got %q", data[COLUMN_CONTENT])
	}
	if data[COLUMN_AUTHOR] != "some author" {
		t.Errorf("Data()[COLUMN_AUTHOR] mismatch, got %q", data[COLUMN_AUTHOR])
	}
}

// TestStore_LinkCreate_WithContentAuthor verifies that LinkCreate stores
// content and author in the database.
func TestStore_LinkCreate_WithContentAuthor(t *testing.T) {
	db := initDB(":memory:")
	defer db.Close()
	store := createTestStore(t, db, "feed_content_author", "link_content_author")
	ctx := context.Background()

	link := NewLink().
		SetTitle("Test Article").
		SetFeedID("feedA").
		SetStatus(LINK_STATUS_ACTIVE).
		SetURL("https://example.com/test")
	link.SetContent("Full article content here")
	link.SetAuthor("Test Author")

	if err := store.LinkCreate(ctx, link); err != nil {
		t.Fatalf("LinkCreate failed: %v", err)
	}

	// Retrieve and verify
	found, err := store.LinkFindByID(ctx, link.ID())
	if err != nil {
		t.Fatalf("LinkFindByID failed: %v", err)
	}
	if found == nil {
		t.Fatal("Link not found after creation")
	}
	if found.Content() != "Full article content here" {
		t.Errorf("expected Content to be stored, got %q", found.Content())
	}
	if found.Author() != "Test Author" {
		t.Errorf("expected Author to be stored, got %q", found.Author())
	}
}

// TestStore_LinkCreate_LongContent verifies that long content (TEXT column)
// is stored and retrieved correctly.
func TestStore_LinkCreate_LongContent(t *testing.T) {
	db := initDB(":memory:")
	defer db.Close()
	store := createTestStore(t, db, "feed_long_content", "link_long_content")
	ctx := context.Background()

	// Generate content longer than VARCHAR(1024)
	longContent := ""
	for i := 0; i < 200; i++ {
		longContent += "This is sentence " + string(rune('A'+i%26)) + " in a very long article. "
	}
	if len(longContent) <= 1024 {
		t.Fatalf("test content must be longer than 1024 chars, got %d", len(longContent))
	}

	link := NewLink().
		SetTitle("Long Content Article").
		SetFeedID("feedA").
		SetStatus(LINK_STATUS_ACTIVE).
		SetURL("https://example.com/long")
	link.SetContent(longContent)

	if err := store.LinkCreate(ctx, link); err != nil {
		t.Fatalf("LinkCreate failed: %v", err)
	}

	found, err := store.LinkFindByID(ctx, link.ID())
	if err != nil {
		t.Fatalf("LinkFindByID failed: %v", err)
	}
	if found.Content() != longContent {
		t.Errorf("expected long content to be stored intact (len=%d), got len=%d", len(longContent), len(found.Content()))
	}
}

// TestStore_LinkUpdate_WithContentAuthor verifies that LinkUpdate correctly
// updates content and author fields.
func TestStore_LinkUpdate_WithContentAuthor(t *testing.T) {
	db := initDB(":memory:")
	defer db.Close()
	store := createTestStore(t, db, "feed_update_content", "link_update_content")
	ctx := context.Background()

	link := NewLink().
		SetTitle("Original Title").
		SetFeedID("feedA").
		SetStatus(LINK_STATUS_ACTIVE).
		SetURL("https://example.com/update-test")

	if err := store.LinkCreate(ctx, link); err != nil {
		t.Fatalf("LinkCreate failed: %v", err)
	}

	// Update content and author
	link.SetContent("Updated content")
	link.SetAuthor("Updated Author")
	if err := store.LinkUpdate(ctx, link); err != nil {
		t.Fatalf("LinkUpdate failed: %v", err)
	}

	// Retrieve and verify
	found, err := store.LinkFindByID(ctx, link.ID())
	if err != nil {
		t.Fatalf("LinkFindByID failed: %v", err)
	}
	if found.Content() != "Updated content" {
		t.Errorf("expected updated Content, got %q", found.Content())
	}
	if found.Author() != "Updated Author" {
		t.Errorf("expected updated Author, got %q", found.Author())
	}
}

// TestStore_LinkList_WithContentAuthor verifies that LinkList returns
// links with content and author fields populated.
func TestStore_LinkList_WithContentAuthor(t *testing.T) {
	db := initDB(":memory:")
	defer db.Close()
	store := createTestStore(t, db, "feed_list_content", "link_list_content")
	ctx := context.Background()

	link := NewLink().
		SetTitle("List Test").
		SetFeedID("feedA").
		SetStatus(LINK_STATUS_ACTIVE).
		SetURL("https://example.com/list-test")
	link.SetContent("Content for list test")
	link.SetAuthor("Author for list test")

	if err := store.LinkCreate(ctx, link); err != nil {
		t.Fatalf("LinkCreate failed: %v", err)
	}

	links, err := store.LinkList(ctx, LinkQuery().SetFeedID("feedA").SetLimit(10))
	if err != nil {
		t.Fatalf("LinkList failed: %v", err)
	}
	if len(links) != 1 {
		t.Fatalf("expected 1 link, got %d", len(links))
	}
	if links[0].Content() != "Content for list test" {
		t.Errorf("expected Content in list result, got %q", links[0].Content())
	}
	if links[0].Author() != "Author for list test" {
		t.Errorf("expected Author in list result, got %q", links[0].Author())
	}
}

// TestNewLinkFromExistingData_WithContentAuthor verifies that
// NewLinkFromExistingData correctly loads content and author from DB data.
func TestNewLinkFromExistingData_WithContentAuthor(t *testing.T) {
	data := map[string]string{
		COLUMN_ID:          "test123",
		COLUMN_FEED_ID:     "feedA",
		COLUMN_STATUS:      LINK_STATUS_ACTIVE,
		COLUMN_TITLE:       "Test Title",
		COLUMN_DESCRIPTION: "Test Description",
		COLUMN_CONTENT:     "Test content from data map",
		COLUMN_AUTHOR:      "Test author from data map",
		COLUMN_URL:         "https://example.com/test",
		COLUMN_VIEWS:       "0",
		COLUMN_VOTES_UP:    "0",
		COLUMN_VOTES_DOWN:  "0",
		COLUMN_TIME:        "2026-08-31 10:00:00",
		COLUMN_CREATED_AT:  "2026-08-31 10:00:00",
		COLUMN_UPDATED_AT:  "2026-08-31 10:00:00",
	}

	link := NewLinkFromExistingData(data)

	if link.Content() != "Test content from data map" {
		t.Errorf("expected Content from data map, got %q", link.Content())
	}
	if link.Author() != "Test author from data map" {
		t.Errorf("expected Author from data map, got %q", link.Author())
	}
}

// TestNewLinkFromExistingData_MissingContentAuthor verifies that
// NewLinkFromExistingData handles missing content/author gracefully
// (should default to empty string, not panic).
func TestNewLinkFromExistingData_MissingContentAuthor(t *testing.T) {
	data := map[string]string{
		COLUMN_ID:      "test456",
		COLUMN_FEED_ID: "feedA",
		COLUMN_STATUS:  LINK_STATUS_ACTIVE,
		COLUMN_TITLE:   "No Content",
		COLUMN_URL:     "https://example.com/no-content",
	}

	link := NewLinkFromExistingData(data)

	if link.Content() != "" {
		t.Errorf("expected empty Content when not in data, got %q", link.Content())
	}
	if link.Author() != "" {
		t.Errorf("expected empty Author when not in data, got %q", link.Author())
	}
}
