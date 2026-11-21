// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package library

import (
	"encoding/binary"
	"errors"
	"path/filepath"
	"strings"
	"unsafe"

	"github.com/richardwilkes/toolbox/v2/errs"
	"github.com/rjeczalik/notify"
	"golang.org/x/sys/windows"
)

var procOpenFileByID = windows.NewLazySystemDLL("kernel32.dll").NewProc("OpenFileById")

// The FILE_ID_TYPE values of the ids folderID holds.
const (
	fileIDType         = 0
	extendedFileIDType = 2
)

// folderID identifies a folder on its volume, wherever on that volume it is moved to.
type folderID struct {
	volume uint64
	kind   uint32   // fileIDType or extendedFileIDType
	id     [16]byte // A 128-bit id, or for fileIDType a 64-bit one in the first 8 bytes
}

// fileIDInfo is the FILE_ID_INFO that GetFileInformationByHandleEx fills in.
type fileIDInfo struct {
	volume uint64
	id     [16]byte
}

// fileIDDescriptor is the FILE_ID_DESCRIPTOR that OpenFileById takes.
type fileIDDescriptor struct {
	size uint32
	kind uint32
	id   [16]byte
}

// idOfFolder returns the id of the folder at p.
func idOfFolder(p string) (folderID, error) {
	h, err := openFolder(p)
	if err != nil {
		return folderID{}, err
	}
	defer closeFolder(h)
	return idOfHandle(h)
}

// findFolder returns the path of the folder with the id, looking for it on the volume that holds the folder at hint, or
// an empty string if no folder there has the id.
func findFolder(id folderID, hint string) (string, error) {
	h, err := openFolder(hint)
	if err != nil {
		return "", err
	}
	defer closeFolder(h)
	var hintID folderID
	if hintID, err = idOfHandle(h); err != nil {
		return "", err
	}
	if hintID.volume != id.volume {
		return "", errors.New("the folder is not on the volume looked on")
	}
	var flags uint32
	if err = windows.GetVolumeInformationByHandle(h, nil, 0, nil, nil, &flags, nil, 0); err != nil {
		return "", err
	}
	if flags&windows.FILE_SUPPORTS_OPEN_BY_FILE_ID == 0 {
		return "", errors.New("the volume cannot open a folder by its id")
	}
	desc := fileIDDescriptor{size: uint32(unsafe.Sizeof(fileIDDescriptor{})), kind: id.kind, id: id.id}
	r, _, callErr := procOpenFileByID.Call(uintptr(h), uintptr(unsafe.Pointer(&desc)), 0,
		windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE|windows.FILE_SHARE_DELETE, 0, windows.FILE_FLAG_BACKUP_SEMANTICS)
	if windows.Handle(r) == windows.InvalidHandle {
		if errors.Is(callErr, windows.ERROR_INVALID_PARAMETER) {
			return "", nil // OpenFileById's answer, on a volume that supports it, for an id no file has
		}
		return "", callErr
	}
	defer closeFolder(windows.Handle(r))
	buf := make([]uint16, windows.MAX_PATH)
	for {
		var n uint32
		if n, err = windows.GetFinalPathNameByHandle(windows.Handle(r), &buf[0], uint32(len(buf)), 0); err != nil {
			return "", err
		}
		if n < uint32(len(buf)) {
			return windows.UTF16ToString(buf[:n]), nil
		}
		buf = make([]uint16, n) // n is the size needed, terminator included
	}
}

// shortPath returns p with each name in it that has a short (8.3) one given by that, or p itself if that can't be had.
func shortPath(p string) string {
	name, err := windows.UTF16PtrFromString(p)
	if err != nil {
		return p
	}
	buf := make([]uint16, windows.MAX_PATH)
	for {
		var n uint32
		if n, err = windows.GetShortPathName(name, &buf[0], uint32(len(buf))); err != nil || n == 0 {
			return p
		}
		if n < uint32(len(buf)) {
			return windows.UTF16ToString(buf[:n])
		}
		buf = make([]uint16, n) // n is the size needed, terminator included
	}
}

