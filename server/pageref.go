package server

import (
	"net/http"

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/toolbox/xio/network/xhttp"
)

func (s *Server) installPageRefHandlers() {
	s.mux.HandleFunc("GET /ref/{key}/{name}", s.pageRefHandler)
}

func (s *Server) pageRefHandler(w http.ResponseWriter, r *http.Request) {
	ref := gurps.GlobalSettings().PageRefs.Lookup(r.PathValue("key"))
	if ref == nil {
		xhttp.ErrorStatus(w, http.StatusNotFound)
		return
	}
	http.ServeFile(w, r, ref.Path)
}
