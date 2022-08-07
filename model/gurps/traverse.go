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

package gurps

import "golang.org/x/exp/slices"

type traversalData[T NodeTypes] struct {
	list  []T
	index int
}

// Traverse calls the function 'f' for each node and its children in the input list, recursively. Return true from the
// function to abort early. If excludeContainers is true, then nodes that are containers will not be passed to 'f',
// although their children will still be processed as usual.
func Traverse[T NodeTypes](f func(T) bool, excludeContainers, excludeChildrenOfDisabledNodes bool, in ...T) {
	tracking := []*traversalData[T]{
		{
			list:  in,
			index: 0,
		},
	}
	for len(tracking) != 0 {
		current := tracking[len(tracking)-1]
		if current.index >= len(current.list) {
			tracking = tracking[:len(tracking)-1]
			continue
		}
		one := current.list[current.index]
		node := AsNode(one)
		current.index++
		enabled := node.Enabled()
		if enabled && (!excludeContainers || !node.Container()) && f(one) {
			return
		}
		if (enabled || !excludeChildrenOfDisabledNodes) && node.HasChildren() {
			tracking = append(tracking, &traversalData[T]{list: slices.Clone(node.NodeChildren())})
		}
	}
}
