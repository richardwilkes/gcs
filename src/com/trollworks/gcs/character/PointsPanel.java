/*
 * Copyright (c) 1998-2016 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public License,
 * version 2.0. If a copy of the MPL was not distributed with this file, You
 * can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as defined
 * by the Mozilla Public License, version 2.0.
 */

package com.trollworks.gcs.character;

import com.trollworks.gcs.preferences.SheetPreferences;
import com.trollworks.toolkit.annotation.Localize;
import com.trollworks.toolkit.ui.UIUtilities;
import com.trollworks.toolkit.ui.layout.ColumnLayout;
import com.trollworks.toolkit.ui.layout.RowDistribution;
import com.trollworks.toolkit.ui.widget.Wrapper;
import com.trollworks.toolkit.utility.Localization;
import com.trollworks.toolkit.utility.Preferences;
import com.trollworks.toolkit.utility.notification.NotifierTarget;
import com.trollworks.toolkit.utility.text.Numbers;

import java.awt.Color;
import java.awt.Dimension;
import java.text.MessageFormat;

import javax.swing.SwingConstants;

/** The character points panel. */
public class PointsPanel extends DropPanel implements NotifierTarget {
	@Localize("{0} Points")
	@Localize(locale = "de", value = "{0} Punkte")
	@Localize(locale = "ru", value = "{0} очков")
	@Localize(locale = "es", value = "{0} Puntos")
	private static String	POINTS;
	@Localize("Attributes:")
	@Localize(locale = "de", value = "Attribute:")
	@Localize(locale = "ru", value = "Атрибуты:")
	@Localize(locale = "es", value = "Atributos:")
	private static String	ATTRIBUTE_POINTS;
	@Localize("A summary of all points spent on attributes for this character")
	@Localize(locale = "de", value = "Die Summe der Punkte, die für Attribute dieses Charakters aufgewendet wurden")
	@Localize(locale = "ru", value = "Очки, потраченные на атрибуты")
	@Localize(locale = "es", value = "Suma de todos los puntos consumidos en atributos para el personaje")
	private static String	ATTRIBUTE_POINTS_TOOLTIP;
	@Localize("Advantages:")
	@Localize(locale = "de", value = "Vorteile:")
	@Localize(locale = "ru", value = "Преимущ-во:")
	@Localize(locale = "es", value = "Ventajas:")
	private static String	ADVANTAGE_POINTS;
	@Localize("A summary of all points spent on advantages for this character")
	@Localize(locale = "de", value = "Die Summe der Punkte, die für Vorteile dieses Charakters aufgewendet wurden")
	@Localize(locale = "ru", value = "Очки, потраченные на преимущества")
	@Localize(locale = "es", value = "Suma de todos los puntos consumidos en ventajas para el personaje")
	private static String	ADVANTAGE_POINTS_TOOLTIP;
	@Localize("Disadvantages:")
	@Localize(locale = "de", value = "Nachteile:")
	@Localize(locale = "ru", value = "Недостатки:")
	@Localize(locale = "es", value = "Desventajas:")
	private static String	DISADVANTAGE_POINTS;
	@Localize("A summary of all points spent on disadvantages for this character")
	@Localize(locale = "de", value = "Die Summe der Punkte, die für Nachteile dieses Charakters aufgewendet wurden")
	@Localize(locale = "ru", value = "Очки, потраченные на недостатки")
	@Localize(locale = "es", value = "Suma de todos los puntos consumidos en desventajas para el personaje")
	private static String	DISADVANTAGE_POINTS_TOOLTIP;
	@Localize("Quirks:")
	@Localize(locale = "de", value = "Marotten:")
	@Localize(locale = "ru", value = "Причуды:")
	@Localize(locale = "es", value = "Singularidades:")
	private static String	QUIRK_POINTS;
	@Localize("A summary of all points spent on quirks for this character")
	@Localize(locale = "de", value = "Die Summe der Punkte, die für Marotten dieses Charakters aufgewendet wurden")
	@Localize(locale = "ru", value = "Очки, потраченные на причуды")
	@Localize(locale = "es", value = "Suma de todos los puntos consumidos en singularidades para el personaje")
	private static String	QUIRK_POINTS_TOOLTIP;
	@Localize("Skills:")
	@Localize(locale = "de", value = "Fertigkeiten:")
	@Localize(locale = "ru", value = "Умения:")
	@Localize(locale = "es", value = "Habilidades:")
	private static String	SKILL_POINTS;
	@Localize("A summary of all points spent on skills for this character")
	@Localize(locale = "de", value = "Die Summe der Punkte, die für Fertigkeiten dieses Charakters aufgewendet wurden")
	@Localize(locale = "ru", value = "Очки, потраченные на умения")
	@Localize(locale = "es", value = "Suma de todos los puntos consumidos en habilidades para el personaje")
	private static String	SKILL_POINTS_TOOLTIP;
	@Localize("Spells:")
	@Localize(locale = "de", value = "Zauber:")
	@Localize(locale = "ru", value = "Заклинания:")
	@Localize(locale = "es", value = "Sortilegios:")
	private static String	SPELL_POINTS;
	@Localize("A summary of all points spent on spells for this character")
	@Localize(locale = "de", value = "Die Summe der Punkte, die für Zauber dieses Charakters aufgewendet wurden")
	@Localize(locale = "ru", value = "Очки, потраченные на заклинания")
	@Localize(locale = "es", value = "Suma de todos los puntos consumidos en sortilegios para el personaje")
	private static String	SPELL_POINTS_TOOLTIP;
	@Localize("Race:")
	@Localize(locale = "de", value = "Rasse:")
	@Localize(locale = "ru", value = "Раса:")
	@Localize(locale = "es", value = "Raza:")
	private static String	RACE_POINTS;
	@Localize("A summary of all points spent on a racial package for this character")
	@Localize(locale = "de", value = "Die Summe der Punkte, die für ein Rassenpaket dieses Charakters aufgewendet wurden")
	@Localize(locale = "ru", value = "Очки, потраченные на расовый пакет")
	@Localize(locale = "es", value = "Suma de todos los puntos consumidos en el paquete racial para el personaje")
	private static String	RACE_POINTS_TOOLTIP;
	@Localize("Earned:")
	@Localize(locale = "de", value = "Verdient:")
	@Localize(locale = "ru", value = "Заработано:")
	@Localize(locale = "es", value = "Ganados:")
	private static String	EARNED_POINTS;
	@Localize("Points that have been earned but not yet been spent")
	@Localize(locale = "de", value = "Punkte, die verdient aber noch nicht ausgegeben wurden")
	@Localize(locale = "ru", value = "Нераспределенные очки")
	@Localize(locale = "es", value = "Puntos ganados pero aún no consumidos")
	private static String	EARNED_POINTS_TOOLTIP;

