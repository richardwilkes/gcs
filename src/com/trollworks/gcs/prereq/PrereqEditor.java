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

package com.trollworks.gcs.prereq;

import com.trollworks.gcs.common.EditorPanel;
import com.trollworks.gcs.widgets.outline.ListRow;
import com.trollworks.toolkit.annotation.Localize;
import com.trollworks.toolkit.ui.UIUtilities;
import com.trollworks.toolkit.ui.image.StdImage;
import com.trollworks.toolkit.ui.layout.FlexGrid;
import com.trollworks.toolkit.ui.layout.FlexRow;
import com.trollworks.toolkit.ui.widget.IconButton;
import com.trollworks.toolkit.utility.Localization;

import java.awt.event.ActionEvent;

import javax.swing.JComboBox;
import javax.swing.JComponent;

/** A generic prerequisite editor panel. */
public abstract class PrereqEditor extends EditorPanel {
	@Localize("Remove this prerequisite")
	@Localize(locale = "de", value = "Diese Bedingung entfernen")
	@Localize(locale = "ru", value = "Удалить это требование")
	private static String			REMOVE_PREREQ_TOOLTIP;
	@Localize("Remove this prerequisite list")
	@Localize(locale = "de", value = "Diese Bedingungsliste entfernen")
	@Localize(locale = "ru", value = "Удалить этот список требований")
	private static String			REMOVE_PREREQ_LIST_TOOLTIP;
	@Localize("has")
	@Localize(locale = "de", value = "hat")
	@Localize(locale = "ru", value = "имеет")
	private static String			HAS;
	@Localize("doesn't have")
	@Localize(locale = "de", value = "hat nicht")
	@Localize(locale = "ru", value = "не имеет")
	private static String			DOES_NOT_HAVE;
	@Localize("advantage")
	@Localize(locale = "de", value = "Vorteil")
	@Localize(locale = "ru", value = "преимущество")
	private static String			ADVANTAGE;
	@Localize("attribute")
	@Localize(locale = "de", value = "Attribut")
	@Localize(locale = "ru", value = "атрибут")
	private static String			ATTRIBUTE;
	@Localize("contained weight")
	@Localize(locale = "de", value = "Zuladung")
	@Localize(locale = "ru", value = "содержит вес")
	private static String			CONTAINED_WEIGHT;
	@Localize("skill")
	@Localize(locale = "de", value = "Fertigkeit")
	@Localize(locale = "ru", value = "умение")
	private static String			SKILL;
	@Localize("spell(s)")
	@Localize(locale = "de", value = "Zauber")
	@Localize(locale = "ru", value = "заклинание(я)")
	private static String			SPELL;

	static {
		Localization.initialize();
	}

	private static final String		CHANGE_BASE_TYPE	= "ChangeBaseType";																													//$NON-NLS-1$
	private static final String		CHANGE_HAS			= "ChangeHas";																															//$NON-NLS-1$
	private static final Class<?>[]	BASE_TYPES			= new Class<?>[] { AttributePrereq.class, AdvantagePrereq.class, SkillPrereq.class, SpellPrereq.class, ContainedWeightPrereq.class };
	/** The prerequisite this panel represents. */
	protected Prereq				mPrereq;
	/** The row this prerequisite will be attached to. */
	protected ListRow				mRow;
	private int						mDepth;
	private JComboBox<Object>		mBaseTypeCombo;

	/**
	 * Creates a new prerequisite editor panel.
	 *
	 * @param row The owning row.
	 * @param prereq The prerequisite to edit.
	 * @param depth The depth of this prerequisite.
	 * @return The newly created editor panel.
	 */
	public static PrereqEditor create(ListRow row, Prereq prereq, int depth) {
		if (prereq instanceof PrereqList) {
			return new ListPrereqEditor(row, (PrereqList) prereq, depth);
		}
		if (prereq instanceof AdvantagePrereq) {
			return new AdvantagePrereqEditor(row, (AdvantagePrereq) prereq, depth);
		}
		if (prereq instanceof SkillPrereq) {
			return new SkillPrereqEditor(row, (SkillPrereq) prereq, depth);
		}
		if (prereq instanceof SpellPrereq) {
			return new SpellPrereqEditor(row, (SpellPrereq) prereq, depth);
		}
		if (prereq instanceof AttributePrereq) {
			return new AttributePrereqEditor(row, (AttributePrereq) prereq, depth);
		}
		if (prereq instanceof ContainedWeightPrereq) {
			return new ContainedWeightPrereqEditor(row, (ContainedWeightPrereq) prereq, depth);
		}
		return null;
	}

	/**
	 * Creates a new generic prerequisite editor panel.
	 *
	 * @param row The owning row.
	 * @param prereq The prerequisite to edit.
	 * @param depth The depth of this prerequisite.
	 */
	protected PrereqEditor(ListRow row, Prereq prereq, int depth) {
		super(20 * depth);
		mRow = row;
		mPrereq = prereq;
		mDepth = depth;
		rebuild();
	}

