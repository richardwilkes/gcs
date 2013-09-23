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
 * Portions created by the Initial Developer are Copyright (C) 1998-2013 the
 * Initial Developer. All Rights Reserved.
 *
 * Contributor(s):
 *
 * ***** END LICENSE BLOCK ***** */

package com.trollworks.gcs.modifier;

import static com.trollworks.gcs.modifier.ModifierListEditor_LS.*;

import com.trollworks.gcs.advantage.Advantage;
import com.trollworks.gcs.common.DataFile;
import com.trollworks.gcs.common.ListFile;
import com.trollworks.gcs.library.LibraryFile;
import com.trollworks.gcs.widgets.outline.ListRow;
import com.trollworks.gcs.widgets.outline.RowEditor;
import com.trollworks.ttk.annotation.LS;
import com.trollworks.ttk.annotation.Localized;
import com.trollworks.ttk.collections.FilteredIterator;
import com.trollworks.ttk.collections.FilteredList;
import com.trollworks.ttk.image.ToolkitImage;
import com.trollworks.ttk.widgets.ActionPanel;
import com.trollworks.ttk.widgets.IconButton;
import com.trollworks.ttk.widgets.outline.Outline;
import com.trollworks.ttk.widgets.outline.OutlineModel;

import java.awt.BorderLayout;
import java.awt.Component;
import java.awt.event.ActionEvent;
import java.awt.event.ActionListener;
import java.util.ArrayList;
import java.util.List;

import javax.swing.JScrollPane;
import javax.swing.ScrollPaneConstants;

@Localized({
				@LS(key = "MODIFIERS", msg = "Modifiers"),
})
/** Editor for {@link ModifierList}s. */
public class ModifierListEditor extends ActionPanel implements ActionListener {
	private DataFile	mOwner;
	private Outline		mOutline;
	IconButton			mAddButton;
	boolean				mModified;

	/**
	 * @param advantage The {@link Advantage} to edit.
	 * @return An instance of {@link ModifierListEditor}.
	 */
	static public ModifierListEditor createEditor(Advantage advantage) {
		return new ModifierListEditor(advantage);
	}

	/**
	 * Creates a new {@link ModifierListEditor} editor.
	 * 
	 * @param owner The owning row.
	 * @param readOnlyModifiers The list of {@link Modifier}s from parents, which are not to be
	 *            modified.
	 * @param modifiers The list of {@link Modifier}s to modify.
	 */
	public ModifierListEditor(DataFile owner, List<Modifier> readOnlyModifiers, List<Modifier> modifiers) {
		super(new BorderLayout());
		mOwner = owner;
		add(createOutline(readOnlyModifiers, modifiers), BorderLayout.CENTER);
		setName(toString());
	}

	/**
	 * Creates a new {@link ModifierListEditor}.
	 * 
	 * @param advantage Associated advantage
	 */
	public ModifierListEditor(Advantage advantage) {
		this(advantage.getDataFile(), advantage.getParent() != null ? ((Advantage) advantage.getParent()).getAllModifiers() : null, advantage.getModifiers());
	}

	/** @return Whether a {@link Modifier} was modified. */
	public boolean wasModified() {
		return mModified;
	}

	@Override
	public void actionPerformed(ActionEvent event) {
		Object source = event.getSource();

		if (mAddButton == source) {
			addModifier();
		} else if (mOutline == source) {
			handleOutline(event.getActionCommand());
		}
	}

	private void handleOutline(String cmd) {
		if (Outline.CMD_OPEN_SELECTION.equals(cmd)) {
			openDetailEditor();
		}
	}

	private Component createOutline(List<Modifier> readOnlyModifiers, List<Modifier> modifiers) {
		JScrollPane scroller;
		OutlineModel model;

		mAddButton = new IconButton(ToolkitImage.getAddIcon());
		mAddButton.addActionListener(this);

		mOutline = new ModifierOutline();
		model = mOutline.getModel();
		ModifierColumnID.addColumns(mOutline, true);

		if (readOnlyModifiers != null) {
			for (Modifier modifier : readOnlyModifiers) {
				if (modifier.isEnabled()) {
					Modifier romod = modifier.cloneModifier();
					romod.setReadOnly(true);
					model.addRow(romod);
				}
			}
		}
		for (Modifier modifier : modifiers) {
			model.addRow(modifier.cloneModifier());
		}
		mOutline.addActionListener(this);

		scroller = new JScrollPane(mOutline, ScrollPaneConstants.VERTICAL_SCROLLBAR_ALWAYS, ScrollPaneConstants.HORIZONTAL_SCROLLBAR_ALWAYS);
		scroller.setColumnHeaderView(mOutline.getHeaderPanel());
		scroller.setCorner(ScrollPaneConstants.UPPER_RIGHT_CORNER, mAddButton);
		return scroller;
	}

	private void openDetailEditor() {
		ArrayList<ListRow> rows = new ArrayList<>();
		for (Modifier row : new FilteredIterator<>(mOutline.getModel().getSelectionAsList(), Modifier.class)) {
			if (!row.isReadOnly()) {
				rows.add(row);
			}
		}
		if (!rows.isEmpty()) {
			mOutline.getModel().setLocked(!mAddButton.isEnabled());
			if (RowEditor.edit(getTopLevelAncestor(), rows)) {
				mModified = true;
				for (ListRow row : rows) {
					row.update();
				}
				mOutline.updateRowHeights(rows);
				mOutline.sizeColumnsToFit();
				notifyActionListeners();
			}
		}
	}

	private void addModifier() {
		Modifier modifier = new Modifier(mOwner);
		OutlineModel model = mOutline.getModel();

		if (mOwner instanceof ListFile || mOwner instanceof LibraryFile) {
			modifier.setEnabled(false);
		}
		model.addRow(modifier);
		mOutline.sizeColumnsToFit();
		model.select(modifier, false);
		mOutline.revalidate();
		mOutline.scrollSelectionIntoView();
		mOutline.requestFocus();
		mModified = true;
		openDetailEditor();
	}

	/** @return Modifiers edited by this editor */
	public List<Modifier> getModifiers() {
		ArrayList<Modifier> modifiers = new ArrayList<>();
		for (Modifier modifier : new FilteredIterator<>(mOutline.getModel().getRows(), Modifier.class)) {
			if (!modifier.isReadOnly()) {
				modifiers.add(modifier);
			}
		}
		return modifiers;
	}

	/** @return Modifiers edited by this editor plus inherited Modifiers */
	public List<Modifier> getAllModifiers() {
		return new FilteredList<>(mOutline.getModel().getRows(), Modifier.class);
	}

	@Override
	public String toString() {
		return MODIFIERS;
	}

	class ModifierOutline extends Outline {
		ModifierOutline() {
			super(false);
			setAllowColumnDrag(false);
			setAllowColumnResize(false);
			setAllowRowDrag(false);
		}

		@Override
		public boolean canDeleteSelection() {
			OutlineModel model = getModel();
			boolean can = mAddButton.isEnabled() && model.hasSelection();
			if (can) {
				for (Modifier row : new FilteredIterator<>(model.getSelectionAsList(), Modifier.class)) {
					if (row.isReadOnly()) {
						return false;
					}
				}
			}
			return can;
		}

		@Override
		public void deleteSelection() {
			if (canDeleteSelection()) {
				getModel().removeSelection();
				sizeColumnsToFit();
				mModified = true;
				notifyActionListeners();
			}
		}
	}
}
