/*
 * Copyright ©1998-2022 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, version 2.0. If a copy of the MPL was not distributed with
 * this file, You can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as
 * defined by the Mozilla Public License, version 2.0.
 */

package ui

import (
	"bytes"
	"fmt"
	"image"
	"os"
	"path/filepath"
	"syscall"

	"github.com/richardwilkes/gcs/v5/model/library"
	"github.com/richardwilkes/gcs/v5/ui/svglayer"
	"github.com/richardwilkes/toolbox/cmdline"
	"github.com/richardwilkes/toolbox/errs"
	"github.com/richardwilkes/toolbox/formats/icon"
	"github.com/richardwilkes/toolbox/formats/icon/ico"
	"github.com/richardwilkes/toolbox/log/jot"
	"github.com/richardwilkes/toolbox/xio/fs/paths"
	"golang.org/x/sys/windows/registry"
)

const (
	shcneAssocChanged = 0x08000000
	shcfnIDlist       = 0
)

var (
	shell32            = syscall.NewLazyDLL("shell32.dll")
	shChangeNotifyProc = shell32.NewProc("SHChangeNotify")
	softwareClasses    = `Software\Classes\`
)

func performPlatformStartup() {
	if err := configureRegistry(); err != nil {
		jot.Error(err)
	}
}

func configureRegistry() error {
	exePath, err := os.Executable()
	if err != nil {
		return errs.Wrap(err)
	}
	if exePath, err = filepath.Abs(exePath); err != nil {
		return errs.Wrap(err)
	}
	var docBaseIcon image.Image
	if docBaseIcon, _, err = image.Decode(bytes.NewBuffer(docIconBytes)); err != nil {
		return errs.Wrap(err)
	}
	for i := range library.KnownFileTypes {
		if fi := &library.KnownFileTypes[i]; fi.IsGCSData {
			// Create the doc icon
			var overlay image.Image
			if overlay, err = svglayer.CreateImageFromSVG(fi, 128); err != nil {
				return err
			}
			docPath := filepath.Join(paths.AppDataDir(), fi.Extensions[0][1:]+".ico")
			if err = writeIco(icon.Stack(docBaseIcon, overlay), docPath); err != nil {
				return err
			}

			// Create the entry that points to the app's information for the extension
			appExtKey := cmdline.AppIdentifier + fi.Extensions[0]
			if err = setKey(softwareClasses+fi.Extensions[0], "", appExtKey); err != nil {
				return err
			}

			// Create the entry for the extension
			path := softwareClasses + appExtKey
			if err = setKey(path, "", fi.Name); err != nil {
				return err
			}
			if err = setKey(path+`\DefaultIcon`, "", docPath); err != nil {
				return err
			}
			if err = setKey(path+`\Shell`, "", ""); err != nil {
				return err
			}
			if err = setKey(path+`\Shell\Open`, "", ""); err != nil {
				return err
			}
			if err = setKey(path+`\Shell\Open\Command`, "", fmt.Sprintf(`"%s" "%%1"`, exePath)); err != nil {
				return err
			}
		}
	}
	shChangeNotifyProc.Call(shcneAssocChanged, shcfnIDlist, 0, 0)
	return nil
}

func writeIco(img image.Image, path string) (err error) {
	var f *os.File
	if f, err = os.Create(path); err != nil {
		return errs.Wrap(err)
	}
	defer func() {
		if cerr := f.Close(); cerr != nil && err == nil {
			err = errs.Wrap(cerr)
		}
	}()
	err = errs.Wrap(ico.Encode(f, img))
	return
}

func setKey(path, name, value string) error {
	k, _, err := registry.CreateKey(registry.CURRENT_USER, path, registry.READ|registry.WRITE)
	if err != nil {
		return errs.Wrap(err)
	}
	if err = k.SetStringValue(name, value); err != nil {
		return errs.Wrap(err)
	}
	if err = k.Close(); err != nil {
		return errs.Wrap(err)
	}
	return nil
}

/*
void RegSet( HKEY hkeyHive, const char* pszVar, const char* pszVa

lue ) {

  HKEY hkey;

  char szValueCurrent[1000];
  DWORD dwType;
  DWORD dwSize = sizeof( szValueCurrent );

  int iRC = RegGetValue( hkeyHive, pszVar, NULL, RRF_RT_ANY, &dwType, szValueCurrent, &dwSize );

  bool bDidntExist = iRC == ERROR_FILE_NOT_FOUND;

  if ( iRC != ERROR_SUCCESS && !bDidntExist )
      AKS( AKSFatal, "RegGetValue( %s ): %s", pszVar, strerror( iRC ) );

  DWORD dwDisposition;
  iRC = RegCreateKeyEx( hkeyHive, pszVar, 0, 0, 0, KEY_ALL_ACCESS, NULL, &hkey, &dwDisposition );
  if ( iRC != ERROR_SUCCESS )
      AKS( AKSFatal, "RegCreateKeyEx( %s ): %s", pszVar, strerror( iRC ) );

  iRC = RegSetValueEx( hkey, "", 0, REG_SZ, (BYTE*) pszValue, strlen( pszValue ) + 1 );
  if ( iRC != ERROR_SUCCESS )
      AKS( AKSFatal, "RegSetValueEx( %s ): %s", pszVar, strerror( iRC ) );

  RegCloseKey(hkey);
}
*/