	/** Rebuilds the contents of this panel with the current prerequisite settings. */
	protected final void rebuild() {
		removeAll();
		FlexGrid grid = new FlexGrid();
		FlexRow left = new FlexRow();
		FlexRow right = new FlexRow();
		if (mPrereq.getParent() != null) {
			AndOrLabel andOrLabel = new AndOrLabel(mPrereq);
			add(andOrLabel);
			left.add(andOrLabel);
		}
		grid.add(left, 0, 0);
		rebuildSelf(left, grid, right);
		if (mDepth > 0) {
			IconButton button = new IconButton(StdImage.REMOVE, mPrereq instanceof PrereqList ? REMOVE_PREREQ_LIST_TOOLTIP : REMOVE_PREREQ_TOOLTIP, () -> remove());
			add(button);
			right.add(button);
		}
		grid.add(right, 0, 2);
		grid.apply(this);
		revalidate();
		repaint();
	}

	/**
	 * Sub-classes must implement this method to add any components they want to be visible.
	 *
	 * @param left The left-side {@link FlexRow}, situated in grid row 0, column 0.
	 * @param grid The general {@link FlexGrid}. Add items in column 1.
	 * @param right The right-side {@link FlexRow}, situated in grid row 0, column 2.
	 */
	protected abstract void rebuildSelf(FlexRow left, FlexGrid grid, FlexRow right);

	/**
	 * @param has The current value of the "has" attribute.
	 * @return The {@link JComboBox} that allows the "has" attribute to be changed.
	 */
	protected JComboBox<Object> addHasCombo(boolean has) {
		return addComboBox(CHANGE_HAS, new Object[] { HAS, DOES_NOT_HAVE }, has ? HAS : DOES_NOT_HAVE);
	}

	/** @return The {@link JComboBox} that allows the base prereq type to be changed. */
	protected JComboBox<Object> addChangeBaseTypeCombo() {
		Object[] choices = new Object[] { ATTRIBUTE, ADVANTAGE, SKILL, SPELL, CONTAINED_WEIGHT };
		Class<?> type = mPrereq.getClass();
		Object current = choices[0];
		for (int i = 0; i < BASE_TYPES.length; i++) {
			if (type.equals(BASE_TYPES[i])) {
				current = choices[i];
				break;
			}
		}
		mBaseTypeCombo = addComboBox(CHANGE_BASE_TYPE, choices, current);
		return mBaseTypeCombo;
	}

	/** @return The depth of this prerequisite. */
	public int getDepth() {
		return mDepth;
	}

	/** @return The underlying prerequisite. */
	public Prereq getPrereq() {
		return mPrereq;
	}

	private void remove() {
		JComponent parent = (JComponent) getParent();
		int index = UIUtilities.getIndexOf(parent, this);
		int count = countSelfAndDescendents(mPrereq);
		for (int i = 0; i < count; i++) {
			parent.remove(index);
		}
		mPrereq.removeFromParent();
		parent.revalidate();
		parent.repaint();
	}

	@Override
	public void actionPerformed(ActionEvent event) {
		String command = event.getActionCommand();
		if (CHANGE_BASE_TYPE.equals(command)) {
			Class<?> type = BASE_TYPES[mBaseTypeCombo.getSelectedIndex()];
			if (!mPrereq.getClass().equals(type)) {
				JComponent parent = (JComponent) getParent();
				PrereqList list = mPrereq.getParent();
				int lIndex = list.getIndexOf(mPrereq);

				try {
					Prereq prereq = (Prereq) type.getConstructor(PrereqList.class).newInstance(list);
					if (prereq instanceof HasPrereq && mPrereq instanceof HasPrereq) {
						((HasPrereq) prereq).has(((HasPrereq) mPrereq).has());
					}
					list.add(lIndex, prereq);
					list.remove(mPrereq);
					parent.add(create(mRow, prereq, mDepth), UIUtilities.getIndexOf(parent, this));
				} catch (Exception exception) {
					// Shouldn't have a failure...
					exception.printStackTrace(System.err);
				}
				parent.remove(this);
				parent.revalidate();
				parent.repaint();
				ListPrereqEditor.setLastItemType(type);
			}
		} else if (CHANGE_HAS.equals(command)) {
			((HasPrereq) mPrereq).has(((JComboBox<?>) event.getSource()).getSelectedIndex() == 0);
		} else {
			super.actionPerformed(event);
		}
	}

	private int countSelfAndDescendents(Prereq prereq) {
		int count = 1;

		if (prereq instanceof PrereqList) {
			for (Prereq one : ((PrereqList) prereq).getChildren()) {
				count += countSelfAndDescendents(one);
			}
		}
		return count;
	}
}
