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
import java.awt.Container;
import java.awt.Dimension;
import java.text.MessageFormat;

import javax.swing.SwingConstants;

/** The character hit location panel. */
public class HitLocationPanel extends DropPanel {
	@Localize("Hit Location")
	@Localize(locale = "de", value = "Trefferzonen")
	@Localize(locale = "ru", value = "Зоны попадания")
	private static String			HIT_LOCATION;
	@Localize("Roll")
	@Localize(locale = "de", value = "Wurf")
	@Localize(locale = "ru", value = "ДБ")
	private static String			ROLL;
	@Localize("<html><body>The random roll needed to hit the <b>{0}</b> hit location</body></html>")
	@Localize(locale = "de", value = "<html><body>Der Würfelwurf, um die Trefferzone <b>{0}</b> zu treffen</body></html>")
	@Localize(locale = "ru", value = "<html><body>Для попадания в <b>{0}</b>, необходимо сделать дополнительный бросок (ДБ) и выбросить указанные числа</body></html>")
	private static String			ROLL_TOOLTIP;
	@Localize("Where")
	@Localize(locale = "de", value = "Zone")
	@Localize(locale = "ru", value = "Где")
	private static String			LOCATION;
	@Localize("-")
	@Localize(locale = "de", value = "-")
	private static String			PENALTY;
	@Localize("The hit penalty for targeting a specific hit location")
	@Localize(locale = "de", value = "Der Treffernachteil für das Zielen auf eine spezifische Trefferzone")
	@Localize(locale = "ru", value = "Штраф для попадания в указанную зону")
	private static String			PENALTY_TITLE_TOOLTIP;
	@Localize("<html><body>The hit penalty for targeting the <b>{0}</b> hit location</body></html>")
	@Localize(locale = "de", value = "<html><body>Der Treffernachteil für das Zielen auf die Trefferzone <b>{0}</b></body></html>")
	@Localize(locale = "ru", value = "<html><body>Штраф за прицеливания в зону попадания <b>{0}</b></body></html>")
	private static String			PENALTY_TOOLTIP;
	@Localize("DR")
	@Localize(locale = "de", value = "SR")
	@Localize(locale = "ru", value = "СП")
	private static String			DR;
	@Localize("<html><body>The total DR protecting the <b>{0}</b> hit location</body></html>")
	@Localize(locale = "de", value = "<html><body>Die Gesamte Schadensresistenz, die die Trefferzone <b>{0}</b> schützt</body></html>")
	@Localize(locale = "ru", value = "<html><body>Суммарное СП, защищающее зону попадания: <b>{0}</b></body></html>")
	private static String			DR_TOOLTIP;
	@Localize("Eye")
	@Localize(locale = "de", value = "Auge")
	@Localize(locale = "ru", value = "Глаз")
	private static String			EYE;
	@Localize("Skull")
	@Localize(locale = "de", value = "Schädel")
	@Localize(locale = "ru", value = "Череп")
	private static String			SKULL;
	@Localize("Face")
	@Localize(locale = "de", value = "Gesicht")
	@Localize(locale = "ru", value = "Лицо")
	private static String			FACE;
	@Localize("R. Leg")
	@Localize(locale = "de", value = "R. Bein")
	@Localize(locale = "ru", value = "Пр. нога")
	private static String			RIGHT_LEG;
	@Localize("R. Arm")
	@Localize(locale = "de", value = "R. Arm")
	@Localize(locale = "ru", value = "Пр. рука")
	private static String			RIGHT_ARM;
	@Localize("Torso")
	@Localize(locale = "de", value = "Torso")
	@Localize(locale = "ru", value = "Туловище")
	private static String			TORSO;
	@Localize("Groin")
	@Localize(locale = "de", value = "Leiste")
	@Localize(locale = "ru", value = "Пах")
	private static String			GROIN;
	@Localize("L. Arm")
	@Localize(locale = "de", value = "L. Arm")
	@Localize(locale = "ru", value = "Л. рука")
	private static String			LEFT_ARM;
	@Localize("L. Leg")
	@Localize(locale = "de", value = "L. Bein")
	@Localize(locale = "ru", value = "Л. нога")
	private static String			LEFT_LEG;
	@Localize("Hand")
	@Localize(locale = "de", value = "Hand")
	@Localize(locale = "ru", value = "Рука")
	private static String			HAND;
	@Localize("Foot")
	@Localize(locale = "de", value = "Fuß")
	@Localize(locale = "ru", value = "Нога")
	private static String			FOOT;
	@Localize("Neck")
	@Localize(locale = "de", value = "Hals")
	@Localize(locale = "ru", value = "Шея")
	private static String			NECK;
	@Localize("Vitals")
	@Localize(locale = "de", value = "Organe")
	@Localize(locale = "ru", value = "Жиз.орг.")
	private static String			VITALS;

