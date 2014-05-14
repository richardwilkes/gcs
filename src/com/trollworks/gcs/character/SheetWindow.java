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

import com.trollworks.gcs.advantage.Advantage;
import com.trollworks.gcs.equipment.Equipment;
import com.trollworks.gcs.skill.Skill;
import com.trollworks.gcs.skill.Technique;
import com.trollworks.gcs.spell.Spell;
import com.trollworks.gcs.widgets.GCSWindow;
import com.trollworks.gcs.widgets.outline.ListOutline;
import com.trollworks.gcs.widgets.outline.ListRow;
import com.trollworks.gcs.widgets.outline.RowItemRenderer;
import com.trollworks.gcs.widgets.outline.RowPostProcessor;
import com.trollworks.gcs.widgets.search.Search;
import com.trollworks.gcs.widgets.search.SearchTarget;
import com.trollworks.toolkit.annotation.Localize;
import com.trollworks.toolkit.ui.layout.FlexRow;
import com.trollworks.toolkit.ui.menu.file.Saveable;
import com.trollworks.toolkit.ui.print.PrintManager;
import com.trollworks.toolkit.ui.widget.AppWindow;
import com.trollworks.toolkit.ui.widget.BaseWindow;
import com.trollworks.toolkit.ui.widget.DataModifiedListener;
import com.trollworks.toolkit.ui.widget.ModifiedMarker;
import com.trollworks.toolkit.ui.widget.WindowUtils;
import com.trollworks.toolkit.ui.widget.outline.Outline;
import com.trollworks.toolkit.ui.widget.outline.OutlineModel;
import com.trollworks.toolkit.ui.widget.outline.Row;
import com.trollworks.toolkit.ui.widget.outline.RowIterator;
import com.trollworks.toolkit.utility.Localization;
import com.trollworks.toolkit.utility.PathUtils;

import java.awt.Color;
import java.awt.EventQueue;
import java.awt.Graphics;
import java.awt.KeyboardFocusManager;
import java.awt.print.PageFormat;
import java.awt.print.Printable;
import java.io.File;
import java.io.IOException;
import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;

import javax.swing.JScrollPane;
import javax.swing.JToolBar;
import javax.swing.ListCellRenderer;
import javax.swing.undo.StateEdit;

/** The character sheet window. */
public class SheetWindow extends GCSWindow implements Saveable, Printable, SearchTarget {
	@Localize("An error occurred while trying to save the sheet as a PNG.")
	private static String		SAVE_AS_PNG_ERROR;
	@Localize("An error occurred while trying to save the sheet as a PDF.")
	private static String		SAVE_AS_PDF_ERROR;
	@Localize("An error occurred while trying to save the sheet as HTML.")
	private static String		SAVE_AS_HTML_ERROR;
	@Localize("An error occurred while trying to save the sheet.")
	private static String		SAVE_ERROR;
	@Localize("Untitled Sheet")
	private static String		UNTITLED_SHEET;
	@Localize("Add Rows")
	private static String		ADD_ROWS;

	static {
		Localization.initialize();
	}

	/** The extension for character sheets. */
	public static final String	SHEET_EXTENSION	= "gcs";	//$NON-NLS-1$
	/** The PNG extension. */
	public static final String	PNG_EXTENSION	= "png";	//$NON-NLS-1$
	/** The PDF extension. */
	public static final String	PDF_EXTENSION	= "pdf";	//$NON-NLS-1$
	/** The HTML extension. */
	public static final String	HTML_EXTENSION	= "html";	//$NON-NLS-1$
	private CharacterSheet		mSheet;
	private GURPSCharacter		mCharacter;
	private Search				mSearch;
	private PrerequisitesThread	mPrereqThread;

	/** @return The top character sheet window, if any. */
	public static SheetWindow getTopSheet() {
		ArrayList<SheetWindow> list = AppWindow.getActiveWindows(SheetWindow.class);

		return list.isEmpty() ? null : list.get(0);
	}

	/**
	 * Looks for an existing character sheet window for the specified character.
	 *
	 * @param character The character to look for.
	 * @return The character sheet window for the specified character, if any.
	 */
	public static SheetWindow findSheetWindow(GURPSCharacter character) {
		for (SheetWindow window : BaseWindow.getWindows(SheetWindow.class)) {
			if (window.getCharacter() == character) {
				return window;
			}
		}

		return null;
	}

	/**
	 * Looks for an existing character sheet window for the specified file.
	 *
	 * @param file The character sheet file to look for.
	 * @return The character sheet window for the specified file, if any.
	 */
	public static SheetWindow findSheetWindow(File file) {
		String fullPath = PathUtils.getFullPath(file);

		for (SheetWindow window : BaseWindow.getWindows(SheetWindow.class)) {
			File wFile = window.getCharacter().getFile();

			if (wFile != null) {
				if (PathUtils.getFullPath(wFile).equals(fullPath)) {
					return window;
				}
			}
		}
		return null;
	}

