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
import com.trollworks.toolkit.ui.layout.RowDistribution;
import com.trollworks.toolkit.ui.widget.Wrapper;
import com.trollworks.toolkit.utility.Localization;

import java.awt.Color;
import java.awt.Container;
import java.awt.Dimension;

import javax.swing.SwingConstants;

/** The character hit points panel. */
public class HitPointsPanel extends DropPanel {
	@Localize("Fatigue/Hit Points")
	@Localize(locale = "de", value = "Erschöpfung/Treffer")
	@Localize(locale = "ru", value = "Очки усталости/жизни")
	@Localize(locale = "es", value = "Puntos de Fatiga/Vida")
	private static String	FP_HP;
	@Localize("Basic HP:")
	@Localize(locale = "de", value = "Normale TP:")
	@Localize(locale = "ru", value = "Баз. ЕЖ:")
	@Localize(locale = "es", value = "PV Básicos")
	private static String	HP;
	@Localize("<html><body>Normal (i.e. unharmed) hit points.<br><b>{0} points</b> have been spent to modify <b>HP</b></body></html>")
	@Localize(locale = "de", value = "<html><body>Normale (d.h. unverletzt) Trefferpunke.<br><b>{0} Punkte</b> wurden für die Veränderung von <b>TP</b> aufgewendet</body></html>")
	@Localize(locale = "ru", value = "<html><body>Обычные единицы жизни (т.е. персонаж невредимый).<br><b>{0} очков</b> использовано на изменение <b>ЗД</b></body></html>")
	@Localize(locale = "es", value = "<html><body>Puntos de Vida normales (p.e. ileso).<br><b>{0} puntos</b> que se han consumido para modificar <b>PV</b></body></html>")
	private static String	HP_TOOLTIP;
	@Localize("Current HP:")
	@Localize(locale = "de", value = "Aktuelle TP:")
	@Localize(locale = "ru", value = "Тек. ЕЖ:")
	@Localize(locale = "es", value = "PV actuales:")
	private static String	HP_CURRENT;
	@Localize("Current hit points")
	@Localize(locale = "de", value = "Aktuelle Trefferpunkte")
	@Localize(locale = "ru", value = "Текущие очки (единицы) жизни")
	@Localize(locale = "es", value = "Puntos de Vida actuales")
	private static String	HP_CURRENT_TOOLTIP;
	@Localize("Reeling:")
	@Localize(locale = "de", value = "Taumeln:")
	@Localize(locale = "ru", value = "Измотан:")
	@Localize(locale = "es", value = "Aturdido:")
	private static String	HP_REELING;
	@Localize("<html><body>Current hit points at or below this point indicate the character<br>is reeling from the pain, halving move, speed and dodge</body></html>")
	@Localize(locale = "de", value = "<html><body>Wenn die aktuellen Trefferpunkte auf diesem Wert oder geringer sind,<br>bedeutet dies, dass der Charakter vor Schmerzen taumelt,<br>was Bewegungspunkte, Geschwindigkeit und Ausweichen halbiert</body></html>")
	@Localize(locale = "ru", value = "<html><body>ЕЖ, меньшее или равное этому числу означает, что персонаж<br>измотан, и имеет половину движения, скорости и уклонения</body></html>")
	@Localize(locale = "es", value = "<html><body>Con estos puntos de vida o menos, el personaje está aturdido<br>por el dolor, por lo que reduce movimiento, velocidad y parada a la mitad</body></html>")
	private static String	HP_REELING_TOOLTIP;
	@Localize("Collapse:")
	@Localize(locale = "de", value = "Kollaps:")
	@Localize(locale = "ru", value = "Без сил:")
	@Localize(locale = "es", value = "Colapso:")
	private static String	HP_UNCONSCIOUS_CHECKS;
	@Localize("<html><body>Current hit points at or below this point indicate the character<br>is on the verge of collapse, causing the character to <b>roll vs. HT</b><br>(at -1 per full multiple of HP below zero) every second to avoid<br>falling unconscious</body></html>")
	@Localize(locale = "de", value = "<html><body>Wenn die aktuellen Trefferpunkte auf diesem Wert oder geringer sind,<br>bedeutet dies, dass der Charakter vor dem Kollaps steht,<br>wodurch der Charakter jede Sekunde <b>eine KO-Probe</b><br>(mit -1 für jedes Vielfache der TP unter Null) bestehen muss,<br>um nicht in Ohnmacht zu fallen.</body></html>")
	@Localize(locale = "ru", value = "<html><body>ЕЖ, меньшее или равное этому числу означает, что персонаж <br>почти истощен и должен <b>делать бросок ЗД</b> (-1 за каждое полное <br>кратное ЗД ниже нуля) каждую секунду, чтобы не упасть без сознания</body></html>")
	@Localize(locale = "es", value = "<html><body>Con estos Puntos de Fatiga o menos, el personaje está al<br>borde del colapso, por lo que realizará una <b>tirada contra SL</b><br>(con -1 por cada múltiplo completo de los PV totales por debajo de cero)<br>cada segundo para evitar caer inconsciente</body></html>")
	private static String	HP_UNCONSCIOUS_CHECKS_TOOLTIP;
	@Localize("Check #1:")
	@Localize(locale = "de", value = "Todesprobe 1:")
	@Localize(locale = "ru", value = "Бросок#1:")
	@Localize(locale = "es", value = "Chequeo nº1:")
	private static String	HP_DEATH_CHECK_1;
	@Localize("<html><body>Current hit points at or below this point cause<br>the character to <b>roll vs. HT</b> to avoid death</body></html>")
	@Localize(locale = "de", value = "<html><body>Wenn die aktuellen Trefferpunkte auf diesem Wert oder geringer sind,<br>muss der Charakter <b>eine KO-Probe</b> bestehen, um nicht zu sterben.</body></html>")
	@Localize(locale = "ru", value = "<html><body>ЕЖ, меньшее или равное этому числу означает, что персонажу<br>нужно <b>сделать бросок ЗД</b> для избежания смерти</body></html>")
	@Localize(locale = "es", value = "<html><body>Con estos puntos de Vida o menos, el personaje<br>realizará una <b>tirada contra SL</b> para evitar caer muerto</body></html>")
	private static String	HP_DEATH_CHECK_1_TOOLTIP;
	@Localize("Check #2:")
	@Localize(locale = "de", value = "Todesprobe 2:")
	@Localize(locale = "ru", value = "Бросок#2:")
	@Localize(locale = "es", value = "Chequeo nº2:")
	private static String	HP_DEATH_CHECK_2;
	@Localize("<html><body>Current hit points at or below this point cause<br>the character to <b>roll vs. HT</b> to avoid death</body></html>")
	@Localize(locale = "de", value = "<html><body>Wenn die aktuellen Trefferpunkte auf diesem Wert oder geringer sind,<br>muss der Charakter <b>eine KO-Probe</b> bestehen, um nicht zu sterben.</body></html>")
	@Localize(locale = "ru", value = "<html><body>ЕЖ, меньшее или равное этому числу означает, что персонажу<br>нужно <b>сделать бросок ЗД</b> для избежания смерти</body></html>")
	@Localize(locale = "es", value = "<html><body>Con estos puntos de Vida o menos, el personaje<br>realizará una <b>tirada contra SL</b> para evitar caer muerto</body></html>")
	private static String	HP_DEATH_CHECK_2_TOOLTIP;
	@Localize("Check #3:")
	@Localize(locale = "de", value = "Todesprobe 3:")
	@Localize(locale = "ru", value = "Бросок#3:")
	@Localize(locale = "es", value = "Chequeo nº3:")
	private static String	HP_DEATH_CHECK_3;
	@Localize("<html><body>Current hit points at or below this point cause<br>the character to <b>roll vs. HT</b> to avoid death</body></html>")
	@Localize(locale = "de", value = "<html><body>Wenn die aktuellen Trefferpunkte auf diesem Wert oder geringer sind,<br>muss der Charakter <b>eine KO-Probe</b> bestehen, um nicht zu sterben.</body></html>")
	@Localize(locale = "ru", value = "<html><body>ЕЖ, меньшее или равное этому числу означает, что персонажу<br>нужно <b>сделать бросок ЗД</b> для избежания смерти</body></html>")
	@Localize(locale = "es", value = "<html><body>Con estos puntos de Vida o menos, el personaje<br>realizará una <b>tirada contra SL</b> para evitar caer muerto</body></html>")
	private static String	HP_DEATH_CHECK_3_TOOLTIP;
	@Localize("Check #4:")
	@Localize(locale = "de", value = "Todesprobe 4:")
	@Localize(locale = "ru", value = "Бросок#4:")
	@Localize(locale = "es", value = "Chequeo nº4:")
	private static String	HP_DEATH_CHECK_4;
	@Localize("<html><body>Current hit points at or below this point cause<br>the character to <b>roll vs. HT</b> to avoid death</body></html>")
	@Localize(locale = "de", value = "<html><body>Wenn die aktuellen Trefferpunkte auf diesem Wert oder geringer sind,<br>muss der Charakter <b>eine KO-Probe</b> bestehen, um nicht zu sterben.</body></html>")
	@Localize(locale = "ru", value = "<html><body>ЕЖ, меньшее или равное этому числу означает, что персонажу<br>нужно <b>сделать бросок ЗД</b> для избежания смерти</body></html>")
	@Localize(locale = "es", value = "<html><body>Con estos puntos de Vida o menos, el personaje<br>realizará una <b>tirada contra SL</b> para evitar caer muerto</body></html>")
	private static String	HP_DEATH_CHECK_4_TOOLTIP;
	@Localize("Dead:")
	@Localize(locale = "de", value = "Tod:")
	@Localize(locale = "ru", value = "Мертв:")
	@Localize(locale = "es", value = "Muerto:")
	private static String	HP_DEAD;
	@Localize("<html><body>Current hit points at or below this<br>point cause the character to die</body></html>")
	@Localize(locale = "de", value = "<html><body>Wenn die aktuellen Trefferpunkte auf diesem Wert oder geringer sind,<br>stirbt der Charakter.</body></html>")
	@Localize(locale = "ru", value = "<html><body>ЕЖ, меньшее или равное этому числу означает<br>смерть персонажа</body></html>")
	@Localize(locale = "es", value = "<html><body>Con estos puntos de Vida o menos, el personaje caerá muerto</body></html>")
	private static String	HP_DEAD_TOOLTIP;
	@Localize("Basic FP:")
	@Localize(locale = "de", value = "Normale EP")
	@Localize(locale = "ru", value = "Баз. ЕУ:")
	@Localize(locale = "es", value = "PF Básicos:")
	private static String	FP;
	@Localize("<html><body>Normal (i.e. fully rested) fatigue points.<br><b>{0} points</b> have been spent to modify <b>FP</b></body></html>")
	@Localize(locale = "de", value = "<html><body>Normale (d.h. ausgeruhte) Erschöpfungspunkte.<br><b>{0} Punkte</b> wurden für die Veränderung von <b>FP</b> aufgewendet</body></html>")
	@Localize(locale = "ru", value = "<html><body>Обычные единицы усталости (т.е. полностью отдохнувший персонаж).<br><b>{0} очков</b> использовано на изменение <b>ЕУ</b></body></html>")
	@Localize(locale = "es", value = "<html><body>Puntos de Fatiga normales (p.e. descansado).<br><b>{0} puntos</b> que se han consumido para modificar <b>PV</b></body></html>")
	private static String	FP_TOOLTIP;
	@Localize("Current FP:")
	@Localize(locale = "de", value = "Aktuelle EP")
	@Localize(locale = "ru", value = "Тек. ЕУ:")
	@Localize(locale = "es", value = "PF Actuales:")
	private static String	FP_CURRENT;
	@Localize("Current fatigue points")
	@Localize(locale = "de", value = "Aktuelle Erschöpfungspunkte")
	@Localize(locale = "ru", value = "Текущие очки (единицы) усталости")
	@Localize(locale = "es", value = "Puntos de Fatiga Actuales:")
	private static String	FP_CURRENT_TOOLTIP;
	@Localize("Tired:")
	@Localize(locale = "de", value = "Müde")
	@Localize(locale = "ru", value = "Утомлен:")
	@Localize(locale = "es", value = "Cansado:")
	private static String	FP_TIRED;
	@Localize("<html><body>Current fatigue points at or below this point indicate the<br>character is very tired, halving move, dodge and strength</body></html>")
	@Localize(locale = "de", value = "<html><body>Wenn die aktuellen Trefferpunkte auf diesem Wert oder geringer sind,<br>bedeutet dies, dass der Charakter sehr müde ist, was seine Bewegung,<br>ausweichen und Stärke halbiert.</body></html>")
	@Localize(locale = "ru", value = "<html><body>ЕУ, меньшее или равное этому числу означает, что персонаж почти<br>истощен. Движение, уклонение и сила уменьшаются вдвое</body></html>")
	@Localize(locale = "es", value = "<html><body>Con estos Puntos de Fatiga o menos, el personaje está muy<br>cansado, , por lo que reduce movimiento, parada y fuerza a la mitad</body></html>")
	private static String	FP_TIRED_TOOLTIP;
	@Localize("Collapse:")
	@Localize(locale = "de", value = "Kollaps:")
	@Localize(locale = "ru", value = "Без сил:")
	@Localize(locale = "es", value = "Colapso:")
	private static String	FP_UNCONSCIOUS_CHECKS;
	@Localize("<html><body>Current fatigue points at or below this point indicate the<br>character is on the verge of collapse, causing the character<br>to roll vs. Will to do anything besides talk or rest</body></html>")
	@Localize(locale = "de", value = "<html><body>Wenn die aktuellen Erschöpfungspunkte auf diesem Wert oder geringer sind,<br>bedeutet dies, dass der Charakter vor dem Kollaps steht,<br>wodurch der Charakter jedes mal, wenn er etwas anderes macht<br>als reden oder ausruhen, <b>eine Willens-Probe</b> bestehen muss.</body></html>")
	@Localize(locale = "ru", value = "<html><body>ЕУ, меньшее или равное этому числу означает, что персонаж<br>почти истощен, и персонаж делает бросок Воли,<br>чтобы сделать что-либо, кроме отдыха и разговора.</body></html>")
	@Localize(locale = "es", value = "<html><body>Con estos Puntos de Fatiga o menos, el personaje está al<br>borde del colapso, por lo que deberá tirar contra<br>Voluntad para realizar cualquier actividad más allá de hablar o descansar</body></html>")
	private static String	FP_UNCONSCIOUS_CHECKS_TOOLTIP;
	@Localize("Unconscious:")
	@Localize(locale = "de", value = "Bewusstlos:")
	@Localize(locale = "ru", value = "Обморок:")
	@Localize(locale = "es", value = "Incosnciente:")
	private static String	FP_UNCONSCIOUS;
	@Localize("<html><body>Current fatigue points at or below this point<br>cause the character to fall unconscious</body></html>")
	@Localize(locale = "de", value = "<html><body>Wenn die aktuellen Trefferpunkte auf diesem Wert oder geringer sind,<br>wird der Charakter bewusstlos.</body></html>")
	@Localize(locale = "ru", value = "<html><body>ЕУ, меньшее или равное этому числу<br>означает, что персонаж теряет сознание</body></html>")
	@Localize(locale = "es", value = "<html><body>Con estos Puntos de Fatiga o menos<br>, el personaje cae incosnciente</body></html>")
	private static String	FP_UNCONSCIOUS_TOOLTIP;

