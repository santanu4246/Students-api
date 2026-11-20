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

	// MarkAttendance records a status for a student on a date.
	// Marking the same student and date again overwrites the earlier status.
	MarkAttendance(studentId int64, date string, status string) error
	// GetAttendance lists a student's records, oldest first.
	// An empty from or to means no limit on that side.
	GetAttendance(studentId int64, from string, to string) ([]types.Attendance, error)
}
