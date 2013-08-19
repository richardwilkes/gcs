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

import com.trollworks.toolkit.text.TKDocument;
import com.trollworks.toolkit.text.TKDocumentListener;
import com.trollworks.toolkit.utility.TKAlignment;
import com.trollworks.toolkit.utility.TKFont;
import com.trollworks.toolkit.widget.border.TKCompoundBorder;
import com.trollworks.toolkit.widget.border.TKEmptyBorder;
import com.trollworks.toolkit.widget.border.TKLineBorder;
import com.trollworks.toolkit.widget.layout.TKColumnLayout;
import com.trollworks.toolkit.window.TKBaseWindow;
import com.trollworks.toolkit.window.TKOptionDialog;
import com.trollworks.toolkit.window.TKDialog;

import java.awt.AWTEvent;
import java.awt.Color;
import java.awt.Dimension;
import java.awt.GradientPaint;
import java.awt.Graphics2D;
import java.awt.Polygon;
import java.awt.Rectangle;
import java.awt.event.InputEvent;
import java.awt.event.KeyEvent;
import java.awt.event.MouseEvent;

/** Provides a standard color chooser. */
public class TKColorChooser extends TKPanel implements TKDocumentListener, TKKeyEventFilter {
	private static final int	DIAMOND_SIZE	= 11;
	private Color				mColor;
	private TKPanel				mCurrentColorDisplay;
	private TKTextField			mRedField;
	private TKTextField			mGreenField;
	private TKTextField			mBlueField;
	private ColorBar			mRedColorBar;
	private ColorBar			mGreenColorBar;
	private ColorBar			mBlueColorBar;

	/**
	 * Creates a new color chooser panel.
	 * 
	 * @param initialColor The initial color to have the panel start with.
	 */
	protected TKColorChooser(Color initialColor) {
		super(new TKColumnLayout());

		mColor = initialColor != null ? initialColor : Color.black;

		TKPanel originalColorDisplay = new OriginalWell();
		originalColorDisplay.setBorder(new TKEmptyBorder(0, 10, 0, 10));
		originalColorDisplay.setPreferredSize(new Dimension(20, 20));
		originalColorDisplay.setOpaque(true);
		originalColorDisplay.setBackground(mColor);
		originalColorDisplay.setBorder(new TKCompoundBorder(new TKCompoundBorder(TKLineBorder.getSharedBorder(false), TKLineBorder.getSharedBorder(true)), TKLineBorder.getSharedBorder(false)));

		mCurrentColorDisplay = new TKPanel();
		mCurrentColorDisplay.setBorder(new TKEmptyBorder(0, 10, 0, 10));
		mCurrentColorDisplay.setPreferredSize(originalColorDisplay.getPreferredSize());
		mCurrentColorDisplay.setOpaque(true);
		mCurrentColorDisplay.setBackground(mColor);
		mCurrentColorDisplay.setBorder(originalColorDisplay.getBorder());

		TKPanel panel = new TKPanel(new TKColumnLayout(2));
		panel.setBorder(new TKEmptyBorder(0, 0, 10, 0));
		panel.add(new TKLabel(Msgs.ORIGINAL, null, TKAlignment.CENTER, false, TKFont.CONTROL_FONT_KEY));
		panel.add(new TKLabel(Msgs.NEW, null, TKAlignment.CENTER, false, TKFont.CONTROL_FONT_KEY));
		panel.add(originalColorDisplay);
		panel.add(mCurrentColorDisplay);
		add(panel);

		panel = new TKPanel(new TKColumnLayout(3));
		panel.add(new TKLabel(Msgs.RED, null, TKAlignment.RIGHT, false, TKFont.CONTROL_FONT_KEY));
		mRedField = createTextField(mColor.getRed());
		panel.add(mRedField);
		mRedColorBar = new ColorBar();
		panel.add(mRedColorBar);

		panel.add(new TKLabel(Msgs.GREEN, null, TKAlignment.RIGHT, false, TKFont.CONTROL_FONT_KEY));
		mGreenField = createTextField(mColor.getGreen());
		panel.add(mGreenField);
		mGreenColorBar = new ColorBar();
		panel.add(mGreenColorBar);

		panel.add(new TKLabel(Msgs.BLUE, null, TKAlignment.RIGHT, false, TKFont.CONTROL_FONT_KEY));
		mBlueField = createTextField(mColor.getBlue());
		panel.add(mBlueField);
		mBlueColorBar = new ColorBar();
		panel.add(mBlueColorBar);
		add(panel);

		adjustColorBars();
	}

