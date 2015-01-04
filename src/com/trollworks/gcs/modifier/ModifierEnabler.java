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

package com.trollworks.gcs.modifier;

import com.trollworks.gcs.advantage.Advantage;
import com.trollworks.gcs.advantage.SelfControlRoll;
import com.trollworks.toolkit.annotation.Localize;
import com.trollworks.toolkit.ui.UIUtilities;
import com.trollworks.toolkit.ui.layout.ColumnLayout;
import com.trollworks.toolkit.ui.widget.WindowUtils;
import com.trollworks.toolkit.utility.Localization;
import com.trollworks.toolkit.utility.text.TextUtility;

import java.awt.BorderLayout;
import java.awt.Color;
import java.awt.Component;
import java.awt.Container;
import java.awt.Dimension;
import java.text.MessageFormat;
import java.util.ArrayList;
import java.util.Arrays;

import javax.swing.JCheckBox;
import javax.swing.JComboBox;
import javax.swing.JLabel;
import javax.swing.JOptionPane;
import javax.swing.JPanel;
import javax.swing.JScrollPane;
import javax.swing.SwingConstants;
import javax.swing.border.CompoundBorder;
import javax.swing.border.EmptyBorder;
import javax.swing.border.LineBorder;

/** Asks the user to enable/disable modifiers. */
public class ModifierEnabler extends JPanel {
	@Localize("Enable Modifiers")
	@Localize(locale = "de", value = "Modifikatoren auswählen")
	@Localize(locale = "ru", value = "Включить модификаторы")
	private static String		MODIFIER_TITLE;
	@Localize("1 advantage remaining to be processed.")
	@Localize(locale = "de", value = "1 weiterer Vorteil zu bearbeiten.")
	@Localize(locale = "ru", value = "осталось обработать 1 преимущество.")
	private static String		MODIFIER_ONE_REMAINING;
	@Localize("{0} advantages remaining to be processed.")
	@Localize(locale = "de", value = "{0} weitere Vorteile zu bearbeiten.")
	@Localize(locale = "ru", value = "{0} преимуществ(а) осталось обработать.")
	private static String		MODIFIER_REMAINING;
	@Localize("Cancel Remaining")
	@Localize(locale = "de", value = "Alles Abbrechen")
	@Localize(locale = "ru", value = "Пропустить остальные")
	private static String		CANCEL_REST;
	@Localize("Cancel")
	@Localize(locale = "de", value = "Abbrechen")
	@Localize(locale = "ru", value = "Отмена")
	private static String		CANCEL;
	@Localize("Apply")
	@Localize(locale = "de", value = "Anwenden")
	@Localize(locale = "ru", value = "Применить")
	private static String		APPLY;

	static {
		Localization.initialize();
	}

	private Advantage			mAdvantage;
	private JCheckBox[]			mEnabled;
	private Modifier[]			mModifiers;
	private JComboBox<String>	mCRCombo;

	/**
	 * Brings up a modal dialog that allows {@link Modifier}s to be enabled or disabled for the
	 * specified {@link Advantage}s.
	 *
	 * @param comp The component to open the dialog over.
	 * @param advantages The {@link Advantage}s to process.
	 * @return Whether anything was modified.
	 */
	static public boolean process(Component comp, ArrayList<Advantage> advantages) {
		ArrayList<Advantage> list = new ArrayList<>();
		boolean modified = false;
		int count;

		for (Advantage advantage : advantages) {
			if (advantage.getCR() != SelfControlRoll.NONE_REQUIRED || !advantage.getModifiers().isEmpty()) {
				list.add(advantage);
			}
		}

		count = list.size();
		for (int i = 0; i < count; i++) {
			Advantage advantage = list.get(i);
			boolean hasMore = i != count - 1;
			ModifierEnabler panel = new ModifierEnabler(advantage, count - i - 1);
			switch (WindowUtils.showOptionDialog(comp, panel, MODIFIER_TITLE, true, hasMore ? JOptionPane.YES_NO_CANCEL_OPTION : JOptionPane.YES_NO_OPTION, JOptionPane.INFORMATION_MESSAGE, advantage.getIcon(true), hasMore ? new String[] { APPLY, CANCEL, CANCEL_REST } : new String[] { APPLY, CANCEL }, APPLY)) {
				case JOptionPane.YES_OPTION:
					panel.applyChanges();
					modified = true;
					break;
				case JOptionPane.NO_OPTION:
					break;
				case JOptionPane.CANCEL_OPTION:
				case JOptionPane.CLOSED_OPTION:
				default:
					return modified;
			}
		}
		return modified;
	}