	static {
		Localization.initialize();
	}

	/** The various hit locations. */
	public static final String[]	LOCATIONS	= new String[] { EYE, SKULL, FACE, RIGHT_LEG, RIGHT_ARM, TORSO, GROIN, LEFT_ARM, LEFT_LEG, HAND, FOOT, NECK, VITALS };
	/** The rolls needed for various hit locations. */
	public static final String[]	ROLLS		= new String[] { "-", "3-4", "5", "6-7", "8", "9-10", "11", "12", "13-14", "15", "16", "17-18", "-" };																																							//$NON-NLS-1$ //$NON-NLS-2$ //$NON-NLS-3$ //$NON-NLS-4$ //$NON-NLS-5$ //$NON-NLS-6$ //$NON-NLS-7$ //$NON-NLS-8$ //$NON-NLS-9$ //$NON-NLS-10$ //$NON-NLS-11$ //$NON-NLS-12$ //$NON-NLS-13$
	/** The to hit penalties for various hit locations. */
	public static final String[]	PENALTIES	= new String[] { "-9", "-7", "-5", "-2", "-2", "0", "-3", "-2", "-2", "-4", "-4", "-5", "-3" };																																								//$NON-NLS-1$ //$NON-NLS-2$ //$NON-NLS-3$ //$NON-NLS-4$ //$NON-NLS-5$ //$NON-NLS-6$ //$NON-NLS-7$ //$NON-NLS-8$ //$NON-NLS-9$ //$NON-NLS-10$ //$NON-NLS-11$ //$NON-NLS-12$ //$NON-NLS-13$
	/** The DR for various hit locations. */
	public static final String[]	DR_KEYS		= new String[] { Armor.ID_EYES_DR, Armor.ID_SKULL_DR, Armor.ID_FACE_DR, Armor.ID_LEG_DR, Armor.ID_ARM_DR, Armor.ID_TORSO_DR, Armor.ID_GROIN_DR, Armor.ID_ARM_DR, Armor.ID_LEG_DR, Armor.ID_HAND_DR, Armor.ID_FOOT_DR, Armor.ID_NECK_DR, Armor.ID_TORSO_DR };

	/**
	 * Creates a new hit location panel.
	 *
	 * @param character The character to display the data for.
	 */
	public HitLocationPanel(GURPSCharacter character) {
		super(new ColumnLayout(7, 2, 0), HIT_LOCATION);

		int i;

		Wrapper wrapper = new Wrapper(new ColumnLayout(1, 2, 0));
		PageHeader header = createHeader(wrapper, ROLL, null);
		addHorizontalBackground(header, Color.black);
		for (i = 0; i < LOCATIONS.length; i++) {
			createLabel(wrapper, ROLLS[i], MessageFormat.format(ROLL_TOOLTIP, LOCATIONS[i]), SwingConstants.CENTER);
		}
		wrapper.setAlignmentY(TOP_ALIGNMENT);
		add(wrapper);

		createDivider();

		wrapper = new Wrapper(new ColumnLayout(1, 2, 0));
		header = createHeader(wrapper, LOCATION, null);
		for (i = 0; i < LOCATIONS.length; i++) {
			wrapper.add(new PageLabel(LOCATIONS[i], header, SwingConstants.CENTER));
		}
		wrapper.setAlignmentY(TOP_ALIGNMENT);
		add(wrapper);

		createDivider();

		wrapper = new Wrapper(new ColumnLayout(1, 2, 0));
		header = createHeader(wrapper, PENALTY, PENALTY_TITLE_TOOLTIP);
		for (i = 0; i < LOCATIONS.length; i++) {
			createLabel(wrapper, PENALTIES[i], MessageFormat.format(PENALTY_TOOLTIP, LOCATIONS[i]), SwingConstants.RIGHT);
		}
		wrapper.setAlignmentY(TOP_ALIGNMENT);
		add(wrapper);

		createDivider();

		wrapper = new Wrapper(new ColumnLayout(1, 2, 0));
		header = createHeader(wrapper, DR, null);
		for (i = 0; i < LOCATIONS.length; i++) {
			createDisabledField(wrapper, character, DR_KEYS[i], MessageFormat.format(DR_TOOLTIP, LOCATIONS[i]), SwingConstants.RIGHT);
		}
		wrapper.setAlignmentY(TOP_ALIGNMENT);
		add(wrapper);
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
		addVerticalBackground(panel, Color.black);
	}

	private static void createLabel(Container panel, String title, String tooltip, int alignment) {
		PageLabel label = new PageLabel(title, null);
		label.setHorizontalAlignment(alignment);
		label.setToolTipText(tooltip);
		panel.add(label);
	}
}
