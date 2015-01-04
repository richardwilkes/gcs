/*
 * Copyright (c) 1998-2015 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public License,
 * version 2.0. If a copy of the MPL was not distributed with this file, You
 * can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as defined
 * by the Mozilla Public License, version 2.0.
 */

package com.trollworks.gcs.feature;

import com.trollworks.gcs.widgets.outline.ListRow;
import com.trollworks.toolkit.ui.layout.FlexGrid;
import com.trollworks.toolkit.ui.layout.FlexRow;
import com.trollworks.toolkit.ui.layout.FlexSpacer;

/** This is a placeholder feature, used for the empty state. */
public class NoFeature extends FeatureEditor {
	/**
	 * Create a new placeholder.
	 * 
	 * @param row The row this feature will belong to.
	 */
	public NoFeature(ListRow row) {
		super(row, null);

	}

	@Override
	protected void rebuildSelf(FlexGrid grid, FlexRow right) {
		grid.add(new FlexSpacer(0, 0, true, false), 0, 0);
	}
}
