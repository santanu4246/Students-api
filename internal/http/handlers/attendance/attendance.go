package attendance

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/go-playground/validator/v10"

	"github.com/santanu/students-api/internal/storage"
	"github.com/santanu/students-api/internal/types"
	"github.com/santanu/students-api/internal/utils/response"
)

// dateLayout is Go's way of writing the format YYYY-MM-DD.
const dateLayout = "2006-01-02"

// Mark handles POST /api/students/{id}/attendance
// Body: {"date": "2026-11-20", "status": "present"}
func Mark(store storage.Storage) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		slog.Info("marking attendance", slog.String("studentId", id))

		studentId, err := strconv.ParseInt(id, 10, 64)

		if err != nil {
			response.WriteJson(w, http.StatusBadRequest, response.GeneralError(err))
			return
		}

		var record types.Attendance

		err = json.NewDecoder(r.Body).Decode(&record)

		if errors.Is(err, io.EOF) {
			response.WriteJson(w, http.StatusBadRequest, response.GeneralError(fmt.Errorf("empty body")))
			return
		}

		if err != nil {
			response.WriteJson(w, http.StatusBadRequest, response.GeneralError(err))
			return
		}

		if err := validator.New().Struct(record); err != nil {
			validateErrs := err.(validator.ValidationErrors)
			response.WriteJson(w, http.StatusBadRequest, response.ValidationError(validateErrs))
			return
		}

		// Make sure the student exists, so a wrong id gives 404 and not a database error.
		if _, err := store.GetStudentById(studentId); err != nil {
			writeStorageError(w, err)
			return
		}

		err = store.MarkAttendance(studentId, record.Date, record.Status)

		if err != nil {
			response.WriteJson(w, http.StatusInternalServerError, response.GeneralError(err))
			return
		}

		response.WriteJson(w, http.StatusOK, map[string]string{"status": response.StatusOk})
	}
}

// List handles GET /api/students/{id}/attendance?from=YYYY-MM-DD&to=YYYY-MM-DD
// from and to are both optional.
func List(store storage.Storage) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		slog.Info("listing attendance", slog.String("studentId", id))

		studentId, err := strconv.ParseInt(id, 10, 64)

		if err != nil {
			response.WriteJson(w, http.StatusBadRequest, response.GeneralError(err))
			return
		}

		from := r.URL.Query().Get("from")
		to := r.URL.Query().Get("to")

		for _, value := range []string{from, to} {
			if value == "" {
				continue
			}
			if _, err := time.Parse(dateLayout, value); err != nil {
				response.WriteJson(w, http.StatusBadRequest, response.GeneralError(fmt.Errorf("dates must look like YYYY-MM-DD")))
				return
			}
		}

		if _, err := store.GetStudentById(studentId); err != nil {
			writeStorageError(w, err)
			return
		}

		records, err := store.GetAttendance(studentId, from, to)

		if err != nil {
			response.WriteJson(w, http.StatusInternalServerError, response.GeneralError(err))
			return
		}

		response.WriteJson(w, http.StatusOK, records)
	}
}

// writeStorageError sends 404 for "not found" and 500 for anything else.
func writeStorageError(w http.ResponseWriter, err error) {
	if errors.Is(err, storage.ErrNotFound) {
		response.WriteJson(w, http.StatusNotFound, response.GeneralError(err))
		return
	}
	response.WriteJson(w, http.StatusInternalServerError, response.GeneralError(err))
}
