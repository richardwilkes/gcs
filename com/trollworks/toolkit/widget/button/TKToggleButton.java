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
import com.trollworks.toolkit.widget.border.TKBorder;

import java.awt.Graphics2D;
import java.awt.Rectangle;
import java.awt.image.BufferedImage;

/** A standard toggle button. */
public class TKToggleButton extends TKButton {
	private boolean			mSelected;
	private boolean			mPushButtonMode;
	private BufferedImage	mDisabledPressedIcon;

	/** Creates an unselected button with no icon and no text. */
	public TKToggleButton() {
		this(null, null, false, TKAlignment.CENTER);
	}

	/**
	 * Creates a button with no icon and no text.
	 * 
	 * @param selected The initially selected state.
	 */
	public TKToggleButton(boolean selected) {
		this(null, null, selected, TKAlignment.CENTER);
	}

	/**
	 * Creates an unselected button with the specified text.
	 * 
	 * @param text The text to use.
	 */
	public TKToggleButton(String text) {
		this(text, null, false, TKAlignment.CENTER);
	}

	/**
	 * Creates a button with the specified text.
	 * 
	 * @param text The text to use.
	 * @param selected The initially selected state.
	 */
	public TKToggleButton(String text, boolean selected) {
		this(text, null, selected, TKAlignment.CENTER);
	}

	/**
	 * Creates an unselected button with the specified text and horizontal alignment.
	 * 
	 * @param text The text to use.
	 * @param alignment The horizontal alignment to use.
	 */
	public TKToggleButton(String text, int alignment) {
		this(text, null, false, alignment);
	}

	/**
	 * Creates a button with the specified text and horizontal alignment.
	 * 
	 * @param text The text to use.
	 * @param selected The initially selected state.
	 * @param alignment The horizontal alignment to use.
	 */
	public TKToggleButton(String text, boolean selected, int alignment) {
		this(text, null, selected, alignment);
	}

	/**
	 * Creates an unselected button with the specified icon.
	 * 
	 * @param icon The icon to use.
	 */
	public TKToggleButton(BufferedImage icon) {
		this(null, icon, false, TKAlignment.CENTER);
	}

	/**
	 * Creates a button with the specified icon.
	 * 
	 * @param icon The icon to use.
	 * @param selected The initially selected state.
	 */
	public TKToggleButton(BufferedImage icon, boolean selected) {
		this(null, icon, selected, TKAlignment.CENTER);
	}

	/**
	 * Creates an unselected button with the specified icon and horizontal alignment.
	 * 
	 * @param icon The icon to use.
	 * @param alignment The horizontal alignment to use.
	 */
	public TKToggleButton(BufferedImage icon, int alignment) {
		this(null, icon, false, alignment);
	}

	/**
	 * Creates a button with the specified icon and horizontal alignment.
	 * 
	 * @param icon The icon to use.
	 * @param selected The initially selected state.
	 * @param alignment The horizontal alignment to use.
	 */
	public TKToggleButton(BufferedImage icon, boolean selected, int alignment) {
		this(null, icon, selected, alignment);
	}

	/**
	 * Creates an unselected button with the specified text and icon.
	 * 
	 * @param text The text to use.
	 * @param icon The icon to use.
	 */
	public TKToggleButton(String text, BufferedImage icon) {
		this(text, icon, false, TKAlignment.CENTER);
	}

	/**
	 * Creates a button with the specified text and icon.
	 * 
	 * @param text The text to use.
	 * @param icon The icon to use.
	 * @param selected The initially selected state.
	 */
	public TKToggleButton(String text, BufferedImage icon, boolean selected) {
		this(text, icon, selected, TKAlignment.CENTER);
	}

	/**
	 * Creates an unselected button with the specified text, icon and horizontal alignment.
	 * 
	 * @param text The text to use.
	 * @param icon The icon to use.
	 * @param alignment The horizontal alignment to use.
	 */
	public TKToggleButton(String text, BufferedImage icon, int alignment) {
		this(text, icon, false, alignment);
	}

	/**
	 * Creates a button with the specified text, icon and horizontal alignment.
	 * 
	 * @param text The text to use.
	 * @param icon The icon to use.
	 * @param selected The initially selected state.
	 * @param alignment The horizontal alignment to use.
	 */
	public TKToggleButton(String text, BufferedImage icon, boolean selected, int alignment) {
		super(text, icon, alignment);
		mSelected = !selected; // Ensure the following call will "take".
		setSelected(selected);
	}

	@Override public void doClick() {
		setSelected(mPushButtonMode ? true : !isSelected());
	}

	/** @return <code>true</code> if this button is selected. */
	public boolean isSelected() {
		return mSelected;
	}

	@Override public BufferedImage getCurrentImage() {
		if (isSelected()) {
			if (isEnabled()) {
				return mInRollOver ? getPressedRollOverIcon() : getPressedIcon();
			}
			return getDisabledPressedIcon();
		}

		return super.getCurrentImage();
	}

	/** @return The disabled pressed icon. */
	public BufferedImage getDisabledPressedIcon() {
		if (mDisabledPressedIcon == null) {
			mDisabledPressedIcon = TKImage.createDisabledImage(getPressedIcon());
		}
		return mDisabledPressedIcon;
	}

	/** @return <code>true</code> if push button mode is on. */
	public boolean getPushButtonMode() {
		return mPushButtonMode;
	}

	@Override protected void paintBorder(Graphics2D g2d) {
		if (isBorderDisplayed()) {
			TKBorder border = isPressed() || isSelected() ? getPressedBorder() : getBorder();

			if (border != null) {
				border.paintBorder(this, g2d, 0, 0, getWidth(), getHeight());
			}
		}
	}

	@Override protected void paintPanel(Graphics2D g2d, Rectangle[] clips) {
		boolean isPressed = mPressed;

		if (isOpaque() && isSelected() && !isPressed) {
			mPressed = true;
			g2d.setColor(TKColor.darker(getBackground(), 25));
			g2d.fillRect(0, 0, getWidth(), getHeight());
		}
		super.paintPanel(g2d, clips);
		mPressed = isPressed;
	}

	/**
	 * Turns push button mode on/off.
	 * 
	 * @param enabled Whether to enable or disable push button mode.
	 */
	public void setPushButtonMode(boolean enabled) {
		mPushButtonMode = enabled;
	}

	/**
	 * Sets the selected state of the button.
	 * 
	 * @param selected The new selected state.
	 */
	public void setSelected(boolean selected) {
		if (mSelected != selected) {
			mSelected = selected;
			repaint();
			notifyActionListeners();
		}
	}
}
