package storage

import (
	"errors"

	"github.com/santanu/students-api/internal/types"
)

// ErrNotFound is returned when no student matches the given id.
var ErrNotFound = errors.New("student not found")

type Storage interface {
	CreateStudent(name string, email string, age int) (int64, error)
	GetStudentById(id int64) (types.Student, error)
	UpdateStudent(id int64, name string, email string, age int) error
	DeleteStudent(id int64) error
}
