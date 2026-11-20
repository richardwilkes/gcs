// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package ux

import "github.com/richardwilkes/unison"

// cellPanelPoolKey is the key under which a table holds its cellPanelPool.
const cellPanelPoolKey = "cell_panel_pool"

// cellPanelPoolTurn is how many cell panels a cellPanelPool takes in before it lets go of those not used since the turn
// before. It has to exceed the cells a table asks for to draw itself once, or they would be let go while still in view.
var cellPanelPoolTurn = 4096

// cellPanelPool bounds the cell panels the rows of a table that isn't on a page keep, at twice cellPanelPoolTurn. Such
// a table, a list of every trait in the libraries for one, may show tens of thousands of rows, of which only those in
// view are drawn: the rows keep the sizes of their cells for the table to measure them by (see Node.KeptCellSize), so
// the panels of the rest have no further use. The rows of a page, which holds one character's items, keep all of
// theirs.
type cellPanelPool struct {
	recent []*CellCache // The panels built or used since the pool last turned over
	older  []*CellCache // Those of the turn before, let go at the next unless used again by then
	turn   uint64       // Never 0, which marks a panel no pool holds
}

// cellPanelPoolFor returns the pool of the table, which is given one when it has none.
func cellPanelPoolFor(table unison.Paneler) *cellPanelPool {
	data := table.AsPanel().ClientData()
	if pool, ok := data[cellPanelPoolKey].(*cellPanelPool); ok {
		return pool
	}
	pool := &cellPanelPool{turn: 1}
	data[cellPanelPoolKey] = pool
	return pool
}

// use notes that the panel was built or used, which keeps it through the next turn.
func (p *cellPanelPool) use(c *CellCache) {
	if c.turn == p.turn {
		return
	}
	c.turn = p.turn
	p.recent = append(p.recent, c)
	if len(p.recent) >= cellPanelPoolTurn {
		for i, old := range p.older {
			if old.turn != p.turn {
				old.letGo()
			}
			p.older[i] = nil
		}
		p.older, p.recent = p.recent, p.older[:0]
		p.turn++
	}
}

// release lets go of the panel now. Only the one taken in last, as a panel built just to be measured is, is taken out
// of the pool, which lets go of any other again, to no effect, in its turn.
func (p *cellPanelPool) release(c *CellCache) {
	if last := len(p.recent) - 1; last >= 0 && p.recent[last] == c {
		p.recent[last] = nil
		p.recent = p.recent[:last]
	}
	c.letGo()
}
