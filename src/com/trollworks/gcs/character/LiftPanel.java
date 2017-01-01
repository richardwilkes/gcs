/*
 * Copyright (c) 1998-2017 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, version 2.0. If a copy of the MPL was not distributed with
 * this file, You can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as
 * defined by the Mozilla Public License, version 2.0.
 */

package com.trollworks.gcs.character;

import com.trollworks.gcs.page.DropPanel;
import com.trollworks.toolkit.annotation.Localize;
import com.trollworks.toolkit.ui.layout.ColumnLayout;
import com.trollworks.toolkit.utility.Localization;

import javax.swing.SwingConstants;

/** The character damage panel. */
public class LiftPanel extends DropPanel {
	@Localize("Lifting & Moving Things")
	@Localize(locale = "de", value = "Gegenstände heben und bewegen")
	@Localize(locale = "ru", value = "Поднятие и перемещение предметов")
	@Localize(locale = "es", value = "Levantando y Moviendo objetos")
	private static String	LIFT_MOVE;
	@Localize("Basic Lift:")
	@Localize(locale = "de", value = "Grundtragkraft")
	@Localize(locale = "ru", value = "Базовый Груз:")
	@Localize(locale = "es", value = "Levantar:")
	private static String	BASIC_LIFT;
	@Localize("The weight the character can lift overhead with one hand in one second")
	@Localize(locale = "de", value = "Das Gewicht, das der Charakter mit einer Hand in einer Sekunde überkopf heben kann")
	@Localize(locale = "ru", value = "Вес, который может поднять персонаж одной рукой на 1 сек")
	@Localize(locale = "es", value = "Peso que el personaje puede levantar por encima de su cabeza con una mano en un segundo")
	private static String	BASIC_LIFT_TOOLTIP;
	@Localize("One-Handed Lift:")
	@Localize(locale = "de", value = "Einhändig heben")
	@Localize(locale = "ru", value = "Подъём одной руки:")
	@Localize(locale = "es", value = "Levantar con una mano:")
	private static String	ONE_HANDED_LIFT;
	@Localize("The weight the character can lift overhead with one hand in two seconds")
	@Localize(locale = "de", value = "Das Gewicht, das der Charakter mit einer Hand in zwei Sekunden überkopf heben kann")
	@Localize(locale = "ru", value = "Вес, который может поднять персонаж одной рукой на 2 сек")
	@Localize(locale = "es", value = "Peso que el personaje puede levantar por encima de su cabeza con una mano en dos segundos")
	private static String	ONE_HANDED_LIFT_TOOLTIP;
	@Localize("Two-Handed Lift:")
	@Localize(locale = "de", value = "Zweihändig heben")
	@Localize(locale = "ru", value = "Подъём двумя руками:")
	@Localize(locale = "es", value = "Levantar con las dos manos:")
	private static String	TWO_HANDED_LIFT;
	@Localize("The weight the character can lift overhead with both hands in four seconds")
	@Localize(locale = "de", value = "Das Gewicht, das der Charakter mit beiden Händen in vier Sekunden überkopf heben kann")
	@Localize(locale = "ru", value = "Вес, который может поднять персонаж двумя руками на 4 сек")
	@Localize(locale = "es", value = "Peso que el personaje puede levantar por encima de su cabeza con dos manos en cuatro segundos")
	private static String	TWO_HANDED_LIFT_TOOLTIP;
	@Localize("Shove & Knock Over:")
	@Localize(locale = "de", value = "Schieben & Umstoßen")
	@Localize(locale = "ru", value = "Толчок и опрокид-ние:")
	@Localize(locale = "es", value = "Empujar y Derribar")
	private static String	SHOVE_KNOCK_OVER;
	@Localize("The weight of an object the character can shove and knock over")
	@Localize(locale = "de", value = "Das Gewicht eines Objektes, das der Charakter schieben und umstoßen kann")
	@Localize(locale = "ru", value = "Вес обьектов, который персонаж может столкнуть и опрокинуть")
	@Localize(locale = "es", value = "Peso de un objeto que el personaje puede empujar y derribar")
	private static String	SHOVE_KNOCK_OVER_TOOLTIP;
	@Localize("Running Shove & Knock Over:")
	@Localize(locale = "de", value = "Sch. & Umst. mit Anlauf")
	@Localize(locale = "ru", value = "Толчок в движ. и опр.:")
	@Localize(locale = "es", value = "Emp. y Der. con carrerilla:")
	private static String	RUNNING_SHOVE;
	@Localize("The weight of an object the character can shove  and knock over with a running start")
	@Localize(locale = "de", value = "Das Gewicht eines Objektes, das der Charakter mit Anlauf schieben und umstoßen kann")
	@Localize(locale = "ru", value = "Вес обьектов, который персонаж может столкнуть и опрокинуть с разбегу")
	@Localize(locale = "es", value = "Peso de un objeto que el personaje puede empujar y derribar con carrerilla")
	private static String	RUNNING_SHOVE_TOOLTIP;
	@Localize("Carry On Back:")
	@Localize(locale = "de", value = "Auf dem Rücken tragen")
	@Localize(locale = "ru", value = "Нести на спине:")
	@Localize(locale = "es", value = "Cargar a la espalda:")
	private static String	CARRY_ON_BACK;
	@Localize("The weight the character can carry slung across the back")
	@Localize(locale = "de", value = "Das Gewicht, das der Charakter auf den Rücken gebunden tragen kann")
	@Localize(locale = "ru", value = "Вес, который может персонаж нести, перекинув через спину")
	@Localize(locale = "es", value = "Peso que el personaje puede cargar a la espalda")
	private static String	CARRY_ON_BACK_TOOLTIP;
	@Localize("Shift Slightly:")
	@Localize(locale = "de", value = "Geringfügig verschieben")
	@Localize(locale = "ru", value = "Тащить:")
	@Localize(locale = "es", value = "Desplazar Ligeramente")
	private static String	SHIFT_SLIGHTLY;
	@Localize("The weight of an object the character can shift slightly on a floor")
	@Localize(locale = "de", value = "Das Gewicht eines Objektes, das der Charakter auf einem Boden geringfügig verschieben kann")
	@Localize(locale = "ru", value = "Вес обьектов, который персонаж может немного сдвинуть по полу")
	@Localize(locale = "es", value = "Peso de un objeto que el personaje puede desplazar ligeramente por el suelo")
	private static String	SHIFT_SLIGHTLY_TOOLTIP;

