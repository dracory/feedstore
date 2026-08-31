package feedstore

import (
	"context"
	"testing"
)

// TestLink_Priority_GettersSetters verifies the Priority getter/setter.
func TestLink_Priority_GettersSetters(t *testing.T) {
	link := NewLink()

	// Default should be false
	if link.Priority() {
		t.Error("expected Priority=false on new link")
	}

	// Set to true
	link.SetPriority(true)
	if !link.Priority() {
		t.Error("expected Priority=true after SetPriority(true)")
	}

	// Set back to false
	link.SetPriority(false)
	if link.Priority() {
		t.Error("expected Priority=false after SetPriority(false)")
	}
}

// TestLink_Priority_FluentChaining verifies SetPriority returns the link.
func TestLink_Priority_FluentChaining(t *testing.T) {
	link := NewLink()
	result := link.SetPriority(true)
	if result != link {
		t.Error("SetPriority should return the same LinkInterface for chaining")
	}
}

// TestLink_Data_ContainsPriority verifies Data() includes priority.
func TestLink_Data_ContainsPriority(t *testing.T) {
	link := NewLink()
	link.SetPriority(true)

	data := link.Data()

	if data[COLUMN_PRIORITY] != "1" {
		t.Errorf("expected Data()[COLUMN_PRIORITY]='1', got %q", data[COLUMN_PRIORITY])
	}

	link.SetPriority(false)
	data = link.Data()
	if data[COLUMN_PRIORITY] != "0" {
		t.Errorf("expected Data()[COLUMN_PRIORITY]='0', got %q", data[COLUMN_PRIORITY])
	}
}

// TestStore_LinkCreate_WithPriority verifies LinkCreate stores priority.
func TestStore_LinkCreate_WithPriority(t *testing.T) {
	db := initDB(":memory:")
	defer db.Close()
	store := createTestStore(t, db, "feed_priority", "link_priority")
	ctx := context.Background()

	link := NewLink().
		SetTitle("Priority Article").
		SetFeedID("feedA").
		SetStatus(LINK_STATUS_ACTIVE).
		SetURL("https://example.com/priority")
	link.SetPriority(true)

	if err := store.LinkCreate(ctx, link); err != nil {
		t.Fatalf("LinkCreate failed: %v", err)
	}

	found, err := store.LinkFindByID(ctx, link.ID())
	if err != nil {
		t.Fatalf("LinkFindByID failed: %v", err)
	}
	if !found.Priority() {
		t.Error("expected Priority=true after create+find")
	}
}

// TestStore_LinkUpdate_WithPriority verifies LinkUpdate updates priority.
func TestStore_LinkUpdate_WithPriority(t *testing.T) {
	db := initDB(":memory:")
	defer db.Close()
	store := createTestStore(t, db, "feed_update_priority", "link_update_priority")
	ctx := context.Background()

	link := NewLink().
		SetTitle("Test").
		SetFeedID("feedA").
		SetStatus(LINK_STATUS_ACTIVE).
		SetURL("https://example.com/test")

	if err := store.LinkCreate(ctx, link); err != nil {
		t.Fatalf("LinkCreate failed: %v", err)
	}

	// Update priority to true
	link.SetPriority(true)
	if err := store.LinkUpdate(ctx, link); err != nil {
		t.Fatalf("LinkUpdate failed: %v", err)
	}

	found, err := store.LinkFindByID(ctx, link.ID())
	if err != nil {
		t.Fatalf("LinkFindByID failed: %v", err)
	}
	if !found.Priority() {
		t.Error("expected Priority=true after update")
	}

	// Update back to false
	link.SetPriority(false)
	if err := store.LinkUpdate(ctx, link); err != nil {
		t.Fatalf("LinkUpdate failed: %v", err)
	}

	found, err = store.LinkFindByID(ctx, link.ID())
	if err != nil {
		t.Fatalf("LinkFindByID failed: %v", err)
	}
	if found.Priority() {
		t.Error("expected Priority=false after second update")
	}
}

