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
import com.trollworks.toolkit.ui.layout.Alignment;
import com.trollworks.toolkit.ui.layout.FlexComponent;
import com.trollworks.toolkit.ui.layout.FlexGrid;
import com.trollworks.toolkit.ui.layout.FlexRow;
import com.trollworks.toolkit.ui.layout.FlexSpacer;
import com.trollworks.toolkit.ui.widget.Wrapper;
import com.trollworks.toolkit.utility.Localization;

import java.awt.Color;
import java.awt.Dimension;

import javax.swing.SwingConstants;

/** The character attributes panel. */
public class AttributesPanel extends DropPanel {
	@Localize("Attributes")
	@Localize(locale = "de", value = "Attribute")
	@Localize(locale = "ru", value = "Атрибуты")
	private static String	ATTRIBUTES;
	@Localize("Strength (ST):")
	@Localize(locale = "de", value = "Stärke (ST):")
	@Localize(locale = "ru", value = "Сила (СЛ):")
	private static String	ST;
	@Localize("<html><body><b>{0} points</b> have been spent to modify <b>Strength</b></body></html>")
	@Localize(locale = "de", value = "<html><body><b>{0} Punkte</b> wurden ausgegeben, um die <b>Stärke</b> zu verändern</body></html>")
	@Localize(locale = "ru", value = "<html><body><b>{0} очков</b> было потрачено на <b>Силу</b><br>или по-английски Strength, кратко <b>ST</b></body></html>")
	private static String	ST_TOOLTIP;
	@Localize("Dexterity (DX):")
	@Localize(locale = "de", value = "Geschick (GE):")
	@Localize(locale = "ru", value = "Ловкость (ЛВ):")
	private static String	DX;
	@Localize("<html><body><b>{0} points</b> have been spent to modify <b>Dexterity</b></body></html>")
	@Localize(locale = "de", value = "<html><body><b>{0} Punkte</b> wurden ausgegeben, um das <b>Geschick</b> zu verändern</body></html>")
	@Localize(locale = "ru", value = "<html><body><b>{0} очков</b> было потрачено на <b>Ловкость </b><br>или по-английски Dexterity, кратко <b>DX</b></body></html>")
	private static String	DX_TOOLTIP;
	@Localize("Intelligence (IQ):")
	@Localize(locale = "de", value = "Intelligenz (IQ):")
	@Localize(locale = "ru", value = "Интеллект (ИН):")
	private static String	IQ;
	@Localize("<html><body><b>{0} points</b> have been spent to modify <b>Intelligence</b></body></html>")
	@Localize(locale = "de", value = "<html><body><b>{0} Punkte</b> wurden ausgegeben, um die <b>Intelligenz</b> zu verändern</body></html>")
	@Localize(locale = "ru", value = "<html><body><b>{0} очков</b> было потрачено на <b>Интеллект</b><br>или по-английски Intelligence, кратко <b>IQ</b></body></html>")
	private static String	IQ_TOOLTIP;
	@Localize("Health (HT):")
	@Localize(locale = "de", value = "Konstitution (KO):")
	@Localize(locale = "ru", value = "Здоровье (ЗД):")
	private static String	HT;
	@Localize("<html><body><b>{0} points</b> have been spent to modify <b>Health</b></body></html>")
	@Localize(locale = "de", value = "<html><body><b>{0} Punkte</b> wurden ausgegeben, um die <b>Konstitution</b> zu verändern</body></html>")
	@Localize(locale = "ru", value = "<html><body><b>{0} очков</b> было потрачено на <b>Здоровье</b><br>или по-английски Health, кратко <b>HT</b></body></html>")
	private static String	HT_TOOLTIP;
	@Localize("Will:")
	@Localize(locale = "de", value = "Wille:")
	@Localize(locale = "ru", value = "Воля:")
	private static String	WILL;
	@Localize("<html><body><b>{0} points</b> have been spent to modify <b>Will</b></body></html>")
	@Localize(locale = "de", value = "<html><body><b>{0} Punkte</b> wurden ausgegeben, um den <b>Willen</b> zu verändern</body></html>")
	@Localize(locale = "ru", value = "<html><body><b>{0} очков</b> было потрачено на <b>Волю</b> (по-английски <b>Will</b>)</body></html>")
	private static String	WILL_TOOLTIP;
	@Localize("Fright Check:")
	@Localize(locale = "de", value = "Schreckprobe:")
	@Localize(locale = "ru", value = "Бросок страха:")
	private static String	FRIGHT_CHECK;
	@Localize("Perception:")
	@Localize(locale = "de", value = "Wahrnehmung:")
	@Localize(locale = "ru", value = "Восприятие:")
	private static String	PERCEPTION;
	@Localize("<html><body><b>{0} points</b> have been spent to modify <b>Perception</b></body></html>")
	@Localize(locale = "de", value = "<html><body><b>{0} Punkte</b> wurden ausgegeben, um die <b>Wahrnehmung</b> zu verändern</body></html>")
	@Localize(locale = "ru", value = "<html><body><b>{0} очков</b> было потрачено на <b>Восприятие</b><br>или по-английски Perception, кратко <b>Per</b></body></html>")
	private static String	PERCEPTION_TOOLTIP;
	@Localize("Vision:")
	@Localize(locale = "de", value = "Sehen:")
	@Localize(locale = "ru", value = "Зрение:")
	private static String	VISION;
	@Localize("Hearing:")
	@Localize(locale = "de", value = "Hören:")
	@Localize(locale = "ru", value = "Слух:")
	private static String	HEARING;
	@Localize("Touch:")
	@Localize(locale = "de", value = "Fühlen:")
	@Localize(locale = "ru", value = "Осязание:")
	private static String	TOUCH;
	@Localize("Taste & Smell:")
	@Localize(locale = "de", value = "Schm. & Riechen:")
	@Localize(locale = "ru", value = "Вкус и запах:")
	private static String	TASTE_SMELL;
	@Localize("Basic Speed:")
	@Localize(locale = "de", value = "Grundgeschw.:")
	@Localize(locale = "ru", value = "Баз. скорость:")
	private static String	BASIC_SPEED;
	@Localize("<html><body><b>{0} points</b> have been spent to modify <b>Basic Speed</b></body></html>")
	@Localize(locale = "de", value = "<html><body><b>{0} Punkte</b> wurden ausgegeben, um die <b>Grundgeschwindigkeit</b> zu verändern</body></html>")
	@Localize(locale = "ru", value = "<html><body><b>{0} очков</b> было потрачено на <b>Базовую скорость</b> (по-английски Basic Speed)</body></html>")
	private static String	BASIC_SPEED_TOOLTIP;
	@Localize("Basic Move:")
	@Localize(locale = "de", value = "Grundbewegung:")
	@Localize(locale = "ru", value = "Баз. движение:")
	private static String	BASIC_MOVE;
	@Localize("<html><body><b>{0} points</b> have been spent to modify <b>Basic Move</b></body></html>")
	@Localize(locale = "de", value = "<html><body><b>{0} Punkte</b> wurden ausgegeben, um die <b>Grundbewegung</b> zu verändern</body></html>")
	@Localize(locale = "ru", value = "<html><body><b>{0} очков</b> было потрачено на <b>Базовое движение</b> (по-английски Basic Move)</body></html>")
	private static String	BASIC_MOVE_TOOLTIP;
	@Localize("thr:")
	@Localize(locale = "de", value = "stoß:")
	@Localize(locale = "ru", value = "прм:")
	private static String	BASIC_THRUST;
	@Localize("The basic damage value for thrust attacks")
	@Localize(locale = "de", value = "Der Grundschaden für Stoßangriffe")
	@Localize(locale = "ru", value = "Величина базового урона для прямых атак")
	private static String	BASIC_THRUST_TOOLTIP;
	@Localize("sw:")
	@Localize(locale = "de", value = "schw:")
	@Localize(locale = "ru", value = "амп:")
	private static String	BASIC_SWING;
	@Localize("The basic damage value for swing attacks")
	@Localize(locale = "de", value = "Der Grundschaden für Schwungangriffe")
	@Localize(locale = "ru", value = "Величина базового урона для амплитудных атак")
	private static String	BASIC_SWING_TOOLTIP;

