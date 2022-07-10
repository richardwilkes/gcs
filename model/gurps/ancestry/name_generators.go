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

package ancestry

import (
	"embed"
	"sort"

	"github.com/richardwilkes/gcs/v5/model/library"
	"github.com/richardwilkes/toolbox/txt"
)

//go:embed data
var embeddedFS embed.FS

// NameGeneratorRef holds a reference to a NameGenerator.
type NameGeneratorRef struct {
	FileRef   *library.NamedFileRef
	generator *NameGenerator
}

// AvailableNameGenerators scans the libraries and returns the available name generators.
func AvailableNameGenerators(libraries library.Libraries) []*NameGeneratorRef {
	var list []*NameGeneratorRef
	seen := make(map[string]bool)
	for _, set := range library.ScanForNamedFileSets(embeddedFS, "data", true, libraries, ".names") {
		for _, one := range set.List {
			if seen[one.Name] {
				continue
			}
			seen[one.Name] = true
			list = append(list, &NameGeneratorRef{FileRef: one})
		}
	}
	sort.Slice(list, func(i, j int) bool { return txt.NaturalLess(list[i].FileRef.Name, list[j].FileRef.Name, true) })
	return list
}

// Generator returns the NameGenerator, loading it if needed.
func (n *NameGeneratorRef) Generator() (*NameGenerator, error) {
	if n.generator == nil {
		g, err := NewNameGeneratorFromFS(n.FileRef.FileSystem, n.FileRef.FilePath)
		if err != nil {
			return nil, err
		}
		n.generator = g
	}
	return n.generator, nil
}