	/**
	 * Displays a character sheet for the specified character.
	 *
	 * @param character The character to display.
	 * @return The window.
	 */
	public static SheetWindow displaySheetWindow(GURPSCharacter character) {
		SheetWindow window = findSheetWindow(character);

		if (window == null) {
			window = new SheetWindow(character);
		}
		window.setVisible(true);
		return window;
	}

	/**
	 * Creates character sheet window.
	 *
	 * @param file The file to display.
	 */
	public SheetWindow(File file) throws IOException {
		this(new GURPSCharacter(file));
	}

	/**
	 * Creates character sheet window.
	 *
	 * @param character The character to display.
	 */
	public SheetWindow(GURPSCharacter character) {
		super(null, character.getFileIcons());
		mCharacter = character;
		mSheet = new CharacterSheet(mCharacter);
		adjustWindowTitle();
		JScrollPane scrollPane = new JScrollPane(mSheet);
		scrollPane.getViewport().setBackground(Color.LIGHT_GRAY);
		add(scrollPane);
		createToolBar();
		mSheet.rebuild();
		scrollPane.getViewport().addChangeListener(mSheet);
		restoreBounds();
		mPrereqThread = new PrerequisitesThread(mSheet);
		mPrereqThread.start();
		PrerequisitesThread.waitForProcessingToFinish(character);
		mCharacter.setModified(false);
		getUndoManager().discardAllEdits();
		mCharacter.setUndoManager(getUndoManager());
	}

	/** Notify background threads of prereq or feature modifications. */
	public void notifyOfPrereqOrFeatureModification() {
		mPrereqThread.markForUpdate();
	}

	private void adjustWindowTitle() {
		File file = mCharacter.getFile();
		String title;

		if (file == null) {
			title = BaseWindow.getNextUntitledWindowName(SheetWindow.class, UNTITLED_SHEET, this);
		} else {
			title = PathUtils.getLeafName(file.getName(), false);
		}
		setTitle(title);
		getRootPane().putClientProperty("Window.documentFile", file); //$NON-NLS-1$
	}

	@Override
	public void dispose() {
		mSheet.dispose();
		super.dispose();
	}

	/** @return The character associated with this window. */
	public GURPSCharacter getCharacter() {
		return mCharacter;
	}

	@Override
	public void adjustToPageSetupChanges() {
		mSheet.rebuild();
	}

	@Override
	public int print(Graphics graphics, PageFormat pageFormat, int pageIndex) {
		return mSheet.print(graphics, pageFormat, pageIndex);
	}

	@Override
	public String getWindowPrefsPrefix() {
		return "SheetWindow:" + mCharacter.getUniqueID() + "."; //$NON-NLS-1$ //$NON-NLS-2$
	}

	/**
	 * Adds rows to the sheet.
	 *
	 * @param rows The rows to add.
	 */
	public void addRows(List<Row> rows) {
		HashMap<ListOutline, StateEdit> map = new HashMap<>();
		HashMap<Outline, ArrayList<Row>> selMap = new HashMap<>();
		HashMap<Outline, ArrayList<ListRow>> nameMap = new HashMap<>();
		ListOutline outline = null;

		for (Row row : rows) {
			if (row instanceof Advantage) {
				outline = mSheet.getAdvantageOutline();
				if (!map.containsKey(outline)) {
					map.put(outline, new StateEdit(outline.getModel(), ADD_ROWS));
				}
				row = new Advantage(mCharacter, (Advantage) row, true);
				addCompleteRow(outline, row, selMap);
			} else if (row instanceof Technique) {
				outline = mSheet.getSkillOutline();
				if (!map.containsKey(outline)) {
					map.put(outline, new StateEdit(outline.getModel(), ADD_ROWS));
				}
				row = new Technique(mCharacter, (Technique) row, true);
				addCompleteRow(outline, row, selMap);
			} else if (row instanceof Skill) {
				outline = mSheet.getSkillOutline();
				if (!map.containsKey(outline)) {
					map.put(outline, new StateEdit(outline.getModel(), ADD_ROWS));
				}
				row = new Skill(mCharacter, (Skill) row, true, true);
				addCompleteRow(outline, row, selMap);
			} else if (row instanceof Spell) {
				outline = mSheet.getSpellOutline();
				if (!map.containsKey(outline)) {
					map.put(outline, new StateEdit(outline.getModel(), ADD_ROWS));
				}
				row = new Spell(mCharacter, (Spell) row, true, true);
				addCompleteRow(outline, row, selMap);
			} else if (row instanceof Equipment) {
				outline = mSheet.getEquipmentOutline();
				if (!map.containsKey(outline)) {
					map.put(outline, new StateEdit(outline.getModel(), ADD_ROWS));
				}
				row = new Equipment(mCharacter, (Equipment) row, true);
				addCompleteRow(outline, row, selMap);
			} else {
				row = null;
			}
			if (row instanceof ListRow) {
				ArrayList<ListRow> process = nameMap.get(outline);
				if (process == null) {
					process = new ArrayList<>();
					nameMap.put(outline, process);
				}
				addRowsToBeProcessed(process, (ListRow) row);
			}
		}
		for (ListOutline anOutline : map.keySet()) {
			OutlineModel model = anOutline.getModel();
			model.select(selMap.get(anOutline), false);
			StateEdit edit = map.get(anOutline);
			edit.end();
			anOutline.postUndo(edit);
			anOutline.scrollSelectionIntoView();
			anOutline.requestFocus();
		}
		if (!nameMap.isEmpty()) {
			EventQueue.invokeLater(new RowPostProcessor(nameMap));
		}
	}

