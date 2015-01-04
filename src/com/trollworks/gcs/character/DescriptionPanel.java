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

import java.awt.Color;
import java.awt.Dimension;

import javax.swing.SwingConstants;

/** The character description panel. */
public class DescriptionPanel extends DropPanel {
	@Localize("Description")
	@Localize(locale = "de", value = "Beschreibung")
	@Localize(locale = "ru", value = "Описание")
	private static String	DESCRIPTION;
	@Localize("Race:")
	@Localize(locale = "de", value = "Rasse")
	@Localize(locale = "ru", value = "Раса:")
	private static String	RACE;
	@Localize("Size:")
	@Localize(locale = "de", value = "Größe")
	@Localize(locale = "ru", value = "Разм.:")
	private static String	SIZE_MODIFIER;
	@Localize("The character's size modifier")
	@Localize(locale = "de", value = "Der Größenmodifikator des Charakters")
	@Localize(locale = "ru", value = "Модификатор размера")
	private static String	SIZE_MODIFIER_TOOLTIP;
	@Localize("TL:")
	@Localize(locale = "de", value = "TL:")
	@Localize(locale = "ru", value = "ТУ:")
	private static String	TECH_LEVEL;
	@Localize("<html><body>TL0: Stone Age<br>TL1: Bronze Age<br>TL2: Iron Age<br>TL3: Medieval<br>TL4: Age of Sail<br>TL5: Industrial Revolution<br>TL6: Mechanized Age<br>TL7: Nuclear Age<br>TL8: Digital Age<br>TL9: Microtech Age<br>TL10: Robotic Age<br>TL11: Age of Exotic Matter<br>TL12: Anything Goes</body></html>")
	@Localize(locale = "de", value = "<html><body>TL0: Steinzeit (Vorzeit und älter)<br>TL1: Bronzezeit (ab 3500 v.Chr.)<br>TL2: Eisenzeit (ab 1200 v. Chr.)<br>TL3: Mittelalter(ab 600 n. Chr.)<br>TL4: Renaissance / Kolonialzeitalter (ab 1450)<br>TL5: Industrielle Revolution (ab 1730)<br>TL6: Zeitalter der Mechanik (ab 1880)<br>TL7: Ära der Kernkraft (ab 1940)<br>TL8: Ära der Digitalisierung (ab 1980)<br>TL9: Ära der Mikrotechnik (ab 2025?)<br>TL10: Ära der Robotik (ab 2070?)<br>TL11: Ära der Exotischen Materie<br>TL12: Was ihr wollt!</body></html>")
	@Localize(locale = "ru", value = "<html><body>ТУ0: Каменный век<br>ТУ1: Бронзовый век<br>ТУ2: Железный век<br>ТУ3: Средневековье<br>ТУ4: Эпоха парусов<br>ТУ5: Промышленный переворот<br>ТУ6: Эпоха механики<br>ТУ7: Атомная эпоха<br>ТУ8: Цифровая эпоха<br>ТУ9: Эпоха микротехники<br>ТУ10: Эпоха роботизации<br>ТУ11: Эпоха экзотических материалов<br>ТУ12: Всё, что угодно</body></html>")
	private static String	TECH_LEVEL_TOOLTIP;
	@Localize("Age:")
	@Localize(locale = "de", value = "Alter:")
	@Localize(locale = "ru", value = "Возраст:")
	static String			AGE;
	@Localize("Gender:")
	@Localize(locale = "de", value = "Geschl.:")
	@Localize(locale = "ru", value = "Пол:")
	static String			GENDER;
	@Localize("Birthday:")
	@Localize(locale = "de", value = "Geburt:")
	@Localize(locale = "ru", value = "Д. р.:")
	static String			BIRTHDAY;
	@Localize("Height:")
	@Localize(locale = "de", value = "Höhe:")
	@Localize(locale = "ru", value = "Рост:")
	static String			HEIGHT_FIELD;
	@Localize("Weight:")
	@Localize(locale = "de", value = "Gew.:")
	@Localize(locale = "ru", value = "Вес:")
	static String			WEIGHT;
	@Localize("Hair:")
	@Localize(locale = "de", value = "Haar:")
	@Localize(locale = "ru", value = "Волосы:")
	static String			HAIR;
	@Localize("The character's hair style and color")
	@Localize(locale = "de", value = "Die Frisur und Haarfarbe des Charakters")
	@Localize(locale = "ru", value = "Прическа и цвет волос")
	static String			HAIR_TOOLTIP;
	@Localize("Eyes:")
	@Localize(locale = "de", value = "Augen:")
	@Localize(locale = "ru", value = "Глаза:")
	static String			EYE_COLOR;
	@Localize("The character's eye color")
	@Localize(locale = "de", value = "Die Augenfarbe des Charakters")
	@Localize(locale = "ru", value = "Цвет глаз")
	static String			EYE_COLOR_TOOLTIP;
	@Localize("Skin:")
	@Localize(locale = "de", value = "Haut:")
	@Localize(locale = "ru", value = "Кожа:")
	static String			SKIN_COLOR;
	@Localize("The character's skin color")
	@Localize(locale = "de", value = "Die Hautfarbe des Charakters")
	@Localize(locale = "ru", value = "Цвет кожы")
	static String			SKIN_COLOR_TOOLTIP;
	@Localize("Hand:")
	@Localize(locale = "de", value = "Händig:")
	@Localize(locale = "ru", value = "Рука:")
	static String			HANDEDNESS;
	@Localize("The character's preferred hand")
	@Localize(locale = "de", value = "Die bevorzugte Hand des Charakters")
	@Localize(locale = "ru", value = "Основная рука")
	static String			HANDEDNESS_TOOLTIP;

