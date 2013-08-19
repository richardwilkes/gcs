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
 * 2005-2009 the Initial Developer. All Rights Reserved.
 *
 * Contributor(s):
 *
 * ***** END LICENSE BLOCK ***** */

package com.trollworks.gcs.skill;

import com.trollworks.gcs.character.GURPSCharacter;
import com.trollworks.gcs.common.DataFile;
import com.trollworks.gcs.common.ListFile;
import com.trollworks.gcs.library.LibraryFile;
import com.trollworks.gcs.template.Template;
import com.trollworks.gcs.widgets.outline.ListHeaderCell;
import com.trollworks.gcs.widgets.outline.ListTextCell;
import com.trollworks.gcs.widgets.outline.MultiCell;
import com.trollworks.ttk.text.Numbers;
import com.trollworks.ttk.utility.LocalizedMessages;
import com.trollworks.ttk.widgets.outline.Cell;
import com.trollworks.ttk.widgets.outline.Column;
import com.trollworks.ttk.widgets.outline.Outline;
import com.trollworks.ttk.widgets.outline.OutlineModel;

import javax.swing.SwingConstants;

/** Definitions for skill columns. */
public enum SkillColumn {
	/** The skill name/description. */
	DESCRIPTION {
		@Override
		public String toString() {
			return MSG_SKILLS;
		}

		@Override
		public String getToolTip() {
			return MSG_SKILLS_TOOLTIP;
		}

		@Override
		public Cell getCell() {
			return new MultiCell();
		}

		@Override
		public boolean shouldDisplay(DataFile dataFile) {
			return true;
		}

		@Override
		public Object getData(Skill skill) {
			return getDataAsText(skill);
		}

		@Override
		public String getDataAsText(Skill skill) {
			StringBuilder builder = new StringBuilder();
			String notes = skill.getNotes();

			builder.append(skill.toString());
			if (notes.length() > 0) {
				builder.append(" - "); //$NON-NLS-1$
				builder.append(notes);
			}
			return builder.toString();
		}
	},
	/** The skill difficulty. */
	DIFFICULTY {
		@Override
		public String toString() {
			return MSG_DIFFICULTY;
		}

		@Override
		public String getToolTip() {
			return MSG_DIFFICULTY_TOOLTIP;
		}

		@Override
		public Cell getCell() {
			return new ListTextCell(SwingConstants.RIGHT, false);
		}

		@Override
		public boolean shouldDisplay(DataFile dataFile) {
			return dataFile instanceof ListFile || dataFile instanceof LibraryFile;
		}

		@Override
		public Object getData(Skill skill) {
			return new Integer(skill.canHaveChildren() ? -1 : skill.getDifficulty().ordinal() + (skill.getAttribute().ordinal() << 8));
		}

		@Override
		public String getDataAsText(Skill skill) {
			return skill.getDifficultyAsText();
		}
	},
	/** The skill level. */
	LEVEL {
		@Override
		public String toString() {
			return MSG_LEVEL;
		}

		@Override
		public String getToolTip() {
			return MSG_LEVEL_TOOLTIP;
		}

		@Override
		public Cell getCell() {
			return new ListTextCell(SwingConstants.RIGHT, false);
		}

		@Override
		public boolean shouldDisplay(DataFile dataFile) {
			return dataFile instanceof GURPSCharacter;
		}

		@Override
		public Object getData(Skill skill) {
			return new Integer(skill.canHaveChildren() ? -1 : skill.getLevel());
		}

		@Override
		public String getDataAsText(Skill skill) {
			int level;

			if (skill.canHaveChildren()) {
				return ""; //$NON-NLS-1$
			}
			level = skill.getLevel();
			if (level < 0) {
				return "-"; //$NON-NLS-1$
			}
			return Numbers.format(level);
		}
	},
	/** The relative skill level. */
	RELATIVE_LEVEL {
		@Override
		public String toString() {
			return MSG_RELATIVE_LEVEL;
		}

		@Override
		public String getToolTip() {
			return MSG_RELATIVE_LEVEL_TOOLTIP;
		}

		@Override
		public Cell getCell() {
			return new ListTextCell(SwingConstants.RIGHT, false);
		}

		@Override
		public boolean shouldDisplay(DataFile dataFile) {
			return dataFile instanceof Template || dataFile instanceof GURPSCharacter;
		}

		@Override
		public Object getData(Skill skill) {
			return new Integer(getRelativeLevel(skill));
		}

		private int getRelativeLevel(Skill skill) {
			if (!skill.canHaveChildren()) {
				if (skill.getCharacter() != null) {
					if (skill.getLevel() < 0) {
						return Integer.MIN_VALUE;
					}
					if (skill instanceof Technique) {
						return skill.getRelativeLevel() + ((Technique) skill).getDefault().getModifier();
					}
					return skill.getRelativeLevel();
				} else if (skill.getTemplate() != null) {
					int points = skill.getPoints();

					if (points > 0) {
						SkillDifficulty difficulty = skill.getDifficulty();
						int level;

						if (skill instanceof Technique) {
							if (difficulty != SkillDifficulty.A) {
								points--;
							}
							return points + ((Technique) skill).getDefault().getModifier();
						}

						level = difficulty.getBaseRelativeLevel();
						if (points > 1) {
							if (points < 4) {
								level++;
							} else {
								level += 1 + points / 4;
							}
						}
						return level;
					}
				}
			}
			return Integer.MIN_VALUE;
		}

		@Override
		public String getDataAsText(Skill skill) {
			if (!skill.canHaveChildren()) {
				int level = getRelativeLevel(skill);
				StringBuilder builder;

				if (level == Integer.MIN_VALUE) {
					return "-"; //$NON-NLS-1$
				}
				builder = new StringBuilder();
				if (!(skill instanceof Technique)) {
					builder.append(skill.getAttribute());
				}
				builder.append(Numbers.formatWithForcedSign(level));
				return builder.toString();
			}
			return ""; //$NON-NLS-1$
		}
	},
	/** The points spent in the skill. */
	POINTS {
		@Override
		public String toString() {
			return MSG_POINTS;
		}

		@Override
		public String getToolTip() {
			return MSG_POINTS_TOOLTIP;
		}

		@Override
		public Cell getCell() {
			return new ListTextCell(SwingConstants.RIGHT, false);
		}

		@Override
		public boolean shouldDisplay(DataFile dataFile) {
			return dataFile instanceof Template || dataFile instanceof GURPSCharacter;
		}

		@Override
		public Object getData(Skill skill) {
			return new Integer(skill.canHaveChildren() ? -1 : skill.getPoints());
		}

		@Override
		public String getDataAsText(Skill skill) {
			if (skill.canHaveChildren()) {
				return ""; //$NON-NLS-1$
			}
			return Numbers.format(skill.getPoints());
		}
	},
	/** The category. */
	CATEGORY {
		@Override
		public String toString() {
			return MSG_CATEGORY;
		}

		@Override
		public String getToolTip() {
			return MSG_CATEGORY_TOOLTIP;
		}

		@Override
		public Cell getCell() {
			return new ListTextCell(SwingConstants.LEFT, true);
		}

		@Override
		public boolean shouldDisplay(DataFile dataFile) {
			return dataFile instanceof ListFile || dataFile instanceof LibraryFile;
		}

		@Override
		public Object getData(Skill skill) {
			return getDataAsText(skill);
		}

		@Override
		public String getDataAsText(Skill skill) {
			return skill.getCategoriesAsString();
		}
	},
	/** The page reference. */
	REFERENCE {
		@Override
		public String toString() {
			return MSG_REFERENCE;
		}

		@Override
		public String getToolTip() {
			return MSG_REFERENCE_TOOLTIP;
		}

		@Override
		public Cell getCell() {
			return new ListTextCell(SwingConstants.RIGHT, false);
		}

		@Override
		public boolean shouldDisplay(DataFile dataFile) {
			return true;
		}

		@Override
		public Object getData(Skill skill) {
			return getDataAsText(skill);
		}

		@Override
		public String getDataAsText(Skill skill) {
			return skill.getReference();
		}
	};