	static {
		Localization.initialize();
	}

	private GURPSCharacter	mCharacter;

	/**
	 * Creates a new points panel.
	 *
	 * @param character The character to display the data for.
	 */
	public PointsPanel(GURPSCharacter character) {
		super(new ColumnLayout(2, 2, 0, RowDistribution.DISTRIBUTE_HEIGHT), getTitle(character));
		mCharacter = character;
		createLabelAndDisabledField(this, character, GURPSCharacter.ID_RACE_POINTS, RACE_POINTS, RACE_POINTS_TOOLTIP, SwingConstants.RIGHT);
		createLabelAndDisabledField(this, character, GURPSCharacter.ID_ATTRIBUTE_POINTS, ATTRIBUTE_POINTS, ATTRIBUTE_POINTS_TOOLTIP, SwingConstants.RIGHT);
		createLabelAndDisabledField(this, character, GURPSCharacter.ID_ADVANTAGE_POINTS, ADVANTAGE_POINTS, ADVANTAGE_POINTS_TOOLTIP, SwingConstants.RIGHT);
		createLabelAndDisabledField(this, character, GURPSCharacter.ID_DISADVANTAGE_POINTS, DISADVANTAGE_POINTS, DISADVANTAGE_POINTS_TOOLTIP, SwingConstants.RIGHT);
		createLabelAndDisabledField(this, character, GURPSCharacter.ID_QUIRK_POINTS, QUIRK_POINTS, QUIRK_POINTS_TOOLTIP, SwingConstants.RIGHT);
		createLabelAndDisabledField(this, character, GURPSCharacter.ID_SKILL_POINTS, SKILL_POINTS, SKILL_POINTS_TOOLTIP, SwingConstants.RIGHT);
		createLabelAndDisabledField(this, character, GURPSCharacter.ID_SPELL_POINTS, SPELL_POINTS, SPELL_POINTS_TOOLTIP, SwingConstants.RIGHT);
		createDivider();
		createLabelAndField(this, character, GURPSCharacter.ID_EARNED_POINTS, EARNED_POINTS, EARNED_POINTS_TOOLTIP, SwingConstants.RIGHT);
		mCharacter.addTarget(this, GURPSCharacter.ID_TOTAL_POINTS);
		Preferences.getInstance().getNotifier().add(this, SheetPreferences.TOTAL_POINTS_DISPLAY_PREF_KEY);
	}

	@Override
	public Dimension getMaximumSize() {
		Dimension size = super.getMaximumSize();

		size.width = getPreferredSize().width;
		return size;
	}

	private void createDivider() {
		Wrapper panel = new Wrapper();
		UIUtilities.setOnlySize(panel, new Dimension(1, 1));
		add(panel);
		addHorizontalBackground(panel, Color.black);
		panel = new Wrapper();
		UIUtilities.setOnlySize(panel, new Dimension(1, 1));
		add(panel);
	}

	@Override
	public void handleNotification(Object producer, String type, Object data) {
		getBoxedDropShadowBorder().setTitle(getTitle(mCharacter));
		invalidate();
		repaint();
	}

	private static String getTitle(GURPSCharacter character) {
		return MessageFormat.format(POINTS, Numbers.format(SheetPreferences.shouldIncludeUnspentPointsInTotalPointDisplay() ? character.getTotalPoints() : character.getSpentPoints()));
	}

	@Override
	public int getNotificationPriority() {
		return 0;
	}
}
