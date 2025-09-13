package feedstore

import (
	"github.com/dracory/sb"
)

// sqlFeedTableCreate returns a SQL string for creating the Feed table
func (st *storeImplementation) sqlFeedTableCreate() string {
	sql := sb.NewBuilder(sb.DatabaseDriverName(st.db)).
		Table(st.feedTableName).
		Column(sb.Column{
			Name:       COLUMN_ID,
			Type:       sb.COLUMN_TYPE_STRING,
			PrimaryKey: true,
			Length:     40,
		}).
		Column(sb.Column{
			Name:   COLUMN_STATUS,
			Type:   sb.COLUMN_TYPE_STRING,
			Length: 40,
		}).
		Column(sb.Column{
			Name: COLUMN_NAME,
			Type: sb.COLUMN_TYPE_STRING,
		}).
		Column(sb.Column{
			Name: COLUMN_DESCRIPTION,
			Type: sb.COLUMN_TYPE_TEXT,
		}).
		Column(sb.Column{
			Name: COLUMN_URL,
			Type: sb.COLUMN_TYPE_STRING,
		}).
		Column(sb.Column{
			Name: COLUMN_FETCH_INTERVAL,
			Type: sb.COLUMN_TYPE_INTEGER,
		}).
		Column(sb.Column{
			Name: COLUMN_LAST_FETCHED_AT,
			Type: sb.COLUMN_TYPE_DATETIME,
		}).
		Column(sb.Column{
			Name: COLUMN_MEMO,
			Type: sb.COLUMN_TYPE_TEXT,
		}).
		Column(sb.Column{
			Name: COLUMN_CREATED_AT,
			Type: sb.COLUMN_TYPE_DATETIME,
		}).
		Column(sb.Column{
			Name: COLUMN_UPDATED_AT,
			Type: sb.COLUMN_TYPE_DATETIME,
		}).
		Column(sb.Column{
			Name:    COLUMN_SOFT_DELETED_AT,
			Type:    sb.COLUMN_TYPE_DATETIME,
			Default: sb.MAX_DATETIME,
		}).
		CreateIfNotExists()

	return sql
}
