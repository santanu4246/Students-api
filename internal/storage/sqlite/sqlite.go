package sqlite

import (
	"database/sql"
	"fmt"
	"strings"

	config "github.com/santanu/students-api/internal"
	"github.com/santanu/students-api/internal/storage"
	"github.com/santanu/students-api/internal/types"

	_ "github.com/mattn/go-sqlite3"
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

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS students (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	name TEXT,
	email TEXT,
	age INTEGER
	)`)

	if err != nil {
		return nil, err
	}

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS attendance (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	student_id INTEGER NOT NULL,
	date TEXT NOT NULL,
	status TEXT NOT NULL,
	FOREIGN KEY (student_id) REFERENCES students(id) ON DELETE CASCADE,
	UNIQUE (student_id, date)
	)`)

	if err != nil {
		return nil, err
	}

	return &Sqlite{
		Db: db,
	} , nil
}

func (s *Sqlite) CreateStudent(name string, email string, age int) (int64 , error){

	stmt, err := s.Db.Prepare("INSERT INTO students (name, email, age) VALUES(?,?,?)")

	if err != nil{
		return 0, err
	}
	defer stmt.Close()

	result , err := stmt.Exec(name, email, age)

	if err != nil {
		return 0, err
	}

	lastId, err := result.LastInsertId()

	if err != nil {
		return 0, nil
	}



	return lastId, nil
}

func (s *Sqlite) GetStudentById(id int64) (types.Student, error) {

	stmt, err := s.Db.Prepare("SELECT id, name, email, age FROM students WHERE id = ? LIMIT 1")

	if err != nil {

		return types.Student{}, err
	}

	defer stmt.Close()

	var student types.Student

	err = stmt.QueryRow(id).Scan(&student.Id, &student.Name, &student.Email, &student.Age)

	if err != nil {
		if err == sql.ErrNoRows {
			return types.Student{}, storage.ErrNotFound
		}
		return types.Student{}, fmt.Errorf("query error: %w", err)
	}

	return student, nil
}

func (s *Sqlite) UpdateStudent(id int64, name string, email string, age int) error {

	stmt, err := s.Db.Prepare("UPDATE students SET name = ?, email = ?, age = ? WHERE id = ?")

	if err != nil {
		return err
	}
	defer stmt.Close()

	result, err := stmt.Exec(name, email, age, id)

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
