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

package com.trollworks.gcs.character;

import com.trollworks.gcs.utility.Fonts;
import com.trollworks.gcs.utility.Platform;
import com.trollworks.gcs.utility.notification.NotifierTarget;
import com.trollworks.gcs.utility.text.DateFormatter;
import com.trollworks.gcs.utility.text.DiceFormatter;
import com.trollworks.gcs.utility.text.DoubleFormatter;
import com.trollworks.gcs.utility.text.HeightFormatter;
import com.trollworks.gcs.utility.text.IntegerFormatter;
import com.trollworks.gcs.utility.text.NumberUtils;
import com.trollworks.gcs.utility.text.WeightFormatter;
import com.trollworks.gcs.widgets.GraphicsUtilities;

import java.awt.Color;
import java.awt.Dimension;
import java.awt.Graphics;
import java.awt.Rectangle;
import java.awt.event.FocusEvent;
import java.beans.PropertyChangeEvent;
import java.beans.PropertyChangeListener;
import java.text.DateFormat;
import java.text.MessageFormat;
import java.util.HashMap;

import javax.swing.JFormattedTextField;
import javax.swing.SwingConstants;
import javax.swing.UIManager;
import javax.swing.plaf.basic.BasicTextFieldUI;
import javax.swing.text.DefaultFormatter;
import javax.swing.text.DefaultFormatterFactory;

/** A generic field for a page. */
public class PageField extends JFormattedTextField implements NotifierTarget, PropertyChangeListener {
	private GURPSCharacter	mCharacter;
	private String			mConsumedType;
	private String			mCustomToolTip;

	/**
	 * Creates a new, left-aligned, text input field.
	 * 
	 * @param character The character to listen to.
	 * @param consumedType The field to listen to.
	 */
	public PageField(GURPSCharacter character, String consumedType) {
		this(character, consumedType, SwingConstants.LEFT, true, null);
	}

	/**
	 * Creates a new, left-aligned, text input field.
	 * 
	 * @param character The character to listen to.
	 * @param consumedType The field to listen to.
	 * @param tooltip The tooltip to set.
	 */
	public PageField(GURPSCharacter character, String consumedType, String tooltip) {
		this(character, consumedType, SwingConstants.LEFT, true, tooltip);
	}

