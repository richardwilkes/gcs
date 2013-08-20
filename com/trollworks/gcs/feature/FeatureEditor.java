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
 * 2005-2013 the Initial Developer. All Rights Reserved.
 *
 * Contributor(s):
 *
 * ***** END LICENSE BLOCK ***** */

package com.trollworks.gcs.feature;

import com.trollworks.gcs.common.EditorPanel;
import com.trollworks.gcs.widgets.outline.ListRow;
import com.trollworks.ttk.image.ToolkitImage;
import com.trollworks.ttk.layout.FlexGrid;
import com.trollworks.ttk.layout.FlexRow;
import com.trollworks.ttk.text.DoubleFormatter;
import com.trollworks.ttk.text.IntegerFormatter;
import com.trollworks.ttk.utility.LocalizedMessages;
import com.trollworks.ttk.utility.UIUtilities;
import com.trollworks.ttk.widgets.CommitEnforcer;
import com.trollworks.ttk.widgets.EditorField;

import java.awt.event.ActionEvent;
import java.beans.PropertyChangeEvent;

import javax.swing.JComboBox;
import javax.swing.JComponent;
import javax.swing.JFormattedTextField.AbstractFormatter;
import javax.swing.SwingConstants;
import javax.swing.text.DefaultFormatterFactory;

/** A generic feature editor panel. */
public abstract class FeatureEditor extends EditorPanel {
	private static String			MSG_ADD_FEATURE_TOOLTIP;
	private static String			MSG_REMOVE_FEATURE_TOOLTIP;
	private static String			MSG_ATTRIBUTE_BONUS;
	private static String			MSG_COST_REDUCTION;
	private static String			MSG_DR_BONUS;
	private static String			MSG_SKILL_BONUS;
	private static String			MSG_SPELL_BONUS;
	private static String			MSG_WEAPON_BONUS;
	private static String			MSG_PER_LEVEL;
	private static String			MSG_PER_DIE;
	private static final String		ADD					= "Add";																																//$NON-NLS-1$
	private static final String		REMOVE				= "Remove";																															//$NON-NLS-1$
	private static final String		CHANGE_BASE_TYPE	= "ChangeBaseType";																													//$NON-NLS-1$
	private static final String		BLANK				= " ";																																	//$NON-NLS-1$
	private static final Class<?>[]	BASE_TYPES			= new Class<?>[] { AttributeBonus.class, DRBonus.class, SkillBonus.class, SpellBonus.class, WeaponBonus.class, CostReduction.class };
	private static Class<?>			LAST_ITEM_TYPE		= SkillBonus.class;
	private ListRow					mRow;
	private Feature					mFeature;
	private JComboBox<Object>		mBaseTypeCombo;
	private JComboBox<Object>		mLeveledAmountCombo;

	static {
		LocalizedMessages.initialize(FeatureEditor.class);
	}

	/**
	 * Creates a new {@link FeatureEditor}.
	 * 
	 * @param row The row this feature will belong to.
	 * @param feature The feature to edit.
	 * @return The newly created editor panel.
	 */
	public static FeatureEditor create(ListRow row, Feature feature) {
		if (feature instanceof AttributeBonus) {
			return new AttributeBonusEditor(row, (AttributeBonus) feature);
		}
		if (feature instanceof DRBonus) {
			return new DRBonusEditor(row, (DRBonus) feature);
		}
		if (feature instanceof SkillBonus) {
			return new SkillBonusEditor(row, (SkillBonus) feature);
		}
		if (feature instanceof SpellBonus) {
			return new SpellBonusEditor(row, (SpellBonus) feature);
		}
		if (feature instanceof WeaponBonus) {
			return new WeaponBonusEditor(row, (WeaponBonus) feature);
		}
		if (feature instanceof CostReduction) {
			return new CostReductionEditor(row, (CostReduction) feature);
		}
		return null;
	}

	/**
	 * Creates a new {@link FeatureEditor}.
	 * 
	 * @param row The row this feature will belong to.
	 * @param feature The feature to edit.
	 */
	public FeatureEditor(ListRow row, Feature feature) {
		super();
		mRow = row;
		mFeature = feature;
		rebuild();
	}

	/** Rebuilds the contents of this panel with the current feature settings. */
	protected void rebuild() {
		removeAll();
		FlexGrid grid = new FlexGrid();
		FlexRow right = new FlexRow();
		rebuildSelf(grid, right);
		if (mFeature != null) {
			right.add(addButton(ToolkitImage.getRemoveIcon(), REMOVE, MSG_REMOVE_FEATURE_TOOLTIP));
		}
		right.add(addButton(ToolkitImage.getAddIcon(), ADD, MSG_ADD_FEATURE_TOOLTIP));
		grid.add(right, 0, 1);
		grid.apply(this);
		revalidate();
		repaint();
	}

