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

package com.trollworks.gcs.character;

import com.trollworks.gcs.advantage.Advantage;
import com.trollworks.gcs.common.CommonDockable;
import com.trollworks.gcs.equipment.Equipment;
import com.trollworks.gcs.skill.Skill;
import com.trollworks.gcs.skill.Technique;
import com.trollworks.gcs.spell.Spell;
import com.trollworks.gcs.widgets.outline.ListOutline;
import com.trollworks.gcs.widgets.outline.ListRow;
import com.trollworks.gcs.widgets.outline.RowItemRenderer;
import com.trollworks.gcs.widgets.outline.RowPostProcessor;
import com.trollworks.gcs.widgets.search.Search;
import com.trollworks.gcs.widgets.search.SearchTarget;
import com.trollworks.toolkit.annotation.Localize;
import com.trollworks.toolkit.ui.UIUtilities;
import com.trollworks.toolkit.ui.menu.RetargetableFocus;
import com.trollworks.toolkit.ui.menu.file.ExportToCommand;
import com.trollworks.toolkit.ui.widget.Toolbar;
import com.trollworks.toolkit.ui.widget.WindowUtils;
import com.trollworks.toolkit.ui.widget.dock.Dock;
import com.trollworks.toolkit.ui.widget.outline.Outline;
import com.trollworks.toolkit.ui.widget.outline.OutlineModel;
import com.trollworks.toolkit.ui.widget.outline.Row;
import com.trollworks.toolkit.ui.widget.outline.RowIterator;
import com.trollworks.toolkit.utility.Localization;
import com.trollworks.toolkit.utility.PathUtils;
import com.trollworks.toolkit.utility.PrintProxy;
import com.trollworks.toolkit.utility.undo.StdUndoManager;

import java.awt.BorderLayout;
import java.awt.Color;
import java.awt.Component;
import java.awt.EventQueue;
import java.awt.KeyboardFocusManager;
import java.io.File;
import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;

import javax.swing.JScrollPane;
import javax.swing.ListCellRenderer;
import javax.swing.undo.StateEdit;

/** A list of advantages and disadvantages from a library. */
public class SheetDockable extends CommonDockable implements SearchTarget, RetargetableFocus {
	@Localize("Untitled Sheet")
	@Localize(locale = "de", value = "Unbenanntes Charakterblatt")
	@Localize(locale = "ru", value = "Лист без названия")
	private static String			UNTITLED;
	@Localize("An error occurred while trying to save the sheet as a PNG.")
	@Localize(locale = "de", value = "Ein Fehler ist beim Exportieren des Charakterblatts als PNG aufgetreten.")
	@Localize(locale = "ru", value = "Произошла ошибка при попытке сохранить лист как PNG.")
	private static String			SAVE_AS_PNG_ERROR;
	@Localize("An error occurred while trying to save the sheet as a PDF.")
	@Localize(locale = "de", value = "Ein Fehler ist beim Exportieren des Charakterblatts als PDF aufgetreten.")
	@Localize(locale = "ru", value = "Произошла ошибка при попытке сохранить лист в формате PDF.")
	private static String			SAVE_AS_PDF_ERROR;
	@Localize("An error occurred while trying to save the sheet as HTML.")
	@Localize(locale = "de", value = "Ein Fehler ist beim Exportieren des Charakterblatts als HTML aufgetreten.")
	@Localize(locale = "ru", value = "Произошла ошибка при попытке сохранить лист в формате HTML.")
	private static String			SAVE_AS_HTML_ERROR;
	@Localize("Add Rows")
	@Localize(locale = "de", value = "Zeilen hinzufügen")
	@Localize(locale = "ru", value = "Добавить строки")
	private static String			ADD_ROWS;

	static {
		Localization.initialize();
	}

	private static SheetDockable	LAST_ACTIVATED;
	private CharacterSheet			mSheet;
	private Toolbar					mToolbar;
	private Search					mSearch;
	private PrerequisitesThread		mPrereqThread;

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
		dataFile.setModified(false);
		StdUndoManager undoManager = getUndoManager();
		undoManager.discardAllEdits();
		dataFile.setUndoManager(undoManager);
	}

	@Override
	public Component getRetargetedFocus() {
		return mSheet;
	}

	/** @return The last activated {@link SheetDockable}. */
	public static SheetDockable getLastActivated() {
		if (LAST_ACTIVATED != null) {
			Dock dock = UIUtilities.getAncestorOfType(LAST_ACTIVATED, Dock.class);
			if (dock == null) {
				LAST_ACTIVATED = null;
			}
		}
		return LAST_ACTIVATED;
	}

	@Override
	public void activated() {
		super.activated();
		LAST_ACTIVATED = this;
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
				row = new Advantage(getDataFile(), (Advantage) row, true);
				addCompleteRow(outline, row, selMap);
			} else if (row instanceof Technique) {
				outline = mSheet.getSkillOutline();
				if (!map.containsKey(outline)) {
					map.put(outline, new StateEdit(outline.getModel(), ADD_ROWS));
				}
				row = new Technique(getDataFile(), (Technique) row, true);
				addCompleteRow(outline, row, selMap);
			} else if (row instanceof Skill) {
				outline = mSheet.getSkillOutline();
				if (!map.containsKey(outline)) {
					map.put(outline, new StateEdit(outline.getModel(), ADD_ROWS));
				}
				row = new Skill(getDataFile(), (Skill) row, true, true);
				addCompleteRow(outline, row, selMap);
			} else if (row instanceof Spell) {
				outline = mSheet.getSpellOutline();
				if (!map.containsKey(outline)) {
					map.put(outline, new StateEdit(outline.getModel(), ADD_ROWS));
				}
				row = new Spell(getDataFile(), (Spell) row, true, true);
				addCompleteRow(outline, row, selMap);
			} else if (row instanceof Equipment) {
				outline = mSheet.getEquipmentOutline();
				if (!map.containsKey(outline)) {
					map.put(outline, new StateEdit(outline.getModel(), ADD_ROWS));
				}
				row = new Equipment(getDataFile(), (Equipment) row, true);
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

	/** Notify background threads of prereq or feature modifications. */
	public void notifyOfPrereqOrFeatureModification() {
		mPrereqThread.markForUpdate();
	}
}
