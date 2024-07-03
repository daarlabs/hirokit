package auth

import (
	"github.com/daarlabs/hirokit/efeqt"
)

var (
	Entity = efeqt.Entity[UserEntity]()
)

type UserEntity struct {
	efeqt.EntityBuilder
}

func (e UserEntity) Table() string {
	return "users"
}

func (e UserEntity) Alias() string {
	return "u"
}

func (e UserEntity) Fields() []efeqt.Field {
	return []efeqt.Field{
		e.Id(),
	}
}

func (e UserEntity) Id() efeqt.Field {
	return e.Field(efeqt.Id).Serial().PrimaryKey()
}
