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

package com.trollworks.gcs.ui.sheet;

import com.trollworks.gcs.model.CMCharacter;
import com.trollworks.gcs.model.CMRow;
import com.trollworks.gcs.model.advantage.CMAdvantage;
import com.trollworks.gcs.model.equipment.CMEquipment;
import com.trollworks.gcs.model.names.USCensusNames;
import com.trollworks.gcs.model.skill.CMSkill;
import com.trollworks.gcs.model.skill.CMTechnique;
import com.trollworks.gcs.model.spell.CMSpell;
import com.trollworks.gcs.ui.common.CSFileOpener;
import com.trollworks.gcs.ui.common.CSFont;
import com.trollworks.gcs.ui.common.CSNamePostProcessor;
import com.trollworks.gcs.ui.common.CSOutline;
import com.trollworks.gcs.ui.common.CSRowItemRenderer;
import com.trollworks.gcs.ui.common.CSWindow;
import com.trollworks.toolkit.io.TKFileFilter;
import com.trollworks.toolkit.io.TKPath;
import com.trollworks.toolkit.print.TKPrintManager;
import com.trollworks.toolkit.widget.TKItemRenderer;
import com.trollworks.toolkit.widget.TKPanel;
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
import java.awt.Graphics;
import java.awt.print.PageFormat;
import java.awt.print.Printable;
import java.io.File;
import java.util.ArrayList;
import java.util.Collection;
import java.util.HashMap;
import java.util.List;

/** The character sheet window. */
public class CSSheetWindow extends CSWindow implements Printable, TKSearchTarget {
	/** The PNG extension. */
	public static final String		PNG_EXTENSION	= ".png";	//$NON-NLS-1$
	/** The PDF extension. */
	public static final String		PDF_EXTENSION	= ".pdf";	//$NON-NLS-1$
	/** The HTML extension. */
	public static final String		HTML_EXTENSION	= ".html";	//$NON-NLS-1$
	private CSSheet					mSheet;
	private CMCharacter				mCharacter;
	private CSPrerequisitesThread	mPrereqThread;

	/**
	 * @param character The character to get the consumer group for.
	 * @return The consumer group for the specified character.
	 */
	static String getConsumerGroup(CMCharacter character) {
		return "CSSheetWindow:" + character.getUniqueID(); //$NON-NLS-1$
	}

	/** @return The top character sheet window, if any. */
	public static CSSheetWindow getTopSheet() {
		ArrayList<CSSheetWindow> list = TKWindow.getActiveWindows(CSSheetWindow.class);

		return list.isEmpty() ? null : list.get(0);
	}