	/**
	 * Sub-classes must implement this method to add any components they want to be visible.
	 * 
	 * @param grid The general {@link FlexGrid}. Add items in column 0.
	 * @param right The right-side {@link FlexRow}, situated in grid row 0, column 1.
	 */
	protected abstract void rebuildSelf(FlexGrid grid, FlexRow right);

	/** @return The {@link JComboBox} that allows the base feature type to be changed. */
	protected JComboBox<Object> addChangeBaseTypeCombo() {
		Object[] choices = new Object[] { MSG_ATTRIBUTE_BONUS, MSG_DR_BONUS, MSG_SKILL_BONUS, MSG_SPELL_BONUS, MSG_WEAPON_BONUS, MSG_COST_REDUCTION };
		Class<?> type = mFeature.getClass();
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

	/**
	 * @param amt The current {@link LeveledAmount}.
	 * @param min The minimum value to allow.
	 * @param max The maximum value to allow.
	 * @return The {@link EditorField} that allows a {@link LeveledAmount} to be changed.
	 */
	protected EditorField addLeveledAmountField(LeveledAmount amt, int min, int max) {
		AbstractFormatter formatter;
		Object value;
		Object prototype;
		if (amt.isIntegerOnly()) {
			formatter = new IntegerFormatter(min, max, true);
			value = new Integer(amt.getIntegerAmount());
			prototype = new Integer(max);
		} else {
			formatter = new DoubleFormatter(min, max, true);
			value = new Double(amt.getAmount());
			prototype = new Double(max + 0.25);
		}
		EditorField field = new EditorField(new DefaultFormatterFactory(formatter), this, SwingConstants.LEFT, value, prototype, null);
		field.putClientProperty(LeveledAmount.class, amt);
		UIUtilities.setOnlySize(field, field.getPreferredSize());
		add(field);
		return field;
	}

	/**
	 * @param amt The current leveled amount object.
	 * @param usePerDie Whether to use the "per die" message or the "per level" message.
	 * @return The {@link JComboBox} that allows a {@link LeveledAmount} to be changed.
	 */
	protected JComboBox<Object> addLeveledAmountCombo(LeveledAmount amt, boolean usePerDie) {
		String per = usePerDie ? MSG_PER_DIE : MSG_PER_LEVEL;
		mLeveledAmountCombo = addComboBox(LeveledAmount.ATTRIBUTE_PER_LEVEL, new Object[] { BLANK, per }, amt.isPerLevel() ? per : BLANK);
		mLeveledAmountCombo.putClientProperty(LeveledAmount.class, amt);
		return mLeveledAmountCombo;
	}

	/** @return The underlying feature. */
	public Feature getFeature() {
		return mFeature;
	}

	@Override
	public void actionPerformed(ActionEvent event) {
		String command = event.getActionCommand();
		JComponent parent = (JComponent) getParent();

		if (LeveledAmount.ATTRIBUTE_PER_LEVEL.equals(command)) {
			((Bonus) mFeature).getAmount().setPerLevel(mLeveledAmountCombo.getSelectedIndex() == 1);
		} else if (CHANGE_BASE_TYPE.equals(command)) {
			LAST_ITEM_TYPE = BASE_TYPES[mBaseTypeCombo.getSelectedIndex()];
			if (!mFeature.getClass().equals(LAST_ITEM_TYPE)) {
				CommitEnforcer.forceFocusToAccept();
				try {
					parent.add(create(mRow, (Feature) LAST_ITEM_TYPE.newInstance()), UIUtilities.getIndexOf(parent, this));
				} catch (Exception exception) {
					// Shouldn't have a failure...
					exception.printStackTrace(System.err);
				}
				parent.remove(this);
				parent.revalidate();
				parent.repaint();
			}
		} else if (ADD.equals(command)) {
			try {
				parent.add(create(mRow, (Feature) LAST_ITEM_TYPE.newInstance()));
			} catch (Exception exception) {
				// Shouldn't have a failure...
				exception.printStackTrace(System.err);
			}
			if (mFeature == null) {
				parent.remove(this);
			}
			parent.revalidate();
		} else if (REMOVE.equals(command)) {
			parent.remove(this);
			if (parent.getComponentCount() == 0) {
				parent.add(new NoFeature(mRow));
			}
			parent.revalidate();
			parent.repaint();
		} else {
			super.actionPerformed(event);
		}
	}

	@Override
	public void propertyChange(PropertyChangeEvent event) {
		if ("value".equals(event.getPropertyName())) { //$NON-NLS-1$
			EditorField field = (EditorField) event.getSource();
			LeveledAmount amt = (LeveledAmount) field.getClientProperty(LeveledAmount.class);
			if (amt != null) {
				if (amt.isIntegerOnly()) {
					amt.setAmount(((Integer) field.getValue()).intValue());
				} else {
					amt.setAmount(((Double) field.getValue()).doubleValue());
				}
				notifyActionListeners();
			} else {
				super.propertyChange(event);
			}
		} else {
			super.propertyChange(event);
		}
	}
}
