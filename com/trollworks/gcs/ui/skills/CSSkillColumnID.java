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

package com.trollworks.gcs.ui.skills;

import com.trollworks.gcs.model.CMCharacter;
import com.trollworks.gcs.model.CMDataFile;
import com.trollworks.gcs.model.CMListFile;
import com.trollworks.gcs.model.CMTemplate;
import com.trollworks.gcs.model.skill.CMSkill;
import com.trollworks.gcs.model.skill.CMSkillDifficulty;
import com.trollworks.gcs.model.skill.CMTechnique;
import com.trollworks.gcs.ui.common.CSHeaderCell;
import com.trollworks.gcs.ui.common.CSMultiCell;
import com.trollworks.gcs.ui.common.CSTextCell;
import com.trollworks.toolkit.utility.TKAlignment;
import com.trollworks.toolkit.utility.TKNumberUtils;
import com.trollworks.toolkit.widget.outline.TKOutline;
import com.trollworks.toolkit.widget.outline.TKCell;
import com.trollworks.toolkit.widget.outline.TKColumn;
import com.trollworks.toolkit.widget.outline.TKOutlineModel;
import com.trollworks.toolkit.widget.outline.TKTextCell;

/** Definitions for skill columns. */
public enum CSSkillColumnID {
	/** The skill name/description. */
	DESCRIPTION(Msgs.SKILLS, Msgs.SKILLS_TOOLTIP) {
		@Override public TKCell getCell() {
			return new CSMultiCell();
		}

		@Override public boolean shouldDisplay(CMDataFile dataFile) {
			return true;
		}

		@Override public Object getData(CMSkill skill) {
			return getDataAsText(skill);
		}

		@Override public String getDataAsText(CMSkill skill) {
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
	DIFFICULTY(Msgs.DIFFICULTY, Msgs.DIFFICULTY_TOOLTIP) {
		@Override public TKCell getCell() {
			return new CSTextCell(TKAlignment.RIGHT, TKTextCell.COMPARE_AS_TEXT, null, false);
		}

		@Override public boolean shouldDisplay(CMDataFile dataFile) {
			return dataFile instanceof CMListFile;
		}

		@Override public Object getData(CMSkill skill) {
			return new Integer(skill.canHaveChildren() ? -1 : skill.getDifficulty().ordinal() + (skill.getAttribute().ordinal() << 8));
		}

		@Override public String getDataAsText(CMSkill skill) {
			return skill.getDifficultyAsText();
		}
	},
	/** The skill level. */
	LEVEL(Msgs.LEVEL, Msgs.LEVEL_TOOLTIP) {
		@Override public TKCell getCell() {
			return new CSTextCell(TKAlignment.RIGHT, TKTextCell.COMPARE_AS_INTEGER, null, false);
		}

		@Override public boolean shouldDisplay(CMDataFile dataFile) {
			return dataFile instanceof CMCharacter;
		}

		@Override public Object getData(CMSkill skill) {
			return new Integer(skill.canHaveChildren() ? -1 : skill.getLevel());
		}

		@Override public String getDataAsText(CMSkill skill) {
			int level;

			if (skill.canHaveChildren()) {
				return ""; //$NON-NLS-1$
			}
			level = skill.getLevel();
			if (level < 0) {
				return "-"; //$NON-NLS-1$
			}
			return TKNumberUtils.format(level);
		}
	},
	/** The relative skill level. */
	RELATIVE_LEVEL(Msgs.RELATIVE_LEVEL, Msgs.RELATIVE_LEVEL_TOOLTIP) {
		@Override public TKCell getCell() {
			return new CSTextCell(TKAlignment.RIGHT, TKTextCell.COMPARE_AS_TEXT, null, false);
		}

		@Override public boolean shouldDisplay(CMDataFile dataFile) {
			return dataFile instanceof CMTemplate || dataFile instanceof CMCharacter;
		}

		@Override public Object getData(CMSkill skill) {
			return new Integer(getRelativeLevel(skill));
		}

		private int getRelativeLevel(CMSkill skill) {
			if (!skill.canHaveChildren()) {
				if (skill.getCharacter() != null) {
					if (skill.getLevel() < 0) {
						return Integer.MIN_VALUE;
					}
					if (skill instanceof CMTechnique) {
						return skill.getRelativeLevel() + ((CMTechnique) skill).getDefault().getModifier();
					}
					return skill.getRelativeLevel();
				} else if (skill.getTemplate() != null) {
					int points = skill.getPoints();

					if (points > 0) {
						CMSkillDifficulty difficulty = skill.getDifficulty();
						int level;

						if (skill instanceof CMTechnique) {
							if (difficulty != CMSkillDifficulty.A) {
								points--;
							}
							return points + ((CMTechnique) skill).getDefault().getModifier();
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

		@Override public String getDataAsText(CMSkill skill) {
			if (!skill.canHaveChildren()) {
				int level = getRelativeLevel(skill);
				StringBuilder builder;

				if (level == Integer.MIN_VALUE) {
					return "-"; //$NON-NLS-1$
				}
				builder = new StringBuilder();
				if (!(skill instanceof CMTechnique)) {
					builder.append(skill.getAttribute());
				}
				builder.append(TKNumberUtils.format(level, true));
				return builder.toString();
			}
			return ""; //$NON-NLS-1$
		}
	},
	/** The points spent in the skill. */
	POINTS(Msgs.POINTS, Msgs.POINTS_TOOLTIP) {
		@Override public TKCell getCell() {
			return new CSTextCell(TKAlignment.RIGHT, TKTextCell.COMPARE_AS_INTEGER, null, false);
		}

		@Override public boolean shouldDisplay(CMDataFile dataFile) {
			return dataFile instanceof CMTemplate || dataFile instanceof CMCharacter;
		}

		@Override public Object getData(CMSkill skill) {
			return new Integer(skill.canHaveChildren() ? -1 : skill.getPoints());
		}

		@Override public String getDataAsText(CMSkill skill) {
			if (skill.canHaveChildren()) {
				return ""; //$NON-NLS-1$
			}
			return TKNumberUtils.format(skill.getPoints());
		}
	},
	/** The page reference. */
	REFERENCE(Msgs.REFERENCE, Msgs.REFERENCE_TOOLTIP) {
		@Override public TKCell getCell() {
			return new CSTextCell(TKAlignment.RIGHT, TKTextCell.COMPARE_AS_TEXT, null, false);
		}

		@Override public boolean shouldDisplay(CMDataFile dataFile) {
			return true;
		}

		@Override public Object getData(CMSkill skill) {
			return getDataAsText(skill);
		}

		@Override public String getDataAsText(CMSkill skill) {
			return skill.getReference();
		}
	};

	private String	mTitle;
	private String	mToolTip;

	private CSSkillColumnID(String title, String tooltip) {
		mTitle = title;
		mToolTip = tooltip;
	}

	@Override public String toString() {
		return mTitle;
	}

	/**
	 * @param skill The {@link CMSkill} to get the data from.
	 * @return An object representing the data for this column.
	 */
	public abstract Object getData(CMSkill skill);

	/**
	 * @param skill The {@link CMSkill} to get the data from.
	 * @return Text representing the data for this column.
	 */
	public abstract String getDataAsText(CMSkill skill);

	/** @return The tooltip for the column. */
	public String getToolTip() {
		return mToolTip;
	}

	/** @return The {@link TKCell} used to display the data. */
	public abstract TKCell getCell();

	/**
	 * @param dataFile The {@link CMDataFile} to use.
	 * @return Whether this column should be displayed for the specified data file.
	 */
	public abstract boolean shouldDisplay(CMDataFile dataFile);

	/**
	 * Adds all relevant {@link TKColumn}s to a {@link TKOutline}.
	 * 
	 * @param outline The {@link TKOutline} to use.
	 * @param dataFile The {@link CMDataFile} that data is being displayed for.
	 */
	public static void addColumns(TKOutline outline, CMDataFile dataFile) {
		boolean sheetOrTemplate = dataFile instanceof CMCharacter || dataFile instanceof CMTemplate;
		TKOutlineModel model = outline.getModel();

		for (CSSkillColumnID one : values()) {
			if (one.shouldDisplay(dataFile)) {
				TKColumn column = new TKColumn(one.ordinal(), one.toString(), one.getToolTip(), one.getCell());

				column.setHeaderCell(new CSHeaderCell(sheetOrTemplate));
				model.addColumn(column);
			}
		}
	}
}
