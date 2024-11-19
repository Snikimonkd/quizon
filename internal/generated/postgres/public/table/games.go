//
// Code generated by go-jet DO NOT EDIT.
//
// WARNING: Changes to this file may cause incorrect behavior
// and will be lost if the code is regenerated
//

package table

import (
	"github.com/go-jet/jet/v2/postgres"
)

var Games = newGamesTable("public", "games", "")

type gamesTable struct {
	postgres.Table

	// Columns
	ID                   postgres.ColumnInteger
	CreatedAt            postgres.ColumnTimestampz
	StartTime            postgres.ColumnTimestampz
	Location             postgres.ColumnString
	Name                 postgres.ColumnString
	MainAmount           postgres.ColumnInteger
	ReserveAmount        postgres.ColumnInteger
	RegistrationOpenTime postgres.ColumnTimestampz

	AllColumns     postgres.ColumnList
	MutableColumns postgres.ColumnList
}

type GamesTable struct {
	gamesTable

	EXCLUDED gamesTable
}

// AS creates new GamesTable with assigned alias
func (a GamesTable) AS(alias string) *GamesTable {
	return newGamesTable(a.SchemaName(), a.TableName(), alias)
}

// Schema creates new GamesTable with assigned schema name
func (a GamesTable) FromSchema(schemaName string) *GamesTable {
	return newGamesTable(schemaName, a.TableName(), a.Alias())
}

// WithPrefix creates new GamesTable with assigned table prefix
func (a GamesTable) WithPrefix(prefix string) *GamesTable {
	return newGamesTable(a.SchemaName(), prefix+a.TableName(), a.TableName())
}

// WithSuffix creates new GamesTable with assigned table suffix
func (a GamesTable) WithSuffix(suffix string) *GamesTable {
	return newGamesTable(a.SchemaName(), a.TableName()+suffix, a.TableName())
}

func newGamesTable(schemaName, tableName, alias string) *GamesTable {
	return &GamesTable{
		gamesTable: newGamesTableImpl(schemaName, tableName, alias),
		EXCLUDED:   newGamesTableImpl("", "excluded", ""),
	}
}

func newGamesTableImpl(schemaName, tableName, alias string) gamesTable {
	var (
		IDColumn                   = postgres.IntegerColumn("id")
		CreatedAtColumn            = postgres.TimestampzColumn("created_at")
		StartTimeColumn            = postgres.TimestampzColumn("start_time")
		LocationColumn             = postgres.StringColumn("location")
		NameColumn                 = postgres.StringColumn("name")
		MainAmountColumn           = postgres.IntegerColumn("main_amount")
		ReserveAmountColumn        = postgres.IntegerColumn("reserve_amount")
		RegistrationOpenTimeColumn = postgres.TimestampzColumn("registration_open_time")
		allColumns                 = postgres.ColumnList{IDColumn, CreatedAtColumn, StartTimeColumn, LocationColumn, NameColumn, MainAmountColumn, ReserveAmountColumn, RegistrationOpenTimeColumn}
		mutableColumns             = postgres.ColumnList{IDColumn, CreatedAtColumn, StartTimeColumn, LocationColumn, NameColumn, MainAmountColumn, ReserveAmountColumn, RegistrationOpenTimeColumn}
	)

	return gamesTable{
		Table: postgres.NewTable(schemaName, tableName, alias, allColumns...),

		//Columns
		ID:                   IDColumn,
		CreatedAt:            CreatedAtColumn,
		StartTime:            StartTimeColumn,
		Location:             LocationColumn,
		Name:                 NameColumn,
		MainAmount:           MainAmountColumn,
		ReserveAmount:        ReserveAmountColumn,
		RegistrationOpenTime: RegistrationOpenTimeColumn,

		AllColumns:     allColumns,
		MutableColumns: mutableColumns,
	}
}