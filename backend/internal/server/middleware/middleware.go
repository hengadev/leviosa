package middleware

import "net/http"

type Handlerfunc func(w http.ResponseWriter, r *http.Request)

type Middleware func(Handlerfunc) Handlerfunc
