package session

import (
	"database/sql"
	"strings"

	"github.com/sywc670/gee/geeorm/dialect"
	"github.com/sywc670/gee/geeorm/log"
	"github.com/sywc670/gee/geeorm/schema"
)

type Session struct {
	dialect  dialect.Dialect
	refTable *schema.Schema
	sql      strings.Builder
	sqlvars  []any
	db       *sql.DB
}

func New(db *sql.DB, dialect dialect.Dialect) *Session {
	return &Session{db: db, dialect: dialect}
}

func (s *Session) DB() *sql.DB {
	return s.db
}

func (s *Session) Clear() {
	s.sql.Reset()
	clear(s.sqlvars)
}

func (s *Session) Raw(sql string, sqlvars ...any) *Session {
	s.sql.WriteString(sql)
	s.sql.WriteString(" ")
	s.sqlvars = append(s.sqlvars, sqlvars...)
	return s
}

func (s *Session) Exec() (result sql.Result, err error) {
	log.Info(s.sql.String(), s.sqlvars)
	defer s.Clear()
	result, err = s.DB().Exec(s.sql.String(), s.sqlvars...)
	if err != nil {
		log.Error(err)
	}
	return
}

func (s *Session) QueryRows() (rows *sql.Rows, err error) {
	log.Info(s.sql.String(), s.sqlvars)
	defer s.Clear()
	rows, err = s.DB().Query(s.sql.String(), s.sqlvars...)
	if err != nil {
		log.Error(err)
	}
	return
}

func (s *Session) QueryRow() (row *sql.Row) {
	log.Info(s.sql.String(), s.sqlvars)
	defer s.Clear()
	return s.DB().QueryRow(s.sql.String(), s.sqlvars...)
}