	static {
		Localization.initialize();
	}

	/**
	 * Creates a new attributes panel.
	 *
	 * @param character The character to display the data for.
	 */
	public AttributesPanel(GURPSCharacter character) {
		super(null, ATTRIBUTES, true);
		FlexGrid grid = new FlexGrid();
		grid.setVerticalGap(0);
		int row = 0;
		createLabelAndField(grid, row++, character, GURPSCharacter.ID_STRENGTH, ST, ST_TOOLTIP, SwingConstants.RIGHT, true);
		createLabelAndField(grid, row++, character, GURPSCharacter.ID_DEXTERITY, DX, DX_TOOLTIP, SwingConstants.RIGHT, true);
		createLabelAndField(grid, row++, character, GURPSCharacter.ID_INTELLIGENCE, IQ, IQ_TOOLTIP, SwingConstants.RIGHT, true);
		createLabelAndField(grid, row++, character, GURPSCharacter.ID_HEALTH, HT, HT_TOOLTIP, SwingConstants.RIGHT, true);
		createDivider(grid, row++, false);
		createDivider(grid, row++, true);
		createLabelAndField(grid, row++, character, GURPSCharacter.ID_WILL, WILL, WILL_TOOLTIP, SwingConstants.RIGHT, true);
		createLabelAndField(grid, row++, character, GURPSCharacter.ID_FRIGHT_CHECK, FRIGHT_CHECK, null, SwingConstants.RIGHT, false);
		createDivider(grid, row++, false);
		createDivider(grid, row++, true);
		createLabelAndField(grid, row++, character, GURPSCharacter.ID_BASIC_SPEED, BASIC_SPEED, BASIC_SPEED_TOOLTIP, SwingConstants.RIGHT, true);
		createLabelAndField(grid, row++, character, GURPSCharacter.ID_BASIC_MOVE, BASIC_MOVE, BASIC_MOVE_TOOLTIP, SwingConstants.RIGHT, true);
		createDivider(grid, row++, false);
		createDivider(grid, row++, true);
		createLabelAndField(grid, row++, character, GURPSCharacter.ID_PERCEPTION, PERCEPTION, PERCEPTION_TOOLTIP, SwingConstants.RIGHT, true);
		createLabelAndField(grid, row++, character, GURPSCharacter.ID_VISION, VISION, null, SwingConstants.RIGHT, false);
		createLabelAndField(grid, row++, character, GURPSCharacter.ID_HEARING, HEARING, null, SwingConstants.RIGHT, false);
		createLabelAndField(grid, row++, character, GURPSCharacter.ID_TASTE_AND_SMELL, TASTE_SMELL, null, SwingConstants.RIGHT, false);
		createLabelAndField(grid, row++, character, GURPSCharacter.ID_TOUCH, TOUCH, null, SwingConstants.RIGHT, false);
		createDivider(grid, row++, false);
		createDivider(grid, row++, true);
		createDamageFields(grid, row++, character);
		grid.apply(this);
	}

