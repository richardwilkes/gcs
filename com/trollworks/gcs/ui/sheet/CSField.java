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

package com.trollworks.gcs.ui.sheet;

import com.trollworks.gcs.model.CMCharacter;
import com.trollworks.gcs.ui.common.CSFont;
import com.trollworks.toolkit.notification.TKNotifierTarget;
import com.trollworks.toolkit.utility.TKAlignment;
import com.trollworks.toolkit.utility.TKNumberUtils;
import com.trollworks.toolkit.widget.TKTextField;

import java.awt.Color;
import java.awt.Graphics2D;
import java.awt.Rectangle;
import java.text.MessageFormat;

/** A text input field. */
public class CSField extends TKTextField implements TKNotifierTarget {
	private CMCharacter	mCharacter;
	private String		mConsumedType;
	private String		mCustomToolTip;

	/**
	 * Creates a new, left-aligned, text input field.
	 * 
	 * @param character The character to listen to.
	 * @param consumedType The field to listen to.
	 * @param tooltip The tooltip to set.
	 */
	public CSField(CMCharacter character, String consumedType, String tooltip) {
		this(character, consumedType, TKAlignment.LEFT, true, tooltip);
	}

	/**
	 * Creates a new text input field.
	 * 
	 * @param character The character to listen to.
	 * @param consumedType The field to listen to.
	 * @param alignment The alignment of the field.
	 * @param tooltip The tooltip to set.
	 */
	public CSField(CMCharacter character, String consumedType, int alignment, String tooltip) {
		this(character, consumedType, alignment, true, tooltip);
	}

	/**
	 * Creates a new text input field.
	 * 
	 * @param character The character to listen to.
	 * @param consumedType The field to listen to.
	 * @param alignment The alignment of the field.
	 * @param editable Whether or not the user can edit this field.
	 * @param tooltip The tooltip to set.
	 */
	public CSField(CMCharacter character, String consumedType, int alignment, boolean editable, String tooltip) {
		super("", 20, alignment); //$NON-NLS-1$
		mCharacter = character;
		mConsumedType = consumedType;
		setFontKey(CSFont.KEY_FIELD);
		setBorder(null);
		setOpaque(false);
		if (tooltip != null && tooltip.indexOf("{") != -1) { //$NON-NLS-1$
			mCustomToolTip = tooltip;
		}
		setToolTipText(tooltip);
		if (!editable) {
			setEnabled(false);
		}
		if (this.getClass() == CSField.class) {
			initialize();
		}
	}

	/** @return The character this field edits and responds to. */
	public CMCharacter getCharacter() {
		return mCharacter;
	}

	/** @return The type being consumed in notifications. */
	public String getConsumedType() {
		return mConsumedType;
	}

	/**
	 * Sub-classes must call this method before returning from their constructor.
	 */
	protected void initialize() {
		handleNotification(mCharacter, mConsumedType, mCharacter.getValueForID(mConsumedType));
		mCharacter.addTarget(this, mConsumedType);
	}

	@Override public void setText(String text) {
		if (!getText().equals(text)) {
			super.setText(text);
			selectAll();
		}
	}

	@Override public String getToolTipText() {
		if (mCustomToolTip != null) {
			return MessageFormat.format(mCustomToolTip, TKNumberUtils.format(((Integer) mCharacter.getValueForID(CMCharacter.POINTS_PREFIX + mConsumedType)).intValue()));
		}
		return super.getToolTipText();
	}

	public void handleNotification(Object producer, String type, Object data) {
		setText(data.toString());
		invalidate();
	}

	/** @return The object to pass to the setter. */
	protected Object getObjectToSet() {
		return getText();
	}

	@Override public void notifyActionListeners() {
		if (isEnabled()) {
			mCharacter.setValueForID(mConsumedType, getObjectToSet());
		}
		super.notifyActionListeners();
	}

	@Override protected void paintPanel(Graphics2D g2d, Rectangle[] clips) {
		if (isEnabled()) {
			drawUnderline(g2d);
		}
		super.paintPanel(g2d, clips);
	}

	/**
	 * Draws the field underline.
	 * 
	 * @param g2d The graphics object to use.
	 */
	protected void drawUnderline(Graphics2D g2d) {
		Rectangle bounds = getLocalBounds();

		g2d.setColor(Color.lightGray);
		g2d.drawLine(bounds.x, bounds.y + bounds.height - 1, bounds.x + bounds.width - 1, bounds.y + bounds.height - 1);
		g2d.setColor(getForeground());
	}
}
