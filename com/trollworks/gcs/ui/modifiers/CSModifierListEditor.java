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
 * 2005-2007 the Initial Developer. All Rights Reserved.
 *
 * Contributor(s):
 *
 * ***** END LICENSE BLOCK ***** */

package com.trollworks.gcs.ui.modifiers;

import com.trollworks.gcs.model.CMDataFile;
import com.trollworks.gcs.model.CMListFile;
import com.trollworks.gcs.model.CMRow;
import com.trollworks.gcs.model.advantage.CMAdvantage;
import com.trollworks.gcs.model.modifier.CMModifier;
import com.trollworks.gcs.model.modifier.CMModifierList;
import com.trollworks.gcs.ui.common.CSImage;
import com.trollworks.gcs.ui.editor.CSRowEditor;
import com.trollworks.toolkit.collections.TKFilteredIterator;
import com.trollworks.toolkit.collections.TKFilteredList;
import com.trollworks.toolkit.utility.TKColor;
import com.trollworks.toolkit.widget.TKPanel;
import com.trollworks.toolkit.widget.border.TKLineBorder;
import com.trollworks.toolkit.widget.button.TKButton;
import com.trollworks.toolkit.widget.layout.TKCompassLayout;
import com.trollworks.toolkit.widget.layout.TKCompassPosition;
import com.trollworks.toolkit.widget.menu.TKMenuItem;
import com.trollworks.toolkit.widget.menu.TKMenuTarget;
import com.trollworks.toolkit.widget.outline.TKOutline;
import com.trollworks.toolkit.widget.outline.TKOutlineModel;
import com.trollworks.toolkit.widget.scroll.TKScrollPanel;
import com.trollworks.toolkit.window.TKWindow;

import java.awt.Dimension;
import java.awt.event.ActionEvent;
import java.awt.event.ActionListener;
import java.util.ArrayList;
import java.util.List;

/** Editor for {@link CMModifierList}s. */
public class CSModifierListEditor extends TKPanel implements ActionListener, TKMenuTarget {
	private CMDataFile	mOwner;
	private TKOutline	mOutline;
	private TKButton	mAddButton;
	private boolean		mModified;

	/**
	 * @param advantage The {@link CMAdvantage} to edit.
	 * @return An instance of {@link CSModifierListEditor}.
	 */
	static public CSModifierListEditor createEditor(CMAdvantage advantage) {
		return new CSModifierListEditor(advantage);
	}

	/**
	 * Creates a new {@link CSModifierListEditor} editor.
	 * 
	 * @param owner The owning row.
	 * @param readOnlyModifiers The list of {@link CMModifier}s from parents, which are not to be
	 *            modified.
	 * @param modifiers The list of {@link CMModifier}s to modify.
	 */
	public CSModifierListEditor(CMDataFile owner, List<CMModifier> readOnlyModifiers, List<CMModifier> modifiers) {
		super(new TKCompassLayout());
		mOwner = owner;
		add(createOutline(readOnlyModifiers, modifiers), TKCompassPosition.CENTER);
	}

	/**
	 * Creates a new {@link CSModifierListEditor}.
	 * 
	 * @param advantage Associated advantage
	 */
	public CSModifierListEditor(CMAdvantage advantage) {
		this(advantage.getDataFile(), advantage.getParent() != null ? ((CMAdvantage) advantage.getParent()).getAllModifiers() : null, advantage.getModifiers());
	}

	/** @return Whether a {@link CMModifier} was modified. */
	public boolean wasModified() {
		return mModified;
	}

	public void actionPerformed(ActionEvent event) {
		Object source = event.getSource();

		if (mAddButton == source) {
			addModifier();
		} else if (mOutline == source) {
			handleOutline(event.getActionCommand());
		}
	}

	private void handleOutline(String cmd) {
		if (TKOutline.CMD_OPEN_SELECTION.equals(cmd)) {
			openDetailEditor();
		} else if (TKOutline.CMD_DELETE_SELECTION.equals(cmd)) {
			if (canClear()) {
				clear();
			} else {
				getToolkit().beep();
			}
		}
	}

	public boolean adjustMenuItem(String command, TKMenuItem item) {
		if (TKWindow.CMD_CLEAR.equals(command)) {
			item.setEnabled(canClear());
			return true;
		}
		return mOutline.adjustMenuItem(command, item);
	}

	public void menusWereAdjusted() {
		mOutline.menusWereAdjusted();
	}

	public void menusWillBeAdjusted() {
		mOutline.menusWillBeAdjusted();
	}