	static {
		Localization.initialize();
	}

	/**
	 * Creates a new damage panel.
	 *
	 * @param sheet The sheet to display the data for.
	 */
	public LiftPanel(CharacterSheet sheet) {
		super(new ColumnLayout(2, 2, 0), LIFT_MOVE);
		createLabelAndDisabledField(this, sheet, GURPSCharacter.ID_BASIC_LIFT, BASIC_LIFT, BASIC_LIFT_TOOLTIP, SwingConstants.RIGHT);
		createLabelAndDisabledField(this, sheet, GURPSCharacter.ID_ONE_HANDED_LIFT, ONE_HANDED_LIFT, ONE_HANDED_LIFT_TOOLTIP, SwingConstants.RIGHT);
		createLabelAndDisabledField(this, sheet, GURPSCharacter.ID_TWO_HANDED_LIFT, TWO_HANDED_LIFT, TWO_HANDED_LIFT_TOOLTIP, SwingConstants.RIGHT);
		createLabelAndDisabledField(this, sheet, GURPSCharacter.ID_SHOVE_AND_KNOCK_OVER, SHOVE_KNOCK_OVER, SHOVE_KNOCK_OVER_TOOLTIP, SwingConstants.RIGHT);
		createLabelAndDisabledField(this, sheet, GURPSCharacter.ID_RUNNING_SHOVE_AND_KNOCK_OVER, RUNNING_SHOVE, RUNNING_SHOVE_TOOLTIP, SwingConstants.RIGHT);
		createLabelAndDisabledField(this, sheet, GURPSCharacter.ID_CARRY_ON_BACK, CARRY_ON_BACK, CARRY_ON_BACK_TOOLTIP, SwingConstants.RIGHT);
		createLabelAndDisabledField(this, sheet, GURPSCharacter.ID_SHIFT_SLIGHTLY, SHIFT_SLIGHTLY, SHIFT_SLIGHTLY_TOOLTIP, SwingConstants.RIGHT);
	}
}
