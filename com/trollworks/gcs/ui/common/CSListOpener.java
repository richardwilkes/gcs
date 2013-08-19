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

package com.trollworks.gcs.ui.common;

import com.trollworks.gcs.model.advantage.CMAdvantageList;
import com.trollworks.gcs.model.equipment.CMEquipmentList;
import com.trollworks.gcs.model.skill.CMSkillList;
import com.trollworks.gcs.model.spell.CMSpellList;
import com.trollworks.gcs.ui.advantage.CSAdvantageListWindow;
import com.trollworks.gcs.ui.equipment.CSEquipmentListWindow;
import com.trollworks.gcs.ui.skills.CSSkillListWindow;
import com.trollworks.gcs.ui.spell.CSSpellListWindow;
import com.trollworks.toolkit.io.TKFileFilter;
import com.trollworks.toolkit.io.TKPath;
import com.trollworks.toolkit.utility.TKDebug;
import com.trollworks.toolkit.utility.TKFont;
import com.trollworks.toolkit.widget.TKLabel;
import com.trollworks.toolkit.widget.TKPanel;
import com.trollworks.toolkit.widget.border.TKEmptyBorder;
import com.trollworks.toolkit.widget.button.TKRadioButton;
import com.trollworks.toolkit.widget.button.TKRadioButtonGroup;
import com.trollworks.toolkit.widget.layout.TKColumnLayout;
import com.trollworks.toolkit.widget.layout.TKCompassLayout;
import com.trollworks.toolkit.widget.layout.TKCompassPosition;
import com.trollworks.toolkit.window.TKDialog;
import com.trollworks.toolkit.window.TKFileDialog;
import com.trollworks.toolkit.window.TKOptionDialog;
import com.trollworks.toolkit.window.TKWindow;
import com.trollworks.toolkit.window.TKWindowOpener;

import java.io.File;
import java.util.List;

/** Handles opening of list files. */
public class CSListOpener implements TKWindowOpener {
	/** The filter index for (Dis)Advantage lists. */
	public static final int				ADVANTAGE_FILTER	= 0;
	/** The extension for (Dis)Advantage lists. */
	public static final String			ADVANTAGE_EXTENSION	= ".adq";																																																																				//$NON-NLS-1$
	/** The filter index for Equipment lists. */
	public static final int				EQUIPMENT_FILTER	= 1;
	/** The extension for Equipment lists. */
	public static final String			EQUIPMENT_EXTENSION	= ".eqp";																																																																				//$NON-NLS-1$
	/** The filter index for Skill lists. */
	public static final int				SKILL_FILTER		= 2;
	/** The filter index for Skill lists. */
	public static final String			SKILL_EXTENSION		= ".skl";																																																																				//$NON-NLS-1$
	/** The filter index for Spell lists. */
	public static final int				SPELL_FILTER		= 3;
	/** The filter index for Spell lists. */
	public static final String			SPELL_EXTENSION		= ".spl";																																																																				//$NON-NLS-1$
	/** The file filters for the lists. */
	public static TKFileFilter[]		FILTERS				= new TKFileFilter[] { new TKFileFilter(Msgs.ADVANTAGES_DESCRIPTION, ADVANTAGE_EXTENSION), new TKFileFilter(Msgs.EQUIPMENT_DESCRIPTION, EQUIPMENT_EXTENSION), new TKFileFilter(Msgs.SKILLS_DESCRIPTION, SKILL_EXTENSION), new TKFileFilter(Msgs.SPELLS_DESCRIPTION, SPELL_EXTENSION) };
	private static int					LAST_CHOICE			= 0;
	private static final CSListOpener	INSTANCE			= new CSListOpener();

	/** @return The one and only instance of the list opener. */
	public static final CSListOpener getInstance() {
		return INSTANCE;
	}

	/**
	 * @param id One of {@link #ADVANTAGE_FILTER}, {@link #EQUIPMENT_FILTER},
	 *            {@link #SKILL_FILTER}, {@link #SPELL_FILTER}.
	 * @param filters The filters to choose from.
	 * @return The preferred file filter to use from the array of filters.
	 */
	public static TKFileFilter getPreferredFileFilter(int id, TKFileFilter[] filters) {
		if (filters != null) {
			for (TKFileFilter element : filters) {
				if (element == FILTERS[id]) {
					return element;
				}
			}
		}
		return null;
	}

