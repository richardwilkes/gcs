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

package com.trollworks.gcs.skill;

import com.trollworks.gcs.common.EditorPanel;
import com.trollworks.ttk.image.ToolkitImage;
import com.trollworks.ttk.layout.Alignment;
import com.trollworks.ttk.layout.FlexGrid;
import com.trollworks.ttk.layout.FlexRow;
import com.trollworks.ttk.layout.FlexSpacer;
import com.trollworks.ttk.text.IntegerFormatter;
import com.trollworks.ttk.utility.LocalizedMessages;
import com.trollworks.ttk.utility.UIUtilities;
import com.trollworks.ttk.widgets.CommitEnforcer;
import com.trollworks.ttk.widgets.EditorField;

import java.awt.Insets;
import java.awt.event.ActionEvent;
import java.beans.PropertyChangeEvent;

import javax.swing.JComboBox;
import javax.swing.JComponent;
import javax.swing.SwingConstants;
import javax.swing.text.DefaultFormatter;
import javax.swing.text.DefaultFormatterFactory;

/** A skill default editor panel. */
public class SkillDefaultEditor extends EditorPanel {
	private static String			MSG_ADD_DEFAULT;
	private static String			MSG_REMOVE_DEFAULT;
	private static String			MSG_SPECIALIZATION_TOOLTIP;
	static String					MSG_OR;
	private static final String		ADD				= "Add";				//$NON-NLS-1$
	private static final String		REMOVE			= "Remove";			//$NON-NLS-1$
	private static SkillDefaultType	LAST_ITEM_TYPE	= SkillDefaultType.DX;
	private SkillDefault			mDefault;
	private JComboBox				mTypeCombo;
	private EditorField				mSkillNameField;
	private EditorField				mSpecializationField;
	private EditorField				mModifierField;

	static {
		LocalizedMessages.initialize(SkillDefaultEditor.class);
	}

	/** @param type The last item type created or switched to. */
	public static void setLastItemType(SkillDefaultType type) {
		LAST_ITEM_TYPE = type;
	}

	/** Creates a new placeholder {@link SkillDefaultEditor}. */
	public SkillDefaultEditor() {
		this(null);
	}

	/**
	 * Creates a new skill default editor panel.
	 * 
	 * @param skillDefault The skill default to edit.
	 */
	public SkillDefaultEditor(SkillDefault skillDefault) {
		super();
		mDefault = skillDefault;
		rebuild();
	}

	/** Rebuilds the contents of this panel with the current feature settings. */
	protected void rebuild() {
		removeAll();

		if (mDefault != null) {
			FlexGrid grid = new FlexGrid();

			OrLabel or = new OrLabel(this);
			add(or);
			grid.add(or, 0, 0);

			FlexRow row = new FlexRow();
			SkillDefaultType current = mDefault.getType();
			mTypeCombo = new JComboBox(SkillDefaultType.values());
			mTypeCombo.setOpaque(false);
			mTypeCombo.setSelectedItem(current);
			mTypeCombo.setActionCommand(SkillDefault.TAG_TYPE);
			mTypeCombo.addActionListener(this);
			UIUtilities.setOnlySize(mTypeCombo, mTypeCombo.getPreferredSize());
			add(mTypeCombo);
			row.add(mTypeCombo);
			grid.add(row, 0, 1);

			mModifierField = new EditorField(new DefaultFormatterFactory(new IntegerFormatter(-99, 99, true)), this, SwingConstants.LEFT, new Integer(mDefault.getModifier()), new Integer(99), null);
			UIUtilities.setOnlySize(mModifierField, mModifierField.getPreferredSize());
			add(mModifierField);

			if (current.isSkillBased()) {
				row.add(new FlexSpacer(0, 0, true, false));
				row = new FlexRow();
				row.setInsets(new Insets(0, 20, 0, 0));
				DefaultFormatter formatter = new DefaultFormatter();
				formatter.setOverwriteMode(false);
				mSkillNameField = new EditorField(new DefaultFormatterFactory(formatter), this, SwingConstants.LEFT, mDefault.getName(), null);
				add(mSkillNameField);
				row.add(mSkillNameField);
				mSpecializationField = new EditorField(new DefaultFormatterFactory(formatter), this, SwingConstants.LEFT, mDefault.getSpecialization(), MSG_SPECIALIZATION_TOOLTIP);
				mSpecializationField.setHint(MSG_SPECIALIZATION_TOOLTIP);
				add(mSpecializationField);
				row.add(mSpecializationField);
				row.add(mModifierField);
				grid.add(row, 1, 1);
			} else {
				row.add(mModifierField);
				row.add(new FlexSpacer(0, 0, true, false));
			}

			row = new FlexRow();
			row.setHorizontalAlignment(Alignment.RIGHT_BOTTOM);
			row.add(addButton(ToolkitImage.getRemoveIcon(), REMOVE, MSG_REMOVE_DEFAULT));
			row.add(addButton(ToolkitImage.getAddIcon(), ADD, MSG_ADD_DEFAULT));
			grid.add(row, 0, 2);
			grid.apply(this);
		} else {
			FlexRow row = new FlexRow();
			row.setHorizontalAlignment(Alignment.RIGHT_BOTTOM);
			row.add(new FlexSpacer(0, 0, true, false));
			row.add(addButton(ToolkitImage.getAddIcon(), ADD, MSG_ADD_DEFAULT));
			row.apply(this);
		}

		revalidate();
		repaint();
	}

	/** @return The underlying skill default. */
	public SkillDefault getSkillDefault() {
		return mDefault;
	}

	@Override
	public void propertyChange(PropertyChangeEvent event) {
		Object src = event.getSource();
		if (src == mSkillNameField) {
			mDefault.setName((String) mSkillNameField.getValue());
			notifyActionListeners();
		} else if (src == mSpecializationField) {
			mDefault.setSpecialization((String) mSpecializationField.getValue());
			notifyActionListeners();
		} else if (src == mModifierField) {
			mDefault.setModifier(((Integer) mModifierField.getValue()).intValue());
			notifyActionListeners();
		} else {
			super.propertyChange(event);
		}
	}

	@Override
	public void actionPerformed(ActionEvent event) {
		Object src = event.getSource();
		String command = event.getActionCommand();
		JComponent parent = (JComponent) getParent();

		if (ADD.equals(command)) {
			SkillDefault skillDefault = new SkillDefault(LAST_ITEM_TYPE, LAST_ITEM_TYPE.isSkillBased() ? "" : null, null, 0); //$NON-NLS-1$
			parent.add(new SkillDefaultEditor(skillDefault));
			if (mDefault == null) {
				parent.remove(this);
			}
			parent.revalidate();
			parent.repaint();
			notifyActionListeners();
		} else if (REMOVE.equals(command)) {
			parent.remove(this);
			if (parent.getComponentCount() == 0) {
				parent.add(new SkillDefaultEditor());
			}
			parent.revalidate();
			parent.repaint();
			notifyActionListeners();
		} else if (SkillDefault.TAG_TYPE.equals(command)) {
			SkillDefaultType current = mDefault.getType();
			SkillDefaultType value = (SkillDefaultType) ((JComboBox) src).getSelectedItem();
			if (!current.equals(value)) {
				CommitEnforcer.forceFocusToAccept();
				mDefault.setType(value);
				rebuild();
				notifyActionListeners();
			}
			setLastItemType(value);
		} else {
			super.actionPerformed(event);
		}
	}
}