	static String	MSG_SKILLS;
	static String	MSG_SKILLS_TOOLTIP;
	static String	MSG_LEVEL;
	static String	MSG_LEVEL_TOOLTIP;
	static String	MSG_RELATIVE_LEVEL;
	static String	MSG_RELATIVE_LEVEL_TOOLTIP;
	static String	MSG_POINTS;
	static String	MSG_POINTS_TOOLTIP;
	static String	MSG_DIFFICULTY;
	static String	MSG_DIFFICULTY_TOOLTIP;
	static String	MSG_CATEGORY;
	static String	MSG_CATEGORY_TOOLTIP;
	static String	MSG_REFERENCE;
	static String	MSG_REFERENCE_TOOLTIP;

	static {
		LocalizedMessages.initialize(SkillColumn.class);
	}

	/**
	 * @param skill The {@link Skill} to get the data from.
	 * @return An object representing the data for this column.
	 */
	public abstract Object getData(Skill skill);

	/**
	 * @param skill The {@link Skill} to get the data from.
	 * @return Text representing the data for this column.
	 */
	public abstract String getDataAsText(Skill skill);

	/** @return The tooltip for the column. */
	public abstract String getToolTip();

	/** @return The {@link Cell} used to display the data. */
	public abstract Cell getCell();

	/**
	 * @param dataFile The {@link DataFile} to use.
	 * @return Whether this column should be displayed for the specified data file.
	 */
	public abstract boolean shouldDisplay(DataFile dataFile);

	/**
	 * Adds all relevant {@link Column}s to a {@link Outline}.
	 * 
	 * @param outline The {@link Outline} to use.
	 * @param dataFile The {@link DataFile} that data is being displayed for.
	 */
	public static void addColumns(Outline outline, DataFile dataFile) {
		boolean sheetOrTemplate = dataFile instanceof GURPSCharacter || dataFile instanceof Template;
		OutlineModel model = outline.getModel();
		for (SkillColumn one : values()) {
			if (one.shouldDisplay(dataFile)) {
				Column column = new Column(one.ordinal(), one.toString(), one.getToolTip(), one.getCell());

				column.setHeaderCell(new ListHeaderCell(sheetOrTemplate));
				model.addColumn(column);
			}
		}
	}
}
