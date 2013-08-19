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

package com.trollworks.toolkit.widget.button;

import com.trollworks.toolkit.io.TKImage;
import com.trollworks.toolkit.utility.TKAlignment;

/** A standard radio button. */
public class TKRadioButton extends TKBaseButton {
	private boolean	mSelected;

	/** Creates an unselected radio button with no text. */
	public TKRadioButton() {
		this(null, false, TKAlignment.LEFT);
	}

	/**
	 * Creates a radio button with no text.
	 * 
	 * @param selected The initial selection state.
	 */
	public TKRadioButton(boolean selected) {
		this(null, selected, TKAlignment.LEFT);
	}

	/**
	 * Creates an unselected radio button with the specified text.
	 * 
	 * @param text The text to use.
	 */
	public TKRadioButton(String text) {
		this(text, false, TKAlignment.LEFT);
	}

	/**
	 * Creates a radio button with the specified text.
	 * 
	 * @param text The text to use.
	 * @param selected The initial selection state.
	 */
	public TKRadioButton(String text, boolean selected) {
		this(text, selected, TKAlignment.LEFT);
	}

	/**
	 * Creates an unselected radio button with the specified text and horizontal alignment.
	 * 
	 * @param text The text to use.
	 * @param alignment The horizontal alignment to use.
	 */
	public TKRadioButton(String text, int alignment) {
		this(text, false, alignment);
	}

	/**
	 * Creates an unselected radio button with the specified text and horizontal alignment.
	 * 
	 * @param text The text to use.
	 * @param selected The initial selection state.
	 * @param alignment The horizontal alignment to use.
	 */
	public TKRadioButton(String text, boolean selected, int alignment) {
		super(text, TKImage.getRadioButtonEmptyIcon(), alignment);
		mSelected = !selected; // Ensure the following call will "take".
		setSelected(selected);
	}

	@Override public void doClick() {
		setSelected(true);
	}

	/** @return <code>true</code> if this radio button is selected. */
	public boolean isSelected() {
		return mSelected;
	}

	/** @param selected The selected state of the radio button. */
	public void setSelected(boolean selected) {
		if (mSelected != selected) {
			mSelected = selected;
			setImage(selected ? TKImage.getRadioButtonSelectedIcon() : TKImage.getRadioButtonEmptyIcon());
			setPressedIcon(selected ? TKImage.getRadioButtonPressedSelectedIcon() : TKImage.getRadioButtonPressedEmptyIcon());
			notifyActionListeners();
		}
	}
}
