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
	"github.com/richardwilkes/gcs/v5/model/jio"
	"github.com/richardwilkes/gcs/v5/server/sheet"
	"github.com/richardwilkes/gcs/v5/server/websettings"
	"github.com/richardwilkes/toolbox/collection/dict"
	"github.com/richardwilkes/toolbox/errs"
	"github.com/richardwilkes/toolbox/txt"
	"github.com/richardwilkes/toolbox/xio/network/xhttp"
)

type webEntity struct {
	ClientPath    string
	AccessPath    string
	Entity        *gurps.Entity
	OriginalCRC64 uint64
	CurrentCRC64  uint64
}

// Dir is a directory listing.
type Dir struct {
	Name  string   `json:"name"`
	Files []string `json:"files,omitempty"`
	Dirs  []*Dir   `json:"dirs,omitempty"`
}

type sheetUpdate struct {
	Kind      string
	FieldKey  string
	FieldText string
}

func (s *Server) installSheetHandlers() {
	s.mux.HandleFunc("GET /api/sheets", s.sheetsHandler)
	s.mux.HandleFunc("GET /api/sheet/{path...}", s.sheetHandler)
	s.mux.HandleFunc("POST /api/sheet/{path...}", s.updateSheetHandler)
	s.mux.HandleFunc("PUT /api/sheet/{path...}", s.saveSheetHandler)
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
	entity, access, ok := s.loadSheet(w, r)
	if !ok {
		return
	}
	response := sheet.NewSheetFromEntity(entity.Entity, entity.OriginalCRC64 != entity.CurrentCRC64, access.ReadOnly)
	CompressedJSONResponse(w, http.StatusOK, response)
}

func (s *Server) updateSheetHandler(w http.ResponseWriter, r *http.Request) {
	entity, access, ok := s.loadSheet(w, r)
	if !ok {
		return
	}
	var update sheetUpdate
	if err := jio.Load(r.Context(), r.Body, &update); err != nil {
		xhttp.ErrorStatus(w, http.StatusBadRequest)
		return
	}
	switch update.Kind {
	case "field.text":
		if err := s.updateFieldText(&entity, &update); err != nil {
			slog.Error("error updating sheet", "path", entity.ClientPath, "error", err)
			xhttp.ErrorStatus(w, http.StatusBadRequest)
			return
		}
	}
	response := sheet.NewSheetFromEntity(entity.Entity, entity.OriginalCRC64 != entity.CurrentCRC64, access.ReadOnly)
	CompressedJSONResponse(w, http.StatusOK, response)
}

func (s *Server) saveSheetHandler(w http.ResponseWriter, r *http.Request) {
	entity, access, ok := s.loadSheet(w, r)
	if !ok {
		return
	}
	if entity.OriginalCRC64 != entity.CurrentCRC64 {
		if err := entity.Entity.Save(filepath.Join(access.Dir, entity.AccessPath)); err != nil {
			slog.Error("error saving sheet", "path", entity.ClientPath, "error", err)
			xhttp.ErrorStatus(w, http.StatusInternalServerError)
			return
		}
		entity.OriginalCRC64 = entity.CurrentCRC64
		s.sheetsLock.Lock()
		s.entitiesByPath[entity.ClientPath] = entity
		s.sheetsLock.Unlock()
	}
	response := sheet.NewSheetFromEntity(entity.Entity, entity.OriginalCRC64 != entity.CurrentCRC64, access.ReadOnly)
	CompressedJSONResponse(w, http.StatusOK, response)
}

func (s *Server) updateFieldText(entity *webEntity, update *sheetUpdate) error {
	switch update.FieldKey {
	case "Identity.Name":
		update.FieldText = txt.CollapseSpaces(strings.TrimSpace(update.FieldText))
		if update.FieldText == entity.Entity.Profile.Name {
			return nil
		}
		entity.Entity.Profile.Name = update.FieldText
		entity.Entity.ModifiedOn = jio.Now()
	default:
		return errs.Newf("unknown field key: %s", update.FieldKey)
	}
	entity.CurrentCRC64 = entity.Entity.CRC64()
	s.sheetsLock.Lock()
	s.entitiesByPath[entity.ClientPath] = *entity
	s.sheetsLock.Unlock()
	return nil
}

func (s *Server) loadSheet(w http.ResponseWriter, r *http.Request) (entity webEntity, access websettings.Access, good bool) {
	_, userName, ok := sessionFromRequest(r)
	if !ok {
		xhttp.ErrorStatus(w, http.StatusUnauthorized)
		return entity, access, false
	}
	accessList := gurps.GlobalSettings().WebServer.AccessList(userName)
	parts := strings.SplitN(strings.TrimPrefix(r.URL.Path, "/api/sheet/"), "/", 2)
	if len(parts) != 2 {
		xhttp.ErrorStatus(w, http.StatusBadRequest)
		return entity, access, false
	}
	if access, ok = accessList[parts[0]]; !ok {
		xhttp.ErrorStatus(w, http.StatusNotFound)
		return entity, access, false
	}
	if access.ReadOnly {
		xhttp.ErrorStatus(w, http.StatusForbidden)
		return entity, access, false
	}
	parts[1] = filepath.Clean(parts[1])
	if filepath.IsAbs(parts[1]) {
		xhttp.ErrorStatus(w, http.StatusBadRequest)
		return entity, access, false
	}
	entityPath := filepath.Join(access.Dir, parts[1])
	s.sheetsLock.Lock()
	entity, ok = s.entitiesByPath[entityPath]
	s.sheetsLock.Unlock()
	if !ok {
		loadedEntity, err := gurps.NewEntityFromFile(os.DirFS(access.Dir), parts[1])
		if err != nil {
			slog.Error("error loading sheet", "path", entityPath, "error", err)
			xhttp.ErrorStatus(w, http.StatusNotFound)
			return entity, access, false
		}
		s.sheetsLock.Lock()
		if entity, ok = s.entitiesByPath[entityPath]; !ok {
			entity.ClientPath = entityPath
			entity.AccessPath = parts[1]
			entity.Entity = loadedEntity
			entity.OriginalCRC64 = loadedEntity.CRC64()
			entity.CurrentCRC64 = entity.OriginalCRC64
			s.entitiesByPath[entityPath] = entity
		}
		s.sheetsLock.Unlock()
	}
	return entity, access, true
}
