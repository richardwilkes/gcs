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
 * 2005-2013 the Initial Developer. All Rights Reserved.
 *
 * Contributor(s):
 *
 * ***** END LICENSE BLOCK ***** */

package com.trollworks.gcs.character;

import com.trollworks.ttk.utility.LocalizedMessages;
import com.trollworks.ttk.utility.WindowSizeEnforcer;

import java.awt.BorderLayout;
import java.awt.Dimension;
import java.awt.FlowLayout;
import java.awt.event.ActionEvent;
import java.awt.event.ActionListener;
import java.awt.event.WindowEvent;
import java.awt.event.WindowFocusListener;

import javax.swing.JButton;
import javax.swing.JDialog;
import javax.swing.JOptionPane;
import javax.swing.JPanel;
import javax.swing.JScrollPane;
import javax.swing.JTextArea;
import javax.swing.ScrollPaneConstants;
import javax.swing.border.EmptyBorder;

/** Provides simplistic text editing. */
public class TextEditor extends JDialog implements ActionListener, WindowFocusListener {
	private static String	MSG_CANCEL;
	private static String	MSG_SET;
	private JTextArea		mEditor;
	private JButton			mSetButton;
	private boolean			mSet;

	static {
		LocalizedMessages.initialize(TextEditor.class);
	}

	/**
	 * Puts up a modal text editor.
	 * 
	 * @param title The title for the dialog.
	 * @param text The text to edit.
	 * @return The new text, or <code>null</code> if changes were canceled.
	 */
	public static String edit(String title, String text) {
		TextEditor editor = new TextEditor(title, text);
		editor.setVisible(true);
		return editor.mSet ? editor.mEditor.getText() : null;
	}

	private TextEditor(String title, String text) {
		super(JOptionPane.getRootFrame(), title, true);
		setResizable(true);
		setLayout(new BorderLayout());
		setDefaultCloseOperation(DISPOSE_ON_CLOSE);
		setLocationByPlatform(true);

		JPanel content = new JPanel(new BorderLayout());
		content.setBorder(new EmptyBorder(5, 5, 5, 5));
		add(content, BorderLayout.CENTER);

		mEditor = new JTextArea(text);
		mEditor.setLineWrap(true);
		mEditor.setWrapStyleWord(true);
		JScrollPane scroller = new JScrollPane(mEditor, ScrollPaneConstants.VERTICAL_SCROLLBAR_ALWAYS, ScrollPaneConstants.HORIZONTAL_SCROLLBAR_ALWAYS);
		scroller.setMinimumSize(new Dimension(400, 300));
		content.add(scroller, BorderLayout.CENTER);

		JPanel panel = new JPanel(new FlowLayout(FlowLayout.RIGHT));
		createButton(panel, MSG_CANCEL);
		mSetButton = createButton(panel, MSG_SET);
		content.add(panel, BorderLayout.SOUTH);

		WindowSizeEnforcer.monitor(this);
		addWindowFocusListener(this);

		pack();
	}

	private JButton createButton(JPanel parent, String title) {
		JButton button = new JButton(title);
		button.addActionListener(this);
		parent.add(button);
		return button;
	}

	@Override
	public void actionPerformed(ActionEvent event) {
		mSet = event.getSource() == mSetButton;
		dispose();
	}

	@Override
	public void windowGainedFocus(WindowEvent event) {
		mEditor.requestFocus();
		removeWindowFocusListener(this);
	}

	@Override
	public void windowLostFocus(WindowEvent event) {
		// Not used.
	}
}
