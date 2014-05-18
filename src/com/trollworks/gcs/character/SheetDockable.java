/*
 * Copyright (c) 1998-2014 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public License,
 * version 2.0. If a copy of the MPL was not distributed with this file, You
 * can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as defined
 * by the Mozilla Public License, version 2.0.
 */

package com.trollworks.gcs.character;

import com.trollworks.gcs.app.CommonDockable;
import com.trollworks.gcs.widgets.outline.ListOutline;
import com.trollworks.gcs.widgets.outline.ListRow;
import com.trollworks.gcs.widgets.outline.RowItemRenderer;
import com.trollworks.gcs.widgets.search.Search;
import com.trollworks.gcs.widgets.search.SearchTarget;
import com.trollworks.toolkit.annotation.Localize;
import com.trollworks.toolkit.ui.menu.file.ExportToCommand;
import com.trollworks.toolkit.ui.menu.file.PrintProxy;
import com.trollworks.toolkit.ui.widget.Toolbar;
import com.trollworks.toolkit.ui.widget.WindowUtils;
import com.trollworks.toolkit.ui.widget.outline.Outline;
import com.trollworks.toolkit.ui.widget.outline.OutlineModel;
import com.trollworks.toolkit.ui.widget.outline.Row;
import com.trollworks.toolkit.ui.widget.outline.RowIterator;
import com.trollworks.toolkit.utility.Localization;
import com.trollworks.toolkit.utility.PathUtils;

import java.awt.BorderLayout;
import java.awt.Color;
import java.awt.EventQueue;
import java.awt.KeyboardFocusManager;
import java.io.File;
import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;

import javax.swing.JScrollPane;
import javax.swing.ListCellRenderer;

/** A list of advantages and disadvantages from a library. */
public class SheetDockable extends CommonDockable implements SearchTarget {
	@Localize("Untitled Sheet")
	private static String		UNTITLED;
	@Localize("An error occurred while trying to save the sheet as a PNG.")
	private static String		SAVE_AS_PNG_ERROR;
	@Localize("An error occurred while trying to save the sheet as a PDF.")
	private static String		SAVE_AS_PDF_ERROR;
	@Localize("An error occurred while trying to save the sheet as HTML.")
	private static String		SAVE_AS_HTML_ERROR;

	static {
		Localization.initialize();
	}

	private CharacterSheet		mSheet;
	private Toolbar				mToolbar;
	private Search				mSearch;
	private PrerequisitesThread	mPrereqThread;

	/** Creates a new {@link SheetDockable}. */
	public SheetDockable(GURPSCharacter character) {
		super(character);
		GURPSCharacter dataFile = getDataFile();
		mToolbar = new Toolbar();
		mSearch = new Search(this);
		mToolbar.add(mSearch, Toolbar.LAYOUT_FILL);
		add(mToolbar, BorderLayout.NORTH);
		mSheet = new CharacterSheet(dataFile);
		JScrollPane scroller = new JScrollPane(mSheet);
		scroller.setBorder(null);
		scroller.getViewport().setBackground(Color.LIGHT_GRAY);
		scroller.getViewport().addChangeListener(mSheet);
		add(scroller, BorderLayout.CENTER);
		mSheet.rebuild();
		mPrereqThread = new PrerequisitesThread(mSheet);
		mPrereqThread.start();
		PrerequisitesThread.waitForProcessingToFinish(dataFile);
		getUndoManager().discardAllEdits();
		dataFile.setModified(false);
	}

	@Override
	public GURPSCharacter getDataFile() {
		return (GURPSCharacter) super.getDataFile();
	}

	/** @return The {@link CharacterSheet}. */
	public CharacterSheet getSheet() {
		return mSheet;
	}

	@Override
	protected String getUntitledBaseName() {
		return UNTITLED;
	}

	@Override
	public PrintProxy getPrintProxy() {
		return mSheet;
	}

	@Override
	public String getDescriptor() {
		// RAW: Implement
		return null;
	}