	private void addRowsToBeProcessed(ArrayList<ListRow> list, ListRow row) {
		int count = row.getChildCount();

		list.add(row);
		for (int i = 0; i < count; i++) {
			addRowsToBeProcessed(list, (ListRow) row.getChild(i));
		}
	}

	private void addCompleteRow(Outline outline, Row row, HashMap<Outline, ArrayList<Row>> selMap) {
		ArrayList<Row> selection = selMap.get(outline);

		addCompleteRow(outline.getModel(), row);
		outline.contentSizeMayHaveChanged();
		if (selection == null) {
			selection = new ArrayList<>();
			selMap.put(outline, selection);
		}
		selection.add(row);
	}

	private void addCompleteRow(OutlineModel outlineModel, Row row) {
		outlineModel.addRow(row);
		if (row.isOpen() && row.hasChildren()) {
			for (Row child : row.getChildren()) {
				addCompleteRow(outlineModel, child);
			}
		}
	}

	@Override
	protected void createToolBarContents(JToolBar toolbar, FlexRow row) {
		ModifiedMarker marker = new ModifiedMarker();
		mCharacter.addDataModifiedListener(marker);
		toolbar.add(marker);
		row.add(marker);
		mSearch = new Search(this);
		toolbar.add(mSearch);
		row.add(mSearch);
	}

	@Override
	public ListCellRenderer<Object> getSearchRenderer() {
		return new RowItemRenderer();
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
	public List<Object> search(String text) {
		ArrayList<Object> list = new ArrayList<>();

		text = text.toLowerCase();
		searchOne(mSheet.getAdvantageOutline(), text, list);
		searchOne(mSheet.getSkillOutline(), text, list);
		searchOne(mSheet.getSpellOutline(), text, list);
		searchOne(mSheet.getEquipmentOutline(), text, list);
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
			EventQueue.invokeLater(new ScrollToSelection(primary));
			primary.requestFocus();
		}
	}

	@Override
	protected PrintManager createPageSettings() {
		return mCharacter.getPageSettings();
	}

	/** @return The embedded sheet. */
	public CharacterSheet getSheet() {
		return mSheet;
	}

	/** Helper for scrolling a specific outline's selection into view. */
	class ScrollToSelection implements Runnable {
		private Outline	mOutline;

		/** @param outline The outline to scroll the selection into view for. */
		ScrollToSelection(Outline outline) {
			mOutline = outline;
		}

		@Override
		public void run() {
			mOutline.scrollSelectionIntoView();
		}
	}

	@Override
	public boolean isModified() {
		return mCharacter != null && mCharacter.isModified();
	}

	@Override
	public String[] getAllowedExtensions() {
		return new String[] { SHEET_EXTENSION, PDF_EXTENSION, HTML_EXTENSION, PNG_EXTENSION };
	}

	@Override
	public String getPreferredSavePath() {
		String path = mCharacter.getDescription().getName();
		if (path.length() == 0) {
			path = getTitle();
		}
		return PathUtils.getFullPath(PathUtils.getParent(PathUtils.getFullPath(getBackingFile())), path);
	}

	@Override
	public File getBackingFile() {
		return mCharacter.getFile();
	}

	@Override
	public File[] saveTo(File file) {
		ArrayList<File> result = new ArrayList<>();
		String extension = PathUtils.getExtension(file.getName());

		if (HTML_EXTENSION.equals(extension)) {
			if (mSheet.saveAsHTML(file, null, null)) {
				result.add(file);
			} else {
				WindowUtils.showError(this, SAVE_AS_HTML_ERROR);
			}
		} else if (PNG_EXTENSION.equals(extension)) {
			if (!mSheet.saveAsPNG(file, result)) {
				WindowUtils.showError(this, SAVE_AS_PNG_ERROR);
			}
		} else if (PDF_EXTENSION.equals(extension)) {
			if (mSheet.saveAsPDF(file)) {
				result.add(file);
			} else {
				WindowUtils.showError(this, SAVE_AS_PDF_ERROR);
			}
		} else {
			if (mCharacter.save(file)) {
				result.add(file);
				mCharacter.setFile(file);
				adjustWindowTitle();
			} else {
				WindowUtils.showError(this, SAVE_ERROR);
			}
		}
		return result.toArray(new File[result.size()]);
	}

	@Override
	public void addDataModifiedListener(DataModifiedListener listener) {
		mCharacter.addDataModifiedListener(listener);
	}

	@Override
	public void removeDataModifiedListener(DataModifiedListener listener) {
		mCharacter.removeDataModifiedListener(listener);
	}
}
