package github

import (
	"context"
	"time"

	"github.com/shurcooL/githubv4"
)

type PageInfo struct {
	StartCursor     githubv4.String
	EndCursor       githubv4.String
	HasPreviousPage githubv4.Boolean
	HasNextPage     githubv4.Boolean
}

var EmptyPageInfo = PageInfo{}

type Query struct {
	client     *GithubClient
	target     interface{}
	variables  map[string]interface{}
	cursorName string
}

func (s *GithubClient) Query(target interface{}) *Query {
	return &Query{
		client:    s,
		target:    target,
		variables: make(map[string]interface{}),
	}
}

func (s *Query) Str(name string, value string) *Query {
	s.variables[name] = githubv4.String(value)
	return s
}

func (s *Query) NStr(name string) *Query {
	s.variables[name] = (*githubv4.String)(nil)
	return s
}

func (s *Query) Cursor(name string) *Query {
	s.cursorName = name
	s.variables[name] = (*githubv4.String)(nil)
	return s
}

func (s *Query) Time(name string, value time.Time) *Query {
	s.variables[name] = githubv4.DateTime{Time: value}
	return s
}

func (s *Query) Int(name string, value int) *Query {
	s.variables[name] = githubv4.Int(value)
	return s
}

func (s *Query) Bool(name string, value bool) *Query {
	s.variables[name] = githubv4.Boolean(value)
	return s
}

func (s *Query) Run(handler func() PageInfo) error {
	for {
		if err := s.client.v4Client.Query(context.Background(), s.target, s.variables); err != nil {
			return err
		}
		pageInfo := handler()
		if !pageInfo.HasNextPage {
			break
		}
		s.variables[s.cursorName] = pageInfo.EndCursor
	}
	return nil
}
