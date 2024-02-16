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
	"slices"
	"strings"

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/server/sheet"
	"github.com/richardwilkes/gcs/v5/server/websettings"
	"github.com/richardwilkes/toolbox/collection/dict"
	"github.com/richardwilkes/toolbox/txt"
	"github.com/richardwilkes/toolbox/xio/network/xhttp"
)

// Dir is a directory listing.
type Dir struct {
	Name  string   `json:"name"`
	Files []string `json:"files,omitempty"`
	Dirs  []*Dir   `json:"dirs,omitempty"`
}

func (s *Server) sheetsHandler(w http.ResponseWriter, r *http.Request) {
	_, userName, ok := sessionFromRequest(r)
	if !ok {
		xhttp.ErrorStatus(w, http.StatusUnauthorized)
		return
	}
	rsp := make([]*Dir, 0)
	list := gurps.GlobalSettings().WebServer.AccessList(userName)
	keys := dict.Keys(list)
	slices.SortFunc(keys, func(a, b string) int { return txt.NaturalCmp(a, b, true) })
	for _, k := range keys {
		one := list[k]
		m := make(map[string]*Dir)
		m[k] = &Dir{Name: k}
		if err := fs.WalkDir(os.DirFS(one.Dir), ".", func(p string, d fs.DirEntry, err error) error {
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
			slog.Warn("error walking directory tree", "dir", one.Dir, "user", userName, "error", err)
		}
		rsp = append(rsp, m[k])
	}
	JSONResponse(w, http.StatusOK, prune(rsp))
}

func prune(dirs []*Dir) []*Dir {
	for i := len(dirs) - 1; i >= 0; i-- {
		d := dirs[i]
		d.Dirs = prune(d.Dirs)
		if len(d.Dirs) == 0 && len(d.Files) == 0 {
			dirs = append(dirs[:i], dirs[i+1:]...)
		}
	}
	return dirs
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
	s.sheetsLock.RLock()
	var entity *gurps.Entity
	entity, ok = s.sheets[p]
	s.sheetsLock.RUnlock()
	if !ok {
		loadedEntity, err := gurps.NewEntityFromFile(os.DirFS(access.Dir), parts[1])
		if err != nil {
			slog.Error("error loading sheet", "path", p, "error", err)
			xhttp.ErrorStatus(w, http.StatusNotFound)
			return
		}
		s.sheetsLock.Lock()
		if entity, ok = s.sheets[p]; !ok {
			entity = loadedEntity
			s.sheets[p] = entity
		}
		s.sheetsLock.Unlock()
	}
	JSONResponse(w, http.StatusOK, sheet.NewSheetFromEntity(entity))
}
