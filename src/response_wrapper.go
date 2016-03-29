package main

import "net/http"

// RespondRedirection return response of "302 Found""
func RespondRedirection(rw http.ResponseWriter, url string) {
    rw.Header().Add("Location", url)
    rw.WriteHeader(http.StatusFound)
}