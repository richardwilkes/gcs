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

import com.trollworks.toolkit.annotation.Localize;
import com.trollworks.toolkit.ui.UIUtilities;
import com.trollworks.toolkit.ui.layout.ColumnLayout;
import com.trollworks.toolkit.ui.widget.Wrapper;
import com.trollworks.toolkit.utility.Localization;
import com.trollworks.toolkit.utility.notification.NotifierTarget;
import com.trollworks.toolkit.utility.text.Numbers;

import java.awt.Color;
import java.awt.Container;
import java.awt.Dimension;
import java.text.MessageFormat;

import javax.swing.SwingConstants;

/** The character encumbrance panel. */
public class EncumbrancePanel extends DropPanel implements NotifierTarget {
	@Localize("Encumbrance, Move & Dodge")
	@Localize(locale = "de", value = "Belastung, Bewegung & Ausweichen")
	@Localize(locale = "ru", value = "Нагрузка, движение и уклонение")
	private static String		ENCUMBRANCE_MOVE_DODGE;
	@Localize("Level")
	@Localize(locale = "de", value = "Stufe")
	@Localize(locale = "ru", value = "Уровень")
	private static String		ENCUMBRANCE_LEVEL;
	@Localize("Max Load")
	@Localize(locale = "de", value = "Max. Gew.")
	@Localize(locale = "ru", value = "Мак.нагруз.")
	private static String		MAX_CARRY;
	@Localize("Move")
	@Localize(locale = "de", value = "Bew.")
	@Localize(locale = "ru", value = "Движ.")
	private static String		MOVE;
	@Localize("Dodge")
	@Localize(locale = "de", value = "Ausw.")
	@Localize(locale = "ru", value = "Укл.")
	private static String		DODGE;
	@Localize("The encumbrance level")
	@Localize(locale = "de", value = "Die Belastungsstufe")
	@Localize(locale = "ru", value = "Уровень нагрузки")
	private static String		ENCUMBRANCE_TOOLTIP;
	@Localize("The maximum load a character can carry and still remain within a specific encumbrance level")
	@Localize(locale = "de", value = "Das maximale Gewicht, das ein Charakter tragen und immer noch in einer bestimmten Belastungsstufe bleiben kann")
	@Localize(locale = "ru", value = "Максимальная нагрузка, которую персонаж может нести, \nоставаясь в пределах указанного уровня нагрузки")
	private static String		MAX_CARRY_TOOLTIP;
	@Localize("The character's ground movement rate for a specific encumbrance level")
	@Localize(locale = "de", value = "Die Bewegungswert eines Charakters am Boden für eine bestimmte Belastungsstufe")
	@Localize(locale = "ru", value = "Единицы движения персонажа при указанном уровне нагрузки")
	private static String		MOVE_TOOLTIP;
	@Localize("The character's dodge for a specific encumbrance level")
	@Localize(locale = "de", value = "Der Ausweichen-Wert eines Charakters für eine bestimmte Belastungsstufe")
	@Localize(locale = "ru", value = "Уклонение персонажа для указанного уровня нагрузки")
	private static String		DODGE_TOOLTIP;
	@Localize("{0} ({1})")
	@Localize(locale = "de", value = "{0} ({1})")
	static String				ENCUMBRANCE_FORMAT;
	@Localize("\u2022 {0} ({1})")
	@Localize(locale = "de", value = "\u2022 {0} ({1})")
	static String				CURRENT_ENCUMBRANCE_FORMAT;

	static {
		Localization.initialize();
	}

	private static final Color	CURRENT_ENCUMBRANCE_COLOR	= new Color(252, 242, 196);
	private GURPSCharacter		mCharacter;
	private PageLabel[]			mMarkers;

	/**
	 * Creates a new encumbrance panel.
	 *
	 * @param character The character to display the data for.
	 */
	public EncumbrancePanel(GURPSCharacter character) {
		super(new ColumnLayout(7, 2, 0), ENCUMBRANCE_MOVE_DODGE, true);
		mCharacter = character;
		Encumbrance[] encumbranceValues = Encumbrance.values();
		mMarkers = new PageLabel[encumbranceValues.length];
		PageHeader header = createHeader(this, ENCUMBRANCE_LEVEL, ENCUMBRANCE_TOOLTIP);
		addHorizontalBackground(header, Color.black);
		addVerticalBackground(createDivider(), Color.black);
		createHeader(this, MAX_CARRY, MAX_CARRY_TOOLTIP);
		addVerticalBackground(createDivider(), Color.black);
		createHeader(this, MOVE, MOVE_TOOLTIP);
		addVerticalBackground(createDivider(), Color.black);
		createHeader(this, DODGE, DODGE_TOOLTIP);
		Encumbrance current = character.getEncumbranceLevel();
		for (Encumbrance encumbrance : encumbranceValues) {
			int index = encumbrance.ordinal();
			mMarkers[index] = new PageLabel(getMarkerText(encumbrance, current), header);
			add(mMarkers[index]);
			if (current == encumbrance) {
				addHorizontalBackground(mMarkers[index], CURRENT_ENCUMBRANCE_COLOR);
			}
			createDivider();
			createDisabledField(this, character, GURPSCharacter.MAXIMUM_CARRY_PREFIX + index, MAX_CARRY_TOOLTIP, SwingConstants.RIGHT);
			createDivider();
			createDisabledField(this, character, GURPSCharacter.MOVE_PREFIX + index, MOVE_TOOLTIP, SwingConstants.RIGHT);
			createDivider();
			createDisabledField(this, character, GURPSCharacter.DODGE_PREFIX + index, DODGE_TOOLTIP, SwingConstants.RIGHT);
		}
		character.addTarget(this, GURPSCharacter.ID_CARRIED_WEIGHT, GURPSCharacter.ID_BASIC_LIFT);
	}

	private static String getMarkerText(Encumbrance which, Encumbrance current) {
		return MessageFormat.format(which == current ? CURRENT_ENCUMBRANCE_FORMAT : ENCUMBRANCE_FORMAT, which, Numbers.format(-which.getEncumbrancePenalty()));
	}

	private Container createDivider() {
		Wrapper panel = new Wrapper();
		UIUtilities.setOnlySize(panel, new Dimension(1, 1));
		add(panel);
		return panel;
	}

	@Override
	public void handleNotification(Object producer, String type, Object data) {
		Encumbrance current = mCharacter.getEncumbranceLevel();
		for (Encumbrance encumbrance : Encumbrance.values()) {
			int index = encumbrance.ordinal();
			if (encumbrance == current) {
				addHorizontalBackground(mMarkers[index], CURRENT_ENCUMBRANCE_COLOR);
			} else {
				removeHorizontalBackground(mMarkers[index]);
			}
			mMarkers[index].setText(getMarkerText(encumbrance, current));
		}
		revalidate();
		repaint();
	}

	@Override
	public int getNotificationPriority() {
		return 0;
	}
}
