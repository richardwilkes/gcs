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
 * 2005-2008 the Initial Developer. All Rights Reserved.
 *
 * Contributor(s):
 *
 * ***** END LICENSE BLOCK ***** */

package com.trollworks.gcs.widgets;

import java.awt.Color;
import java.beans.PropertyChangeEvent;
import java.beans.PropertyChangeListener;

import javax.swing.ImageIcon;
import javax.swing.JComponent;
import javax.swing.JLabel;
import javax.swing.SwingConstants;
import javax.swing.UIManager;

/** A label whose tooltip reflects that of another panel. */
public class LinkedLabel extends JLabel implements PropertyChangeListener {
	/** The property key that is monitored for error messages. */
	public static final String	ERROR_MESSAGE_KEY	= "Error Message";	//$NON-NLS-1$
	private JComponent			mLink;
	private Color				mColor;

	/**
	 * Creates a label with the specified icon.
	 * 
	 * @param icon The icon to be displayed.
	 */
	public LinkedLabel(ImageIcon icon) {
		super(icon);
		mColor = getForeground();
	}

	/**
	 * Creates a label with the specified icon.
	 * 
	 * @param icon The icon to be displayed.
	 * @param link The {@link JComponent} to pair with.
	 */
	public LinkedLabel(ImageIcon icon, JComponent link) {
		super(icon);
		mColor = getForeground();
		setLink(link);
	}

	/**
	 * Creates a label with the specified text. The label is right-aligned and centered vertically
	 * in its display area.
	 * 
	 * @param text The text to be displayed.
	 */
	public LinkedLabel(String text) {
		this(text, SwingConstants.RIGHT);
	}

	/**
	 * Creates a label with the specified text and alignment.
	 * 
	 * @param text The text to be displayed.
	 * @param alignment The horizontal alignment to use.
	 */
	public LinkedLabel(String text, int alignment) {
		super(text, alignment);
		mColor = getForeground();
	}

	/**
	 * Creates a label with the specified text. The label is right-aligned and centered vertically
	 * in its display area.
	 * 
	 * @param text The text to be displayed.
	 * @param font The dynamic font to use.
	 */
	public LinkedLabel(String text, String font) {
		this(text);
		setFont(UIManager.getFont(font));
	}

	/**
	 * Creates a label with the specified text. The label is right-aligned and centered vertically
	 * in its display area.
	 * 
	 * @param text The text to be displayed.
	 * @param font The dynamic font to use.
	 * @param alignment The horizontal alignment to use.
	 */
	public LinkedLabel(String text, String font, int alignment) {
		this(text, alignment);
		setFont(UIManager.getFont(font));
	}

	/**
	 * Creates a label with the specified text. The label is right-aligned and centered vertically
	 * in its display area.
	 * 
	 * @param text The text to be displayed.
	 * @param link The {@link JComponent} to pair with.
	 */
	public LinkedLabel(String text, JComponent link) {
		this(text);
		setLink(link);
	}

	/**
	 * Creates a label with the specified text. The label is right-aligned and centered vertically
	 * in its display area.
	 * 
	 * @param text The text to be displayed.
	 * @param font The dynamic font to use.
	 * @param link The {@link JComponent} to pair with.
	 */
	public LinkedLabel(String text, String font, JComponent link) {
		this(text, font);
		setLink(link);
	}

	/**
	 * Creates a label with the specified text. The label is right-aligned and centered vertically
	 * in its display area.
	 * 
	 * @param text The text to be displayed.
	 * @param font The dynamic font to use.
	 * @param link The {@link JComponent} to pair with.
	 * @param alignment The horizontal alignment to use.
	 */
	public LinkedLabel(String text, String font, JComponent link, int alignment) {
		this(text, font, alignment);
		setLink(link);
	}

	/** @return The {@link JComponent} that is being paired with. */
	public JComponent getLink() {
		return mLink;
	}

	/** @param link The {@link JComponent} to pair with. */
	public void setLink(JComponent link) {
		if (mLink != link) {
			if (mLink != null) {
				mLink.removePropertyChangeListener(TOOL_TIP_TEXT_KEY, this);
				mLink.removePropertyChangeListener(ERROR_MESSAGE_KEY, this);
			}
			mLink = link;
			if (mLink != null) {
				mLink.addPropertyChangeListener(TOOL_TIP_TEXT_KEY, this);
				mLink.addPropertyChangeListener(ERROR_MESSAGE_KEY, this);
			}
			adjustToLink();
		}
	}

	private void adjustToLink() {
		String tooltip;
		Color color;

		if (mLink != null) {
			tooltip = (String) mLink.getClientProperty(ERROR_MESSAGE_KEY);
			if (tooltip == null) {
				tooltip = mLink.getToolTipText();
				color = mColor;
			} else {
				color = Color.RED;
			}
		} else {
			tooltip = null;
			color = mColor;
		}
		setToolTipText(tooltip);
		super.setForeground(color);
	}

	public void propertyChange(PropertyChangeEvent event) {
		adjustToLink();
	}

	@Override public void setForeground(Color color) {
		mColor = color;
		if (mLink == null || mLink.getClientProperty(ERROR_MESSAGE_KEY) == null) {
			super.setForeground(color);
		}
	}

	/**
	 * Sets/clears the error message that a {@link LinkedLabel} will respond to.
	 * 
	 * @param comp The {@link JComponent} to set the message on.
	 * @param msg The error message or <code>null</code>.
	 */
	public static void setErrorMessage(JComponent comp, String msg) {
		Object old = comp.getClientProperty(ERROR_MESSAGE_KEY);
		if (old == null ? msg != null : !old.equals(msg)) {
			comp.putClientProperty(ERROR_MESSAGE_KEY, msg);
		}
	}
}