	private void adjustColorBars() {
		int red = mColor.getRed();
		int green = mColor.getGreen();
		int blue = mColor.getBlue();

		mRedColorBar.setValue(red);
		mRedColorBar.setColorRange(new Color(0, green, blue), new Color(255, green, blue));

		mGreenColorBar.setValue(green);
		mGreenColorBar.setColorRange(new Color(red, 0, blue), new Color(red, 255, blue));

		mBlueColorBar.setValue(blue);
		mBlueColorBar.setColorRange(new Color(red, green, 0), new Color(red, green, 255));
	}

	/**
	 * Called by the color bar when it changes.
	 * 
	 * @param bar The color bar to use.
	 * @param value The new value.
	 */
	protected void adjustForOneColorBar(ColorBar bar, int value) {
		if (bar == null) {
			mRedField.setText(String.valueOf(value));
			mGreenField.setText(String.valueOf(value));
			mBlueField.setText(String.valueOf(value));
		} else if (bar == mRedColorBar) {
			mRedField.setText(String.valueOf(value));
		} else if (bar == mGreenColorBar) {
			mGreenField.setText(String.valueOf(value));
		} else {
			mBlueField.setText(String.valueOf(value));
		}
	}

	/**
	 * Display the color chooser dialog and wait for the user to pick a color.
	 * 
	 * @param initialColor The initial color to have the chooser start with.
	 * @return The color chosen, or <code>null</code> if the dialog was cancelled.
	 */
	public static Color chooseColor(Color initialColor) {
		return chooseColor(null, initialColor);
	}

	/**
	 * Display the color chooser dialog and wait for the user to pick a color.
	 * 
	 * @param owner The owning window/dialog.
	 * @param initialColor The initial color to have the chooser start with.
	 * @return The color chosen, or <code>null</code> if the dialog was cancelled.
	 */
	public static Color chooseColor(TKBaseWindow owner, Color initialColor) {
		TKColorChooser chooser = new TKColorChooser(initialColor);

		if (TKOptionDialog.modal(owner, null, Msgs.CHOOSE, TKOptionDialog.TYPE_OK_CANCEL, chooser) == TKDialog.OK) {
			return chooser.mColor;
		}
		return null;
	}

	private TKTextField createTextField(int value) {
		TKTextField field = new TKTextField(String.valueOf(value), 35);

		field.selectAll();
		field.setKeyEventFilter(this);
		field.addDocumentListener(this);
		return field;
	}

	public void documentChanged(TKDocument document) {
		int red = getValue(mRedField.getText());
		int green = getValue(mGreenField.getText());
		int blue = getValue(mBlueField.getText());

		mColor = new Color(red, green, blue);
		mCurrentColorDisplay.setBackground(mColor);
		adjustColorBars();
	}

	public boolean filterKeyEvent(TKPanel owner, KeyEvent event, boolean isReal) {
		if (event.getID() == KeyEvent.KEY_TYPED) {
			char ch = event.getKeyChar();

			return ch != '\b' && ch != KeyEvent.VK_DELETE && (ch < '0' || ch > '9');
		}
		return false;
	}

	private int getValue(String buffer) {
		if (buffer.length() == 0) {
			return 0;
		}

		try {
			int value = Integer.valueOf(buffer).intValue();

			if (value < 0) {
				value = 0;
			} else if (value > 255) {
				value = 255;
			}
			return value;
		} catch (NumberFormatException nfe) {
			return 0;
		}
	}

