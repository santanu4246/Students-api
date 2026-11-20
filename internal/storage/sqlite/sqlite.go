package sqlite

import (
	"database/sql"
	"fmt"

	config "github.com/santanu/students-api/internal"
	"github.com/santanu/students-api/internal/storage"
	"github.com/santanu/students-api/internal/types"

	_ "github.com/mattn/go-sqlite3"
)

type Sqlite struct {
	Db *sql.DB
}

func New(cfg *config.Config)(*Sqlite, error) {

	db, err:= sql.Open("sqlite3", cfg.StoragePath)

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