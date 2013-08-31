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
 * Portions created by the Initial Developer are Copyright (C) 1998-2013 the
 * Initial Developer. All Rights Reserved.
 *
 * Contributor(s):
 *
 * ***** END LICENSE BLOCK ***** */

package com.trollworks.gcs.character;

import static com.trollworks.gcs.character.AttributesPanel_LS.*;

import com.trollworks.ttk.annotation.LS;
import com.trollworks.ttk.annotation.Localized;
import com.trollworks.ttk.layout.Alignment;
import com.trollworks.ttk.layout.FlexComponent;
import com.trollworks.ttk.layout.FlexGrid;
import com.trollworks.ttk.layout.FlexRow;
import com.trollworks.ttk.layout.FlexSpacer;
import com.trollworks.ttk.utility.UIUtilities;
import com.trollworks.ttk.widgets.Wrapper;

import java.awt.Color;
import java.awt.Dimension;

import javax.swing.SwingConstants;

@Localized({
				@LS(key = "ATTRIBUTES", msg = "Attributes"),
				@LS(key = "ST", msg = "Strength (ST):"),
				@LS(key = "ST_TOOLTIP", msg = "<html><body><b>{0} points</b> have been spent to modify <b>Strength</b></body></html>"),
				@LS(key = "DX", msg = "Dexterity (DX):"),
				@LS(key = "DX_TOOLTIP", msg = "<html><body><b>{0} points</b> have been spent to modify <b>Dexterity</b></body></html>"),
				@LS(key = "IQ", msg = "Intelligence (IQ):"),
				@LS(key = "IQ_TOOLTIP", msg = "<html><body><b>{0} points</b> have been spent to modify <b>Intelligence</b></body></html>"),
				@LS(key = "HT", msg = "Health (HT):"),
				@LS(key = "HT_TOOLTIP", msg = "<html><body><b>{0} points</b> have been spent to modify <b>Health</b></body></html>"),
				@LS(key = "WILL", msg = "Will:"),
				@LS(key = "WILL_TOOLTIP", msg = "<html><body><b>{0} points</b> have been spent to modify <b>Will</b></body></html>"),
				@LS(key = "FRIGHT_CHECK", msg = "Fright Check:"),
				@LS(key = "PERCEPTION", msg = "Perception:"),
				@LS(key = "PERCEPTION_TOOLTIP", msg = "<html><body><b>{0} points</b> have been spent to modify <b>Perception</b></body></html>"),
				@LS(key = "VISION", msg = "Vision:"),
				@LS(key = "HEARING", msg = "Hearing:"),
				@LS(key = "TOUCH", msg = "Touch:"),
				@LS(key = "TASTE_SMELL", msg = "Taste & Smell:"),
				@LS(key = "BASIC_SPEED", msg = "Basic Speed:"),
				@LS(key = "BASIC_SPEED_TOOLTIP", msg = "<html><body><b>{0} points</b> have been spent to modify <b>Basic Speed</b></body></html>"),
				@LS(key = "BASIC_MOVE", msg = "Basic Move:"),
				@LS(key = "BASIC_MOVE_TOOLTIP", msg = "<html><body><b>{0} points</b> have been spent to modify <b>Basic Move</b></body></html>"),
				@LS(key = "BASIC_THRUST", msg = "thr:"),
				@LS(key = "BASIC_THRUST_TOOLTIP", msg = "The basic damage value for thrust attacks"),
				@LS(key = "BASIC_SWING", msg = "sw:"),
				@LS(key = "BASIC_SWING_TOOLTIP", msg = "The basic damage value for swing attacks"),
})
/** The character attributes panel. */
public class AttributesPanel extends DropPanel {
	/**
	 * Creates a new attributes panel.
	 * 
	 * @param character The character to display the data for.
	 */
	public AttributesPanel(GURPSCharacter character) {
		super(null, ATTRIBUTES, true);
		FlexGrid grid = new FlexGrid();
		grid.setVerticalGap(0);
		int row = 0;
		createLabelAndField(grid, row++, character, GURPSCharacter.ID_STRENGTH, ST, ST_TOOLTIP, SwingConstants.RIGHT, true);
		createLabelAndField(grid, row++, character, GURPSCharacter.ID_DEXTERITY, DX, DX_TOOLTIP, SwingConstants.RIGHT, true);
		createLabelAndField(grid, row++, character, GURPSCharacter.ID_INTELLIGENCE, IQ, IQ_TOOLTIP, SwingConstants.RIGHT, true);
		createLabelAndField(grid, row++, character, GURPSCharacter.ID_HEALTH, HT, HT_TOOLTIP, SwingConstants.RIGHT, true);
		createDivider(grid, row++, false);
		createDivider(grid, row++, true);
		createLabelAndField(grid, row++, character, GURPSCharacter.ID_WILL, WILL, WILL_TOOLTIP, SwingConstants.RIGHT, true);
		createLabelAndField(grid, row++, character, GURPSCharacter.ID_FRIGHT_CHECK, FRIGHT_CHECK, null, SwingConstants.RIGHT, false);
		createDivider(grid, row++, false);
		createDivider(grid, row++, true);
		createLabelAndField(grid, row++, character, GURPSCharacter.ID_BASIC_SPEED, BASIC_SPEED, BASIC_SPEED_TOOLTIP, SwingConstants.RIGHT, true);
		createLabelAndField(grid, row++, character, GURPSCharacter.ID_BASIC_MOVE, BASIC_MOVE, BASIC_MOVE_TOOLTIP, SwingConstants.RIGHT, true);
		createDivider(grid, row++, false);
		createDivider(grid, row++, true);
		createLabelAndField(grid, row++, character, GURPSCharacter.ID_PERCEPTION, PERCEPTION, PERCEPTION_TOOLTIP, SwingConstants.RIGHT, true);
		createLabelAndField(grid, row++, character, GURPSCharacter.ID_VISION, VISION, null, SwingConstants.RIGHT, false);
		createLabelAndField(grid, row++, character, GURPSCharacter.ID_HEARING, HEARING, null, SwingConstants.RIGHT, false);
		createLabelAndField(grid, row++, character, GURPSCharacter.ID_TASTE_AND_SMELL, TASTE_SMELL, null, SwingConstants.RIGHT, false);
		createLabelAndField(grid, row++, character, GURPSCharacter.ID_TOUCH, TOUCH, null, SwingConstants.RIGHT, false);
		createDivider(grid, row++, false);
		createDivider(grid, row++, true);
		createDamageFields(grid, row++, character);
		grid.apply(this);
	}