	/** @param color The current color value. */
	protected void setValue(Color color) {
		int red = color.getRed();
		int green = color.getGreen();
		int blue = color.getBlue();

		mRedField.setText(String.valueOf(red));
		mGreenField.setText(String.valueOf(green));
		mBlueField.setText(String.valueOf(blue));
	}

	private class ColorBar extends TKPanel {
		private int		mColorBarValue;
		private Color	mColorBarStart;
		private Color	mColorBarEnd;

		/** Creates a new color bar. */
		ColorBar() {
			super();
			enableAWTEvents(AWTEvent.MOUSE_EVENT_MASK | AWTEvent.MOUSE_MOTION_EVENT_MASK);
		}

		private void adjustForMousePosition(int x, boolean doAll) {
			x -= (getWidth() - 254) / 2;
			if (x < 0) {
				x = 0;
			} else if (x > 255) {
				x = 255;
			}
			adjustForOneColorBar(doAll ? null : this, x);
		}

		@Override protected Dimension getMaximumSizeSelf() {
			return getPreferredSizeSelf();
		}

		@Override protected Dimension getMinimumSizeSelf() {
			return getPreferredSizeSelf();
		}

		@Override protected Dimension getPreferredSizeSelf() {
			return new Dimension(255 + DIAMOND_SIZE, DIAMOND_SIZE);
		}

		@Override protected void paintPanel(Graphics2D g2d, Rectangle[] clips) {
			int middle = DIAMOND_SIZE / 2;
			float fMiddle = middle;
			int x[] = new int[] { middle + mColorBarValue, DIAMOND_SIZE + mColorBarValue - 1, middle + mColorBarValue, mColorBarValue };
			int y[] = new int[] { 0, middle, DIAMOND_SIZE - 1, middle };
			Polygon diamond = new Polygon(x, y, 4);

			g2d.translate((getWidth() - (254 + DIAMOND_SIZE)) / 2, (getHeight() - DIAMOND_SIZE) / 2);
			g2d.setPaint(new GradientPaint(fMiddle, fMiddle, mColorBarStart, 255f + fMiddle, fMiddle, mColorBarEnd));
			g2d.fillRect(middle, 0, 255, DIAMOND_SIZE);
			g2d.setPaint(Color.black);
			g2d.drawRect(middle, 0, 255, DIAMOND_SIZE - 1);
			g2d.setPaint(Color.white);
			g2d.fill(diamond);
			g2d.setPaint(Color.black);
			g2d.draw(diamond);
		}

		@Override public void processMouseEventSelf(MouseEvent event) {
			int id = event.getID();

			if (id == MouseEvent.MOUSE_PRESSED || id == MouseEvent.MOUSE_DRAGGED) {
				adjustForMousePosition(event.getX(), (event.getModifiers() & InputEvent.SHIFT_MASK) != 0);
			}
		}

		/**
		 * Sets the range of colors.
		 * 
		 * @param startColor The starting color.
		 * @param endColor The ending color.
		 */
		void setColorRange(Color startColor, Color endColor) {
			if (!startColor.equals(mColorBarStart) || !endColor.equals(mColorBarEnd)) {
				mColorBarStart = startColor;
				mColorBarEnd = endColor;
				repaint();
			}
		}

		/** @return The color bar's value. */
		int getValue() {
			return mColorBarValue;
		}

		/**
		 * Sets the color bar's value.
		 * 
		 * @param value The value to set.
		 */
		void setValue(int value) {
			if (mColorBarValue != value) {
				mColorBarValue = value;
				repaint();
			}
		}
	}

	private class OriginalWell extends TKPanel {
		/** Creates a color well with the original color. */
		OriginalWell() {
			super();
			enableAWTEvents(AWTEvent.MOUSE_EVENT_MASK);
		}

		@Override public void processMouseEventSelf(MouseEvent event) {
			if (event.getID() == MouseEvent.MOUSE_PRESSED) {
				setValue(getBackground());
			}
		}
	}
}
