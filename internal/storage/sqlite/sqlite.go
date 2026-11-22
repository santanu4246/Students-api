package sqlite

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

	config "github.com/santanu/students-api/internal"
	"github.com/santanu/students-api/internal/storage"
	"github.com/santanu/students-api/internal/types"

	sqlite3 "github.com/mattn/go-sqlite3"
)

type Sqlite struct {
	Db *sql.DB
}

func New(cfg *config.Config)(*Sqlite, error) {

	// SQLite ignores foreign keys unless this is switched on for every connection.
	// It is needed for ON DELETE CASCADE on the attendance table.
	dsn := cfg.StoragePath
	if strings.Contains(dsn, "?") {
		dsn += "&_foreign_keys=on"
	} else {
		dsn += "?_foreign_keys=on"
	}

	db, err:= sql.Open("sqlite3", dsn)

	if err != nil {
		return nil, err
	}

	// Tables are created and changed by the files in the migrations folder.
	if err := runMigrations(db); err != nil {
		return nil, fmt.Errorf("run migrations: %w", err)
	}

	return &Sqlite{
		Db: db,
	} , nil
}

// translateError turns a "unique constraint failed" error from SQLite into
// storage.ErrDuplicateRollNumber. Other errors pass through unchanged.
func translateError(err error) error {
	var sqliteErr sqlite3.Error

	if errors.As(err, &sqliteErr) && sqliteErr.ExtendedCode == sqlite3.ErrConstraintUnique {
		return storage.ErrDuplicateRollNumber
	}

	return err
}

func (s *Sqlite) CreateStudent(name string, email string, age int, rollNumber string) (int64 , error){

	stmt, err := s.Db.Prepare("INSERT INTO students (name, email, age, roll_number) VALUES(?,?,?,?)")

	if err != nil{
		return 0, err
	}
	defer stmt.Close()

	result , err := stmt.Exec(name, email, age, rollNumber)

	if err != nil {
		return 0, translateError(err)
	}

	lastId, err := result.LastInsertId()

	if err != nil {
		return 0, nil
	}



	return lastId, nil
}

func (s *Sqlite) GetStudentById(id int64) (types.Student, error) {

	stmt, err := s.Db.Prepare("SELECT id, name, email, age, roll_number FROM students WHERE id = ? LIMIT 1")

	if err != nil {

		return types.Student{}, err
	}

	defer stmt.Close()

	var student types.Student

	// Students created before migration 3 have a NULL roll_number.
	// A plain string cannot hold NULL, so scan into sql.NullString first.
	var rollNumber sql.NullString

	err = stmt.QueryRow(id).Scan(&student.Id, &student.Name, &student.Email, &student.Age, &rollNumber)

	if err != nil {
		if err == sql.ErrNoRows {
			return types.Student{}, storage.ErrNotFound
		}
		return types.Student{}, fmt.Errorf("query error: %w", err)
	}

	student.RollNumber = rollNumber.String

	return student, nil
}

func (s *Sqlite) UpdateStudent(id int64, name string, email string, age int, rollNumber string) error {

	stmt, err := s.Db.Prepare("UPDATE students SET name = ?, email = ?, age = ?, roll_number = ? WHERE id = ?")

	if err != nil {
		return err
	}
	defer stmt.Close()

	result, err := stmt.Exec(name, email, age, rollNumber, id)

	if err != nil {
		return translateError(err)
	}

	rows, err := result.RowsAffected()

	if err != nil {
		return err
	}

	if rows == 0 {
		return storage.ErrNotFound
	}

	return nil
}

func (s *Sqlite) DeleteStudent(id int64) error {

	stmt, err := s.Db.Prepare("DELETE FROM students WHERE id = ?")

	if err != nil {
		return err
	}
	defer stmt.Close()

	result, err := stmt.Exec(id)

	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()

	if err != nil {
		return err
	}

	if rows == 0 {
		return storage.ErrNotFound
	}

	return nil
}

func (s *Sqlite) MarkAttendance(studentId int64, date string, status string) error {

	// UNIQUE (student_id, date) means a second mark for the same day conflicts.
	// ON CONFLICT turns that into an update of the status (an "upsert").
	stmt, err := s.Db.Prepare(`INSERT INTO attendance (student_id, date, status) VALUES (?,?,?)
	ON CONFLICT(student_id, date) DO UPDATE SET status = excluded.status`)

	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(studentId, date, status)

	return err
}

func (s *Sqlite) GetAttendance(studentId int64, from string, to string) ([]types.Attendance, error) {

	// (? = '' OR ...) switches a date filter off when its value is empty.
	rows, err := s.Db.Query(`SELECT id, student_id, date, status FROM attendance
	WHERE student_id = ?
	AND (? = '' OR date >= ?)
	AND (? = '' OR date <= ?)
	ORDER BY date`, studentId, from, from, to, to)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Start with an empty slice (not nil) so JSON shows [] instead of null.
	records := []types.Attendance{}

	for rows.Next() {
		var a types.Attendance

		err := rows.Scan(&a.Id, &a.StudentId, &a.Date, &a.Status)

		if err != nil {
			return nil, err
		}

		records = append(records, a)
	}

	// rows.Next() also stops on an error, so check for one afterwards.
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return records, nil
}