	@Override
	public String[] getAllowedExtensions() {
		return new String[] { GURPSCharacter.EXTENSION, ExportToCommand.PDF_EXTENSION, ExportToCommand.HTML_EXTENSION, ExportToCommand.PNG_EXTENSION };
	}

	@Override
	public String getPreferredSavePath() {
		String name = getDataFile().getDescription().getName();
		if (name.length() == 0) {
			name = getTitle();
		}
		return PathUtils.getFullPath(PathUtils.getParent(PathUtils.getFullPath(getBackingFile())), name);
	}

	@Override
	public File[] saveTo(File file) {
		ArrayList<File> result = new ArrayList<>();
		String extension = PathUtils.getExtension(file.getName());
		if (ExportToCommand.HTML_EXTENSION.equals(extension)) {
			if (mSheet.saveAsHTML(file, null, null)) {
				result.add(file);
			} else {
				WindowUtils.showError(this, SAVE_AS_HTML_ERROR);
			}
		} else if (ExportToCommand.PNG_EXTENSION.equals(extension)) {
			if (!mSheet.saveAsPNG(file, result)) {
				WindowUtils.showError(this, SAVE_AS_PNG_ERROR);
			}
		} else if (ExportToCommand.PDF_EXTENSION.equals(extension)) {
			if (mSheet.saveAsPDF(file)) {
				result.add(file);
			} else {
				WindowUtils.showError(this, SAVE_AS_PDF_ERROR);
			}
		} else {
			return super.saveTo(file);
		}
		return result.toArray(new File[result.size()]);
	}

	@Override
	public boolean isJumpToSearchAvailable() {
		return mSearch.isEnabled() && mSearch != KeyboardFocusManager.getCurrentKeyboardFocusManager().getPermanentFocusOwner();
	}

	@Override
	public void jumpToSearchField() {
		mSearch.requestFocusInWindow();
	}

	@Override
	public ListCellRenderer<Object> getSearchRenderer() {
		return new RowItemRenderer();
	}

	@Override
	public List<Object> search(String filter) {
		ArrayList<Object> list = new ArrayList<>();
		filter = filter.toLowerCase();
		searchOne(mSheet.getAdvantageOutline(), filter, list);
		searchOne(mSheet.getSkillOutline(), filter, list);
		searchOne(mSheet.getSpellOutline(), filter, list);
		searchOne(mSheet.getEquipmentOutline(), filter, list);
		return list;
	}

	private static void searchOne(ListOutline outline, String text, ArrayList<Object> list) {
		for (ListRow row : new RowIterator<ListRow>(outline.getModel())) {
			if (row.contains(text, true)) {
				list.add(row);
			}
		}
	}

	@Override
	public void searchSelect(List<Object> selection) {
		HashMap<OutlineModel, ArrayList<Row>> map = new HashMap<>();
		Outline primary = null;
		ArrayList<Row> list;

		mSheet.getAdvantageOutline().getModel().deselect();
		mSheet.getSkillOutline().getModel().deselect();
		mSheet.getSpellOutline().getModel().deselect();
		mSheet.getEquipmentOutline().getModel().deselect();

		for (Object obj : selection) {
			Row row = (Row) obj;
			Row parent = row.getParent();
			OutlineModel model = row.getOwner();

			while (parent != null) {
				parent.setOpen(true);
				model = parent.getOwner();
				parent = parent.getParent();
			}
			list = map.get(model);
			if (list == null) {
				list = new ArrayList<>();
				list.add(row);
				map.put(model, list);
			} else {
				list.add(row);
			}
			if (primary == null) {
				primary = mSheet.getAdvantageOutline();
				if (model != primary.getModel()) {
					primary = mSheet.getSkillOutline();
					if (model != primary.getModel()) {
						primary = mSheet.getSpellOutline();
						if (model != primary.getModel()) {
							primary = mSheet.getEquipmentOutline();
						}
					}
				}
			}
		}

		for (OutlineModel model : map.keySet()) {
			model.select(map.get(model), false);
		}

		if (primary != null) {
			final Outline outline = primary;
			EventQueue.invokeLater(() -> outline.scrollSelectionIntoView());
			primary.requestFocus();
		}
	}
}
