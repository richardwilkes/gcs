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
import com.trollworks.toolkit.utility.TKColor;

import java.awt.Color;
import java.awt.Graphics2D;
import java.awt.Insets;
import java.awt.Rectangle;
import java.awt.image.BufferedImage;

/** A standard check box. */
public class TKCheckbox extends TKBaseButton {
	/** Constant for the mixed check box state. */
	public static final int	MIXED		= -1;
	/** Constant for the unchecked check box state. */
	public static final int	UNCHECKED	= 0;
	/** Constant for the checked check box state. */
	public static final int	CHECKED		= 1;
	private int				mCheckedState;

	/** Creates an unchecked check box with no text. */
	public TKCheckbox() {
		this(null, UNCHECKED, TKAlignment.LEFT);
	}

	/**
	 * Creates a check box with no text.
	 * 
	 * @param selected Whether or not the checkbox is initially selected.
	 */
	public TKCheckbox(boolean selected) {
		this(null, selected ? CHECKED : UNCHECKED, TKAlignment.LEFT);
	}

	/**
	 * Creates a check box with no text.
	 * 
	 * @param value One of {@link #CHECKED}, {@link #UNCHECKED}, or {@link #MIXED}.
	 */
	public TKCheckbox(int value) {
		this(null, value, TKAlignment.LEFT);
	}

	/**
	 * Creates an unchecked check box with the specified text.
	 * 
	 * @param text The text to use.
	 */
	public TKCheckbox(String text) {
		this(text, UNCHECKED, TKAlignment.LEFT);
	}

	/**
	 * Creates a check box with the specified text.
	 * 
	 * @param text The text to use.
	 * @param selected Whether or not the checkbox is initially selected.
	 */
	public TKCheckbox(String text, boolean selected) {
		this(text, selected ? CHECKED : UNCHECKED, TKAlignment.LEFT);
	}

	/**
	 * Creates a check box with the specified text.
	 * 
	 * @param text The text to use.
	 * @param value One of {@link #CHECKED}, {@link #UNCHECKED}, or {@link #MIXED}.
	 */
	public TKCheckbox(String text, int value) {
		this(text, value, TKAlignment.LEFT);
	}

	/**
	 * Creates a check box with the specified text and horizontal alignment.
	 * 
	 * @param text The text to use.
	 * @param selected Whether or not the checkbox is initially selected.
	 * @param alignment The horizontal alignment to use.
	 */
	public TKCheckbox(String text, boolean selected, int alignment) {
		this(text, selected ? CHECKED : UNCHECKED, alignment);
	}

	/**
	 * Creates a check box with the specified text and horizontal alignment.
	 * 
	 * @param text The text to use.
	 * @param value One of {@link #CHECKED}, {@link #UNCHECKED}, or {@link #MIXED}.
	 * @param alignment The horizontal alignment to use.
	 */
	public TKCheckbox(String text, int value, int alignment) {
		super(text, TKImage.getCheckboxEmptyIcon(), alignment);
		mCheckedState = ~value; // Ensure the following call will "take".
		setCheckedState(value);
	}

	@Override public void doClick() {
		setCheckedState(getCheckedState() <= UNCHECKED ? CHECKED : UNCHECKED);
	}

	/**
	 * @return The checked state of the check box. One of {@link #CHECKED}, {@link #UNCHECKED}, or
	 *         {@link #MIXED}.
	 */
	public int getCheckedState() {
		return mCheckedState;
	}

	/** @return <code>true</code> only if the state is {@link #CHECKED}. */
	public boolean isChecked() {
		return mCheckedState == CHECKED;
	}

	/**
	 * Sets the checked state of the check box and notifies action listeners.
	 * 
	 * @param selected Whether or not the checkbox should be selected.
	 */
	public void setCheckedState(boolean selected) {
		setCheckedState(selected ? CHECKED : UNCHECKED);
	}

	/**
	 * Sets the checked state of the check box and notifies action listeners.
	 * 
	 * @param value One of {@link #CHECKED}, {@link #UNCHECKED}, or {@link #MIXED}.
	 */
	public void setCheckedState(int value) {
		if (mCheckedState != value) {
			BufferedImage normal;
			BufferedImage pressed;

			mCheckedState = value;
			if (value < UNCHECKED) {
				normal = TKImage.getCheckboxMixedIcon();
				pressed = TKImage.getCheckboxPressedMixedIcon();
			} else if (value > UNCHECKED) {
				normal = TKImage.getCheckboxCheckedIcon();
				pressed = TKImage.getCheckboxPressedCheckedIcon();
			} else {
				normal = TKImage.getCheckboxEmptyIcon();
				pressed = TKImage.getCheckboxPressedEmptyIcon();
			}
			setImage(normal);
			setPressedIcon(pressed);
			notifyActionListeners();
		}
	}

	@Override protected void paintPanel(Graphics2D g2d, Rectangle[] clips) {
		Color oldColor = g2d.getColor();
		Insets insets = getInsets();
		int x = insets.left;
		int y = insets.top;
		int w = getWidth() - 1;
		int h = getHeight() - 1;

		g2d.setColor(mInRollOver ? TKColor.CONTROL_ROLL : TKColor.CONTROL_FILL);
		g2d.fillRect(x, y, w, h);

		g2d.setColor(oldColor);
		super.paintPanel(g2d, clips);
	}
}
