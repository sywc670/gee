package session

import (
	"database/sql"
	"strings"

	"github.com/sywc670/gee/geeorm/log"
)

type Session struct {
	sql     strings.Builder
	sqlvars []any
	db      *sql.DB
}

func New(db *sql.DB) *Session {
	return &Session{db: db}
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
