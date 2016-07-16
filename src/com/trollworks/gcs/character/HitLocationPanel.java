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

import com.trollworks.toolkit.annotation.Localize;
import com.trollworks.toolkit.ui.UIUtilities;
import com.trollworks.toolkit.ui.layout.ColumnLayout;
import com.trollworks.toolkit.ui.widget.Wrapper;
import com.trollworks.toolkit.utility.Localization;
import com.trollworks.toolkit.utility.text.Text;

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
	@Localize(locale = "es", value = "Localización de Impactos")
	private static String	HIT_LOCATION;
	@Localize("Roll")
	@Localize(locale = "de", value = "Wurf")
	@Localize(locale = "ru", value = "ДБ")
	@Localize(locale = "es", value = "Tirada")
	private static String	ROLL;
	@Localize("<html><body>The random roll needed to hit the <b>{0}</b> hit location</body></html>")
	@Localize(locale = "de", value = "<html><body>Der Würfelwurf, um die Trefferzone <b>{0}</b> zu treffen</body></html>")
	@Localize(locale = "ru", value = "<html><body>Для попадания в <b>{0}</b>, необходимо сделать дополнительный бросок (ДБ) и выбросить указанные числа</body></html>")
	@Localize(locale = "es", value = "<html><body>Tirada al azar requerida para alcanzar la localización del impacto <b>{0}</b></body></html>")
	private static String	ROLL_TOOLTIP;
	@Localize("Where")
	@Localize(locale = "de", value = "Zone")
	@Localize(locale = "ru", value = "Где")
	@Localize(locale = "es", value = "Localización")
	private static String	LOCATION;
	@Localize("-")
	@Localize(locale = "de", value = "-")
	private static String	PENALTY;
	@Localize("The hit penalty for targeting a specific hit location")
	@Localize(locale = "de", value = "Der Treffernachteil für das Zielen auf eine spezifische Trefferzone")
	@Localize(locale = "ru", value = "Штраф для попадания в указанную зону")
	@Localize(locale = "es", value = "Penalización a la tirada por apuntar a una determinada localización de impacto")
	private static String	PENALTY_TITLE_TOOLTIP;
	@Localize("<html><body>The hit penalty for targeting the <b>{0}</b> hit location</body></html>")
	@Localize(locale = "de", value = "<html><body>Der Treffernachteil für das Zielen auf die Trefferzone <b>{0}</b></body></html>")
	@Localize(locale = "ru", value = "<html><body>Штраф за прицеливания в зону попадания <b>{0}</b></body></html>")
	@Localize(locale = "es", value = "<html><body>penalización a la tirada por apuntar a <b>{0}</b></body></html>")
	private static String	PENALTY_TOOLTIP;
	@Localize("DR")
	@Localize(locale = "de", value = "SR")
	@Localize(locale = "ru", value = "СП")
	@Localize(locale = "es", value = "RD")
	private static String	DR;
	@Localize("<html><body>The total DR protecting the <b>{0}</b> hit location</body></html>")
	@Localize(locale = "de", value = "<html><body>Die Gesamte Schadensresistenz, die die Trefferzone <b>{0}</b> schützt</body></html>")
	@Localize(locale = "ru", value = "<html><body>Суммарное СП, защищающее зону попадания: <b>{0}</b></body></html>")
	@Localize(locale = "es", value = "Total de RD ue protege la localización <b>{0}</b>")
	private static String	DR_TOOLTIP;

	static {
		Localization.initialize();
	}

	/**
	 * Creates a new hit location panel.
	 *
	 * @param character The character to display the data for.
	 */
	public HitLocationPanel(GURPSCharacter character) {
		super(new ColumnLayout(7, 2, 0), HIT_LOCATION);

		HitLocationTable table = character.getDescription().getHitLocationTable();

		Wrapper wrapper = new Wrapper(new ColumnLayout(1, 2, 0));
		PageHeader header = createHeader(wrapper, ROLL, null);
		addHorizontalBackground(header, Color.black);
		for (HitLocationTableEntry entry : table.getEntries()) {
			createLabel(wrapper, entry.getRoll(), MessageFormat.format(ROLL_TOOLTIP, entry.getName()), SwingConstants.CENTER);
		}
		wrapper.setAlignmentY(TOP_ALIGNMENT);
		add(wrapper);

		createDivider();

		wrapper = new Wrapper(new ColumnLayout(1, 2, 0));
		header = createHeader(wrapper, LOCATION, null);
		for (HitLocationTableEntry entry : table.getEntries()) {
			createLabel(wrapper, entry.getName(), Text.wrapPlainTextForToolTip(entry.getLocation().getDescription()), SwingConstants.CENTER);
		}
		wrapper.setAlignmentY(TOP_ALIGNMENT);
		add(wrapper);

		createDivider();

		wrapper = new Wrapper(new ColumnLayout(1, 2, 0));
		header = createHeader(wrapper, PENALTY, PENALTY_TITLE_TOOLTIP);
		for (HitLocationTableEntry entry : table.getEntries()) {
			createLabel(wrapper, Integer.toString(entry.getHitPenalty()), MessageFormat.format(PENALTY_TOOLTIP, entry.getName()), SwingConstants.RIGHT);
		}
		wrapper.setAlignmentY(TOP_ALIGNMENT);
		add(wrapper);

		createDivider();

		wrapper = new Wrapper(new ColumnLayout(1, 2, 0));
		header = createHeader(wrapper, DR, null);
		for (HitLocationTableEntry entry : table.getEntries()) {
			createDisabledField(wrapper, character, entry.getKey(), MessageFormat.format(DR_TOOLTIP, entry.getName()), SwingConstants.RIGHT);
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
