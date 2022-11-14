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
	"github.com/richardwilkes/toolbox/i18n"
)

// Possible NameGenerationType values.
const (
	Simple      = NameGenerationType("simple")
	MarkovChain = NameGenerationType("markov_chain")
)

// AllNameGenerationTypes is the complete set of NameGenerationType values.
var AllNameGenerationTypes = []NameGenerationType{
	Simple,
	MarkovChain,
}

// NameGenerationType holds the type of a name generation technique.
type NameGenerationType string

// EnsureValid ensures this is of a known value.
func (n NameGenerationType) EnsureValid() NameGenerationType {
	for _, one := range AllNameGenerationTypes {
		if one == n {
			return n
		}
	}
	return AllNameGenerationTypes[0]
}

// String implements fmt.Stringer.
func (n NameGenerationType) String() string {
	switch n {
	case Simple:
		return i18n.Text("Simple")
	case MarkovChain:
		return i18n.Text("Markov Chain")
	default:
		return Simple.String()
	}
}
