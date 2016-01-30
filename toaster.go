package toaster

import (
	"encoding/json"
	"fmt"
	"golang.org/x/net/context"
	"io"
	"log"
	"net/http"
)

/* 	Error is the primary way to communicate something has gone wrong when 
	requesting a particular resource. 

	When using json encoding, these errors will be serialized as:
	
	{
		"error":"<Error.Message>"
	}

	The status code will be appropriately set. 
*/
type Error struct {
	StatusCode int
	Message    string
}

/*	Toaster can validate inputs for context aware Handler functions. Using a simple 
	multiplexer like Goji in conjuction with Toaster is easy.

	root := goji.NewMux()
	root.HandleFuncC(pat.Post("/user"), toaster.ValidateInputs([]string{"name","birthday"}, userRoute))
	http.ListenAndServe(":3000", root)
*/

type HandlerFuncC func(ctx context.Context, writer http.ResponseWriter, reader *http.Request)

func ValidateInputsC(parameters []string, fn HandlerFuncC) HandlerFuncC {
	return func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		for i := range parameters {
			if inputValue := r.FormValue(parameters[i]); inputValue == "" {
				err := Error{http.StatusBadRequest, parameters[i] + " is not set"}
				ErrorHandler(err, w)
				return
			}
		}

		fn(ctx, w, r)
	}
}

func ErrorHandler(err Error, w http.ResponseWriter) {
	w.WriteHeader(err.StatusCode)
	var errorMap = make(map[string]string)
	errorMap["error"] = err.Message
	SerializeResponseToWriter(w, errorMap)
}

// NOTE(ajafri): Use this as middleware so that we set the content-type properly before the header is sent
func SetContentType(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
}

func SerializeResponseToWriter(w http.ResponseWriter, v interface{}) {
	err := json.NewEncoder(w).Encode(v)

	// TODO(ajafri): make this error visible to the user of the API with an internal server error. Right now it wont add to the response but we get a PrintLn
	if err != nil {
		log.Println(err)
		io.WriteString(w, fmt.Sprintf("%s", err))
	}
}
