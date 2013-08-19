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

import com.trollworks.gcs.model.advantage.CMAdvantage;
import com.trollworks.gcs.model.modifier.CMModifier;
import com.trollworks.toolkit.text.TKTextUtility;
import com.trollworks.toolkit.utility.TKAlignment;
import com.trollworks.toolkit.utility.TKColor;
import com.trollworks.toolkit.utility.TKFont;
import com.trollworks.toolkit.widget.TKLabel;
import com.trollworks.toolkit.widget.TKPanel;
import com.trollworks.toolkit.widget.border.TKCompoundBorder;
import com.trollworks.toolkit.widget.border.TKEmptyBorder;
import com.trollworks.toolkit.widget.border.TKLineBorder;
import com.trollworks.toolkit.widget.button.TKCheckbox;
import com.trollworks.toolkit.widget.layout.TKColumnLayout;
import com.trollworks.toolkit.widget.layout.TKCompassLayout;
import com.trollworks.toolkit.widget.layout.TKCompassPosition;
import com.trollworks.toolkit.window.TKBaseWindow;
import com.trollworks.toolkit.window.TKDialog;
import com.trollworks.toolkit.window.TKOptionDialog;

import java.text.MessageFormat;
import java.util.ArrayList;
import java.util.Arrays;

/** Asks the user to enable/disable modifiers. */
public class CSModifierEnabler extends TKPanel {
	private CMAdvantage		mAdvantage;
	private TKCheckbox[]	mEnabled;
	private CMModifier[]	mModifiers;

	/**
	 * Brings up a modal dialog that allows {@link CMModifier}s to be enabled or disabled for the
	 * specified {@Link CMAdvantage}s.
	 * 
	 * @param window The window to open the dialog over.
	 * @param advantages The {@Link CMAdvantage}s to process.
	 * @return Whether anything was modified.
	 */
	static public boolean process(TKBaseWindow window, ArrayList<CMAdvantage> advantages) {
		ArrayList<CMAdvantage> list = new ArrayList<CMAdvantage>();
		boolean modified = false;
		int count;

		for (CMAdvantage advantage : advantages) {
			if (!advantage.getModifiers().isEmpty()) {
				list.add(advantage);
			}
		}

		count = list.size();
		for (int i = 0; i < count; i++) {
			CMAdvantage advantage = list.get(i);
			boolean hasMore = i != count - 1;
			TKOptionDialog dialog = TKOptionDialog.create(window, Msgs.MODIFIER_TITLE, hasMore ? TKOptionDialog.TYPE_YES_NO_CANCEL : TKOptionDialog.TYPE_YES_NO);
			CSModifierEnabler panel = new CSModifierEnabler(advantage, count - i - 1);

			if (hasMore) {
				dialog.setCancelButtonTitle(Msgs.CANCEL_REST);
			}
			dialog.setYesButtonTitle(Msgs.APPLY);
			dialog.setNoButtonTitle(Msgs.CANCEL);
			dialog.setResizable(true);
			switch (dialog.doModal(advantage.getImage(true), panel)) {
				case TKOptionDialog.YES:
					panel.applyChanges();
					modified = true;
					break;
				case TKOptionDialog.NO:
					break;
				case TKDialog.CANCEL:
					return modified;
			}
		}
		return modified;
	}

	private CSModifierEnabler(CMAdvantage advantage, int remaining) {
		super(new TKCompassLayout());
		mAdvantage = advantage;
		add(createTop(advantage, remaining), TKCompassPosition.NORTH);
		add(createCenter(), TKCompassPosition.CENTER);
	}

	private TKPanel createTop(CMAdvantage advantage, int remaining) {
		TKPanel top = new TKPanel(new TKColumnLayout());
		TKLabel label = new TKLabel(TKTextUtility.truncateIfNecessary(advantage.toString(), 80, TKAlignment.RIGHT), TKFont.CONTROL_FONT_KEY, TKAlignment.LEFT, true);

		top.setBorder(new TKEmptyBorder(0, 0, 15, 0));
		if (remaining > 0) {
			String msg;

			if (remaining == 1) {
				msg = Msgs.MODIFIER_ONE_REMAINING;
			} else {
				msg = MessageFormat.format(Msgs.MODIFIER_REMAINING, new Integer(remaining));
			}
			top.add(new TKLabel(msg, TKFont.CONTROL_FONT_KEY, TKAlignment.CENTER));
		}
		label.setBorder(new TKCompoundBorder(TKLineBorder.getSharedBorder(true), new TKEmptyBorder(0, 2, 0, 2)));
		label.setBackground(TKColor.CONTROL_ROLL);
		label.setOpaque(true);
		top.add(new TKPanel());
		top.add(label);
		return top;
	}

	private TKPanel createCenter() {
		TKPanel wrapper = new TKPanel(new TKColumnLayout());

		mModifiers = mAdvantage.getModifiers().toArray(new CMModifier[0]);
		Arrays.sort(mModifiers);

		mEnabled = new TKCheckbox[mModifiers.length];
		for (int i = 0; i < mModifiers.length; i++) {
			mEnabled[i] = new TKCheckbox(mModifiers[i].getFullDescription(), mModifiers[i].isEnabled());
			wrapper.add(mEnabled[i]);
		}
		return wrapper;
	}

	private void applyChanges() {
		for (int i = 0; i < mModifiers.length; i++) {
			mModifiers[i].setEnabled(mEnabled[i].isChecked());
		}
	}
}
