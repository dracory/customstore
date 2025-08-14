# customstore <a href="https://gitpod.io/#https://github.com/dracory/customstore" style="float:right;"><img src="https://gitpod.io/button/open-in-gitpod.svg" alt="Open in Gitpod" loading="lazy"></a>

[![Tests Status](https://github.com/dracory/customstore/actions/workflows/tests.yml/badge.svg?branch=main)](https://github.com/dracory/customstore/actions/workflows/tests.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/dracory/customstore)](https://goreportcard.com/report/github.com/dracory/customstore)
[![PkgGoDev](https://pkg.go.dev/badge/github.com/dracory/customstore)](https://pkg.go.dev/github.com/dracory/customstore)

**customstore** is a Go package that provides a flexible way to store and manage
custom records in a database table. It simplifies common database operations like
creating, retrieving, updating, and deleting records.

## Features

- **Easy Setup**: Quickly integrate with your existing database
- **Customizable Records**: Define your own record types and data structures
- **Automatic Migration**: Automatically create the necessary database table
- **CRUD Operations**: Supports standard Create, Read, Update, and Delete operations
- **Flexible Queries**: Query records based on various criteria
- **Soft Deletes**: Option to soft delete records instead of permanent deletion
- **Payload Search**: Search for records based on content within the payload

## What's New

- Multi-ID queries via `SetIDList([]string)`: efficiently fetch multiple records in a single `RecordList` call.

## Quick Start

```go
import (
    "database/sql"
    "github.com/dracory/customstore"
)

func example(db *sql.DB) error {
    store, err := customstore.NewStore(customstore.NewOptions().
        SetDB(db).
        SetTableName("custom_records").
        SetAutomigrateEnabled(true))
    if err != nil { return err }

    // List records by multiple IDs and type
    ids := []string{"rec_1", "rec_2", "rec_3"}
    recs, err := store.RecordList(
        customstore.NewRecordQuery().
            SetType("analysis_report").
            SetIDList(ids),
    )
    if err != nil { return err }

    for _, r := range recs {
        // use r.ID(), r.Type(), r.Payload(), ...
    }
    return nil
}
```

## RecordQuery Cheatsheet

- Filtering
  - `SetID(id string)`
  - `SetIDList(ids []string)`
  - `SetType(recordType string)`
  - Payload contains: `AddPayloadSearch("needle")`
  - Payload not contains: `AddPayloadSearchNot("needle")`

- Pagination and order
  - `SetLimit(n)`, `SetOffset(n)`
  - `SetOrderBy(column)` (default order is descending)

- Soft delete
  - Excluded by default
  - Include via: `SetSoftDeletedIncluded(true)`

## Notes

- `RecordList` returns only non-soft-deleted records by default.
- `SetIDList` ignores empty strings; providing an empty slice is treated as no-op (no filter).
- **Debug Mode**: Enable debug mode for detailed logging
- **Schema-less Payloads**: Store any JSON structure without altering DB schema
- **Document-store Feel on SQL**: Document flexibility with SQL power and tooling

## Document-store versatility on SQL

Customstore lets you keep your data model fluid like a document store while using a single SQL table under the hood. No migrations for shape changes—just evolve your JSON payloads.

- __Any shape__: nested objects, arrays, primitives
- __No schema changes__: add/remove fields freely
- __SQL-compatible__: keep transactions, indexes, and familiar tooling

Example: store different shapes without migrations

```go
// A user document
user := customstore.NewRecord(
    "user",
    customstore.WithPayloadMap(map[string]any{
        "name":  "Ada",
        "roles": []any{"admin", "editor"},
        "prefs": map[string]any{"theme": "dark"},
    }),
)
_ = store.RecordCreate(user)

// A product document with a very different shape
product := customstore.NewRecord(
    "product",
    customstore.WithPayloadMap(map[string]any{
        "sku":   "SKU-001",
        "price": 19.99,
        "tags":  []any{"new", "promo"},
    }),
)
_ = store.RecordCreate(product)
```

Query by payload content (driver-dependent implementation uses JSON string search):

```go
// Find active users named Ada
q := customstore.RecordQuery().
    SetType("user").
    AddPayloadSearch(`"name": "Ada"`).
    AddPayloadSearch(`"active": true`)
list, err := store.RecordList(q)
if err != nil { panic(err) }
```

## Installation

```bash
go get -u github.com/dracory/customstore
```

## Setup

```go
// Example with SQLite
db, err := sql.Open("sqlite3", "mydatabase.db")
if err != nil {
    panic(err)
}
defer db.Close()

// Initialize the store
customStore, err := customstore.NewStore(customstore.NewStoreOptions{
    DB:                 db,
    TableName:          "my_custom_records",
    AutomigrateEnabled: true,
    DebugEnabled:       false,
})

if err != nil {
    panic(err)
}
```

## Core Concepts

### Records

A Record represents a single entry in your custom data store. Each record has:

- Type: A string that categorizes the record (e.g., "user", "product", "order")
- ID: A unique identifier for the record
- Payload: A JSON-encoded string containing the record's data
- CreatedAt: A timestamp indicating when the record was created
- UpdatedAt: A timestamp indicating when the record was last updated
- DeletedAt: A timestamp indicating when the record was soft-deleted (if applicable)

### Store

The Store is the main interface for interacting with your custom data store. It provides methods for:

- Creating records
- Retrieving records by ID
- Updating records
- Deleting records (both hard and soft deletes)
- Listing records based on various criteria
- Counting records

### RecordQuery

The RecordQuery struct allows you to build complex queries to filter and retrieve records. You can specify:

- Record type
- ID
- Limit and offset for pagination
- Order by clause
- Whether to include soft-deleted records
- Payload search terms

## Usage Examples

### Creating a Record

You can now pass functional options to the constructor to initialize fields in one place.

```go
record := customstore.NewRecord(
    "person",
    customstore.WithID("person-123"),
    customstore.WithMemo("seed user"),
    customstore.WithPayloadMap(map[string]any{
        "name": "John Doe",
        "age":  30,
    }),
    customstore.WithMetas(map[string]string{
        "role": "admin",
    }),
)

if err := store.RecordCreate(record); err != nil {
    panic(err)
}
```

Or set a raw JSON payload string at construction time:

```go
record := customstore.NewRecord(
    "order",
    customstore.WithPayload(`{"id":1,"total":19.99}`),
)
```

Legacy/imperative style (still supported):

```go
record := customstore.NewRecord("person")
record.SetID("person-123")            // optional, auto-generated if not set
record.SetMemo("seed user")            // optional memo
record.SetPayloadMap(map[string]any{   // or use SetPayload with a JSON string
    "name": "John Doe",
    "age":  30,
})
record.SetMetas(map[string]string{
    "role": "admin",
})

if err := store.RecordCreate(record); err != nil {
    panic(err)
}
```

### Finding a Record by ID

```go
record, err := store.RecordFindByID("1234567890")
if err != nil {
    panic(err)
}
```

### Updating a Record

```go
record, err := store.RecordFindByID("1234567890")
if err != nil {
    panic(err)
}

record.SetPayloadMap(map[string]interface{}{
    "name": "John Doe",
    "age":  30,
})

err = store.RecordUpdate(record)
if err != nil {
    panic(err)
}
```

### Deleting a Record (Hard Delete)

```go
record, err := store.RecordFindByID("1234567890")
if err != nil {
    panic(err)
}

err = store.RecordDelete(record)
if err != nil {
    panic(err)
}
```

### Soft Deleting a Record

```go
record, err := store.RecordFindByID("1234567890")
if err != nil {
    panic(err)
}

err = store.RecordSoftDelete(record)
if err != nil {
    panic(err)
}
```

### Listing Records

```go
query := customstore.RecordQuery().SetType("person").SetLimit(10)
list, err := store.RecordList(query)
if err != nil {
    panic(err)
}
```

### Counting Records

```go
query := customstore.RecordQuery().SetType("person")
count, err := store.RecordCount(query)
if err != nil {
    panic(err)
}
```

### Payload Search

```go
query := customstore.RecordQuery().SetType("person").
    AddPayloadSearch(`"status": "active"`).
    AddPayloadSearch(`"name": "John"`)
list, err := store.RecordList(query)
if err != nil {
    panic(err)
}
```

### Soft Deleted Records

```go
query := customstore.RecordQuery().SetType("person").SetSoftDeletedIncluded(true)
list, err := store.RecordList(query)
if err != nil {
    panic(err)
}
```

## API Reference

### Store Methods

- [NewStore(options NewStoreOptions)](cci:1://file:///d:/PROJECTs/modules/customstore/store.go:18:0-38:1) - Creates a new store instance
  - options: A NewStoreOptions struct containing the database connection, table name, and other configuration options
- [AutoMigrate()](cci:1://file:///d:/PROJECTs/modules/customstore/store.go:85:0-99:1) - Automigrates (creates) the session table
- [DriverName(db *sql.DB)](cci:1://file:///d:/PROJECTs/modules/customstore/store.go:101:0-104:1) - Finds the driver name from the database
- [EnableDebug(debug bool)](cci:1://file:///d:/PROJECTs/modules/customstore/store.go:106:0-109:1) - Enables/disables the debug option
- [RecordCreate(record *Record)](cci:1://file:///d:/PROJECTs/modules/customstore/store.go:251:0-289:1) - Creates a new record
- [RecordFindByID(id string)](cci:1://file:///d:/PROJECTs/modules/customstore/store.go:332:0-355:1) - Finds a record by its ID
- [RecordUpdate(record *Record)](cci:1://file:///d:/PROJECTs/modules/customstore/store.go:424:0-468:1) - Updates an existing record
- [RecordDelete(record *Record)](cci:1://file:///d:/PROJECTs/modules/customstore/store.go:291:0-298:1) - Deletes a record
- [RecordDeleteByID(id string)](cci:1://file:///d:/PROJECTs/modules/customstore/store.go:300:0-330:1) - Deletes a record by its ID
- [RecordSoftDelete(record *Record)](cci:1://file:///d:/PROJECTs/modules/customstore/store.go:395:0-403:1) - Soft deletes a record
- [RecordSoftDeleteByID(id string)](cci:1://file:///d:/PROJECTs/modules/customstore/store.go:405:0-422:1) - Soft deletes a record by its ID
- [RecordList(query *RecordQuery)](cci:1://file:///d:/PROJECTs/modules/customstore/store.go:357:0-393:1) - Lists records based on a query
- [RecordCount(query *RecordQuery)](cci:1://file:///d:/PROJECTs/modules/customstore/store.go:203:0-249:1) - Counts records based on a query

### RecordQuery Methods

- [SetID(id string)](cci:1://file:///d:/PROJECTs/modules/customstore/record_query_interface.go:229:0-239:1) - Sets the ID to search for
- [SetType(recordType string)](cci:1://file:///d:/PROJECTs/modules/customstore/record_query_interface.go:278:0-282:1) - Sets the record type to search for
- [SetLimit(limit int)](cci:1://file:///d:/PROJECTs/modules/customstore/record_query_interface.go:258:0-262:1) - Sets the maximum number of records to return
- [SetOffset(offset int)](cci:1://file:///d:/PROJECTs/modules/customstore/record_query_interface.go:272:0-276:1) - Sets the offset for the records to return
- [SetOrderBy(orderBy string)](cci:1://file:///d:/PROJECTs/modules/customstore/record_query_interface.go:286:0-290:1) - Sets the order by clause
- [SetSoftDeletedIncluded(softDeletedIncluded bool)](cci:1://file:///d:/PROJECTs/modules/customstore/record_query_interface.go:245:0-248:1) - Sets whether to include soft deleted records
- [AddPayloadSearch(payloadSearch string)](cci:1://file:///d:/PROJECTs/modules/customstore/record_query_interface.go:284:0-290:1) - Adds a payload search term

## Contributing

Contributions are welcome! Please feel free to submit a pull request.
