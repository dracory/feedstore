# Feed Store

Feed Store is a Go library designed to efficiently store and manage feed data
(like RSS or Atom) using SQL database backends. It provides a simple interface
for adding, retrieving, and querying feed items, supporting SQLite, PostgreSQL,
and MySQL.

## License

This project is dual-licensed under the following terms:

- For non-commercial use, you may choose either the GNU Affero General Public License v3.0 (AGPLv3) *or* a separate commercial license (see below). You can find a copy of the AGPLv3 at: https://www.gnu.org/licenses/agpl-3.0.txt

- For commercial use, a separate commercial license is required. Commercial licenses are available for various use cases. Please contact me via my [contact page](https://lesichkov.co.uk/contact) to obtain a commercial license.

## Description

Feed Store provides a persistent storage solution for feed items parsed from
sources like RSS or Atom feeds. It abstracts the underlying database interactions,
allowing you to work with feed data through a consistent Go API.

**Key Features:**

*   **SQL Backends:** Supports SQLite, PostgreSQL, and MySQL databases.
*   **Simple API:** Offers straightforward methods for adding, retrieving, and querying feed items.
*   **Persistence:** Stores feed data durably in your chosen SQL database.

This library focuses *solely* on the storage aspect. Fetching and parsing feeds
should be handled by other libraries (e.g., `gofeed`).

## Installation

```bash
go get github.com/dracory/feedstore
```

## Examples

**1. Store Initialization:**

```go
    // --- Initialize Feed Store ---
    // Provide table names and enable automigration for this example
    store, err := feedstore.NewStore(feedstore.NewStoreOptions{
        DB:                 db,
        FeedTableName:      "feeds", // Choose your feed table name
        LinkTableName:      "links", // Choose your link table name
        AutomigrateEnabled: true,    // Automatically create/update tables
        // DebugEnabled:    true,    // Optional: Enable SQL logging
    })
    if err != nil {
        log.Fatalf("❌ Failed to initialize feed store: %v", err)
    }
    fmt.Println("✅ Feed store initialized (tables migrated if needed).")
```

**2. Feed Operations:**

```go
    // --- Create a Feed ---
    feed1 := feedstore.NewFeed().
        SetName("Example Blog Feed").
        SetURL("https://example.com/blog/rss").
        SetStatus(feedstore.FEED_STATUS_ACTIVE).
        SetFetchInterval("3600") // e.g., 1 hour

    err := store.FeedCreate(feed1)
    if err != nil {
        log.Printf("⚠️ Failed to create feed: %v", err)
        return
    }
    fmt.Printf("✅ Feed created successfully: ID=%s, Name=%s\n", feed1.ID(), feed1.Name())

    // --- Find Feed by ID ---
    foundFeed, err := store.FeedFindByID(feed1.ID())
    if err != nil {
        log.Printf("⚠️ Error finding feed %s: %v", feed1.ID(), err)
    } else if foundFeed == nil {
        fmt.Printf("ℹ️ Feed %s not found.\n", feed1.ID())
    } else {
        fmt.Printf("✅ Found feed by ID: %s (Name: %s)\n", foundFeed.ID(), foundFeed.Name())
    }

    // --- List Active Feeds ---
    activeFeeds, err := store.FeedList(feedstore.FeedQuery().
        SetStatus(feedstore.FEED_STATUS_ACTIVE).
        SetLimit(10)) // Limit results
    if err != nil {
        log.Printf("⚠️ Error listing active feeds: %v", err)
    } else {
        fmt.Printf("✅ Found %d active feed(s):\n", len(activeFeeds))
        for _, f := range activeFeeds {
            fmt.Printf("   - ID: %s, Name: %s, Status: %s\n", f.ID(), f.Name(), f.Status())
        }
    }

    // --- Update a Feed ---
    foundFeed.SetMemo("This feed was updated.")
    err = store.FeedUpdate(foundFeed)
    if err != nil {
        log.Printf("⚠️ Error updating feed %s: %v", foundFeed.ID(), err)
    } else {
        fmt.Printf("✅ Feed %s updated successfully.\n", foundFeed.ID())
        // Verify update
        updatedFeed, _ := store.FeedFindByID(foundFeed.ID())
        if updatedFeed != nil {
            fmt.Printf("   Updated Memo: %s\n", updatedFeed.Memo())
        }
    }

    // --- Soft Delete a Feed ---
    // Create another feed to delete
    feedToDelete := feedstore.NewFeed().SetName("Temporary Feed").SetURL("http://temp.com/rss")
    _ = store.FeedCreate(feedToDelete)
    fmt.Printf("   Created temporary feed: %s\n", feedToDelete.ID())

    err = store.FeedSoftDeleteByID(feedToDelete.ID())
    if err != nil {
        log.Printf("⚠️ Error soft deleting feed %s: %v", feedToDelete.ID(), err)
    } else {
        fmt.Printf("✅ Feed %s soft deleted.\n", feedToDelete.ID())
    }
```

**3. Link Operations:**

```go

// --- Create Links ---
link1 := feedstore.NewLink().
    SetFeedID(feedID).
    SetTitle("First Blog Post").
    SetURL("https://example.com/blog/post1").
    SetStatus(feedstore.LINK_STATUS_ACTIVE)


err = store.LinkCreate(link1)
if err != nil {
    log.Printf("⚠️ Failed to create link1: %v", err)
} else {
    fmt.Printf("✅ Link created successfully: ID=%s, Title=%s\n", link1.ID(), link1.Title())
}

// --- Find Link by ID ---
foundLink, err := store.LinkFindByID(link1.ID())
if err != nil {
    log.Printf("⚠️ Error finding link %s: %v", link1.ID(), err)
} else if foundLink == nil {
    fmt.Printf("ℹ️ Link %s not found.\n", link1.ID())
} else {
    fmt.Printf("✅ Found link by ID: %s (Title: %s)\n", foundLink.ID(), foundLink.Title())
}

// --- List Links for a Specific Feed ---
feedLinks, err := store.LinkList(feedstore.LinkQuery().
    SetFeedID(feedID). // Filter by the feed's ID
    SetLimit(10))
if err != nil {
    log.Printf("⚠️ Error listing links for feed %s: %v", feedID, err)
} else {
    fmt.Printf("✅ Found %d link(s) for feed %s:\n", len(feedLinks), feedID)
    for _, lnk := range feedLinks {
        fmt.Printf("   - ID: %s, Title: %s, URL: %s\n", lnk.ID(), lnk.Title(), lnk.URL())
    }
}

// --- Update a Link ---
foundLink.SetDescription("Added a description.")
err = store.LinkUpdate(foundLink)
if err != nil {
    log.Printf("⚠️ Error updating link %s: %v", foundLink.ID(), err)
} else {
    fmt.Printf("✅ Link %s updated successfully.\n", foundLink.ID())
    // Verify update
    updatedLink, _ := store.LinkFindByID(foundLink.ID())
    if updatedLink != nil {
        fmt.Printf("   Updated Description: %s\n", updatedLink.Description())
    }
}

// --- Soft Delete a Link ---
err = store.LinkSoftDeleteByID(link2.ID())
if err != nil {
    log.Printf("⚠️ Error soft deleting link %s: %v", link2.ID(), err)
} else {
    fmt.Printf("✅ Link %s soft deleted.\n", link2.ID())
}
```