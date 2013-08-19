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

package com.trollworks.gcs.ui.template;

import com.trollworks.gcs.model.CMCharacter;
import com.trollworks.gcs.model.CMRow;
import com.trollworks.gcs.model.CMTemplate;
import com.trollworks.gcs.model.advantage.CMAdvantage;
import com.trollworks.gcs.model.equipment.CMEquipment;
import com.trollworks.gcs.model.skill.CMSkill;
import com.trollworks.gcs.model.skill.CMTechnique;
import com.trollworks.gcs.model.spell.CMSpell;
import com.trollworks.gcs.ui.common.CSFileOpener;
import com.trollworks.gcs.ui.common.CSFont;
import com.trollworks.gcs.ui.common.CSNamePostProcessor;
import com.trollworks.gcs.ui.common.CSOutline;
import com.trollworks.gcs.ui.common.CSRowItemRenderer;
import com.trollworks.gcs.ui.common.CSWindow;
import com.trollworks.gcs.ui.sheet.CSSheetWindow;
import com.trollworks.toolkit.io.TKFileFilter;
import com.trollworks.toolkit.io.TKPath;
import com.trollworks.toolkit.undo.TKMultipleUndo;
import com.trollworks.toolkit.widget.TKItemRenderer;
import com.trollworks.toolkit.widget.TKPanel;
import com.trollworks.toolkit.widget.TKSelection;
import com.trollworks.toolkit.widget.TKToolBar;
import com.trollworks.toolkit.widget.menu.TKMenuItem;
import com.trollworks.toolkit.widget.outline.TKOutline;
import com.trollworks.toolkit.widget.outline.TKOutlineModel;
import com.trollworks.toolkit.widget.outline.TKOutlineModelUndo;
import com.trollworks.toolkit.widget.outline.TKOutlineModelUndoSnapshot;
import com.trollworks.toolkit.widget.outline.TKRow;
import com.trollworks.toolkit.widget.outline.TKRowIterator;
import com.trollworks.toolkit.widget.scroll.TKScrollPanel;
import com.trollworks.toolkit.widget.search.TKSearch;
import com.trollworks.toolkit.widget.search.TKSearchTarget;
import com.trollworks.toolkit.window.TKOptionDialog;
import com.trollworks.toolkit.window.TKWindow;

import java.awt.EventQueue;
import java.awt.Point;
import java.io.File;
import java.util.ArrayList;
import java.util.Collection;
import java.util.HashMap;
import java.util.List;

/** The template window. */
public class CSTemplateWindow extends CSWindow implements TKSearchTarget {
	private CSTemplate	mContent;
	private CMTemplate	mTemplate;

	/** @return The top template sheet window, if any. */
	public static CSTemplateWindow getTopTemplate() {
		ArrayList<CSTemplateWindow> list = TKWindow.getActiveWindows(CSTemplateWindow.class);

		return list.isEmpty() ? null : list.get(0);
	}

	/**
	 * @param template The template to get the consumer group for.
	 * @return The consumer group for the specified template.
	 */
	static String getConsumerGroup(CMTemplate template) {
		return "CSTemplateWindow:" + template.getUniqueID(); //$NON-NLS-1$
	}

	/** @param template The template to apply to the top sheet. */
	public static void applyTemplateToSheet(CMTemplate template) {
		CSSheetWindow sheet = CSSheetWindow.getTopSheet();

		if (sheet != null) {
			TKMultipleUndo undo = new TKMultipleUndo(Msgs.UNDO);
			ArrayList<TKRow> rows = new ArrayList<TKRow>();
			String notes = template.getNotes().trim();
			CMCharacter character = sheet.getCharacter();

			character.addEdit(undo);
			rows.addAll(template.getAdvantagesModel().getTopLevelRows());
			rows.addAll(template.getSkillsModel().getTopLevelRows());
			rows.addAll(template.getSpellsModel().getTopLevelRows());
			rows.addAll(template.getEquipmentModel().getTopLevelRows());
			sheet.addRows(rows);
			if (notes.length() > 0) {
				String prevNotes = character.getNotes().trim();

				if (prevNotes.length() > 0) {
					notes = prevNotes + "\n\n" + notes; //$NON-NLS-1$
				}
				character.setNotes(notes);
			}
			undo.end();
		}
	}

