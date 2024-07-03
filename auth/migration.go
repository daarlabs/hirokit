package auth

import (
	"fmt"
	
	"github.com/daarlabs/hirokit/esquel"
)

var (
	pgUserFields = []esquel.Field{
		{Name: esquel.Id, Props: "serial primary key"},
		{Name: UserActive, Props: "bool not null default false"},
		{Name: UserRoles, Props: "varchar[]"},
		{Name: UserEmail, Props: "varchar(255) not null"},
		{Name: UserPassword, Props: "varchar(128) not null"},
		{Name: UserTfa, Props: "bool not null default false"},
		{Name: UserTfaSecret, Props: "varchar(255)"},
		{Name: UserTfaCodes, Props: "varchar(255)"},
		{Name: UserTfaUrl, Props: "varchar(255)"},
		{Name: esquel.Vectors, Props: "tsvector not null default ''"},
		{Name: UserLastActivity, Props: "timestamp not null default current_timestamp"},
		{Name: esquel.CreatedAt, Props: "timestamp not null default current_timestamp"},
		{Name: esquel.UpdatedAt, Props: "timestamp not null default current_timestamp"},
	}
)

func CreateTable(db *esquel.DB) error {
	fields := make([]esquel.Field, 0)
	switch db.DriverName() {
	case esquel.Postgres:
		for _, f := range pgUserFields {
			fields = append(fields, f)
		}
	}
	return db.Q(
		fmt.Sprintf(
			`CREATE TABLE IF NOT EXISTS %s (%s)`,
			usersTable,
			esquel.CreateTableStructure(fields),
		),
	).Exec()
}

func MustCreateTable(q *esquel.DB) {
	err := CreateTable(q)
	if err != nil {
		panic(err)
	}
}

func DropTable(q *esquel.DB) error {
	return q.Q(fmt.Sprintf(`DROP TABLE IF EXISTS %s CASCADE`, usersTable)).Exec()
}

func MustDropTable(q *esquel.DB) {
	err := DropTable(q)
	if err != nil {
		panic(err)
	}
}
