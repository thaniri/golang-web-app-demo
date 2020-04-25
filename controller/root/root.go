// Package root is used for requests to the application root.
package root

import (
	"fmt"
	"net/http"
)

// IndexHandler is a placeholder that just prints something when / is requested
// TODO: some homepage maybe? a link to /login?
func IndexHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Hello world!")
}