func openFolder(p string) (windows.Handle, error) {
	name, err := windows.UTF16PtrFromString(p)
	if err != nil {
		return windows.InvalidHandle, err
	}
	return windows.CreateFile(name, 0, windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE|windows.FILE_SHARE_DELETE, nil,
		windows.OPEN_EXISTING, windows.FILE_FLAG_BACKUP_SEMANTICS, 0)
}

func closeFolder(h windows.Handle) {
	if err := windows.CloseHandle(h); err != nil {
		errs.Log(err)
	}
}

// idOfHandle returns the id of the folder open as h: its 128-bit one, since the 64-bit one need not be unique on ReFS,
// or its 64-bit one on volumes without 128-bit ids, such as FAT ones.
func idOfHandle(h windows.Handle) (folderID, error) {
	var info fileIDInfo
	if windows.GetFileInformationByHandleEx(h, windows.FileIdInfo, (*byte)(unsafe.Pointer(&info)),
		uint32(unsafe.Sizeof(info))) == nil {
		return folderID{volume: info.volume, kind: extendedFileIDType, id: info.id}, nil
	}
	var byHandle windows.ByHandleFileInformation
	if err := windows.GetFileInformationByHandle(h, &byHandle); err != nil {
		return folderID{}, err
	}
	id := folderID{volume: uint64(byHandle.VolumeSerialNumber), kind: fileIDType}
	binary.LittleEndian.PutUint64(id.id[:], uint64(byHandle.FileIndexHigh)<<32|uint64(byHandle.FileIndexLow))
	return id, nil
}

// fileNotifyInformationSize is the size of a FILE_NOTIFY_INFORMATION record ahead of the name that ends it.
const fileNotifyInformationSize = 12

// entryWatch is the state of a watch made by watchFolderEntry. It is kept on the heap, where it does not move, since
// the system writes to buf and ov while a read is under way.
type entryWatch struct {
	buf    [4096]byte // FILE_NOTIFY_INFORMATION records, which must be DWORD-aligned
	ov     windows.Overlapped
	dir    windows.Handle // The folder holding the watched one
	quit   windows.Handle // Set to stop the watch
	parent string         // The path of dir
	name   string         // The watched folder's
	short  string         // Its short (8.3) name, by which a change made through it is named, or its name
}

// watchFolderEntry calls changed, on a goroutine of its own, for each change made to the name of the folder at p within
// the folder holding it, until the returned function is called, which returns once changed is no longer being called.
// rjeczalik/notify watches a folder through a handle, which follows the folder wherever another program moves it, and
// reports nothing of the move, so its watch goes on reporting the changes made where the folder went, and none in a
// folder put in its place. The watch keeps a handle on the folder holding p, which keeps the folders holding that one
// from being renamed, until it is stopped or ends because that folder is removed or can't be read. A folder at the root
// of a volume, which can't be moved, is not watched for.
func watchFolderEntry(p string, changed func(what notify.Event)) (stop func()) {
	parent := filepath.Dir(p)
	if parent == p {
		return func() {}
	}
	w := &entryWatch{
		dir:    windows.InvalidHandle,
		parent: parent,
		name:   filepath.Base(p),
		short:  filepath.Base(shortPath(p)),
	}
	name, err := windows.UTF16PtrFromString(parent)
	if err == nil {
		w.dir, err = windows.CreateFile(name, windows.FILE_LIST_DIRECTORY,
			windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE|windows.FILE_SHARE_DELETE, nil, windows.OPEN_EXISTING,
			windows.FILE_FLAG_BACKUP_SEMANTICS|windows.FILE_FLAG_OVERLAPPED, 0)
	}
	if err == nil {
		w.ov.HEvent, err = windows.CreateEvent(nil, 1, 0, nil)
	}
	if err == nil {
		w.quit, err = windows.CreateEvent(nil, 1, 0, nil)
	}
	if err == nil {
		err = w.read() // Before returning, since the system records no change made ahead of the first read
	}
	if err != nil {
		errs.Log(errs.NewWithCause("unable to watch filesystem path", err), "path", parent)
		closeHandles(w.dir, w.ov.HEvent, w.quit)
		return func() {}
	}
	done := make(chan struct{})
	go func() {
		defer close(done)
		w.run(changed)
		closeHandles(w.dir, w.ov.HEvent) // No read is under way once run returns
	}()
	return func() {
		if setErr := windows.SetEvent(w.quit); setErr != nil {
			errs.Log(setErr)
		}
		<-done
		closeHandles(w.quit)
	}
}

