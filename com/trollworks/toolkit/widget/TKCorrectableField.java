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

import com.trollworks.toolkit.window.TKCorrectableManager;

/** A correctable text field. */
public class TKCorrectableField extends TKTextField implements TKCorrectable {
	private TKCorrectableLabel	mLabel;
	private boolean				mWasCorrected;
	private String				mSavedToolTip;

	/**
	 * Creates a left-justified text field with the specified text.
	 * 
	 * @param label The {@link TKCorrectableLabel} to pair with.
	 * @param text The initial text for the field.
	 */
	public TKCorrectableField(TKCorrectableLabel label, String text) {
		this(label, text, true);
	}

	/**
	 * Creates a left-justified text field with the specified text.
	 * 
	 * @param label The {@link TKCorrectableLabel} to pair with.
	 * @param text The initial text for the field.
	 * @param singleLineOnly Whether to allow only a single line of text.
	 */
	public TKCorrectableField(TKCorrectableLabel label, String text, boolean singleLineOnly) {
		super(text, singleLineOnly);
		setLabel(label);
	}

	/**
	 * Creates a text field with the specified text and alignment.
	 * 
	 * @param label The {@link TKCorrectableLabel} to pair with.
	 * @param text The initial text for the field.
	 * @param alignment The alignment of the text.
	 */
	public TKCorrectableField(TKCorrectableLabel label, String text, int alignment) {
		this(label, text, alignment, true);
	}

	/**
	 * Creates a text field with the specified text and alignment.
	 * 
	 * @param label The {@link TKCorrectableLabel} to pair with.
	 * @param text The initial text for the field.
	 * @param alignment The alignment of the text.
	 * @param singleLineOnly Whether to allow only a single line of text.
	 */
	public TKCorrectableField(TKCorrectableLabel label, String text, int alignment, boolean singleLineOnly) {
		super(text, 10, alignment, singleLineOnly);
		setLabel(label);
	}

	/** @param label The {@link TKCorrectableLabel} to pair with. */
	public void setLabel(TKCorrectableLabel label) {
		mLabel = label;
		if (mLabel != null) {
			mLabel.setLink(this);
		}
	}

	public boolean wasCorrected() {
		return mWasCorrected;
	}

	public void clearCorrectionState() {
		if (mWasCorrected) {
			mWasCorrected = false;
			setToolTipText(mSavedToolTip);
			mSavedToolTip = null;
			if (mLabel != null) {
				mLabel.repaint();
			}
		}
	}

	public void correct(Object correction, String reason) {
		if (mSavedToolTip == null) {
			mSavedToolTip = getToolTipText();
		}
		mWasCorrected = true;
		setToolTipText(reason);
		setText(correction.toString());
		if (mLabel != null) {
			mLabel.repaint();
		}
		TKCorrectableManager.getInstance().register(this);
	}
}
