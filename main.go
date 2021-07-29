package main

import (
	"encoding/json"
	"fmt"
	"github.com/coopstools-homebrew/the-forgotten/lib"
	"github.com/pkg/errors"
	"github.com/rs/cors"
	"net/http"
	"os"
	"time"
)

var nodeHistogram = lib.BuildNodeHistogram(nil)

func main() {
	args := os.Args
	prefix := ""
	if len(args) > 2 {
		prefix = "/" + args[2]
	}

	nodeHistogram.SetupCron(15 * time.Second)
	err := setupServerWithPathPrefix(":" + args[1], prefix)
	fmt.Printf("%+v\n", errors.Wrap(err, "could not start server"))
}

func setupServerWithPathPrefix(port string, pathPrefix string) error {
	mux := http.NewServeMux()
	mux.HandleFunc(pathPrefix + "/", GetSats)
	handler := logRequestHandler(mux)
	handler = cors.New(cors.Options{
		AllowedOrigins: []string{"http://localhost:*", "https://home.coopstools.com"},
	}).Handler(handler)

	fmt.Println(port)
	return http.ListenAndServe(port, handler)
}

func GetSats(w http.ResponseWriter, r *http.Request) {
	data, err := json.Marshal(nodeHistogram.Data)
	if err != nil {
		w.WriteHeader(500)
		fmt.Printf("\nServer error serializing namespaces: %+v", errors.WithStack(err))
		fmt.Fprint(w, "Internal server error")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	count, _ := w.Write(data)
	fmt.Printf("\n%d bytes returned for GetStats", count)
}

func logRequestHandler(h http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		h.ServeHTTP(w, r)
		uri := r.URL.String()
		method := r.Method
		fmt.Printf("\n%v: %v", method, uri)
	}
	return http.HandlerFunc(fn)
}
