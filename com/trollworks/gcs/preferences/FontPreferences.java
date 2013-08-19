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
 * 2005-2008 the Initial Developer. All Rights Reserved.
 *
 * Contributor(s):
 *
 * ***** END LICENSE BLOCK ***** */

package com.trollworks.gcs.preferences;

import com.trollworks.gcs.utility.Fonts;
import com.trollworks.gcs.utility.io.LocalizedMessages;
import com.trollworks.gcs.widgets.FontPanel;
import com.trollworks.gcs.widgets.GraphicsUtilities;
import com.trollworks.gcs.widgets.layout.Alignment;
import com.trollworks.gcs.widgets.layout.FlexComponent;
import com.trollworks.gcs.widgets.layout.FlexGrid;

import java.awt.Font;
import java.awt.event.ActionEvent;
import java.awt.event.ActionListener;

import javax.swing.UIManager;

/** The font preferences panel. */
public class FontPreferences extends PreferencePanel implements ActionListener {
	private static String			MSG_FONTS;
	private static String			MSG_LABELS_FONT;
	private static String			MSG_FIELDS_FONT;
	private static String			MSG_FIELD_NOTES_FONT;
	private static String			MSG_TECHNIQUE_FIELDS_FONT;
	private static String			MSG_PRIMARY_FOOTER_FONT;
	private static String			MSG_SECONDARY_FOOTER_FONT;
	private static String			MSG_NOTES_FONT;

	static {
		LocalizedMessages.initialize(FontPreferences.class);
	}

	private static final String[]	FONT_INFO	= new String[] { MSG_LABELS_FONT, Fonts.KEY_LABEL, MSG_FIELDS_FONT, Fonts.KEY_FIELD, MSG_FIELD_NOTES_FONT, Fonts.KEY_FIELD_NOTES, MSG_TECHNIQUE_FIELDS_FONT, Fonts.KEY_TECHNIQUE_FIELD, MSG_PRIMARY_FOOTER_FONT, Fonts.KEY_PRIMARY_FOOTER, MSG_SECONDARY_FOOTER_FONT, Fonts.KEY_SECONDARY_FOOTER, MSG_NOTES_FONT, Fonts.KEY_NOTES };
	private FontPanel[]				mFontPanels;
	private boolean					mIgnore;

	/**
	 * Creates a new {@link FontPreferences}.
	 * 
	 * @param owner The owning {@link PreferencesWindow}.
	 */
	public FontPreferences(PreferencesWindow owner) {
		super(MSG_FONTS, owner);
		FlexGrid grid = new FlexGrid();
		mFontPanels = new FontPanel[FONT_INFO.length / 2];
		for (int i = 0; i < FONT_INFO.length / 2; i++) {
			grid.add(new FlexComponent(createLabel(FONT_INFO[i * 2], null), Alignment.RIGHT_BOTTOM, Alignment.CENTER), i, 0);
			String key = FONT_INFO[i * 2 + 1];
			mFontPanels[i] = new FontPanel(UIManager.getFont(key));
			mFontPanels[i].setActionCommand(key);
			mFontPanels[i].addActionListener(this);
			add(mFontPanels[i]);
			grid.add(mFontPanels[i], i, 1);
		}
		grid.apply(this);
	}

	public void actionPerformed(ActionEvent event) {
		if (!mIgnore) {
			Object source = event.getSource();
			if (source instanceof FontPanel) {
				boolean adjusted = false;
				for (FontPanel panel : mFontPanels) {
					if (panel == source) {
						Font font = panel.getCurrentFont();
						if (!font.equals(UIManager.getFont(panel.getActionCommand()))) {
							UIManager.put(panel.getActionCommand(), font);
							adjusted = true;
						}
						break;
					}
				}
				if (adjusted) {
					GraphicsUtilities.forceRepaintAndInvalidate();
					Fonts.notifyOfFontChanges();
				}
			}

			adjustResetButton();
		}
	}

	@Override public void reset() {
		Fonts.restoreDefaults();
		mIgnore = true;
		for (int i = 0; i < mFontPanels.length; i++) {
			mFontPanels[i].setCurrentFont(UIManager.getFont(FONT_INFO[i * 2 + 1]));
		}
		mIgnore = false;
		GraphicsUtilities.forceRepaintAndInvalidate();
		Fonts.notifyOfFontChanges();
	}

	@Override public boolean isSetToDefaults() {
		return Fonts.isSetToDefaults();
	}
}
