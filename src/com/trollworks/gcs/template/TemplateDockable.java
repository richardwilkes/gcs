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

package com.trollworks.gcs.template;

import com.trollworks.gcs.advantage.Advantage;
import com.trollworks.gcs.common.CommonDockable;
import com.trollworks.gcs.equipment.Equipment;
import com.trollworks.gcs.preferences.SheetPreferences;
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
import com.trollworks.toolkit.ui.widget.Toolbar;
import com.trollworks.toolkit.ui.widget.dock.Dock;
import com.trollworks.toolkit.ui.widget.outline.Outline;
import com.trollworks.toolkit.ui.widget.outline.OutlineModel;
import com.trollworks.toolkit.ui.widget.outline.Row;
import com.trollworks.toolkit.ui.widget.outline.RowIterator;
import com.trollworks.toolkit.utility.Localization;
import com.trollworks.toolkit.utility.Preferences;
import com.trollworks.toolkit.utility.PrintProxy;
import com.trollworks.toolkit.utility.notification.NotifierTarget;
import com.trollworks.toolkit.utility.undo.StdUndoManager;

import java.awt.BorderLayout;
import java.awt.Color;
import java.awt.Component;
import java.awt.EventQueue;
import java.awt.KeyboardFocusManager;
import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;

import javax.swing.JScrollPane;
import javax.swing.ListCellRenderer;
import javax.swing.undo.StateEdit;

/** A list of advantages and disadvantages from a library. */
public class TemplateDockable extends CommonDockable implements NotifierTarget, SearchTarget, RetargetableFocus {
	@Localize("Untitled Template")
	@Localize(locale = "de", value = "Unbenannte Vorlage")
	@Localize(locale = "ru", value = "Безымянный шаблон")
	private static String			UNTITLED;
	@Localize("Add Rows")
	@Localize(locale = "de", value = "Zeilen hinzufügen")
	@Localize(locale = "ru", value = "Добавить строки")
	private static String			ADD_ROWS;

	static {
		Localization.initialize();
	}

	private static TemplateDockable	LAST_ACTIVATED;
	private TemplateSheet			mTemplate;
	private Toolbar					mToolbar;
	private Search					mSearch;

	/** Creates a new {@link TemplateDockable}. */
	public TemplateDockable(Template template) {
		super(template);
		Template dataFile = getDataFile();
		mToolbar = new Toolbar();
		mSearch = new Search(this);
		mToolbar.add(mSearch, Toolbar.LAYOUT_FILL);
		add(mToolbar, BorderLayout.NORTH);
		mTemplate = new TemplateSheet(dataFile);
		JScrollPane scroller = new JScrollPane(mTemplate);
		scroller.setBorder(null);
		scroller.getViewport().setBackground(Color.LIGHT_GRAY);
		add(scroller, BorderLayout.CENTER);
		dataFile.setModified(false);
		StdUndoManager undoManager = getUndoManager();
		undoManager.discardAllEdits();
		dataFile.setUndoManager(undoManager);
		Preferences.getInstance().getNotifier().add(this, SheetPreferences.OPTIONAL_MODIFIER_RULES_PREF_KEY);
	}

	@Override
	public Component getRetargetedFocus() {
		return mTemplate;
	}

	/** @return The last activated {@link TemplateDockable}. */
	public static TemplateDockable getLastActivated() {
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
	public Template getDataFile() {
		return (Template) super.getDataFile();
	}

	/** @return The {@link TemplateSheet}. */
	public TemplateSheet getTemplate() {
		return mTemplate;
	}

	@Override
	public PrintProxy getPrintProxy() {
		return null;
	}

	@Override
	public String getDescriptor() {
		// RAW: Implement
		return null;
	}

	@Override
	protected String getUntitledBaseName() {
		return UNTITLED;
	}

	@Override
	public String[] getAllowedExtensions() {
		return new String[] { Template.EXTENSION };
	}

	@Override
	public int getNotificationPriority() {
		return 0;
	}

	@Override
	public void handleNotification(Object producer, String name, Object data) {
		getDataFile().notifySingle(Advantage.ID_LIST_CHANGED, null);
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
		searchOne(mTemplate.getAdvantageOutline(), filter, list);
		searchOne(mTemplate.getSkillOutline(), filter, list);
		searchOne(mTemplate.getSpellOutline(), filter, list);
		searchOne(mTemplate.getEquipmentOutline(), filter, list);
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

		mTemplate.getAdvantageOutline().getModel().deselect();
		mTemplate.getSkillOutline().getModel().deselect();
		mTemplate.getSpellOutline().getModel().deselect();
		mTemplate.getEquipmentOutline().getModel().deselect();

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
				primary = mTemplate.getAdvantageOutline();
				if (model != primary.getModel()) {
					primary = mTemplate.getSkillOutline();
					if (model != primary.getModel()) {
						primary = mTemplate.getSpellOutline();
						if (model != primary.getModel()) {
							primary = mTemplate.getEquipmentOutline();
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
				outline = mTemplate.getAdvantageOutline();
				if (!map.containsKey(outline)) {
					map.put(outline, new StateEdit(outline.getModel(), ADD_ROWS));
				}
				row = new Advantage(getDataFile(), (Advantage) row, true);
				addCompleteRow(outline, row, selMap);
			} else if (row instanceof Technique) {
				outline = mTemplate.getSkillOutline();
				if (!map.containsKey(outline)) {
					map.put(outline, new StateEdit(outline.getModel(), ADD_ROWS));
				}
				row = new Technique(getDataFile(), (Technique) row, true);
				addCompleteRow(outline, row, selMap);
			} else if (row instanceof Skill) {
				outline = mTemplate.getSkillOutline();
				if (!map.containsKey(outline)) {
					map.put(outline, new StateEdit(outline.getModel(), ADD_ROWS));
				}
				row = new Skill(getDataFile(), (Skill) row, true, true);
				addCompleteRow(outline, row, selMap);
			} else if (row instanceof Spell) {
				outline = mTemplate.getSpellOutline();
				if (!map.containsKey(outline)) {
					map.put(outline, new StateEdit(outline.getModel(), ADD_ROWS));
				}
				row = new Spell(getDataFile(), (Spell) row, true, true);
				addCompleteRow(outline, row, selMap);
			} else if (row instanceof Equipment) {
				outline = mTemplate.getEquipmentOutline();
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
}