// read starts a read of the changes to the names of the folders within dir.
func (w *entryWatch) read() error {
	if err := windows.ResetEvent(w.ov.HEvent); err != nil {
		return err
	}
	return windows.ReadDirectoryChanges(w.dir, &w.buf[0], uint32(len(w.buf)), false, windows.FILE_NOTIFY_CHANGE_DIR_NAME,
		nil, &w.ov, 0)
}

// run calls changed for each change to the watched folder's name, with a read of them under way, until quit is set, or
// until the folder holding it is removed or can't be read.
func (w *entryWatch) run(changed func(what notify.Event)) {
	for {
		which, err := windows.WaitForMultipleObjects([]windows.Handle{w.ov.HEvent, w.quit}, false, windows.INFINITE)
		if err != nil || which != windows.WAIT_OBJECT_0 {
			if err != nil {
				errs.Log(err)
			}
			w.cancel()
			return
		}
		var n uint32
		err = windows.GetOverlappedResult(w.dir, &w.ov, &n, false)
		switch {
		case errors.Is(err, windows.ERROR_NOTIFY_ENUM_DIR) || (err == nil && n == 0):
			changed(notify.Rename) // Too much changed to tell what, so the folder's name may have
			err = w.read()
		case err == nil:
			w.report(n, changed)
			err = w.read()
		}
		if err != nil {
			if !errors.Is(err, windows.ERROR_ACCESS_DENIED) { // What the removal of the folder holding the watched one brings
				errs.Log(errs.NewWithCause("unable to watch filesystem path", err), "path", w.parent)
			}
			return
		}
	}
}

// report calls changed for each record among the first n bytes of buf that names the watched folder, by either of its
// names.
func (w *entryWatch) report(n uint32, changed func(what notify.Event)) {
	for offset := uint32(0); offset+fileNotifyInformationSize <= n; {
		record := w.buf[offset:n]
		next := binary.LittleEndian.Uint32(record)
		size := binary.LittleEndian.Uint32(record[8:])
		if fileNotifyInformationSize+size <= uint32(len(record)) {
			name := make([]uint16, size/2)
			for i := range name {
				name[i] = binary.LittleEndian.Uint16(record[fileNotifyInformationSize+2*i:])
			}
			if s := windows.UTF16ToString(name); strings.EqualFold(s, w.name) || strings.EqualFold(s, w.short) {
				switch binary.LittleEndian.Uint32(record[4:]) {
				case windows.FILE_ACTION_ADDED:
					changed(notify.Create)
				case windows.FILE_ACTION_REMOVED:
					changed(notify.Remove)
				default:
					changed(notify.Rename)
				}
			}
		}
		if next == 0 {
			return
		}
		offset += next
	}
}

// cancel cancels the read under way and waits it out, since the system writes to buf and ov until it is done.
func (w *entryWatch) cancel() {
	if err := windows.CancelIoEx(w.dir, &w.ov); err != nil && !errors.Is(err, windows.ERROR_NOT_FOUND) {
		errs.Log(err)
	}
	var n uint32
	if err := windows.GetOverlappedResult(w.dir, &w.ov, &n, true); err != nil &&
		!errors.Is(err, windows.ERROR_OPERATION_ABORTED) {
		errs.Log(err)
	}
}

// closeHandles closes those of the handles that are open.
func closeHandles(handles ...windows.Handle) {
	for _, h := range handles {
		if h != 0 && h != windows.InvalidHandle {
			if err := windows.CloseHandle(h); err != nil {
				errs.Log(err)
			}
		}
	}
}
