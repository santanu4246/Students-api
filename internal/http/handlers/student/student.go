package student

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"github.com/santanu/students-api/internal/types"
	"github.com/santanu/students-api/internal/utils/response"
)

func New() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		var student types.Student

		err:= json.NewDecoder(r.Body).Decode(&student)

		if errors.Is(err, io.EOF ){
			
			response.WriteJson(w, http.StatusBadRequest, response.GeneralError(fmt.Errorf("empty body")))
			return
			
		}

		if err != nil {

		}
		
		slog.Info("createing student")

		response.WriteJson(w, http.StatusCreated, map[string] string {"sucess" : "OK"})

		// w.Write([]byte("Welcome to students api"))
	}
}