	private void createLabelAndField(FlexGrid grid, int row, GURPSCharacter character, String key, String title, String tooltip, int alignment, boolean enabled) {
		PageField field = new PageField(character, key, alignment, enabled, tooltip);
		PageLabel label = new PageLabel(title, field);
		add(label);
		add(field);
		grid.add(new FlexComponent(label, Alignment.RIGHT_BOTTOM, Alignment.CENTER), row, 0);
		grid.add(field, row, 1);
	}

	private void createDamageFields(FlexGrid grid, int rowIndex, GURPSCharacter character) {
		FlexRow row = new FlexRow();
		row.setHorizontalAlignment(Alignment.CENTER);
		createDamageLabelAndField(row, character, GURPSCharacter.ID_BASIC_THRUST, BASIC_THRUST, BASIC_THRUST_TOOLTIP);
		row.add(new FlexSpacer(0, 0, false, false));
		createDamageLabelAndField(row, character, GURPSCharacter.ID_BASIC_SWING, BASIC_SWING, BASIC_SWING_TOOLTIP);
		grid.add(row, rowIndex, 0, 1, 2);
	}

	private void createDamageLabelAndField(FlexRow row, GURPSCharacter character, String key, String title, String tooltip) {
		PageField field = new PageField(character, key, SwingConstants.RIGHT, false, tooltip);
		PageLabel label = new PageLabel(title, field);
		add(label);
		add(field);
		row.add(new FlexComponent(label, true));
		row.add(new FlexComponent(field, true));
	}

	private void createDivider(FlexGrid grid, int row, boolean black) {
		Wrapper panel = new Wrapper();
		UIUtilities.setOnlySize(panel, new Dimension(1, 1));
		add(panel);
		if (black) {
			addHorizontalBackground(panel, Color.black);
		}
		grid.add(panel, row, 0, 1, 2);
	}
}
