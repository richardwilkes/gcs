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

package com.trollworks.gcs.ui.preferences;

import com.trollworks.gcs.ui.common.CSFont;
import com.trollworks.toolkit.io.TKPreferences;
import com.trollworks.toolkit.utility.TKAlignment;
import com.trollworks.toolkit.utility.TKColor;
import com.trollworks.toolkit.utility.TKFont;
import com.trollworks.toolkit.utility.TKGraphics;
import com.trollworks.toolkit.widget.TKFontPanel;
import com.trollworks.toolkit.widget.TKLabel;
import com.trollworks.toolkit.widget.TKPanel;
import com.trollworks.toolkit.widget.TKScrollablePanel;
import com.trollworks.toolkit.widget.border.TKEmptyBorder;
import com.trollworks.toolkit.widget.border.TKLineBorder;
import com.trollworks.toolkit.widget.button.TKCheckbox;
import com.trollworks.toolkit.widget.layout.TKColumnLayout;
import com.trollworks.toolkit.widget.layout.TKCompassLayout;
import com.trollworks.toolkit.widget.layout.TKCompassPosition;
import com.trollworks.toolkit.widget.scroll.TKScrollPanel;

import java.awt.Color;
import java.awt.Dimension;
import java.awt.Font;
import java.awt.Insets;
import java.awt.event.ActionEvent;
import java.awt.event.ActionListener;

/** The font preferences panel. */
public class CSFontPreferences extends CSPreferencePanel implements ActionListener {
	private static final String		MODULE						= "FontGraphics";																																																																																																																											//$NON-NLS-1$
	private static final String		FONT_ANTIALIASING_ENABLED	= "AntiAliasing";																																																																																																																											//$NON-NLS-1$
	private static final String[]	FONT_INFO					= new String[] { Msgs.CONTROLS_FONT, TKFont.CONTROL_FONT_KEY, Msgs.TEXT_FONT, TKFont.TEXT_FONT_KEY, Msgs.MENUS_FONT, TKFont.MENU_FONT_KEY, Msgs.MENU_KEYS_FONT, TKFont.MENU_KEY_FONT_KEY, Msgs.LABELS_FONT, CSFont.KEY_LABEL, Msgs.FIELDS_FONT, CSFont.KEY_FIELD, Msgs.FIELD_NOTES_FONT, CSFont.KEY_FIELD_NOTES, Msgs.TECHNIQUE_FIELDS_FONT, CSFont.KEY_TECHNIQUE_FIELD, Msgs.PRIMARY_FOOTER_FONT, CSFont.KEY_PRIMARY_FOOTER, Msgs.SECONDARY_FOOTER_FONT, CSFont.KEY_SECONDARY_FOOTER, Msgs.NOTES_FONT, CSFont.KEY_NOTES };
	private TKFontPanel[]			mFontPanels;
	private TKCheckbox				mAntiAliasingCheckBox;
	private boolean					mIgnore;

	/** Initializes the services controlled by these preferences. */
	public static void initialize() {
		TKGraphics.setAntiAliasFonts(TKPreferences.getInstance().getBooleanValue(MODULE, FONT_ANTIALIASING_ENABLED, true));
	}

	/** Creates the general preferences panel. */
	public CSFontPreferences() {
		super(Msgs.FONTS);
		setLayout(new TKCompassLayout(5, 5));
		mAntiAliasingCheckBox = new TKCheckbox(Msgs.ANTIALIAS_FONTS, TKGraphics.isAntiAliasFontsOn());
		mAntiAliasingCheckBox.addActionListener(this);
		add(mAntiAliasingCheckBox, TKCompassPosition.NORTH);
		TKScrollPanel scroller = new TKScrollPanel(createFontPanelWrapper());
		scroller.getContentBorderView().setBorder(new TKLineBorder(TKColor.SCROLL_BAR_LINE, 1, TKLineBorder.LEFT_EDGE | TKLineBorder.TOP_EDGE, false));
		add(scroller, TKCompassPosition.CENTER);
	}

	private TKPanel createFontPanelWrapper() {
		TKScrollablePanel wrapper = new TKScrollablePanel(new TKColumnLayout(3, 0, 0));
		Insets insets;
		Dimension size;

		wrapper.setOpaque(true);
		wrapper.setBackground(Color.white);
		wrapper.setBorder(new TKEmptyBorder(5));
		mFontPanels = new TKFontPanel[FONT_INFO.length / 2];
		for (int i = 0; i < FONT_INFO.length / 2; i++) {
			mFontPanels[i] = createFontPanel(wrapper, FONT_INFO[i * 2], FONT_INFO[i * 2 + 1]);
		}
		insets = wrapper.getInsets();
		size = wrapper.getPreferredSize();
		size.height = insets.top + insets.bottom + mFontPanels[0].getPreferredSize().height * 5 + TKColumnLayout.DEFAULT_V_GAP_SIZE * 4;
		wrapper.setPreferredViewportSize(size);
		return wrapper;
	}

	private TKFontPanel createFontPanel(TKPanel wrapper, String labelName, String fontNameKey) {
		TKLabel label = new TKLabel(labelName, TKAlignment.RIGHT);
		TKFontPanel panel = new TKFontPanel(TKFont.lookup(fontNameKey));

		panel.setActionCommand(fontNameKey);
		panel.addActionListener(this);
		wrapper.add(label);
		wrapper.add(panel);
		wrapper.add(new TKPanel());
		return panel;
	}

	public void actionPerformed(ActionEvent event) {
		if (!mIgnore) {
			Object source = event.getSource();

			if (source == mAntiAliasingCheckBox) {
				boolean enabled = mAntiAliasingCheckBox.isChecked();

				TKGraphics.setAntiAliasFonts(enabled);
				TKPreferences.getInstance().setValue(MODULE, FONT_ANTIALIASING_ENABLED, enabled);
			} else if (source instanceof TKFontPanel) {
				for (TKFontPanel panel : mFontPanels) {
					if (panel == source) {
						Font font = panel.getCurrentFont();

						if (!font.equals(TKFont.register(panel.getActionCommand(), font))) {
							TKGraphics.forceRepaintAndInvalidate();
						}
						break;
					}
				}
			}

			adjustResetButton();
		}
	}

	@Override public void reset() {
		TKFont.restoreDefaults();
		mIgnore = true;
		for (int i = 0; i < mFontPanels.length; i++) {
			mFontPanels[i].setCurrentFont(TKFont.lookup(FONT_INFO[i * 2 + 1]));
		}
		mAntiAliasingCheckBox.setCheckedState(true);
		TKPreferences.getInstance().setValue(MODULE, FONT_ANTIALIASING_ENABLED, true);
		mIgnore = false;
		TKGraphics.setAntiAliasFonts(true);
		TKGraphics.forceRepaintAndInvalidate();
	}

	@Override public boolean isSetToDefaults() {
		return TKGraphics.isAntiAliasFontsOn() && TKFont.isSetToDefaults();
	}
}
