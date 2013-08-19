/* ***** BEGIN LICENSE BLOCK *****
 * Version: MPL 1.1
 *
 * The contents of this file are subject to the Mozilla Public License Version
 * 1.1 (the "License"); you may not use this file except in compliance with
 * the License. You may obtain a copy of the License at
 * http://www.mozilla.org/MPL/
 *
 * Software distributed under the License is distributed on an "AS IS" basis,
 * WITHOUT WARRANTY OF ANY KIND, either express or implied. See the License
 * for the specific language governing rights and limitations under the
 * License.
 *
 * The Original Code is GURPS Character Sheet.
 *
 * The Initial Developer of the Original Code is Richard A. Wilkes.
 * Portions created by the Initial Developer are Copyright (C) 1998-2002,
 * 2005-2007 the Initial Developer. All Rights Reserved.
 *
 * Contributor(s):
 *
 * ***** END LICENSE BLOCK ***** */

package com.trollworks.gcs.ui.common;

import com.trollworks.gcs.model.CMRow;
import com.trollworks.toolkit.utility.TKAlignment;
import com.trollworks.toolkit.utility.TKColor;
import com.trollworks.toolkit.widget.TKDefaultItemRenderer;

import java.awt.Color;
import java.awt.image.BufferedImage;

/** An item renderer for rows. */
public class CSRowItemRenderer extends TKDefaultItemRenderer {
	/** Creates a new row renderer. */
	public CSRowItemRenderer() {
		super(CSFont.KEY_FIELD);
		setVerticalAlignment(TKAlignment.CENTER);
	}

	@Override public BufferedImage getImageForItem(Object item, int index) {
		if (item instanceof CMRow) {
			return ((CMRow) item).getImage(false);
		}
		return super.getImageForItem(item, index);
	}

	@Override public Color getBackgroundForItem(Object item, int index, boolean selected, boolean active) {
		return selected ? active ? TKColor.HIGHLIGHT : TKColor.INACTIVE_HIGHLIGHT : index % 2 == 0 ? TKColor.PRIMARY_BANDING : TKColor.SECONDARY_BANDING;
	}

	@Override public String getFontKeyForItem(Object item, int index) {
		return CSFont.KEY_FIELD;
	}
}
