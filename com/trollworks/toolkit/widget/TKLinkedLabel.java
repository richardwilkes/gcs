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

package com.trollworks.toolkit.widget;

import com.trollworks.toolkit.utility.TKAlignment;

/** A label whose tooltip reflects that of another panel. */
public class TKLinkedLabel extends TKLabel {
	private TKPanel	mLink;

	/**
	 * Creates a label with the specified text. The label is right-aligned and centered vertically
	 * in its display area.
	 * 
	 * @param text The text to be displayed.
	 */
	public TKLinkedLabel(String text) {
		super(text, TKAlignment.RIGHT);
	}

	/**
	 * Creates a label with the specified text. The label is right-aligned and centered vertically
	 * in its display area.
	 * 
	 * @param text The text to be displayed.
	 * @param font The dynamic font to use.
	 */
	public TKLinkedLabel(String text, String font) {
		super(text, font, TKAlignment.RIGHT);
	}

	/**
	 * Creates a label with the specified text. The label is right-aligned and centered vertically
	 * in its display area.
	 * 
	 * @param link The {@link TKPanel} to pair with.
	 * @param text The text to be displayed.
	 */
	public TKLinkedLabel(TKPanel link, String text) {
		super(text, TKAlignment.RIGHT);
		setLink(link);
	}

	/**
	 * Creates a label with the specified text. The label is right-aligned and centered vertically
	 * in its display area.
	 * 
	 * @param link The {@link TKPanel} to pair with.
	 * @param text The text to be displayed.
	 * @param font The dynamic font to use.
	 */
	public TKLinkedLabel(TKPanel link, String text, String font) {
		super(text, font, TKAlignment.RIGHT);
		setLink(link);
	}

	@Override public String getToolTipText() {
		return mLink != null ? mLink.getToolTipText() : super.getToolTipText();
	}

	/** @return The {@link TKPanel} that is being paired with. */
	public TKPanel getLink() {
		return mLink;
	}

	/** @param link The {@link TKPanel} to pair with. */
	public void setLink(TKPanel link) {
		mLink = link;
	}
}
