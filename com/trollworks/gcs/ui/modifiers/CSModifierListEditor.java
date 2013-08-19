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
import com.trollworks.gcs.model.advantage.CMAdvantage;
import com.trollworks.gcs.model.modifier.CMModifier;
import com.trollworks.gcs.model.modifier.CMModifierList;
import com.trollworks.gcs.ui.common.CSImage;
import com.trollworks.gcs.ui.common.CSOutline;
import com.trollworks.toolkit.collections.TKFilteredIterator;
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
import com.trollworks.toolkit.widget.outline.TKRow;
import com.trollworks.toolkit.widget.scroll.TKScrollPanel;
import com.trollworks.toolkit.window.TKWindow;

import java.awt.Dimension;
import java.awt.event.ActionEvent;
import java.awt.event.ActionListener;
import java.util.ArrayList;
import java.util.List;

/** Editor for {@link CMModifierList}s. */
public class CSModifierListEditor extends TKPanel implements ActionListener, TKMenuTarget {
	private CMDataFile		mOwner;
	private CSOutline		mOutline;
	private TKButton		mAddButton;
	private TKOutlineModel	mModifierModel;

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
	 * @param modifier The list of {@link CMModifier}s to modify.
	 */
	public CSModifierListEditor(CMDataFile owner, List<CMModifier> modifier) {
		super(new TKCompassLayout());
		mOwner = owner;
		add(createOutline(modifier), TKCompassPosition.CENTER);
	}

	/**
	 * Creates a new {@link CSModifierListEditor}.
	 * 
	 * @param advantage Associated advantage
	 */
	public CSModifierListEditor(CMAdvantage advantage) {
		this(advantage.getDataFile(), advantage.getModifiers());
	}

	public void actionPerformed(ActionEvent event) {
		Object source = event.getSource();

		if (mAddButton == source) {
			addModifier();
		} else if (mOutline == source) {
			handleOutline(event.getActionCommand());
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

	private TKPanel createOutline(List<CMModifier> modifiers) {
		TKScrollPanel scroller;

		TKPanel borderView;
		Dimension minSize;

		mAddButton = new TKButton(CSImage.getAddIcon());
		mAddButton.addActionListener(this);

		mModifierModel = new TKOutlineModel();
		for (CMModifier modifier : new TKFilteredIterator<CMModifier>(modifiers, CMModifier.class)) {
			mModifierModel.addRow(modifier);
		}

		mOutline = new CSModifierOutline(mOwner, mModifierModel, CMModifier.ID_LIST_CHANGED);
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

	private void addModifier() {
		CMModifier modifier = new CMModifier(mOwner);
		TKOutlineModel model = mOutline.getModel();

		model.addRow(modifier);
		mOutline.sizeColumnsToFit();
		model.select(modifier, false);
		mOutline.revalidateImmediately();
		mOutline.scrollSelectionIntoView();
		mOutline.requestFocus();

		mOutline.openDetailEditor(false);
	}

	private void handleOutline(String cmd) {
		if (TKOutline.CMD_UPDATE_FROM_EDITOR.equals(cmd)) {
			// TODO
		}
	}

	private boolean canClear() {
		return mOutline.getModel().hasSelection();
	}

	private void clear() {
		if (canClear()) {
			mOutline.getModel().removeSelection();
			mOutline.sizeColumnsToFit();
		}
	}

	/** @return Modifiers edited by this editor */
	public List<CMModifier> getModifiers() {
		ArrayList<CMModifier> modifiers = new ArrayList<CMModifier>(mOutline.getModel().getRowCount());
		for (TKRow row : mOutline.getModel().getRows()) {
			modifiers.add(((CMModifier) row).cloneModifier());
		}
		return modifiers;
	}

	@Override public String toString() {
		return Msgs.MODIFIERS;
	}
}
