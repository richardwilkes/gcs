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

package com.trollworks.toolkit.window;

import com.trollworks.toolkit.text.TKTextDrawing;
import com.trollworks.toolkit.utility.TKAlignment;
import com.trollworks.toolkit.utility.TKColor;
import com.trollworks.toolkit.utility.TKFont;
import com.trollworks.toolkit.widget.TKLabel;
import com.trollworks.toolkit.widget.border.TKBorder;
import com.trollworks.toolkit.widget.border.TKCompoundBorder;
import com.trollworks.toolkit.widget.border.TKEmptyBorder;
import com.trollworks.toolkit.widget.border.TKLineBorder;

/** A standard, text-only, tooltip. */
public class TKTooltip extends TKWidgetWindow {
	private static final TKBorder	BORDER	= new TKCompoundBorder(TKLineBorder.getSharedBorder(true), new TKEmptyBorder(2, 4, 2, 2));
	private String					mTip;

	/**
	 * @param owner The owning window.
	 * @param tip The tooltip text.
	 */
	public TKTooltip(TKBaseWindow owner, String tip) {
		super(owner);
		mTip = tip;
		TKLabel content = new TKLabel(TKTextDrawing.wrapToPixelWidth(TKFont.lookup(TKFont.TEXT_FONT_KEY), null, tip, getGraphicsConfiguration().getBounds().width / 2), null, TKAlignment.LEFT, true, TKFont.TEXT_FONT_KEY);
		content.setBackground(TKColor.TOOLTIP_BACKGROUND);
		content.setOpaque(true);
		content.setBorder(BORDER);
		setContent(content);
	}

	/** @return The tooltip's text. */
	public String getTipText() {
		return mTip;
	}
}