	private ModifierEnabler(Advantage advantage, int remaining) {
		super(new BorderLayout());
		mAdvantage = advantage;
		add(createTop(advantage, remaining), BorderLayout.NORTH);
		JScrollPane scrollPanel = new JScrollPane(createCenter());
		scrollPanel.setMinimumSize(new Dimension(500, 120));
		add(scrollPanel, BorderLayout.CENTER);
	}

	private static Container createTop(Advantage advantage, int remaining) {
		JPanel top = new JPanel(new ColumnLayout());
		JLabel label = new JLabel(TextUtility.truncateIfNecessary(advantage.toString(), 80, SwingConstants.RIGHT), SwingConstants.LEFT);

		top.setBorder(new EmptyBorder(0, 0, 15, 0));
		if (remaining > 0) {
			String msg;
			if (remaining == 1) {
				msg = MODIFIER_ONE_REMAINING;
			} else {
				msg = MessageFormat.format(MODIFIER_REMAINING, new Integer(remaining));
			}
			top.add(new JLabel(msg, SwingConstants.CENTER));
		}
		label.setBorder(new CompoundBorder(LineBorder.createBlackLineBorder(), new EmptyBorder(0, 2, 0, 2)));
		label.setOpaque(true);
		top.add(new JPanel());
		top.add(label);
		return top;
	}

	private Container createCenter() {
		JPanel wrapper = new JPanel(new ColumnLayout());
		wrapper.setBackground(Color.WHITE);
		SelfControlRoll cr = mAdvantage.getCR();
		if (cr != SelfControlRoll.NONE_REQUIRED) {
			ArrayList<String> possible = new ArrayList<>();
			for (SelfControlRoll one : SelfControlRoll.values()) {
				if (one != SelfControlRoll.NONE_REQUIRED) {
					possible.add(one.getDescriptionWithCost());
				}
			}
			mCRCombo = new JComboBox<>(possible.toArray(new String[possible.size()]));
			mCRCombo.setSelectedItem(cr.getDescriptionWithCost());
			UIUtilities.setOnlySize(mCRCombo, mCRCombo.getPreferredSize());
			wrapper.add(mCRCombo);
		}

		mModifiers = mAdvantage.getModifiers().toArray(new Modifier[0]);
		Arrays.sort(mModifiers);

		mEnabled = new JCheckBox[mModifiers.length];
		for (int i = 0; i < mModifiers.length; i++) {
			mEnabled[i] = new JCheckBox(mModifiers[i].getFullDescription(), mModifiers[i].isEnabled());
			wrapper.add(mEnabled[i]);
		}
		return wrapper;
	}

	private void applyChanges() {
		if (mAdvantage.getCR() != SelfControlRoll.NONE_REQUIRED) {
			Object selectedItem = mCRCombo.getSelectedItem();
			for (SelfControlRoll one : SelfControlRoll.values()) {
				if (one != SelfControlRoll.NONE_REQUIRED) {
					if (selectedItem.equals(one.getDescriptionWithCost())) {
						mAdvantage.setCR(one);
						break;
					}
				}
			}
		}
		for (int i = 0; i < mModifiers.length; i++) {
			mModifiers[i].setEnabled(mEnabled[i].isSelected());
		}
	}
}
