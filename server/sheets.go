// Copyright (c) 1998-2024 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package server

import (
	"encoding/base64"
	"errors"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

	"github.com/richardwilkes/gcs/v5/imgutil"
	"github.com/richardwilkes/gcs/v5/model/fxp"
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
	Kind string
	Key  string
	Data string
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
	if access.ReadOnly {
		xhttp.ErrorStatus(w, http.StatusForbidden)
		return
	}
	var update sheetUpdate
	if err := JSONFromRequest(r, &update); err != nil {
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
	case "field.binary":
		if err := s.updateFieldBinary(&entity, &update); err != nil {
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
	if access.ReadOnly {
		xhttp.ErrorStatus(w, http.StatusForbidden)
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

var errInvalid = errors.New("invalid input")

func (s *Server) updateFieldText(entity *webEntity, update *sheetUpdate) error {
	var stringFieldPtr *string
	var lengthFieldPtr *fxp.Length
	var weightFieldPtr *fxp.Weight
	var fxpIntSetter func(fxp.Int) error
	var intSetter func(int) error
	switch update.Key {
	case "Identity.Name":
		stringFieldPtr = &entity.Entity.Profile.Name
	case "Identity.Title":
		stringFieldPtr = &entity.Entity.Profile.Title
	case "Identity.Organization":
		stringFieldPtr = &entity.Entity.Profile.Organization
	case "Misc.Player":
		stringFieldPtr = &entity.Entity.Profile.PlayerName
	case "Description.Gender":
		stringFieldPtr = &entity.Entity.Profile.Gender
	case "Description.Age":
		stringFieldPtr = &entity.Entity.Profile.Age
	case "Description.Birthday":
		stringFieldPtr = &entity.Entity.Profile.Birthday
	case "Description.Religion":
		stringFieldPtr = &entity.Entity.Profile.Religion
	case "Description.Height":
		lengthFieldPtr = &entity.Entity.Profile.Height
	case "Description.Weight":
		weightFieldPtr = &entity.Entity.Profile.Weight
	case "Description.SizeModifier":
		intSetter = func(v int) error {
			if v < -99 || v > 99 {
				return errInvalid
			}
			entity.Entity.Profile.SetAdjustedSizeModifier(v)
			return nil
		}
	case "Description.TechLevel":
		stringFieldPtr = &entity.Entity.Profile.TechLevel
	case "Description.Hair":
		stringFieldPtr = &entity.Entity.Profile.Hair
	case "Description.Eyes":
		stringFieldPtr = &entity.Entity.Profile.Eyes
	case "Description.Skin":
		stringFieldPtr = &entity.Entity.Profile.Skin
	case "Description.Hand":
		stringFieldPtr = &entity.Entity.Profile.Handedness
	default:
		if strings.HasPrefix(update.Key, "PrimaryAttributes.") {
			if attr, ok := entity.Entity.Attributes.Set[strings.TrimPrefix(update.Key, "PrimaryAttributes.")]; ok {
				if attr.AttributeDef().Primary() {
					fxpIntSetter = func(v fxp.Int) error {
						if v < fxp.Min || v > fxp.Max {
							return errInvalid
						}
						attr.SetMaximum(v)
						return nil
					}
					break
				}
			}
		}
		if strings.HasPrefix(update.Key, "SecondaryAttributes.") {
			if attr, ok := entity.Entity.Attributes.Set[strings.TrimPrefix(update.Key, "SecondaryAttributes.")]; ok {
				if attr.AttributeDef().Secondary() {
					fxpIntSetter = func(v fxp.Int) error {
						if v < fxp.Min || v > fxp.Max {
							return errInvalid
						}
						attr.SetMaximum(v)
						return nil
					}
					break
				}
			}
		}
		if strings.HasPrefix(update.Key, "PointPools.") {
			current := false
			key := strings.TrimPrefix(update.Key, "PointPools.")
			if strings.HasSuffix(key, ".Current") {
				key = strings.TrimSuffix(key, ".Current")
				current = true
			}
			if attr, ok := entity.Entity.Attributes.Set[key]; ok {
				if attr.AttributeDef().Pool() {
					fxpIntSetter = func(v fxp.Int) error {
						if current {
							if v < fxp.Min || v > attr.Maximum() {
								return errInvalid
							}
							attr.Damage = (attr.Maximum() - v).Max(0)
						} else {
							if v < fxp.Min || v > fxp.Max {
								return errInvalid
							}
							attr.SetMaximum(v)
						}
						return nil
					}
					break
				}
			}
		}
		if strings.HasPrefix(update.Key, "HitLocation.") {
			id, err := strconv.Atoi(strings.TrimPrefix(update.Key, "HitLocation."))
			if err != nil {
				return errs.NewWithCause("invalid number", err)
			}
			if loc := sheet.FindHitLocationByID(entity.Entity, gurps.SheetSettingsFor(entity.Entity).BodyType, id); loc != nil {
				stringFieldPtr = &loc.Notes
				break
			}
		}
		return errs.Newf("unknown field key: %q", update.Key)
	}
	update.Data = txt.CollapseSpaces(strings.TrimSpace(update.Data))
	switch {
	case stringFieldPtr != nil:
		slog.Info("string update", "update.Data", update.Data, "*stringFieldPtr", *stringFieldPtr)
		if update.Data == *stringFieldPtr {
			return nil
		}
		*stringFieldPtr = update.Data
	case lengthFieldPtr != nil:
		length, err := fxp.LengthFromString(update.Data, entity.Entity.SheetSettings.DefaultLengthUnits)
		if err != nil || length < 0 || length > fxp.Length(fxp.Max) {
			return errs.Newf("invalid input: %q", update.Data)
		}
		if update.Data == entity.Entity.SheetSettings.DefaultLengthUnits.Format(*lengthFieldPtr) {
			return nil
		}
		*lengthFieldPtr = length
	case weightFieldPtr != nil:
		weight, err := fxp.WeightFromString(update.Data, entity.Entity.SheetSettings.DefaultWeightUnits)
		if err != nil || weight < 0 || weight > fxp.Weight(fxp.Max) {
			return errs.Newf("invalid input: %q", update.Data)
		}
		if update.Data == entity.Entity.SheetSettings.DefaultWeightUnits.Format(*weightFieldPtr) {
			return nil
		}
		*weightFieldPtr = weight
	case intSetter != nil:
		v, err := strconv.ParseInt(update.Data, 10, 64)
		if err == nil {
			err = intSetter(int(v))
		}
		if err != nil {
			return errs.Newf("invalid input: %q", update.Data)
		}
	case fxpIntSetter != nil:
		v, err := fxp.FromString(update.Data)
		if err == nil {
			err = fxpIntSetter(v)
		}
		if err != nil {
			return errs.Newf("invalid input: %q", update.Data)
		}
	}
	entity.Entity.ModifiedOn = jio.Now()
	entity.CurrentCRC64 = entity.Entity.CRC64()
	s.sheetsLock.Lock()
	s.entitiesByPath[entity.ClientPath] = *entity
	s.sheetsLock.Unlock()
	slog.Info("updated sheet", "path", entity.ClientPath, "field", update.Key, "text", update.Data)
	return nil
}

func (s *Server) updateFieldBinary(entity *webEntity, update *sheetUpdate) error {
	switch update.Key {
	case "Portrait":
		if update.Data == "" {
			return errs.New("no data")
		}
		data, err := base64.StdEncoding.DecodeString(update.Data)
		if err != nil {
			return errs.Wrap(err)
		}
		data, err = imgutil.ConvertForPortraitUse(data)
		if err != nil {
			return err
		}
		entity.Entity.Profile.PortraitData = data
		entity.Entity.Profile.PortraitImage = nil
	default:
		return errs.Newf("unknown field key: %q", update.Key)
	}
	entity.Entity.ModifiedOn = jio.Now()
	entity.CurrentCRC64 = entity.Entity.CRC64()
	s.sheetsLock.Lock()
	s.entitiesByPath[entity.ClientPath] = *entity
	s.sheetsLock.Unlock()
	slog.Info("updated sheet", "path", entity.ClientPath, "field", update.Key)
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
		loadedEntity, err := gurps.NewEntityFromFile(os.DirFS(filepath.Dir(entityPath)), filepath.Base(entityPath))
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