	private void createLabelAndField(FlexGrid grid, int row, GURPSCharacter character, String key, String title, String tooltip, int alignment, boolean enabled) {
		PageField field = new PageField(character, key, alignment, enabled, tooltip);
		PageLabel label = new PageLabel(title, field);
		add(label);
		add(field);
		grid.add(new FlexComponent(label, Alignment.RIGHT_BOTTOM, Alignment.CENTER), row, 0);
		grid.add(field, row, 1);
	}

	private void createDamageFields(FlexGrid grid, int rowIndex, GURPSCharacter character) {
		FlexRow row = new FlexRow();
		row.setHorizontalAlignment(Alignment.CENTER);
		createDamageLabelAndField(row, character, GURPSCharacter.ID_BASIC_THRUST, BASIC_THRUST, BASIC_THRUST_TOOLTIP);
		row.add(new FlexSpacer(0, 0, false, false));
		createDamageLabelAndField(row, character, GURPSCharacter.ID_BASIC_SWING, BASIC_SWING, BASIC_SWING_TOOLTIP);
		grid.add(row, rowIndex, 0, 1, 2);
	}

	private void createDamageLabelAndField(FlexRow row, GURPSCharacter character, String key, String title, String tooltip) {
		PageField field = new PageField(character, key, SwingConstants.RIGHT, false, tooltip);
		PageLabel label = new PageLabel(title, field);
		add(label);
		add(field);
		row.add(new FlexComponent(label, true));
		row.add(new FlexComponent(field, true));
	}

	private void createDivider(FlexGrid grid, int row, boolean black) {
		Wrapper panel = new Wrapper();
		UIUtilities.setOnlySize(panel, new Dimension(1, 1));
		add(panel);
		if (black) {
			addHorizontalBackground(panel, Color.black);
		}
		grid.add(panel, row, 0, 1, 2);
	}
}
