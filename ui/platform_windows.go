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
	"fmt"
	"os"
	"path/filepath"
	"syscall"

	"github.com/richardwilkes/gcs/v5/model/library"
	"github.com/richardwilkes/toolbox/cmdline"
	"github.com/richardwilkes/toolbox/errs"
	"github.com/richardwilkes/toolbox/log/jot"
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
	if err = setKey(softwareClasses+cmdline.AppIdentifier, "", ""); err != nil {
		return err
	}
	if err = setKey(softwareClasses+cmdline.AppIdentifier+`\Shell`, "", ""); err != nil {
		return err
	}
	if err = setKey(softwareClasses+cmdline.AppIdentifier+`\Shell\Open`, "", ""); err != nil {
		return err
	}
	if err = setKey(softwareClasses+cmdline.AppIdentifier+`\Shell\Open\Command`, "", fmt.Sprintf(`"%s" "%%*"`, exePath)); err != nil {
		return err
	}
	counter := 1
	for i := range library.KnownFileTypes {
		if fi := &library.KnownFileTypes[i]; fi.IsGCSData {
			// Create the entry that points to the app's information for the extension
			path := softwareClasses + fi.Extensions[0]
			if err = setKey(path, "", cmdline.AppIdentifier); err != nil {
				return err
			}
			if err = setKey(path, "DefaultIcon", fmt.Sprintf("%s,%d", exePath, counter)); err != nil {
				return err
			}
			counter++
			if err = setKey(path, "Content Type", fi.MimeTypes[0]); err != nil {
				return err
			}
		}
	}
	shChangeNotifyProc.Call(shcneAssocChanged, shcfnIDlist, 0, 0)
	return nil
}

func setKey(path, name, value string) error {
	// if err  := registry.DeleteKey(registry.CURRENT_USER, path); err != nil {
	// 	var e *syscall.Errno
	// 	if !errors.As(err, &e) || *e != registry.ErrNotExist {
	// 		return errs.Wrap(err)
	// 	}
	// }
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
