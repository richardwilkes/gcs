package server

import (
	"net/http"

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/toolbox/xio/fs"
	"github.com/richardwilkes/toolbox/xio/network/xhttp"
)

func (s *Server) installPageRefHandlers() {
	s.mux.HandleFunc("GET /ref/{key}", s.pageRefHandler)
}

func (s *Server) pageRefHandler(w http.ResponseWriter, r *http.Request) {
	ref := gurps.GlobalSettings().PageRefs.Lookup(r.PathValue("key"))
	if ref != nil && fs.FileIsReadable(ref.Path) {
		http.ServeFile(w, r, ref.Path)
	} else {
		xhttp.ErrorStatus(w, http.StatusNotFound)
	}
}