	static {
		Localization.initialize();
	}

	/**
	 * Creates a new hit points panel.
	 *
	 * @param character The character to display the data for.
	 */
	public HitPointsPanel(GURPSCharacter character) {
		super(new ColumnLayout(2, 2, 0, RowDistribution.DISTRIBUTE_HEIGHT), FP_HP);
		createLabelAndField(this, character, GURPSCharacter.ID_CURRENT_FATIGUE_POINTS, FP_CURRENT, FP_CURRENT_TOOLTIP, SwingConstants.RIGHT);
		createLabelAndField(this, character, GURPSCharacter.ID_FATIGUE_POINTS, FP, FP_TOOLTIP, SwingConstants.RIGHT);
		createDivider();
		createLabelAndDisabledField(this, character, GURPSCharacter.ID_TIRED_FATIGUE_POINTS, FP_TIRED, FP_TIRED_TOOLTIP, SwingConstants.RIGHT);
		createLabelAndDisabledField(this, character, GURPSCharacter.ID_UNCONSCIOUS_CHECKS_FATIGUE_POINTS, FP_UNCONSCIOUS_CHECKS, FP_UNCONSCIOUS_CHECKS_TOOLTIP, SwingConstants.RIGHT);
		createLabelAndDisabledField(this, character, GURPSCharacter.ID_UNCONSCIOUS_FATIGUE_POINTS, FP_UNCONSCIOUS, FP_UNCONSCIOUS_TOOLTIP, SwingConstants.RIGHT);
		createDivider();
		createLabelAndField(this, character, GURPSCharacter.ID_CURRENT_HIT_POINTS, HP_CURRENT, HP_CURRENT_TOOLTIP, SwingConstants.RIGHT);
		createLabelAndField(this, character, GURPSCharacter.ID_HIT_POINTS, HP, HP_TOOLTIP, SwingConstants.RIGHT);
		createDivider();
		createLabelAndDisabledField(this, character, GURPSCharacter.ID_REELING_HIT_POINTS, HP_REELING, HP_REELING_TOOLTIP, SwingConstants.RIGHT);
		createLabelAndDisabledField(this, character, GURPSCharacter.ID_UNCONSCIOUS_CHECKS_HIT_POINTS, HP_UNCONSCIOUS_CHECKS, HP_UNCONSCIOUS_CHECKS_TOOLTIP, SwingConstants.RIGHT);
		createLabelAndDisabledField(this, character, GURPSCharacter.ID_DEATH_CHECK_1_HIT_POINTS, HP_DEATH_CHECK_1, HP_DEATH_CHECK_1_TOOLTIP, SwingConstants.RIGHT);
		createLabelAndDisabledField(this, character, GURPSCharacter.ID_DEATH_CHECK_2_HIT_POINTS, HP_DEATH_CHECK_2, HP_DEATH_CHECK_2_TOOLTIP, SwingConstants.RIGHT);
		createLabelAndDisabledField(this, character, GURPSCharacter.ID_DEATH_CHECK_3_HIT_POINTS, HP_DEATH_CHECK_3, HP_DEATH_CHECK_3_TOOLTIP, SwingConstants.RIGHT);
		createLabelAndDisabledField(this, character, GURPSCharacter.ID_DEATH_CHECK_4_HIT_POINTS, HP_DEATH_CHECK_4, HP_DEATH_CHECK_4_TOOLTIP, SwingConstants.RIGHT);
		createLabelAndDisabledField(this, character, GURPSCharacter.ID_DEAD_HIT_POINTS, HP_DEAD, HP_DEAD_TOOLTIP, SwingConstants.RIGHT);
	}

	@Override
	public Dimension getMaximumSize() {
		Dimension size = super.getMaximumSize();

		size.width = getPreferredSize().width;
		return size;
	}

	private void createDivider() {
		createOneByOnePanel();
		createOneByOnePanel();
		addHorizontalBackground(createOneByOnePanel(), Color.black);
		createOneByOnePanel();
	}

	private Container createOneByOnePanel() {
		Wrapper panel = new Wrapper();
		UIUtilities.setOnlySize(panel, new Dimension(1, 1));
		add(panel);
		return panel;
	}
}
