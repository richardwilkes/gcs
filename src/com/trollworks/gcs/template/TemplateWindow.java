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

package com.trollworks.gcs.template;

import com.trollworks.toolkit.annotation.Localize;
import com.trollworks.toolkit.utility.Localization;
import com.trollworks.toolkit.utility.PathUtils;
import com.trollworks.toolkit.utility.Preferences;
import com.trollworks.gcs.advantage.Advantage;
import com.trollworks.gcs.equipment.Equipment;
import com.trollworks.gcs.preferences.SheetPreferences;
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
import com.trollworks.toolkit.ui.layout.FlexRow;
import com.trollworks.toolkit.ui.menu.file.Saveable;
import com.trollworks.toolkit.ui.widget.AppWindow;
import com.trollworks.toolkit.ui.widget.BaseWindow;
import com.trollworks.toolkit.ui.widget.ModifiedMarker;
import com.trollworks.toolkit.ui.widget.WindowUtils;
import com.trollworks.toolkit.ui.widget.outline.Outline;
import com.trollworks.toolkit.ui.widget.outline.OutlineModel;
import com.trollworks.toolkit.ui.widget.outline.Row;
import com.trollworks.toolkit.ui.widget.outline.RowIterator;
import com.trollworks.toolkit.utility.notification.NotifierTarget;

import java.awt.EventQueue;
import java.io.File;
import java.io.IOException;
import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;

import javax.swing.JScrollPane;
import javax.swing.JToolBar;
import javax.swing.ListCellRenderer;
import javax.swing.undo.StateEdit;

/** The template window. */
public class TemplateWindow extends GCSWindow implements Saveable, SearchTarget, NotifierTarget {
	@Localize("Untitled Template")
	private static String UNTITLED;
	@Localize("Add Rows")
	private static String ADD_ROWS;
	@Localize("An error occurred while trying to save the template.")
	private static String SAVE_ERROR;

	static {
		Localization.initialize();
	}

	/** The extension for templates. */
	public static final String	EXTENSION	= ".gct";	//$NON-NLS-1$
	private TemplateSheet		mContent;
	private Template			mTemplate;
	private Search				mSearch;

	/** @return The top template sheet window, if any. */
	public static TemplateWindow getTopTemplate() {
		ArrayList<TemplateWindow> list = AppWindow.getActiveWindows(TemplateWindow.class);

		return list.isEmpty() ? null : list.get(0);
	}

	/** @return The {@link TemplateSheet}. */
	public TemplateSheet getSheet() {
		return mContent;
	}

	/**
	 * Looks for an existing template window for the specified template.
	 * 
	 * @param template The template to look for.
	 * @return The template window for the specified template, if any.
	 */
	public static TemplateWindow findTemplateWindow(Template template) {
		for (TemplateWindow window : BaseWindow.getWindows(TemplateWindow.class)) {
			if (window.getTemplate() == template) {
				return window;
			}
		}

		return null;
	}

	/**
	 * Looks for an existing template window for the specified file.
	 * 
	 * @param file The template file to look for.
	 * @return The template window for the specified file, if any.
	 */
	public static TemplateWindow findTemplateWindow(File file) {
		String fullPath = PathUtils.getFullPath(file);

		for (TemplateWindow window : BaseWindow.getWindows(TemplateWindow.class)) {
			File wFile = window.getTemplate().getFile();

			if (wFile != null) {
				if (PathUtils.getFullPath(wFile).equals(fullPath)) {
					return window;
				}
			}
		}
		return null;
	}

	/**
	 * Displays a template window for the specified template.
	 * 
	 * @param template The template to display.
	 * @return The displayed template.
	 */
	public static TemplateWindow displayTemplateWindow(Template template) {
		TemplateWindow window = findTemplateWindow(template);

		if (window == null) {
			window = new TemplateWindow(template);
		}
		window.setVisible(true);
		return window;
	}

	/**
	 * Creates a new {@link TemplateWindow}.
	 * 
	 * @param file The file to display.
	 */
	public TemplateWindow(File file) throws IOException {
		this(new Template(file));
	}

	/**
	 * Creates a template window.
	 * 
	 * @param template The template to display.
	 */
	public TemplateWindow(Template template) {
		super(null, template.getFileIcon(true), template.getFileIcon(false));
		mTemplate = template;
		mContent = new TemplateSheet(mTemplate);
		adjustWindowTitle();
		mContent.setSize(mContent.getPreferredSize());
		add(new JScrollPane(mContent));
		createToolBar();
		restoreBounds();
		getUndoManager().discardAllEdits();
		mTemplate.setUndoManager(getUndoManager());
		Preferences.getInstance().getNotifier().add(this, SheetPreferences.OPTIONAL_MODIFIER_RULES_PREF_KEY);
	}

