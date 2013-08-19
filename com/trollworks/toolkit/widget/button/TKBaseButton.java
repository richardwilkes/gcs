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

import com.trollworks.toolkit.utility.TKFont;
import com.trollworks.toolkit.widget.TKLabel;
import com.trollworks.toolkit.widget.border.TKBorder;

import java.awt.AWTEvent;
import java.awt.Color;
import java.awt.Graphics2D;
import java.awt.event.MouseEvent;
import java.awt.image.BufferedImage;

/** The base class for buttons. */
public class TKBaseButton extends TKLabel {
	/** <code>true</code> if the button is currently pressed. */
	protected boolean		mPressed;
	/** <code>true</code> if the we're currently tracking a press of the button. */
	protected boolean		mInPress;
	/** <code>true</code> if we're currently in a mouse roll over. */
	protected boolean		mInRollOver;
	private TKBorder		mPressedBorder;
	private boolean			mDisplayBorder;
	private boolean			mDisplayBorderOnRollOver;
	private BufferedImage	mRollOverIcon;
	private BufferedImage	mPressedIcon;
	private BufferedImage	mPressedRollOverIcon;
	private boolean			mTurnBorderOff;

	/**
	 * Creates a button.
	 * 
	 * @param text The text to be displayed.
	 * @param icon The icon to be displayed.
	 * @param alignment The horizontal alignment to use.
	 */
	public TKBaseButton(String text, BufferedImage icon, int alignment) {
		super(text, icon, alignment, false, TKFont.CONTROL_FONT_KEY);
		setOpaque(true);
		setDisabledForeground(Color.lightGray);
		enableAWTEvents(AWTEvent.MOUSE_EVENT_MASK);
	}

	/** Handle a click on the button. Notifies action listeners by default. */
	public void doClick() {
		notifyActionListeners();
	}

	/** @return The current icon being displayed, if any. */
	@Override public BufferedImage getCurrentImage() {
		if (isEnabled()) {
			if (isPressed()) {
				return mInRollOver ? getPressedRollOverIcon() : getPressedIcon();
			}
			return mInRollOver ? getRollOverIcon() : getImage();
		}
		return getDisabledImage();
	}

	/**
	 * @return The border of this component when it is in the "pressed" state or <code>null</code>
	 *         if no pressed border is currently set.
	 */
	public TKBorder getPressedBorder() {
		return mPressedBorder;
	}

	/** @return The pressed icon. */
	public BufferedImage getPressedIcon() {
		return mPressedIcon != null ? mPressedIcon : getImage();
	}

	/** @return The pressed icon. */
	public BufferedImage getPressedRollOverIcon() {
		return mPressedRollOverIcon != null ? mPressedRollOverIcon : getPressedIcon();
	}

	/** @return The mouse roll over icon. */
	public BufferedImage getRollOverIcon() {
		return mRollOverIcon != null ? mRollOverIcon : getImage();
	}

	/** @return <code>true</code> if the border should be displayed. */
	public boolean isBorderDisplayed() {
		return mDisplayBorder;
	}

	/**
	 * @return <code>true</code> if the border should be displayed on mouse roll over.
	 */
	public boolean isBorderDisplayedOnRollOver() {
		return mDisplayBorderOnRollOver;
	}

	/** @return <code>true</code> if the button is currently pressed. */
	public boolean isPressed() {
		return mPressed;
	}

	@Override protected void paintBorder(Graphics2D g2d) {
		if (isBorderDisplayed()) {
			TKBorder border = isPressed() ? getPressedBorder() : getBorder();

			if (border != null) {
				border.paintBorder(this, g2d, 0, 0, getWidth(), getHeight());
			}
		}
	}

	@Override public void processMouseEventSelf(MouseEvent event) {
		switch (event.getID()) {
			case MouseEvent.MOUSE_PRESSED:
				mInPress = true;
				setPressed(true);
				break;
			case MouseEvent.MOUSE_RELEASED:
				if (isPressed()) {
					unpress();
					doClick();
				}
				mInPress = false;
				break;
			case MouseEvent.MOUSE_ENTERED:
				mInRollOver = true;
				if (isBorderDisplayedOnRollOver() && !isBorderDisplayed()) {
					setBorderDisplayed(true);
					mTurnBorderOff = true;
				}
				if (mInPress) {
					setPressed(true);
				}

				repaint();
				break;
			case MouseEvent.MOUSE_EXITED:
				mInRollOver = false;
				unpress();
				break;
		}
	}

	@Override public void setEnabled(boolean enabled) {
		if (!enabled) {
			mInRollOver = false;
		}
		super.setEnabled(enabled);
	}

	/** @param enabled Whether the border should be displayed or not. */
	public void setBorderDisplayed(boolean enabled) {
		if (mDisplayBorder != enabled) {
			mDisplayBorder = enabled;
			repaint();
		}
	}

	/**
	 * @param enabled Whether the border should be displayed on mouse roll over or not.
	 */
	public void setBorderDisplayedOnRollOver(boolean enabled) {
		mDisplayBorderOnRollOver = enabled;
	}

	/** @param pressed The "pressed" state of the button. */
	public void setPressed(boolean pressed) {
		if (pressed != mPressed) {
			mPressed = pressed;
			repaint();
		}
	}

	/**
	 * @param border The border of this component when it is in the "pressed" state.
	 */
	public void setPressedBorder(TKBorder border) {
		if (border != mPressedBorder) {
			TKBorder oldBorder = mPressedBorder;

			mPressedBorder = border;
			if (border == null || oldBorder == null || !border.getBorderInsets(this).equals(oldBorder.getBorderInsets(this))) {
				revalidate();
			}
			repaint();
		}
	}

	/** @param icon The pressed icon. */
	public void setPressedIcon(BufferedImage icon) {
		BufferedImage oldValue = mPressedIcon;

		mPressedIcon = icon;
		if (mPressedIcon != oldValue) {
			if (mPressedIcon == null || oldValue == null || mPressedIcon.getWidth() != oldValue.getWidth() || mPressedIcon.getHeight() != oldValue.getHeight()) {
				revalidate();
			}
			repaint();
		}
	}

	/** @param icon The pressed rollover icon. */
	public void setPressedRollOverIcon(BufferedImage icon) {
		BufferedImage oldValue = mPressedRollOverIcon;

		mPressedRollOverIcon = icon;
		if (mPressedRollOverIcon != oldValue) {
			if (mPressedRollOverIcon == null || oldValue == null || mPressedRollOverIcon.getWidth() != oldValue.getWidth() || mPressedRollOverIcon.getHeight() != oldValue.getHeight()) {
				revalidate();
			}
			repaint();
		}
	}

	/** @param icon The mouse roll over icon. */
	public void setRollOverIcon(BufferedImage icon) {
		BufferedImage oldValue = mRollOverIcon;

		mRollOverIcon = icon;
		if (mRollOverIcon != oldValue) {
			if (mRollOverIcon == null || oldValue == null || mRollOverIcon.getWidth() != oldValue.getWidth() || mRollOverIcon.getHeight() != oldValue.getHeight()) {
				revalidate();
			}
			repaint();
		}
	}

	/** "Unpress" the button. */
	protected void unpress() {
		if (mInPress) {
			setPressed(false);
		}
		if (mTurnBorderOff) {
			setBorderDisplayed(false);
		}
		repaint();
	}
}
