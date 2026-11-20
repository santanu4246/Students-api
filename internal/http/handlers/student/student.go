package student

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/go-playground/validator/v10"

	"github.com/santanu/students-api/internal/storage"
	storagepkg "github.com/santanu/students-api/internal/storage"
	"github.com/santanu/students-api/internal/types"
	"github.com/santanu/students-api/internal/utils/response"
)

func New(storage storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		var student types.Student

		err:= json.NewDecoder(r.Body).Decode(&student)

		if errors.Is(err, io.EOF ){
			
			response.WriteJson(w, http.StatusBadRequest, response.GeneralError(fmt.Errorf("empty body")))
			return
			
		}

		if err != nil {
			response.WriteJson(w, http.StatusBadRequest, response.GeneralError(err))
			return
		}

		//request validation
		if err := validator.New().Struct(student); err != nil {
			validateErrs := err.(validator.ValidationErrors)
			response.WriteJson(w, http.StatusBadRequest, response.ValidationError(validateErrs))
			return 
		}
		
		

		lastID, err := storage.CreateStudent(
			student.Name,
			student.Email,
			student.Age,
		)
		slog.Info("user created sucessfully", slog.String("userID", fmt.Sprint(lastID)))
		if err != nil {

			response.WriteJson(w, http.StatusInternalServerError, err)
			return 
		}

		response.WriteJson(w, http.StatusCreated, map[string] int64 {"id" : lastID})

		// w.Write([]byte("Welcome to students api"))
	}
}

func GetByID(storage storage.Storage) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		slog.Info("getting a student", slog.String("id", id))

		intId, err := strconv.ParseInt(id, 10, 64)

		if err != nil {
			response.WriteJson(w, http.StatusBadRequest, response.GeneralError(err))
			return 
		}
		student, err := storage.GetStudentById(intId)

		if err != nil{
			if errors.Is(err, storagepkg.ErrNotFound) {
				response.WriteJson(w, http.StatusNotFound, response.GeneralError(err))
				return
			}
			response.WriteJson(w, http.StatusInternalServerError, response.GeneralError(err) )
			return
		}

		response.WriteJson(w, http.StatusOK, student)
	}
}

func Update(storage storage.Storage) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		slog.Info("updating a student", slog.String("id", id))

		intId, err := strconv.ParseInt(id, 10, 64)

		if err != nil {
			response.WriteJson(w, http.StatusBadRequest, response.GeneralError(err))
			return
		}

		var student types.Student

		err = json.NewDecoder(r.Body).Decode(&student)

		if errors.Is(err, io.EOF) {
			response.WriteJson(w, http.StatusBadRequest, response.GeneralError(fmt.Errorf("empty body")))
			return
		}

		if err != nil {
			response.WriteJson(w, http.StatusBadRequest, response.GeneralError(err))
			return
		}

		if err := validator.New().Struct(student); err != nil {
			validateErrs := err.(validator.ValidationErrors)
			response.WriteJson(w, http.StatusBadRequest, response.ValidationError(validateErrs))
			return
		}

		err = storage.UpdateStudent(intId, student.Name, student.Email, student.Age)

		if err != nil {
			if errors.Is(err, storagepkg.ErrNotFound) {
				response.WriteJson(w, http.StatusNotFound, response.GeneralError(err))
				return
			}
			response.WriteJson(w, http.StatusInternalServerError, response.GeneralError(err))
			return
		}

		response.WriteJson(w, http.StatusOK, map[string]string{"status": response.StatusOk})
	}
}

func Delete(storage storage.Storage) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		slog.Info("deleting a student", slog.String("id", id))

		intId, err := strconv.ParseInt(id, 10, 64)

		if err != nil {
			response.WriteJson(w, http.StatusBadRequest, response.GeneralError(err))
			return
		}

		err = storage.DeleteStudent(intId)

		if err != nil {
			if errors.Is(err, storagepkg.ErrNotFound) {
				response.WriteJson(w, http.StatusNotFound, response.GeneralError(err))
				return
			}
			response.WriteJson(w, http.StatusInternalServerError, response.GeneralError(err))
			return
		}

		response.WriteJson(w, http.StatusOK, map[string]string{"status": response.StatusOk})
	}
}