	private void adjustWindowTitle() {
		File file = mTemplate.getFile();
		String title;

		if (file == null) {
			title = BaseWindow.getNextUntitledWindowName(TemplateWindow.class, UNTITLED, this);
		} else {
			title = PathUtils.getLeafName(file.getName(), false);
		}
		setTitle(title);
		getRootPane().putClientProperty("Window.documentFile", file); //$NON-NLS-1$
	}

	@Override
	public void dispose() {
		Preferences.getInstance().getNotifier().remove(this);
		mTemplate.resetNotifier();
		super.dispose();
	}

	/** @return The template associated with this window. */
	public Template getTemplate() {
		return mTemplate;
	}

	@Override
	public String getWindowPrefsPrefix() {
		return "TemplateWindow:" + mTemplate.getUniqueID() + "."; //$NON-NLS-1$ //$NON-NLS-2$
	}

	/**
	 * Adds rows to the display.
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
				outline = mContent.getAdvantageOutline();
				if (!map.containsKey(outline)) {
					map.put(outline, new StateEdit(outline.getModel(), ADD_ROWS));
				}
				row = new Advantage(mTemplate, (Advantage) row, true);
				addCompleteRow(outline, row, selMap);
			} else if (row instanceof Technique) {
				outline = mContent.getSkillOutline();
				if (!map.containsKey(outline)) {
					map.put(outline, new StateEdit(outline.getModel(), ADD_ROWS));
				}
				row = new Technique(mTemplate, (Technique) row, true);
				addCompleteRow(outline, row, selMap);
			} else if (row instanceof Skill) {
				outline = mContent.getSkillOutline();
				if (!map.containsKey(outline)) {
					map.put(outline, new StateEdit(outline.getModel(), ADD_ROWS));
				}
				row = new Skill(mTemplate, (Skill) row, true, true);
				addCompleteRow(outline, row, selMap);
			} else if (row instanceof Spell) {
				outline = mContent.getSpellOutline();
				if (!map.containsKey(outline)) {
					map.put(outline, new StateEdit(outline.getModel(), ADD_ROWS));
				}
				row = new Spell(mTemplate, (Spell) row, true, true);
				addCompleteRow(outline, row, selMap);
			} else if (row instanceof Equipment) {
				outline = mContent.getEquipmentOutline();
				if (!map.containsKey(outline)) {
					map.put(outline, new StateEdit(outline.getModel(), ADD_ROWS));
				}
				row = new Equipment(mTemplate, (Equipment) row, true);
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
		mTemplate.addDataModifiedListener(marker);
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
	public void jumpToSearchField() {
		mSearch.requestFocusInWindow();
	}

	@Override
	public List<Object> search(String text) {
		ArrayList<Object> list = new ArrayList<>();
		text = text.toLowerCase();
		searchOne(mContent.getAdvantageOutline(), text, list);
		searchOne(mContent.getSkillOutline(), text, list);
		searchOne(mContent.getSpellOutline(), text, list);
		searchOne(mContent.getEquipmentOutline(), text, list);
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

		mContent.getAdvantageOutline().getModel().deselect();
		mContent.getSkillOutline().getModel().deselect();
		mContent.getSpellOutline().getModel().deselect();
		mContent.getEquipmentOutline().getModel().deselect();

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
				primary = mContent.getAdvantageOutline();
				if (model != primary.getModel()) {
					primary = mContent.getSkillOutline();
					if (model != primary.getModel()) {
						primary = mContent.getSpellOutline();
						if (model != primary.getModel()) {
							primary = mContent.getEquipmentOutline();
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
	public boolean isModified() {
		return mTemplate != null && mTemplate.isModified();
	}

	@Override
	public String[] getAllowedExtensions() {
		return new String[] { EXTENSION };
	}

	@Override
	public String getPreferredSavePath() {
		return PathUtils.getFullPath(PathUtils.getParent(PathUtils.getFullPath(getBackingFile())), getTitle());
	}

	@Override
	public File getBackingFile() {
		return mTemplate.getFile();
	}

	@Override
	public File[] saveTo(File file) {
		if (mTemplate.save(file)) {
			mTemplate.setFile(file);
			adjustWindowTitle();
			return new File[] { file };
		}
		WindowUtils.showError(this, SAVE_ERROR);
		return new File[0];
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
	public void handleNotification(Object producer, String name, Object data) {
		mTemplate.notifySingle(Advantage.ID_LIST_CHANGED, null);
	}

	@Override
	public int getNotificationPriority() {
		return 0;
	}
}
