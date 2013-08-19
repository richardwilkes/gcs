/* ***** BEGIN LICENSE BLOCK *****
 * Version: MPL 1.1
 *
 * The contents of this file are subject to the Mozilla Public License Version
 * 1.1 (the "License"); you may not use this file except in compliance with
 * the License. You may obtain a copy of the License at
 * http://www.mozilla.org/MPL/
 *
 * Software distributed under the License is distributed on an "AS IS" basis,
 * WITHOUT WARRANTY OF ANY KIND, either express or implied. See the License
 * for the specific language governing rights and limitations under the
 * License.
 *
 * The Original Code is GURPS Character Sheet.
 *
 * The Initial Developer of the Original Code is Richard A. Wilkes.
 * Portions created by the Initial Developer are Copyright (C) 1998-2002,
 * 2005-2008 the Initial Developer. All Rights Reserved.
 *
 * Contributor(s):
 *
 * ***** END LICENSE BLOCK ***** */

package com.trollworks.gcs.prereq;

import com.trollworks.gcs.common.EditorPanel;
import com.trollworks.gcs.widgets.outline.ListRow;
import com.trollworks.ttk.image.ToolkitImage;
import com.trollworks.ttk.layout.FlexGrid;
import com.trollworks.ttk.layout.FlexRow;
import com.trollworks.ttk.utility.LocalizedMessages;
import com.trollworks.ttk.utility.UIUtilities;

import java.awt.event.ActionEvent;

import javax.swing.JComboBox;
import javax.swing.JComponent;

/** A generic prerequisite editor panel. */
public abstract class PrereqEditor extends EditorPanel {
	private static String			MSG_REMOVE_PREREQ_TOOLTIP;
	private static String			MSG_REMOVE_PREREQ_LIST_TOOLTIP;
	private static String			MSG_HAS;
	private static String			MSG_DOES_NOT_HAVE;
	private static String			MSG_ADVANTAGE;
	private static String			MSG_ATTRIBUTE;
	private static String			MSG_CONTAINED_WEIGHT;
	private static String			MSG_SKILL;
	private static String			MSG_SPELL;
	private static final String		REMOVE				= "Remove";																															//$NON-NLS-1$
	private static final String		CHANGE_BASE_TYPE	= "ChangeBaseType";																													//$NON-NLS-1$
	private static final String		CHANGE_HAS			= "ChangeHas";																															//$NON-NLS-1$
	private static final Class<?>[]	BASE_TYPES			= new Class<?>[] { AttributePrereq.class, AdvantagePrereq.class, SkillPrereq.class, SpellPrereq.class, ContainedWeightPrereq.class };
	/** The prerequisite this panel represents. */
	protected Prereq				mPrereq;
	/** The row this prerequisite will be attached to. */
	protected ListRow				mRow;
	private int						mDepth;
	private JComboBox				mBaseTypeCombo;

	static {
		LocalizedMessages.initialize(PrereqEditor.class);
	}

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
			right.add(addButton(ToolkitImage.getRemoveIcon(), REMOVE, mPrereq instanceof PrereqList ? MSG_REMOVE_PREREQ_LIST_TOOLTIP : MSG_REMOVE_PREREQ_TOOLTIP));
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
	protected JComboBox addHasCombo(boolean has) {
		return addComboBox(CHANGE_HAS, new Object[] { MSG_HAS, MSG_DOES_NOT_HAVE }, has ? MSG_HAS : MSG_DOES_NOT_HAVE);
	}

	/** @return The {@link JComboBox} that allows the base prereq type to be changed. */
	protected JComboBox addChangeBaseTypeCombo() {
		Object[] choices = new Object[] { MSG_ATTRIBUTE, MSG_ADVANTAGE, MSG_SKILL, MSG_SPELL, MSG_CONTAINED_WEIGHT };
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
			((HasPrereq) mPrereq).has(((JComboBox) event.getSource()).getSelectedIndex() == 0);
		} else if (REMOVE.equals(command)) {
			JComponent parent = (JComponent) getParent();
			int index = UIUtilities.getIndexOf(parent, this);
			int count = countSelfAndDescendents(mPrereq);
			for (int i = 0; i < count; i++) {
				parent.remove(index);
			}
			mPrereq.removeFromParent();
			parent.revalidate();
			parent.repaint();
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