// TestStore_LinkList_FilterByPriority verifies LinkList filters by priority.
func TestStore_LinkList_FilterByPriority(t *testing.T) {
	db := initDB(":memory:")
	defer db.Close()
	store := createTestStore(t, db, "feed_list_priority", "link_list_priority")
	ctx := context.Background()

	// Create two links: one priority, one not
	link1 := NewLink().
		SetTitle("Normal").
		SetFeedID("feedA").
		SetStatus(LINK_STATUS_ACTIVE).
		SetURL("https://example.com/normal")

	link2 := NewLink().
		SetTitle("Important").
		SetFeedID("feedA").
		SetStatus(LINK_STATUS_ACTIVE).
		SetURL("https://example.com/important")
	link2.SetPriority(true)

	if err := store.LinkCreate(ctx, link1); err != nil {
		t.Fatalf("LinkCreate link1 failed: %v", err)
	}
	if err := store.LinkCreate(ctx, link2); err != nil {
		t.Fatalf("LinkCreate link2 failed: %v", err)
	}

	// Query priority only
	links, err := store.LinkList(ctx, LinkQuery().SetPriority("1").SetLimit(10))
	if err != nil {
		t.Fatalf("LinkList with priority filter failed: %v", err)
	}
	if len(links) != 1 {
		t.Fatalf("expected 1 priority link, got %d", len(links))
	}
	if links[0].ID() != link2.ID() {
		t.Errorf("expected link2 (priority), got %s", links[0].ID())
	}

	// Query non-priority (all without filter)
	links, err = store.LinkList(ctx, LinkQuery().SetLimit(10))
	if err != nil {
		t.Fatalf("LinkList without filter failed: %v", err)
	}
	if len(links) != 2 {
		t.Errorf("expected 2 links without filter, got %d", len(links))
	}
}

// TestNewLinkFromExistingData_WithPriority verifies loading priority from DB data.
func TestNewLinkFromExistingData_WithPriority(t *testing.T) {
	data := map[string]string{
		COLUMN_ID:          "test789",
		COLUMN_FEED_ID:     "feedA",
		COLUMN_STATUS:      LINK_STATUS_ACTIVE,
		COLUMN_TITLE:       "Priority Test",
		COLUMN_URL:         "https://example.com/test",
		COLUMN_VIEWS:       "0",
		COLUMN_VOTES_UP:    "0",
		COLUMN_VOTES_DOWN:  "0",
		COLUMN_TIME:        "2026-08-31 10:00:00",
		COLUMN_CREATED_AT:  "2026-08-31 10:00:00",
		COLUMN_UPDATED_AT:  "2026-08-31 10:00:00",
		COLUMN_PRIORITY:    "1",
	}

	link := NewLinkFromExistingData(data)
	if !link.Priority() {
		t.Error("expected Priority=true from data with priority=1")
	}

	data[COLUMN_PRIORITY] = "0"
	link = NewLinkFromExistingData(data)
	if link.Priority() {
		t.Error("expected Priority=false from data with priority=0")
	}
}

// TestLinkQuery_Priority_GettersSetters verifies the query priority filter.
func TestLinkQuery_Priority_GettersSetters(t *testing.T) {
	q := LinkQuery()

	if q.IsPrioritySet() {
		t.Error("IsPrioritySet should be false on new query")
	}
	if q.GetPriority() != "" {
		t.Errorf("GetPriority should return empty when not set, got %q", q.GetPriority())
	}

	q.SetPriority("1")
	if !q.IsPrioritySet() {
		t.Error("IsPrioritySet should be true after SetPriority")
	}
	if q.GetPriority() != "1" {
		t.Errorf("GetPriority mismatch, got %q", q.GetPriority())
	}
}

// TestLinkQuery_Priority_Validate verifies validation rejects empty priority.
func TestLinkQuery_Priority_Validate(t *testing.T) {
	q := LinkQuery()
	q.SetPriority("")
	if err := q.Validate(); err == nil {
		t.Error("Validate should fail when Priority is set to empty string")
	}

	q2 := LinkQuery()
	q2.SetPriority("1")
	if err := q2.Validate(); err != nil {
		t.Errorf("Validate should pass with priority='1', got: %v", err)
	}
}