	private CSListOpener() {
		TKFileDialog.setIconForFileExtension(ADVANTAGE_EXTENSION, CSImage.getAdvantageIcon(false, false));
		TKFileDialog.setIconForFileExtension(EQUIPMENT_EXTENSION, CSImage.getEquipmentIcon(false, false));
		TKFileDialog.setIconForFileExtension(SKILL_EXTENSION, CSImage.getSkillIcon(false, false));
		TKFileDialog.setIconForFileExtension(SPELL_EXTENSION, CSImage.getSpellIcon(false, false));
	}

	public TKFileFilter[] getFileFilters() {
		return FILTERS;
	}

	public TKWindow openWindow(Object obj, boolean show, boolean finalChance, List<String> msgs) {
		if (obj instanceof String) {
			obj = new File((String) obj);
		}

		if (obj instanceof File) {
			File file = (File) obj;
			String extension = TKPath.getExtension(file.getName());

			for (int i = 0; i < FILTERS.length; i++) {
				if (finalChance || FILTERS[i].getExtensions()[0].equals(extension)) {
					CSListWindow window = CSListWindow.findListWindow(file);

					if (window == null) {
						try {
							switch (i) {
								case ADVANTAGE_FILTER:
									window = new CSAdvantageListWindow(new CMAdvantageList(file));
									break;
								case SKILL_FILTER:
									window = new CSSkillListWindow(new CMSkillList(file));
									break;
								case EQUIPMENT_FILTER:
									window = new CSEquipmentListWindow(new CMEquipmentList(file));
									break;
								case SPELL_FILTER:
									window = new CSSpellListWindow(new CMSpellList(file));
									break;
								default:
									throw new Exception();
							}
						} catch (Exception exception) {
							if (TKDebug.isKeySet(TKDebug.KEY_DIAGNOSE_LOAD_SAVE)) {
								exception.printStackTrace(System.err);
							}
							// Do nothing... will be dealt with later
						}
					}

					if (window != null) {
						if (show) {
							window.setVisible(true);
						}
						return window;
					}
				}
			}
		}

		return null;
	}

	/** Creates a new list. */
	public void newList() {
		TKOptionDialog dialog = new TKOptionDialog(CSMenuKeys.getTitle(CSWindow.CMD_NEW_LIST), TKOptionDialog.TYPE_OK_CANCEL);
		TKPanel panel = new TKPanel(new TKCompassLayout());
		TKRadioButtonGroup group = new TKRadioButtonGroup();
		panel.add(new TKLabel(Msgs.CHOICE_QUERY, TKFont.CONTROL_FONT_KEY), TKCompassPosition.NORTH);
		TKPanel buttonPanel = new TKPanel(new TKColumnLayout());
		TKRadioButton advantageButton = createButton(buttonPanel, group, Msgs.ADVANTAGE_CHOICE);
		TKRadioButton skillButton = createButton(buttonPanel, group, Msgs.SKILL_CHOICE);
		TKRadioButton spellButton = createButton(buttonPanel, group, Msgs.SPELL_CHOICE);
		TKRadioButton equipmentButton = createButton(buttonPanel, group, Msgs.EQUIPMENT_CHOICE);

		buttonPanel.setBorder(new TKEmptyBorder(10, 25, 0, 0));
		panel.add(buttonPanel, TKCompassPosition.CENTER);
		switch (LAST_CHOICE) {
			case 0:
				advantageButton.setSelected(true);
				break;
			case 1:
				skillButton.setSelected(true);
				break;
			case 2:
				spellButton.setSelected(true);
				break;
			case 3:
				equipmentButton.setSelected(true);
				break;
		}
		if (dialog.doModal(null, panel) == TKDialog.OK) {
			if (advantageButton.isSelected()) {
				LAST_CHOICE = 0;
				CSListWindow.displayListWindow(new CMAdvantageList());
			} else if (skillButton.isSelected()) {
				LAST_CHOICE = 1;
				CSListWindow.displayListWindow(new CMSkillList());
			} else if (spellButton.isSelected()) {
				LAST_CHOICE = 2;
				CSListWindow.displayListWindow(new CMSpellList());
			} else if (equipmentButton.isSelected()) {
				LAST_CHOICE = 3;
				CSListWindow.displayListWindow(new CMEquipmentList());
			}
		}
	}

	private static TKRadioButton createButton(TKPanel panel, TKRadioButtonGroup group, String title) {
		TKRadioButton button = new TKRadioButton(title);

		group.add(button);
		panel.add(button);
		return button;
	}
}