	/**
	 * Looks for an existing template window for the specified template.
	 * 
	 * @param template The template to look for.
	 * @return The template window for the specified template, if any.
	 */
	public static CSTemplateWindow findTemplateWindow(CMTemplate template) {
		for (CSTemplateWindow window : TKWindow.getWindows(CSTemplateWindow.class)) {
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
	public static CSTemplateWindow findTemplateWindow(File file) {
		String fullPath = TKPath.getFullPath(file);

		for (CSTemplateWindow window : TKWindow.getWindows(CSTemplateWindow.class)) {
			File wFile = window.getTemplate().getFile();

			if (wFile != null) {
				if (TKPath.getFullPath(wFile).equals(fullPath)) {
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
	 */
	public static void displayTemplateWindow(CMTemplate template) {
		CSTemplateWindow window = findTemplateWindow(template);

		if (window == null) {
			window = new CSTemplateWindow(template);
		}
		window.setVisible(true);
	}

	/**
	 * Creates a template window.
	 * 
	 * @param template The template to display.
	 */
	public CSTemplateWindow(CMTemplate template) {
		super(null, template.getFileIcon(true), template.getFileIcon(false));
		mTemplate = template;
		mContent = new CSTemplate(mTemplate);
		adjustWindowTitle();
		setContent(new TKScrollPanel(mContent));
		createToolBar();
		restoreBounds();
		mTemplate.discardAllEdits();
		setUndoManager(mTemplate);
	}

	private void adjustWindowTitle() {
		File file = mTemplate.getFile();
		String title;

		if (file == null) {
			title = TKWindow.getNextUntitledWindowName(CSTemplateWindow.class, Msgs.UNTITLED, this);
		} else {
			title = TKPath.getLeafName(file.getName(), false);
		}
		setTitle(title);
	}

	@Override public void dispose() {
		mTemplate.resetNotifier();
		mTemplate.noLongerNeeded();
		super.dispose();
	}

	@Override public ArrayList<File> saveTo(File file) {
		ArrayList<File> result = new ArrayList<File>();

		if (mTemplate.save(file)) {
			result.add(file);
			mTemplate.setFile(file);
			adjustWindowTitle();
		} else {
			TKOptionDialog.error(this, Msgs.SAVE_ERROR);
		}
		return result;
	}

	/** @return The template associated with this window. */
	public CMTemplate getTemplate() {
		return mTemplate;
	}

	@Override public String getWindowPrefsPrefix() {
		return getConsumerGroup(mTemplate) + "."; //$NON-NLS-1$
	}

	@Override public boolean isModified() {
		return mTemplate != null && mTemplate.isModified();
	}

	@Override public File getBackingFile() {
		return mTemplate.getFile();
	}

	@Override public TKFileFilter[] getFileFilters() {
		return CSTemplateOpener.FILTERS;
	}

	@Override public TKFileFilter getPreferredFileFilter(TKFileFilter[] filters) {
		return CSFileOpener.getPreferredFileFilter(filters);
	}

	@Override public boolean adjustMenuItem(String command, TKMenuItem item) {
		if (CMD_NEW_ADVANTAGE.equals(command) || CMD_NEW_ADVANTAGE_CONTAINER.equals(command) || CMD_NEW_SKILL.equals(command) || CMD_NEW_SKILL_CONTAINER.equals(command) || CMD_NEW_TECHNIQUE.equals(command) || CMD_NEW_SPELL.equals(command) || CMD_NEW_SPELL_CONTAINER.equals(command) || CMD_NEW_CARRIED_EQUIPMENT.equals(command) || CMD_NEW_CARRIED_EQUIPMENT_CONTAINER.equals(command)) {
			item.setEnabled(true);
		} else if (CMD_APPLY_TEMPLATE_TO_SHEET.equals(command)) {
			item.setEnabled(CSSheetWindow.getTopSheet() != null);
		} else {
			return super.adjustMenuItem(command, item);
		}
		return true;
	}

	@Override public boolean obeyCommand(String command, TKMenuItem item) {
		if (CMD_NEW_ADVANTAGE.equals(command)) {
			addRow(mContent.getAdvantageOutline(), new CMAdvantage(mTemplate, false), item.getTitle());
		} else if (CMD_NEW_ADVANTAGE_CONTAINER.equals(command)) {
			addRow(mContent.getAdvantageOutline(), new CMAdvantage(mTemplate, true), item.getTitle());
		} else if (CMD_NEW_SKILL.equals(command)) {
			addRow(mContent.getSkillOutline(), new CMSkill(mTemplate, false), item.getTitle());
		} else if (CMD_NEW_SKILL_CONTAINER.equals(command)) {
			addRow(mContent.getSkillOutline(), new CMSkill(mTemplate, true), item.getTitle());
		} else if (CMD_NEW_TECHNIQUE.equals(command)) {
			addRow(mContent.getSkillOutline(), new CMTechnique(mTemplate), item.getTitle());
		} else if (CMD_NEW_SPELL.equals(command)) {
			addRow(mContent.getSpellOutline(), new CMSpell(mTemplate, false), item.getTitle());
		} else if (CMD_NEW_SPELL_CONTAINER.equals(command)) {
			addRow(mContent.getSpellOutline(), new CMSpell(mTemplate, true), item.getTitle());
		} else if (CMD_NEW_CARRIED_EQUIPMENT.equals(command)) {
			addRow(mContent.getEquipmentOutline(), new CMEquipment(mTemplate, false), item.getTitle());
		} else if (CMD_NEW_CARRIED_EQUIPMENT_CONTAINER.equals(command)) {
			addRow(mContent.getEquipmentOutline(), new CMEquipment(mTemplate, true), item.getTitle());
		} else if (CMD_APPLY_TEMPLATE_TO_SHEET.equals(command)) {
			applyTemplateToSheet(mTemplate);
		} else {
			return super.obeyCommand(command, item);
		}
		return true;
	}

	private void addRow(CSOutline outline, CMRow row, String name) {
		TKOutlineModel model = outline.getModel();
		TKOutlineModelUndoSnapshot before = new TKOutlineModelUndoSnapshot(model);
		Point cell = outline.getCellLocationOfEditor();
		TKSelection selection = model.getSelection();
		int count = selection.getCount();
		int insertAt;
		TKRow parentRow;

		if (count > 0 || cell != null) {
			insertAt = cell != null ? cell.y : count == 1 ? selection.firstSelectedIndex() : selection.lastSelectedIndex();
			parentRow = model.getRowAtIndex(insertAt++);
			if (!parentRow.canHaveChildren() || !parentRow.isOpen()) {
				parentRow = parentRow.getParent();
			}
			if (parentRow != null && parentRow.canHaveChildren()) {
				parentRow.addChild(row);
			}
		} else {
			insertAt = model.getRowCount();
		}

		model.addRow(insertAt, row);
		outline.postUndo(new TKOutlineModelUndo(name, model, before, new TKOutlineModelUndoSnapshot(model)));
		mContent.revalidate();
		outline.getModel().select(row, false);
		outline.openDetailEditor(true);
	}

	/**
	 * Adds rows to the display.
	 * 
	 * @param rows The rows to add.
	 */
	public void addRows(List<TKRow> rows) {
		HashMap<CSOutline, TKOutlineModelUndoSnapshot> map = new HashMap<CSOutline, TKOutlineModelUndoSnapshot>();
		HashMap<TKOutline, ArrayList<TKRow>> selMap = new HashMap<TKOutline, ArrayList<TKRow>>();
		HashMap<TKOutline, ArrayList<CMRow>> nameMap = new HashMap<TKOutline, ArrayList<CMRow>>();
		CSOutline outline = null;

		for (TKRow row : rows) {
			if (row instanceof CMAdvantage) {
				outline = mContent.getAdvantageOutline();
				if (!map.containsKey(outline)) {
					map.put(outline, new TKOutlineModelUndoSnapshot(outline.getModel()));
				}
				row = new CMAdvantage(mTemplate, (CMAdvantage) row, true);
				addCompleteRow(outline, row, selMap);
			} else if (row instanceof CMTechnique) {
				outline = mContent.getSkillOutline();
				if (!map.containsKey(outline)) {
					map.put(outline, new TKOutlineModelUndoSnapshot(outline.getModel()));
				}
				row = new CMTechnique(mTemplate, (CMTechnique) row, true);
				addCompleteRow(outline, row, selMap);
			} else if (row instanceof CMSkill) {
				outline = mContent.getSkillOutline();
				if (!map.containsKey(outline)) {
					map.put(outline, new TKOutlineModelUndoSnapshot(outline.getModel()));
				}
				row = new CMSkill(mTemplate, (CMSkill) row, true, true);
				addCompleteRow(outline, row, selMap);
			} else if (row instanceof CMSpell) {
				outline = mContent.getSpellOutline();
				if (!map.containsKey(outline)) {
					map.put(outline, new TKOutlineModelUndoSnapshot(outline.getModel()));
				}
				row = new CMSpell(mTemplate, (CMSpell) row, true, true);
				addCompleteRow(outline, row, selMap);
			} else if (row instanceof CMEquipment) {
				outline = mContent.getEquipmentOutline();
				if (!map.containsKey(outline)) {
					map.put(outline, new TKOutlineModelUndoSnapshot(outline.getModel()));
				}
				row = new CMEquipment(mTemplate, (CMEquipment) row, true);
				addCompleteRow(outline, row, selMap);
			} else {
				row = null;
			}
			if (row instanceof CMRow) {
				ArrayList<CMRow> process = nameMap.get(outline);

				if (process == null) {
					process = new ArrayList<CMRow>();
					nameMap.put(outline, process);
				}
				addRowsToBeProcessed(process, (CMRow) row);
			}
		}
		for (CSOutline anOutline : map.keySet()) {
			TKOutlineModel model = anOutline.getModel();

			model.select(selMap.get(anOutline), false);
			anOutline.postUndo(new TKOutlineModelUndo(Msgs.ADD_ROWS, model, map.get(anOutline), new TKOutlineModelUndoSnapshot(model)));
			anOutline.scrollSelectionIntoView();
			anOutline.requestFocus();
		}
		if (!nameMap.isEmpty()) {
			EventQueue.invokeLater(new CSNamePostProcessor(nameMap));
		}
	}

	private void addRowsToBeProcessed(ArrayList<CMRow> list, CMRow row) {
		int count = row.getChildCount();

		list.add(row);

		for (int i = 0; i < count; i++) {
			addRowsToBeProcessed(list, (CMRow) row.getChild(i));
		}
	}

	private void addCompleteRow(TKOutline outline, TKRow row, HashMap<TKOutline, ArrayList<TKRow>> selMap) {
		ArrayList<TKRow> selection = selMap.get(outline);

		addCompleteRow(outline.getModel(), row);
		outline.contentSizeMayHaveChanged();
		if (selection == null) {
			selection = new ArrayList<TKRow>();
			selMap.put(outline, selection);
		}
		selection.add(row);
	}

	private void addCompleteRow(TKOutlineModel outlineModel, TKRow row) {
		outlineModel.addRow(row);
		if (row.isOpen() && row.hasChildren()) {
			for (TKRow child : row.getChildren()) {
				addCompleteRow(outlineModel, child);
			}
		}
	}

	private void createToolBar() {
		TKToolBar toolbar = new TKToolBar(this);
		TKSearch search = new TKSearch(toolbar, CSFont.KEY_FIELD, CSFont.KEY_LABEL);

		toolbar.addControl(search, -1, TKSearch.CMD_SEARCH);
		setTKToolBar(toolbar);
	}

	@Override public boolean adjustToolBarItem(String command, TKPanel item) {
		if (TKSearch.CMD_SEARCH.equals(command)) {
			item.setEnabled(true);
		} else {
			return super.adjustToolBarItem(command, item);
		}
		return true;
	}

	public TKItemRenderer getSearchRenderer() {
		return new CSRowItemRenderer();
	}

	public Collection<Object> search(String text) {
		ArrayList<Object> list = new ArrayList<Object>();

		text = text.toLowerCase();
		searchOne(mContent.getAdvantageOutline(), text, list);
		searchOne(mContent.getSkillOutline(), text, list);
		searchOne(mContent.getSpellOutline(), text, list);
		searchOne(mContent.getEquipmentOutline(), text, list);
		return list;
	}

	private void searchOne(CSOutline outline, String text, ArrayList<Object> list) {
		for (CMRow row : new TKRowIterator<CMRow>(outline.getModel())) {
			if (row.contains(text, true)) {
				list.add(row);
			}
		}
	}

	public void searchSelect(Collection<Object> selection) {
		HashMap<TKOutlineModel, ArrayList<TKRow>> map = new HashMap<TKOutlineModel, ArrayList<TKRow>>();
		TKOutline primary = null;
		ArrayList<TKRow> list;

		mContent.getAdvantageOutline().getModel().deselect();
		mContent.getSkillOutline().getModel().deselect();
		mContent.getSpellOutline().getModel().deselect();
		mContent.getEquipmentOutline().getModel().deselect();

		for (Object obj : selection) {
			TKRow row = (TKRow) obj;
			TKRow parent = row.getParent();
			TKOutlineModel model = row.getOwner();

			while (parent != null) {
				parent.setOpen(true);
				model = parent.getOwner();
				parent = parent.getParent();
			}
			list = map.get(model);
			if (list == null) {
				list = new ArrayList<TKRow>();
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

		for (TKOutlineModel model : map.keySet()) {
			model.select(map.get(model), false);
		}

		if (primary != null) {
			EventQueue.invokeLater(new ScrollToSelection(primary));
			primary.requestFocus();
		}
	}

	/** Helper for scrolling a specific outline's selection into view. */
	class ScrollToSelection implements Runnable {
		private TKOutline	mOutline;

		/** @param outline The outline to scroll the selection into view for. */
		ScrollToSelection(TKOutline outline) {
			mOutline = outline;
		}

		public void run() {
			mOutline.scrollSelectionIntoView();
		}
	}
}
