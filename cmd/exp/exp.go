package main

import (
	"fmt"
	"strings"

	sq "github.com/Masterminds/squirrel"
)

// Builder pattern:
type Employee struct {
	Name string
	Role string
	MinSalary int
	MaxSalary int
}

type Builder struct {
	e Employee
	err error
}

func (b *Builder) Build() (Employee, error) {
	if b.err != nil {
		return Employee{}, b.err
	}
	return b.e, nil
}

func (b *Builder) Name(name string) *Builder {
	b.e.Name = name
	return b
}

func (b *Builder) Role(role string) *Builder {
	if role == "Manager" {
		b.e.MinSalary = 20000
		b.e.MaxSalary = 60000
	}
	b.e.Role = role
	return b
}


// Functional options pattern:
type Option func(e *Employee) error

func EmployeeName(name string) Option {
	return func(e *Employee) error {
		e.Name = name
		return nil
	}
}

func EmployeeRole(role string) Option {
	return func(e *Employee) error {
		if role == "Manager" {
			e.MinSalary = 20000
			e.MaxSalary = 60000
		}
		e.Role = role
		return nil
	}
}

func NewEmployee(opts ...Option) (*Employee, error) {
	e := &Employee{}
	for _, opt := range opts {
		err := opt(e)
		if err != nil {
			return nil, err
		}
	}
	return e, nil
}


func Join(vals ...string) string {
	var sb strings.Builder
	for i, s := range vals {
		sb.WriteString(s)
		if i < len(vals) - 1 {
			sb.WriteString(", ")
		}
	}
	return sb.String()
}


func main() {
	b := &Builder{}
	e1, err := b.Name("Robert Miller").Role("Web Developer II").Build()
	if err != nil {
		panic(err)
	}

	e2, err := NewEmployee(EmployeeName("Robert Miller"), EmployeeRole("Manager"))
	if err != nil {
		panic(err)
	}

	fmt.Println("Employee 1:", e1)
	fmt.Println("Employee 2:", e2)

	// Squirrel builder version:
	sql, args, err := sq.Insert("users").Columns("name", "age").Values("Robert", 44).Values("Sharon", sq.Expr("? + 5", 44)).ToSql()
	if err != nil {
		panic(err)
	}
	fmt.Println(sql)
	fmt.Println(args)
	
	// strings := []string{"the", "quick", "brown", "fox"}
	// fmt.Println(Join(strings...))
}