	static {
		Localization.initialize();
	}

	/**
	 * Creates a new description panel.
	 *
	 * @param character The character to display the data for.
	 */
	public DescriptionPanel(GURPSCharacter character) {
		super(new ColumnLayout(5, 2, 0), DESCRIPTION);

		Wrapper wrapper = new Wrapper(new ColumnLayout(2, 2, 0));
		createLabelAndField(wrapper, character, Profile.ID_RACE, RACE, null, SwingConstants.LEFT);
		createLabelAndField(wrapper, character, Profile.ID_GENDER, GENDER, null, SwingConstants.LEFT);
		createLabelAndField(wrapper, character, Profile.ID_AGE, AGE, null, SwingConstants.LEFT);
		createLabelAndField(wrapper, character, Profile.ID_BIRTHDAY, BIRTHDAY, null, SwingConstants.LEFT);
		add(wrapper);

		createDivider();

		wrapper = new Wrapper(new ColumnLayout(2, 2, 0));
		createLabelAndField(wrapper, character, Profile.ID_HEIGHT, HEIGHT_FIELD, null, SwingConstants.LEFT);
		createLabelAndField(wrapper, character, Profile.ID_WEIGHT, WEIGHT, null, SwingConstants.LEFT);
		createLabelAndField(wrapper, character, Profile.ID_SIZE_MODIFIER, SIZE_MODIFIER, SIZE_MODIFIER_TOOLTIP, SwingConstants.LEFT);
		createLabelAndField(wrapper, character, Profile.ID_TECH_LEVEL, TECH_LEVEL, TECH_LEVEL_TOOLTIP, SwingConstants.LEFT);
		add(wrapper);

		createDivider();

		wrapper = new Wrapper(new ColumnLayout(2, 2, 0));
		createLabelAndField(wrapper, character, Profile.ID_HAIR, HAIR, HAIR_TOOLTIP, SwingConstants.LEFT);
		createLabelAndField(wrapper, character, Profile.ID_EYE_COLOR, EYE_COLOR, EYE_COLOR_TOOLTIP, SwingConstants.LEFT);
		createLabelAndField(wrapper, character, Profile.ID_SKIN_COLOR, SKIN_COLOR, SKIN_COLOR_TOOLTIP, SwingConstants.LEFT);
		createLabelAndField(wrapper, character, Profile.ID_HANDEDNESS, HANDEDNESS, HANDEDNESS_TOOLTIP, SwingConstants.LEFT);
		add(wrapper);
	}

	private void createDivider() {
		Wrapper panel = new Wrapper();
		UIUtilities.setOnlySize(panel, new Dimension(1, 1));
		add(panel);
		addVerticalBackground(panel, Color.black);
	}
}
