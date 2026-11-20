package types

type Student struct {
	Id int64 
	Name string `json:"name" validate:"required"`
	Email string `json:"email" validate:"required"`
	Age int `json:"age" validate:"required"`
}

// Attendance is one student's status on one day.
// Date uses the YYYY-MM-DD format.
type Attendance struct {
	Id        int64  `json:"id"`
	StudentId int64  `json:"student_id"`
	Date      string `json:"date" validate:"required,datetime=2006-01-02"`
	Status    string `json:"status" validate:"required,oneof=present absent late excused"`
}
