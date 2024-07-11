package user

// func (s *Server) signOutHandler(w http.ResponseWriter, r *http.Request) {
// 	cookie, err := r.Cookie(types.SessionCookieName)
// 	if err == http.ErrNoCookie {
// 		w.WriteHeader(http.StatusUnauthorized)
// 		return
// 	}
// 	if err := s.Store.DeleteSessionByID(cookie.Value); err != nil {
// 		log.Fatalf("Can not find the session with id %q - %s", cookie.Value, err)
// 	}
// 	// TODO:
// 	// 3. redirect to the sign in page.
// }