	/**
	 * Looks for an existing character sheet window for the specified character.
	 * 
	 * @param character The character to look for.
	 * @return The character sheet window for the specified character, if any.
	 */
	public static CSSheetWindow findSheetWindow(CMCharacter character) {
		for (CSSheetWindow window : TKWindow.getWindows(CSSheetWindow.class)) {
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
	public static CSSheetWindow findSheetWindow(File file) {
		String fullPath = TKPath.getFullPath(file);

		for (CSSheetWindow window : TKWindow.getWindows(CSSheetWindow.class)) {
			File wFile = window.getCharacter().getFile();

			if (wFile != null) {
				if (TKPath.getFullPath(wFile).equals(fullPath)) {
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
	 */
	public static void displaySheetWindow(CMCharacter character) {
		CSSheetWindow window = findSheetWindow(character);

		if (window == null) {
			window = new CSSheetWindow(character);
		}
		window.setVisible(true);
	}

	/**
	 * Creates character sheet window.
	 * 
	 * @param character The character to display.
	 */
	public CSSheetWindow(CMCharacter character) {
		super(null, character.getFileIcon(true), character.getFileIcon(false));
		mCharacter = character;
		mSheet = new CSSheet(mCharacter);
		adjustWindowTitle();
		setContent(new TKScrollPanel(mSheet));
		mSheet.rebuild();
		createToolBar();
		restoreBounds();
		mPrereqThread = new CSPrerequisitesThread(mSheet);
		mPrereqThread.start();
		CSPrerequisitesThread.waitForProcessingToFinish(character);
		mCharacter.discardAllEdits();
		setUndoManager(mCharacter);
		mCharacter.setModified(false);
	}

	/** Notify background threads of prereq or feature modifications. */
	public void notifyOfPrereqOrFeatureModification() {
		mPrereqThread.markForUpdate();
	}

	private void adjustWindowTitle() {
		File file = mCharacter.getFile();
		String title;

		if (file == null) {
			title = TKWindow.getNextUntitledWindowName(CSSheetWindow.class, Msgs.UNTITLED_SHEET, this);
		} else {
			title = TKPath.getLeafName(file.getName(), false);
		}
		setTitle(title);
	}

	@Override public void dispose() {
		mSheet.dispose();
		mCharacter.noLongerNeeded();
		super.dispose();
	}

	/** @return The character associated with this window. */
	public CMCharacter getCharacter() {
		return mCharacter;
	}

	@Override public void adjustToPageSetupChanges() {
		mSheet.rebuild();
		mSheet.revalidateImmediately();
	}

	public int print(Graphics graphics, PageFormat pageFormat, int pageIndex) {
		return mSheet.print(graphics, pageFormat, pageIndex);
	}

	@Override protected String getSaveAsName() {
		String name = mCharacter.getName();

		return name.length() > 0 ? name : super.getSaveAsName();
	}

	@Override public ArrayList<File> saveTo(File file) {
		ArrayList<File> result = new ArrayList<File>();
		String extension = TKPath.getExtension(file.getName());

		if (HTML_EXTENSION.equals(extension)) {
			if (mSheet.saveAsHTML(file)) {
				result.add(file);
			} else {
				TKOptionDialog.error(this, Msgs.SAVE_AS_HTML_ERROR);
			}
		} else if (PNG_EXTENSION.equals(extension)) {
			if (!mSheet.saveAsPNG(file, result)) {
				TKOptionDialog.error(this, Msgs.SAVE_AS_PNG_ERROR);
			}
		} else if (PDF_EXTENSION.equals(extension)) {
			if (mSheet.saveAsPDF(file)) {
				result.add(file);
			} else {
				TKOptionDialog.error(this, Msgs.SAVE_AS_PDF_ERROR);
			}
		} else {
			if (mCharacter.save(file)) {
				result.add(file);
				mCharacter.setFile(file);
				adjustWindowTitle();
			} else {
				TKOptionDialog.error(this, Msgs.SAVE_ERROR);
			}
		}
		return result;
	}

	@Override public String getWindowPrefsPrefix() {
		return getConsumerGroup(mCharacter) + "."; //$NON-NLS-1$
	}

	@Override public boolean isModified() {
		return mCharacter != null && mCharacter.isModified();
	}

	@Override public File getBackingFile() {
		return mCharacter.getFile();
	}

	@Override public TKFileFilter[] getFileFilters() {
		return new TKFileFilter[] { CSSheetOpener.FILTERS[0], new TKFileFilter(Msgs.HTML_DESCRIPTION, HTML_EXTENSION), new TKFileFilter(Msgs.PDF_DESCRIPTION, PDF_EXTENSION), new TKFileFilter(Msgs.PNG_DESCRIPTION, PNG_EXTENSION) };
	}

	@Override public TKFileFilter getPreferredFileFilter(TKFileFilter[] filters) {
		return CSFileOpener.getPreferredFileFilter(filters);
	}

	@Override public boolean adjustMenuItem(String command, TKMenuItem item) {
		if (CMD_NEW_ADVANTAGE.equals(command) || CMD_NEW_ADVANTAGE_CONTAINER.equals(command) || CMD_NEW_SKILL.equals(command) || CMD_NEW_SKILL_CONTAINER.equals(command) || CMD_NEW_TECHNIQUE.equals(command) || CMD_NEW_SPELL.equals(command) || CMD_NEW_SPELL_CONTAINER.equals(command) || CMD_NEW_CARRIED_EQUIPMENT.equals(command) || CMD_NEW_CARRIED_EQUIPMENT_CONTAINER.equals(command) || CMD_NEW_EQUIPMENT.equals(command) || CMD_NEW_EQUIPMENT_CONTAINER.equals(command) || CMD_RANDOMIZE_DESCRIPTION.equals(command) || CMD_RANDOMIZE_FEMALE_NAME.equals(command) || CMD_RANDOM_MALE_NAME.equals(command)) {
			item.setEnabled(true);
		} else if (CMD_ADD_NATURAL_PUNCH.equals(command)) {
			item.setMarked(mCharacter.includePunch());
			item.setEnabled(true);
		} else if (CMD_ADD_NATURAL_KICK.equals(command)) {
			item.setMarked(mCharacter.includeKick());
			item.setEnabled(true);
		} else if (CMD_ADD_NATURAL_KICK_WITH_BOOTS.equals(command)) {
			item.setMarked(mCharacter.includeKickBoots());
			item.setEnabled(true);
		} else {
			return super.adjustMenuItem(command, item);
		}
		return true;
	}

	@Override public boolean obeyCommand(String command, TKMenuItem item) {
		if (CMD_NEW_ADVANTAGE.equals(command)) {
			addRow(mSheet.getAdvantageOutline(), new CMAdvantage(mCharacter, false), item.getTitle());
		} else if (CMD_NEW_ADVANTAGE_CONTAINER.equals(command)) {
			addRow(mSheet.getAdvantageOutline(), new CMAdvantage(mCharacter, true), item.getTitle());
		} else if (CMD_NEW_SKILL.equals(command)) {
			addRow(mSheet.getSkillOutline(), new CMSkill(mCharacter, false), item.getTitle());
		} else if (CMD_NEW_SKILL_CONTAINER.equals(command)) {
			addRow(mSheet.getSkillOutline(), new CMSkill(mCharacter, true), item.getTitle());
		} else if (CMD_NEW_TECHNIQUE.equals(command)) {
			addRow(mSheet.getSkillOutline(), new CMTechnique(mCharacter), item.getTitle());
		} else if (CMD_NEW_SPELL.equals(command)) {
			addRow(mSheet.getSpellOutline(), new CMSpell(mCharacter, false), item.getTitle());
		} else if (CMD_NEW_SPELL_CONTAINER.equals(command)) {
			addRow(mSheet.getSpellOutline(), new CMSpell(mCharacter, true), item.getTitle());
		} else if (CMD_NEW_CARRIED_EQUIPMENT.equals(command)) {
			addRow(mSheet.getCarriedEquipmentOutline(), new CMEquipment(mCharacter, false), item.getTitle());
		} else if (CMD_NEW_CARRIED_EQUIPMENT_CONTAINER.equals(command)) {
			addRow(mSheet.getCarriedEquipmentOutline(), new CMEquipment(mCharacter, true), item.getTitle());
		} else if (CMD_NEW_EQUIPMENT.equals(command)) {
			addRow(mSheet.getOtherEquipmentOutline(), new CMEquipment(mCharacter, false), item.getTitle());
		} else if (CMD_NEW_EQUIPMENT_CONTAINER.equals(command)) {
			addRow(mSheet.getOtherEquipmentOutline(), new CMEquipment(mCharacter, true), item.getTitle());
		} else if (CMD_RANDOMIZE_DESCRIPTION.equals(command)) {
			CSDescriptionRandomizer.randomize(getCharacter());
		} else if (CMD_RANDOMIZE_FEMALE_NAME.equals(command)) {
			mCharacter.setName(USCensusNames.INSTANCE.getFullName(false));
		} else if (CMD_RANDOM_MALE_NAME.equals(command)) {
			mCharacter.setName(USCensusNames.INSTANCE.getFullName(true));
		} else if (CMD_ADD_NATURAL_PUNCH.equals(command)) {
			mCharacter.setIncludePunch(!mCharacter.includePunch());
		} else if (CMD_ADD_NATURAL_KICK.equals(command)) {
			mCharacter.setIncludeKick(!mCharacter.includeKick());
		} else if (CMD_ADD_NATURAL_KICK_WITH_BOOTS.equals(command)) {
			mCharacter.setIncludeKickBoots(!mCharacter.includeKickBoots());
		} else if (CMD_REDO.equals(command) || CMD_UNDO.equals(command)) {
			if (super.obeyCommand(command, item)) {
				notifyOfPrereqOrFeatureModification();
				repaint();
			}
		} else {
			return super.obeyCommand(command, item);
		}
		return true;
	}

	private void addRow(CSOutline outline, CMRow row, String name) {
		outline.addRow(row, name, false);
		outline.getModel().select(row, false);
		outline.openDetailEditor(true);
	}

	/**
	 * Adds rows to the sheet.
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
				outline = mSheet.getAdvantageOutline();
				if (!map.containsKey(outline)) {
					map.put(outline, new TKOutlineModelUndoSnapshot(outline.getModel()));
				}
				row = new CMAdvantage(mCharacter, (CMAdvantage) row, true);
				addCompleteRow(outline, row, selMap);
			} else if (row instanceof CMTechnique) {
				outline = mSheet.getSkillOutline();
				if (!map.containsKey(outline)) {
					map.put(outline, new TKOutlineModelUndoSnapshot(outline.getModel()));
				}
				row = new CMTechnique(mCharacter, (CMTechnique) row, true);
				addCompleteRow(outline, row, selMap);
			} else if (row instanceof CMSkill) {
				outline = mSheet.getSkillOutline();
				if (!map.containsKey(outline)) {
					map.put(outline, new TKOutlineModelUndoSnapshot(outline.getModel()));
				}
				row = new CMSkill(mCharacter, (CMSkill) row, true, true);
				addCompleteRow(outline, row, selMap);
			} else if (row instanceof CMSpell) {
				outline = mSheet.getSpellOutline();
				if (!map.containsKey(outline)) {
					map.put(outline, new TKOutlineModelUndoSnapshot(outline.getModel()));
				}
				row = new CMSpell(mCharacter, (CMSpell) row, true, true);
				addCompleteRow(outline, row, selMap);
			} else if (row instanceof CMEquipment) {
				outline = mSheet.getCarriedEquipmentOutline();
				if (!map.containsKey(outline)) {
					map.put(outline, new TKOutlineModelUndoSnapshot(outline.getModel()));
				}
				row = new CMEquipment(mCharacter, (CMEquipment) row, true);
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
		searchOne(mSheet.getAdvantageOutline(), text, list);
		searchOne(mSheet.getSkillOutline(), text, list);
		searchOne(mSheet.getSpellOutline(), text, list);
		searchOne(mSheet.getCarriedEquipmentOutline(), text, list);
		searchOne(mSheet.getOtherEquipmentOutline(), text, list);
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

		mSheet.getAdvantageOutline().getModel().deselect();
		mSheet.getSkillOutline().getModel().deselect();
		mSheet.getSpellOutline().getModel().deselect();
		mSheet.getCarriedEquipmentOutline().getModel().deselect();
		mSheet.getOtherEquipmentOutline().getModel().deselect();

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
				primary = mSheet.getAdvantageOutline();
				if (model != primary.getModel()) {
					primary = mSheet.getSkillOutline();
					if (model != primary.getModel()) {
						primary = mSheet.getSpellOutline();
						if (model != primary.getModel()) {
							primary = mSheet.getCarriedEquipmentOutline();
							if (model != primary.getModel()) {
								primary = mSheet.getOtherEquipmentOutline();
							}
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

	@Override public void forceRepaintAndInvalidate() {
		mSheet.rebuild();
		super.forceRepaintAndInvalidate();
	}

	@Override protected TKPrintManager createPageSettings() {
		return mCharacter.getPageSettings();
	}

	/** @return The embedded sheet. */
	public CSSheet getSheet() {
		return mSheet;
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
