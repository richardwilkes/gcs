/*
 * Copyright (c) 1998-2015 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public License,
 * version 2.0. If a copy of the MPL was not distributed with this file, You
 * can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as defined
 * by the Mozilla Public License, version 2.0.
 */

package com.trollworks.gcs.character;

import com.trollworks.gcs.app.GCSFonts;
import com.trollworks.toolkit.ui.GraphicsUtilities;
import com.trollworks.toolkit.utility.Platform;
import com.trollworks.toolkit.utility.notification.NotifierTarget;
import com.trollworks.toolkit.utility.text.DateFormatter;
import com.trollworks.toolkit.utility.text.DiceFormatter;
import com.trollworks.toolkit.utility.text.DoubleFormatter;
import com.trollworks.toolkit.utility.text.HeightFormatter;
import com.trollworks.toolkit.utility.text.IntegerFormatter;
import com.trollworks.toolkit.utility.text.Numbers;
import com.trollworks.toolkit.utility.text.WeightFormatter;

import java.awt.Color;
import java.awt.Dimension;
import java.awt.Graphics;
import java.awt.Rectangle;
import java.awt.event.ActionEvent;
import java.awt.event.ActionListener;
import java.awt.event.FocusEvent;
import java.beans.PropertyChangeEvent;
import java.beans.PropertyChangeListener;
import java.text.DateFormat;
import java.text.MessageFormat;
import java.text.ParseException;
import java.util.HashMap;

import javax.swing.JFormattedTextField;
import javax.swing.SwingConstants;
import javax.swing.UIManager;
import javax.swing.plaf.basic.BasicTextFieldUI;
import javax.swing.text.DefaultFormatter;
import javax.swing.text.DefaultFormatterFactory;

/** A generic field for a page. */
public class PageField extends JFormattedTextField implements NotifierTarget, PropertyChangeListener, ActionListener {
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
		setFont(UIManager.getFont(GCSFonts.KEY_FIELD));
		setBorder(null);
		setOpaque(false);
		// Just setting opaque to false isn't enough for some reason, so I'm also setting the
		// background color to a 100% transparent value.
		setBackground(new Color(255, 255, 255, 0));
		setHorizontalAlignment(alignment);
		setEditable(editable);
		setEnabled(editable);
		setForeground(editable ? Color.BLACK : Color.GRAY);
		if (tooltip != null) {
			setToolTipText(tooltip);
			if (tooltip.indexOf("{") != -1) { //$NON-NLS-1$
				mCustomToolTip = tooltip;
			}
		}
		mCharacter.addTarget(this, mConsumedType);
		addPropertyChangeListener("value", this); //$NON-NLS-1$
		addActionListener(this);

		// Reset the selection colors back to what is standard for text fields.
		// This is necessary, since (at least on the Mac) JFormattedTextField
		// has the wrong values by default.
		setCaretColor(UIManager.getColor("TextField.caretForeground")); //$NON-NLS-1$
		setSelectionColor(UIManager.getColor("TextField.selectionBackground")); //$NON-NLS-1$
		setSelectedTextColor(UIManager.getColor("TextField.selectionForeground")); //$NON-NLS-1$
		setDisabledTextColor(UIManager.getColor("TextField.inactiveForeground")); //$NON-NLS-1$
	}

	@Override
	public Dimension getPreferredSize() {
		Dimension size = super.getPreferredSize();
		// Don't know why this is needed, but it seems to be. Without it, text is being truncated by
		// about 2 pixels.
		size.width += 2;
		return size;
	}

	@Override
	public String getToolTipText() {
		if (mCustomToolTip != null) {
			return MessageFormat.format(mCustomToolTip, Numbers.format(((Integer) mCharacter.getValueForID(GURPSCharacter.POINTS_PREFIX + mConsumedType)).intValue()));
		}
		return super.getToolTipText();
	}

	@Override
	public void handleNotification(Object producer, String name, Object data) {
		setValue(data);
		invalidate();
		repaint();
	}

	@Override
	protected void processFocusEvent(FocusEvent event) {
		super.processFocusEvent(event);
		if (event.getID() == FocusEvent.FOCUS_GAINED) {
			selectAll();
		}
	}

	@Override
	protected void paintComponent(Graphics gc) {
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

	@Override
	public void propertyChange(PropertyChangeEvent event) {
		if (isEditable()) {
			mCharacter.setValueForID(mConsumedType, getValue());
		}
	}

	private static final HashMap<String, AbstractFormatterFactory>	FACTORY_MAP	= new HashMap<>();
	private static final AbstractFormatterFactory					DEFAULT_FACTORY;

	static {
		DefaultFormatterFactory factory = new DefaultFormatterFactory(new WeightFormatter(true));
		FACTORY_MAP.put(Profile.ID_WEIGHT, factory);
		FACTORY_MAP.put(GURPSCharacter.ID_BASIC_LIFT, factory);
		FACTORY_MAP.put(GURPSCharacter.ID_ONE_HANDED_LIFT, factory);
		FACTORY_MAP.put(GURPSCharacter.ID_TWO_HANDED_LIFT, factory);
		FACTORY_MAP.put(GURPSCharacter.ID_SHOVE_AND_KNOCK_OVER, factory);
		FACTORY_MAP.put(GURPSCharacter.ID_RUNNING_SHOVE_AND_KNOCK_OVER, factory);
		FACTORY_MAP.put(GURPSCharacter.ID_CARRY_ON_BACK, factory);
		FACTORY_MAP.put(GURPSCharacter.ID_SHIFT_SLIGHTLY, factory);
		for (Encumbrance encumbrance : Encumbrance.values()) {
			FACTORY_MAP.put(GURPSCharacter.MAXIMUM_CARRY_PREFIX + encumbrance.ordinal(), factory);
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
		for (Encumbrance encumbrance : Encumbrance.values()) {
			int index = encumbrance.ordinal();
			FACTORY_MAP.put(GURPSCharacter.MOVE_PREFIX + index, factory);
			FACTORY_MAP.put(GURPSCharacter.DODGE_PREFIX + index, factory);
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

		FACTORY_MAP.put(Profile.ID_SIZE_MODIFIER, new DefaultFormatterFactory(new IntegerFormatter(-99, 9999, true)));
		FACTORY_MAP.put(Profile.ID_AGE, new DefaultFormatterFactory(new IntegerFormatter(0, Integer.MAX_VALUE, false, true)));
		FACTORY_MAP.put(Profile.ID_HEIGHT, new DefaultFormatterFactory(new HeightFormatter(true)));
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

	@Override
	public int getNotificationPriority() {
		return 0;
	}

	@Override
	public void actionPerformed(ActionEvent event) {
		try {
			commitEdit();
		} catch (ParseException exception) {
			invalidEdit();
		}
	}
}
