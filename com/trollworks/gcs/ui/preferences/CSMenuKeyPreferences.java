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

import com.trollworks.gcs.ui.common.CSMenuKeys;
import com.trollworks.toolkit.utility.TKAlignment;
import com.trollworks.toolkit.utility.TKKeystroke;
import com.trollworks.toolkit.utility.TKColor;
import com.trollworks.toolkit.widget.TKLabel;
import com.trollworks.toolkit.widget.TKScrollablePanel;
import com.trollworks.toolkit.widget.TKPanel;
import com.trollworks.toolkit.widget.border.TKEmptyBorder;
import com.trollworks.toolkit.widget.border.TKLineBorder;
import com.trollworks.toolkit.widget.button.TKButton;
import com.trollworks.toolkit.widget.layout.TKColumnLayout;
import com.trollworks.toolkit.widget.layout.TKCompassLayout;
import com.trollworks.toolkit.widget.scroll.TKScrollPanel;

import java.awt.Color;
import java.awt.Dimension;
import java.awt.Insets;
import java.awt.event.ActionEvent;
import java.awt.event.ActionListener;
import java.util.ArrayList;

/** The menu key preferences panel. */
public class CSMenuKeyPreferences extends CSPreferencePanel implements ActionListener {
	private ArrayList<TKButton>	mButtons;

	/** Creates the general preferences panel. */
	public CSMenuKeyPreferences() {
		super(Msgs.MENU_KEYS);
		setLayout(new TKCompassLayout());
		TKScrollPanel scroller = new TKScrollPanel(createKeyPanelWrapper());
		scroller.getContentBorderView().setBorder(new TKLineBorder(TKColor.SCROLL_BAR_LINE, 1, TKLineBorder.LEFT_EDGE | TKLineBorder.TOP_EDGE, false));
		add(scroller);
	}

	private TKPanel createKeyPanelWrapper() {
		TKScrollablePanel wrapper = new TKScrollablePanel(new TKColumnLayout(3));
		int height = 0;
		Dimension size;
		Insets insets;

		wrapper.setOpaque(true);
		wrapper.setBackground(Color.white);
		wrapper.setBorder(new TKEmptyBorder(5));
		mButtons = new ArrayList<TKButton>();
		for (String cmd : CSMenuKeys.getCommands()) {
			TKLabel label = new TKLabel(CSMenuKeys.getTitle(cmd), TKAlignment.RIGHT);
			TKKeystroke keyStroke = CSMenuKeys.getKeyStroke(cmd);
			TKButton button = new TKButton(keyStroke == null ? Msgs.NOT_ASSIGNED : keyStroke.toString());
			int tmp;

			button.setActionCommand(cmd);
			button.addActionListener(this);
			mButtons.add(button);
			tmp = label.getPreferredSize().height;
			if (tmp > height) {
				height = tmp;
			}
			tmp = button.getPreferredSize().height;
			if (tmp > height) {
				height = tmp;
			}
			wrapper.add(label);
			wrapper.add(button);
			wrapper.add(new TKPanel());
		}
		insets = wrapper.getInsets();
		size = wrapper.getPreferredSize();
		size.height = insets.top + insets.bottom + height * 8 + TKColumnLayout.DEFAULT_V_GAP_SIZE * 7;
		wrapper.setPreferredViewportSize(size);
		return wrapper;
	}

	public void actionPerformed(ActionEvent event) {
		TKButton button = (TKButton) event.getSource();
		String cmd = event.getActionCommand();
		TKKeystroke keyStroke = CSKeystrokeDialog.getKeyStroke(cmd);
		TKKeystroke old = CSMenuKeys.getKeyStroke(cmd);

		if (keyStroke == null ? old != null : !keyStroke.equals(old)) {
			if (keyStroke != null) {
				String oldCmd = CSMenuKeys.getCommand(keyStroke);

				if (oldCmd != null) {
					CSMenuKeys.put(oldCmd, null);
					for (TKButton other : mButtons) {
						if (other.getActionCommand().equals(oldCmd)) {
							adjustButton(other, null);
							break;
						}
					}
				}
			}
			CSMenuKeys.put(cmd, keyStroke);
			adjustButton(button, keyStroke);
		}
		adjustResetButton();
	}

	private void adjustButton(TKButton button, TKKeystroke keyStroke) {
		button.setText(keyStroke == null ? Msgs.NOT_ASSIGNED : keyStroke.toString());
	}

	@Override public void reset() {
		CSMenuKeys.reset();
		for (TKButton button : mButtons) {
			adjustButton(button, CSMenuKeys.getKeyStroke(button.getActionCommand()));
		}
	}

	@Override public boolean isSetToDefaults() {
		return CSMenuKeys.isSetToDefaults();
	}
}
