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

// Package ux holds the user experience implementation for GCS. Due to the incestuous nature of user interface code, I
// ultimately decided to combine all of the sub-packages for the user experience into a single large package. Prior to
// this, I had to introduce a number of artificial function trampolines in order to avoid circular references. The
// package structure at that time also prevented some the features I ultimately wanted to add from being implemented,
// again, thanks to circular references that were difficult or even impossible to resolve. A flat package structure
// eliminates those problems at the expense of making it more difficult to find things and requiring sometimes awkwardly
// long naming conventions.
package ux
