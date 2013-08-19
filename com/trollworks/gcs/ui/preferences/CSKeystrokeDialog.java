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
import com.trollworks.toolkit.utility.TKFont;
import com.trollworks.toolkit.widget.TKLabel;
import com.trollworks.toolkit.widget.TKPanel;
import com.trollworks.toolkit.widget.TKTextField;
import com.trollworks.toolkit.widget.border.TKEmptyBorder;
import com.trollworks.toolkit.widget.layout.TKColumnLayout;
import com.trollworks.toolkit.window.TKDialog;

import java.awt.Color;
import java.awt.event.KeyEvent;
import java.awt.event.KeyListener;
import java.text.MessageFormat;

/** Provides a simple editor for changing a key stroke. */
public class CSKeystrokeDialog extends TKDialog implements KeyListener {
	private String		mCmd;
	private TKKeystroke	mKeyStroke;
	private TKLabel		mAssignment;
	private TKLabel		mAlreadyAssigned;
	private TKLabel		mNoModifier;

	/**
	 * @param cmd The command to define a {@link TKKeystroke} for.
	 * @return The {@link TKKeystroke} to use.
	 */
	public static TKKeystroke getKeyStroke(String cmd) {
		CSKeystrokeDialog dialog = new CSKeystrokeDialog(cmd);

		dialog.doModal();
		return dialog.mKeyStroke;
	}

	private CSKeystrokeDialog(String cmd) {
		super(null, true);
		mCmd = cmd;
		mKeyStroke = CSMenuKeys.getKeyStroke(cmd);
		setContent(createContent());
	}

	private TKPanel createContent() {
		TKPanel content = new TKPanel(new TKColumnLayout(1, 0, 0));
		TKLabel prompt = new TKLabel(Msgs.PROMPT, TKAlignment.CENTER, true);

		mAlreadyAssigned = new TKLabel(Msgs.ALREADY_ASSIGNED, TKAlignment.CENTER, true);
		mNoModifier = new TKLabel(MessageFormat.format(Msgs.NO_MODIFIER, TKKeystroke.getCommandName()), TKAlignment.CENTER, true);
		mAssignment = new TKLabel(TKAlignment.CENTER);
		mAssignment.setBorder(TKTextField.BORDER);
		mAssignment.setOpaque(true);
		mAssignment.setBackground(Color.WHITE);
		mAssignment.setFontKey(TKFont.CONTROL_FONT_KEY);
		mAssignment.addKeyListener(this);
		mAssignment.setFocusable(true);
		updateAssignment();
		content.setBorder(new TKEmptyBorder(10));
		content.setOpaque(true);
		content.setBackground(TKColor.LIGHT_BACKGROUND);
		content.add(mAssignment);
		content.add(prompt);
		content.add(mAlreadyAssigned);
		content.add(mNoModifier);
		return content;
	}

	private void updateAssignment() {
		String existingCmd = CSMenuKeys.getCommand(mKeyStroke);

		mAssignment.setText(mKeyStroke == null ? Msgs.NOT_ASSIGNED : mKeyStroke.toString());

		if (existingCmd != null && !mCmd.equals(existingCmd)) {
			mAlreadyAssigned.setForeground(Color.RED);
		} else {
			mAlreadyAssigned.setForeground(TKColor.LIGHT_BACKGROUND);
		}

		if (mKeyStroke != null && (mKeyStroke.getModifiers() & TKKeystroke.getCommandMask()) == 0) {
			mNoModifier.setForeground(Color.RED);
		} else {
			mNoModifier.setForeground(TKColor.LIGHT_BACKGROUND);
		}
	}

	public void keyTyped(KeyEvent event) {
		// Not used
	}

	public void keyPressed(KeyEvent event) {
		int keyCode = event.getKeyCode();
		int modifiers = event.getModifiers();

		if (keyCode == KeyEvent.VK_ENTER) {
			attemptClose();
		} else if (keyCode == KeyEvent.VK_ESCAPE) {
			mKeyStroke = null;
			updateAssignment();
		} else if (keyCode != KeyEvent.VK_CONTROL && keyCode != KeyEvent.VK_META && keyCode != KeyEvent.VK_ALT && keyCode != KeyEvent.VK_SHIFT) {
			mKeyStroke = new TKKeystroke(keyCode, modifiers);
			updateAssignment();
		}
	}

	public void keyReleased(KeyEvent event) {
		// Not used
	}
}
