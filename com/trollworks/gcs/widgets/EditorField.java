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

import com.trollworks.gcs.utility.text.TextDrawing;

import java.awt.Color;
import java.awt.Graphics;
import java.awt.Rectangle;
import java.awt.event.FocusEvent;
import java.beans.PropertyChangeListener;

import javax.swing.JFormattedTextField;
import javax.swing.SwingConstants;
import javax.swing.UIManager;

/** Provides a standard editor field. */
public class EditorField extends JFormattedTextField {
	private String	mHint;

	/**
	 * Creates a new {@link EditorField}.
	 * 
	 * @param formatter The formatter to use.
	 * @param listener The listener to use.
	 * @param alignment The alignment to use.
	 * @param value The initial value.
	 * @param tooltip The tooltip to use.
	 */
	public EditorField(AbstractFormatterFactory formatter, PropertyChangeListener listener, int alignment, Object value, String tooltip) {
		this(formatter, listener, alignment, value, null, tooltip);
	}

	/**
	 * Creates a new {@link EditorField}.
	 * 
	 * @param formatter The formatter to use.
	 * @param listener The listener to use.
	 * @param alignment The alignment to use.
	 * @param value The initial value.
	 * @param protoValue The prototype value to use to set the preferred size.
	 * @param tooltip The tooltip to use.
	 */
	public EditorField(AbstractFormatterFactory formatter, PropertyChangeListener listener, int alignment, Object value, Object protoValue, String tooltip) {
		super(formatter, protoValue != null ? protoValue : value);
		setHorizontalAlignment(alignment);
		setToolTipText(tooltip);
		if (protoValue != null) {
			setPreferredSize(getPreferredSize());
			setValue(value);
		}
		if (listener != null) {
			addPropertyChangeListener("value", listener); //$NON-NLS-1$
		}

		// Reset the selection colors back to what is standard for text fields.
		// This is necessary, since (at least on the Mac) JFormattedTextField
		// has the wrong values by default.
		setCaretColor(UIManager.getColor("TextField.caretForeground")); //$NON-NLS-1$
		setSelectionColor(UIManager.getColor("TextField.selectionBackground")); //$NON-NLS-1$
		setSelectedTextColor(UIManager.getColor("TextField.selectionForeground")); //$NON-NLS-1$
		setDisabledTextColor(UIManager.getColor("TextField.inactiveForeground")); //$NON-NLS-1$
	}

	@Override protected void processFocusEvent(FocusEvent event) {
		super.processFocusEvent(event);
		if (event.getID() == FocusEvent.FOCUS_GAINED) {
			selectAll();
		}
	}

	/** @param hint The hint to use, or <code>null</code>. */
	public void setHint(String hint) {
		mHint = hint;
		repaint();
	}

	@Override protected void paintComponent(Graphics gc) {
		super.paintComponent(gc);
		if (mHint != null && getText().length() == 0) {
			Rectangle bounds = getBounds();
			bounds.x = 0;
			bounds.y = 0;
			gc.setColor(Color.GRAY);
			TextDrawing.draw(gc, bounds, mHint, SwingConstants.CENTER, SwingConstants.CENTER);
		}
	}
}
