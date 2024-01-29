/*
 * Copyright ©1998-2024 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, version 2.0. If a copy of the MPL was not distributed with
 * this file, You can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as
 * defined by the Mozilla Public License, version 2.0.
 */

package server

import (
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/server/websettings"
	"github.com/richardwilkes/toolbox/xio"
	"github.com/richardwilkes/toolbox/xio/network/xhttp"
)

func (s *Server) sheetsHandler(w http.ResponseWriter, r *http.Request) {
	_, userName, ok := sessionFromRequest(r)
	if !ok {
		xhttp.ErrorStatus(w, http.StatusUnauthorized)
		return
	}
	type Dir struct {
		Name  string   `json:"name"`
		Files []string `json:"files,omitempty"`
		Dirs  []*Dir   `json:"dirs,omitempty"`
	}
	rsp := make([]*Dir, 0)
	for k, v := range gurps.GlobalSettings().WebServer.AccessList(userName) {
		m := make(map[string]*Dir)
		m[k] = &Dir{Name: k}
		if err := fs.WalkDir(os.DirFS(v.Dir), ".", func(p string, d fs.DirEntry, err error) error {
			if err != nil {
				return nil
			}
			name := d.Name()
			if strings.HasPrefix(name, ".") {
				if name != "." && d.IsDir() {
					return fs.SkipDir
				}
				return nil
			}
			full := path.Clean(k + "/" + p)
			parent := m[path.Dir(full)]
			if d.IsDir() {
				dir := &Dir{Name: name}
				parent.Dirs = append(parent.Dirs, dir)
				m[full] = dir
				return nil
			}
			if strings.EqualFold(gurps.SheetExt, path.Ext(name)) {
				parent.Files = append(parent.Files, name)
			}
			return nil
		}); err != nil {
			slog.Warn("error walking directory tree", "dir", v.Dir, "user", userName, "error", err)
		}
		rsp = append(rsp, m[k])
	}
	JSONResponse(w, http.StatusOK, rsp)
}

func (s *Server) sheetHandler(w http.ResponseWriter, r *http.Request) {
	_, userName, ok := sessionFromRequest(r)
	if !ok {
		xhttp.ErrorStatus(w, http.StatusUnauthorized)
		return
	}
	accessList := gurps.GlobalSettings().WebServer.AccessList(userName)
	parts := strings.SplitN(strings.TrimPrefix(r.URL.Path, "/api/sheet/"), "/", 2)
	if len(parts) != 2 {
		xhttp.ErrorStatus(w, http.StatusBadRequest)
		return
	}
	var access websettings.Access
	if access, ok = accessList[parts[0]]; !ok {
		xhttp.ErrorStatus(w, http.StatusNotFound)
		return
	}
	parts[1] = filepath.Clean(parts[1])
	if filepath.IsAbs(parts[1]) {
		xhttp.ErrorStatus(w, http.StatusBadRequest)
		return
	}
	p := filepath.Join(access.Dir, parts[1])
	f, err := os.Open(p)
	if err != nil {
		xhttp.ErrorStatus(w, http.StatusNotFound)
		return
	}
	defer xio.CloseIgnoringErrors(f)
	var fi os.FileInfo
	if fi, err = f.Stat(); err != nil {
		xhttp.ErrorStatus(w, http.StatusNotFound)
		return
	}
	http.ServeContent(w, r, p, fi.ModTime(), f)
}