	/**
	 * Creates a new text input field.
	 * 
	 * @param character The character to listen to.
	 * @param consumedType The field to listen to.
	 * @param alignment The alignment of the field.
	 * @param tooltip The tooltip to set.
	 */
	public PageField(GURPSCharacter character, String consumedType, int alignment, String tooltip) {
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
	public PageField(GURPSCharacter character, String consumedType, int alignment, boolean editable, String tooltip) {
		super(getFormatterFactoryForType(consumedType), character.getValueForID(consumedType));
		if (Platform.isLinux()) {
			// I override the UI here since the GTK UI on Linux has no way to turn off the border
			// around text fields.
			setUI(new BasicTextFieldUI());
		}
		mCharacter = character;
		mConsumedType = consumedType;
		setFont(UIManager.getFont(Fonts.KEY_FIELD));
		setBorder(null);
		setOpaque(false);
		// Just setting opaque to false isn't enough for some reason, so I'm also setting the
		// background color to a 100% transparent value.
		setBackground(new Color(255, 255, 255, 0));
		setHorizontalAlignment(alignment);
		setEditable(editable);
		setEnabled(editable);
		if (!editable) {
			setForeground(Color.gray);
		}
		if (tooltip != null) {
			setToolTipText(tooltip);
			if (tooltip.indexOf("{") != -1) { //$NON-NLS-1$
				mCustomToolTip = tooltip;
			}
		}
		mCharacter.addTarget(this, mConsumedType);
		addPropertyChangeListener("value", this); //$NON-NLS-1$

		// Reset the selection colors back to what is standard for text fields.
		// This is necessary, since (at least on the Mac) JFormattedTextField
		// has the wrong values by default.
		setCaretColor(UIManager.getColor("TextField.caretForeground")); //$NON-NLS-1$
		setSelectionColor(UIManager.getColor("TextField.selectionBackground")); //$NON-NLS-1$
		setSelectedTextColor(UIManager.getColor("TextField.selectionForeground")); //$NON-NLS-1$
		setDisabledTextColor(UIManager.getColor("TextField.inactiveForeground")); //$NON-NLS-1$
	}

	@Override public Dimension getPreferredSize() {
		Dimension size = super.getPreferredSize();
		// Don't know why this is needed, but it seems to be. Without it, text is being truncated by
		// about 2 pixels.
		size.width += 2;
		return size;
	}

	@Override public String getToolTipText() {
		if (mCustomToolTip != null) {
			return MessageFormat.format(mCustomToolTip, NumberUtils.format(((Integer) mCharacter.getValueForID(GURPSCharacter.POINTS_PREFIX + mConsumedType)).intValue()));
		}
		return super.getToolTipText();
	}

	public void handleNotification(Object producer, String name, Object data) {
		setValue(data);
		invalidate();
		repaint();
	}

	@Override protected void processFocusEvent(FocusEvent event) {
		super.processFocusEvent(event);
		if (event.getID() == FocusEvent.FOCUS_GAINED) {
			selectAll();
		}
	}

	@Override protected void paintComponent(Graphics gc) {
		if (isEditable()) {
			Rectangle bounds = getBounds();
			bounds.x = 0;
			bounds.y = 0;
			gc.setColor(Color.lightGray);
			gc.drawLine(bounds.x, bounds.y + bounds.height - 1, bounds.x + bounds.width - 1, bounds.y + bounds.height - 1);
			gc.setColor(getForeground());
		}
		super.paintComponent(GraphicsUtilities.prepare(gc));
	}

	/** @return The consumed type. */
	public String getConsumedType() {
		return mConsumedType;
	}

	public void propertyChange(PropertyChangeEvent event) {
		if (isEditable()) {
			mCharacter.setValueForID(mConsumedType, getValue());
		}
	}

	private static final HashMap<String, AbstractFormatterFactory>	FACTORY_MAP	= new HashMap<String, AbstractFormatterFactory>();
	private static final AbstractFormatterFactory					DEFAULT_FACTORY;

	static {
		DefaultFormatterFactory factory = new DefaultFormatterFactory(new WeightFormatter());
		FACTORY_MAP.put(Profile.ID_WEIGHT, factory);
		FACTORY_MAP.put(GURPSCharacter.ID_BASIC_LIFT, factory);
		FACTORY_MAP.put(GURPSCharacter.ID_ONE_HANDED_LIFT, factory);
		FACTORY_MAP.put(GURPSCharacter.ID_TWO_HANDED_LIFT, factory);
		FACTORY_MAP.put(GURPSCharacter.ID_SHOVE_AND_KNOCK_OVER, factory);
		FACTORY_MAP.put(GURPSCharacter.ID_RUNNING_SHOVE_AND_KNOCK_OVER, factory);
		FACTORY_MAP.put(GURPSCharacter.ID_CARRY_ON_BACK, factory);
		FACTORY_MAP.put(GURPSCharacter.ID_SHIFT_SLIGHTLY, factory);
		for (int i = 0; i < GURPSCharacter.ENCUMBRANCE_LEVELS; i++) {
			FACTORY_MAP.put(GURPSCharacter.MAXIMUM_CARRY_PREFIX + i, factory);
		}

		factory = new DefaultFormatterFactory(new IntegerFormatter(0, 9999, false));
		FACTORY_MAP.put(GURPSCharacter.ID_STRENGTH, factory);
		FACTORY_MAP.put(GURPSCharacter.ID_DEXTERITY, factory);
		FACTORY_MAP.put(GURPSCharacter.ID_INTELLIGENCE, factory);
		FACTORY_MAP.put(GURPSCharacter.ID_HEALTH, factory);
		FACTORY_MAP.put(GURPSCharacter.ID_WILL, factory);
		FACTORY_MAP.put(GURPSCharacter.ID_FRIGHT_CHECK, factory);
		FACTORY_MAP.put(GURPSCharacter.ID_BASIC_MOVE, factory);
		FACTORY_MAP.put(GURPSCharacter.ID_PERCEPTION, factory);
		FACTORY_MAP.put(GURPSCharacter.ID_VISION, factory);
		FACTORY_MAP.put(GURPSCharacter.ID_HEARING, factory);
		FACTORY_MAP.put(GURPSCharacter.ID_TASTE_AND_SMELL, factory);
		FACTORY_MAP.put(GURPSCharacter.ID_TOUCH, factory);
		FACTORY_MAP.put(Armor.ID_EYES_DR, factory);
		FACTORY_MAP.put(Armor.ID_SKULL_DR, factory);
		FACTORY_MAP.put(Armor.ID_FACE_DR, factory);
		FACTORY_MAP.put(Armor.ID_LEG_DR, factory);
		FACTORY_MAP.put(Armor.ID_ARM_DR, factory);
		FACTORY_MAP.put(Armor.ID_TORSO_DR, factory);
		FACTORY_MAP.put(Armor.ID_GROIN_DR, factory);
		FACTORY_MAP.put(Armor.ID_HAND_DR, factory);
		FACTORY_MAP.put(Armor.ID_FOOT_DR, factory);
		FACTORY_MAP.put(Armor.ID_NECK_DR, factory);
		for (int i = 0; i < GURPSCharacter.ENCUMBRANCE_LEVELS; i++) {
			FACTORY_MAP.put(GURPSCharacter.MOVE_PREFIX + i, factory);
			FACTORY_MAP.put(GURPSCharacter.DODGE_PREFIX + i, factory);
		}

		factory = new DefaultFormatterFactory(new IntegerFormatter(-99999, 99999, false));
		FACTORY_MAP.put(GURPSCharacter.ID_ATTRIBUTE_POINTS, factory);
		FACTORY_MAP.put(GURPSCharacter.ID_ADVANTAGE_POINTS, factory);
		FACTORY_MAP.put(GURPSCharacter.ID_DISADVANTAGE_POINTS, factory);
		FACTORY_MAP.put(GURPSCharacter.ID_QUIRK_POINTS, factory);
		FACTORY_MAP.put(GURPSCharacter.ID_SKILL_POINTS, factory);
		FACTORY_MAP.put(GURPSCharacter.ID_SPELL_POINTS, factory);
		FACTORY_MAP.put(GURPSCharacter.ID_RACE_POINTS, factory);
		FACTORY_MAP.put(GURPSCharacter.ID_EARNED_POINTS, factory);
		FACTORY_MAP.put(GURPSCharacter.ID_TIRED_FATIGUE_POINTS, factory);
		FACTORY_MAP.put(GURPSCharacter.ID_UNCONSCIOUS_CHECKS_FATIGUE_POINTS, factory);
		FACTORY_MAP.put(GURPSCharacter.ID_UNCONSCIOUS_FATIGUE_POINTS, factory);
		FACTORY_MAP.put(GURPSCharacter.ID_REELING_HIT_POINTS, factory);
		FACTORY_MAP.put(GURPSCharacter.ID_UNCONSCIOUS_CHECKS_HIT_POINTS, factory);
		FACTORY_MAP.put(GURPSCharacter.ID_DEATH_CHECK_1_HIT_POINTS, factory);
		FACTORY_MAP.put(GURPSCharacter.ID_DEATH_CHECK_2_HIT_POINTS, factory);
		FACTORY_MAP.put(GURPSCharacter.ID_DEATH_CHECK_3_HIT_POINTS, factory);
		FACTORY_MAP.put(GURPSCharacter.ID_DEATH_CHECK_4_HIT_POINTS, factory);
		FACTORY_MAP.put(GURPSCharacter.ID_DEAD_HIT_POINTS, factory);

		factory = new DefaultFormatterFactory(new IntegerFormatter(0, 999999, false));
		FACTORY_MAP.put(GURPSCharacter.ID_FATIGUE_POINTS, factory);
		FACTORY_MAP.put(GURPSCharacter.ID_HIT_POINTS, factory);

		factory = new DefaultFormatterFactory(new DiceFormatter());
		FACTORY_MAP.put(GURPSCharacter.ID_BASIC_THRUST, factory);
		FACTORY_MAP.put(GURPSCharacter.ID_BASIC_SWING, factory);

		FACTORY_MAP.put(Profile.ID_SIZE_MODIFIER, new DefaultFormatterFactory(new IntegerFormatter(-8, 20, true)));
		FACTORY_MAP.put(Profile.ID_AGE, new DefaultFormatterFactory(new IntegerFormatter(0, Integer.MAX_VALUE, false)));
		FACTORY_MAP.put(Profile.ID_HEIGHT, new DefaultFormatterFactory(new HeightFormatter()));
		FACTORY_MAP.put(GURPSCharacter.ID_CREATED_ON, new DefaultFormatterFactory(new DateFormatter(DateFormat.MEDIUM)));
		FACTORY_MAP.put(GURPSCharacter.ID_BASIC_SPEED, new DefaultFormatterFactory(new DoubleFormatter(0, 9999, false)));

		DefaultFormatter formatter = new DefaultFormatter();
		formatter.setOverwriteMode(false);
		DEFAULT_FACTORY = new DefaultFormatterFactory(formatter);
	}

	private static AbstractFormatterFactory getFormatterFactoryForType(String type) {
		AbstractFormatterFactory factory = FACTORY_MAP.get(type);
		return factory != null ? factory : DEFAULT_FACTORY;
	}
}