	public boolean obeyCommand(String command, TKMenuItem item) {
		if (TKWindow.CMD_CLEAR.equals(command)) {
			clear();
			return true;
		}
		return mOutline.obeyCommand(command, item);
	}

	private TKPanel createOutline(List<CMModifier> readOnlyModifiers, List<CMModifier> modifiers) {
		TKScrollPanel scroller;
		TKPanel borderView;
		Dimension minSize;
		TKOutlineModel model;

		mAddButton = new TKButton(CSImage.getAddIcon());
		mAddButton.addActionListener(this);

		mOutline = new TKOutline(false);
		model = mOutline.getModel();
		CSModifierColumnID.addColumns(mOutline, true);
		mOutline.setAllowColumnDrag(false);
		mOutline.setAllowColumnResize(false);
		mOutline.setAllowRowDrag(false);
		mOutline.setMenuTargetDelegate(this);

		if (readOnlyModifiers != null) {
			for (CMModifier modifier : readOnlyModifiers) {
				if (modifier.isEnabled()) {
					CMModifier romod = modifier.cloneModifier();
					romod.setReadOnly(true);
					model.addRow(romod);
				}
			}
		}
		for (CMModifier modifier : modifiers) {
			model.addRow(modifier.cloneModifier());
		}
		mOutline.addActionListener(this);

		scroller = new TKScrollPanel(mOutline);
		borderView = scroller.getContentBorderView();
		scroller.setHorizontalHeader(mOutline.getHeaderPanel());
		scroller.setTopRightCorner(mAddButton);
		borderView.setBorder(new TKLineBorder(TKColor.SCROLL_BAR_LINE, 1, TKLineBorder.TOP_EDGE, false));
		scroller.setBorder(new TKLineBorder(TKColor.SCROLL_BAR_LINE, 1, TKLineBorder.LEFT_EDGE | TKLineBorder.TOP_EDGE | TKLineBorder.RIGHT_EDGE, false));
		minSize = borderView.getMinimumSize();
		if (minSize.height < 48) {
			minSize.height = 48;
		}
		borderView.setMinimumSize(minSize);
		return scroller;
	}

	private void openDetailEditor() {
		ArrayList<CMRow> rows = new ArrayList<CMRow>();
		for (CMModifier row : new TKFilteredIterator<CMModifier>(mOutline.getModel().getSelectionAsList(), CMModifier.class)) {
			if (!row.isReadOnly()) {
				rows.add(row);
			}
		}
		if (!rows.isEmpty()) {
			mOutline.getModel().setLocked(!mAddButton.isEnabled());
			if (CSRowEditor.edit(getBaseWindowAsWindow(), rows)) {
				mModified = true;
				for (CMRow row : rows) {
					row.update();
				}
				mOutline.updateRowHeights(rows);
				mOutline.sizeColumnsToFit();
				notifyActionListeners();
			}
		}
	}

	private void addModifier() {
		CMModifier modifier = new CMModifier(mOwner);
		TKOutlineModel model = mOutline.getModel();

		if (mOwner instanceof CMListFile) {
			modifier.setEnabled(false);
		}
		model.addRow(modifier);
		mOutline.sizeColumnsToFit();
		model.select(modifier, false);
		mOutline.revalidateImmediately();
		mOutline.scrollSelectionIntoView();
		mOutline.requestFocus();
		mModified = true;
		openDetailEditor();
	}

	private boolean canClear() {
		boolean can = mAddButton.isEnabled() && mOutline.getModel().hasSelection();
		if (can) {
			for (CMModifier row : new TKFilteredIterator<CMModifier>(mOutline.getModel().getSelectionAsList(), CMModifier.class)) {
				if (row.isReadOnly()) {
					return false;
				}
			}
		}
		return can;
	}

	private void clear() {
		if (canClear()) {
			mOutline.getModel().removeSelection();
			mOutline.sizeColumnsToFit();
			mModified = true;
			notifyActionListeners();
		}
	}

	/** @return Modifiers edited by this editor */
	public List<CMModifier> getModifiers() {
		ArrayList<CMModifier> modifiers = new ArrayList<CMModifier>();
		for (CMModifier modifier : new TKFilteredIterator<CMModifier>(mOutline.getModel().getRows(), CMModifier.class)) {
			if (!modifier.isReadOnly()) {
				modifiers.add(modifier);
			}
		}
		return modifiers;
	}

	/** @return Modifiers edited by this editor plus inherited Modifiers */
	public List<CMModifier> getAllModifiers() {
		return new TKFilteredList<CMModifier>(mOutline.getModel().getRows(), CMModifier.class);
	}

	@Override public String toString() {
		return Msgs.MODIFIERS;
	}
}
