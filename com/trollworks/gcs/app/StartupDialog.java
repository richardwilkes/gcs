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

package com.trollworks.gcs.app;

import com.trollworks.gcs.menu.data.DataMenu;
import com.trollworks.gcs.menu.file.NewCharacterSheetCommand;
import com.trollworks.gcs.menu.file.NewCharacterTemplateCommand;
import com.trollworks.gcs.menu.file.NewLibraryCommand;
import com.trollworks.gcs.menu.file.OpenCommand;
import com.trollworks.gcs.menu.file.RecentFilesMenu;
import com.trollworks.gcs.utility.io.Images;
import com.trollworks.gcs.utility.io.LocalizedMessages;
import com.trollworks.gcs.utility.io.Path;
import com.trollworks.gcs.widgets.AppWindow;
import com.trollworks.gcs.widgets.UIUtilities;
import com.trollworks.gcs.widgets.WindowSizeEnforcer;
import com.trollworks.gcs.widgets.layout.Alignment;
import com.trollworks.gcs.widgets.layout.FlexColumn;
import com.trollworks.gcs.widgets.layout.FlexComponent;
import com.trollworks.gcs.widgets.layout.FlexRow;
import com.trollworks.gcs.widgets.layout.FlexSpacer;

import java.awt.Component;
import java.awt.Container;
import java.awt.Cursor;
import java.awt.Dimension;
import java.awt.Insets;
import java.awt.event.ActionEvent;
import java.awt.event.ActionListener;
import java.awt.event.KeyEvent;
import java.awt.event.KeyListener;
import java.awt.event.MouseEvent;
import java.awt.event.MouseListener;
import java.awt.event.WindowEvent;
import java.awt.event.WindowFocusListener;
import java.awt.image.BufferedImage;
import java.io.File;

import javax.swing.DefaultListCellRenderer;
import javax.swing.ImageIcon;
import javax.swing.JButton;
import javax.swing.JDialog;
import javax.swing.JList;
import javax.swing.JOptionPane;
import javax.swing.JPanel;
import javax.swing.JScrollPane;
import javax.swing.SwingConstants;
import javax.swing.border.TitledBorder;

/** The initial dialog user's see upon launch if no files have been specified. */
public class StartupDialog extends JDialog implements WindowFocusListener, ActionListener, MouseListener, KeyListener {
	private static String	MSG_TITLE;
	private static String	MSG_NEW;
	private static String	MSG_OPEN;
	private static String	MSG_NEW_CHARACTER_SHEET;
	private static String	MSG_NEW_LIBRARY;
	private static String	MSG_NEW_TEMPLATE;
	private static String	MSG_OTHER;
	private JButton			mSheetButton;
	private JButton			mLibraryButton;
	private JButton			mTemplateButton;
	private JButton			mOpenButton;
	private JList			mRecentFilesList;

	static {
		LocalizedMessages.initialize(StartupDialog.class);
	}

	/** Creates a new {@link StartupDialog}. */
	public StartupDialog() {
		super(JOptionPane.getRootFrame(), MSG_TITLE, true);

		Container content = getContentPane();

		JPanel newPanel = new JPanel();
		newPanel.setBorder(new TitledBorder(MSG_NEW));
		mSheetButton = createButton(newPanel, MSG_NEW_CHARACTER_SHEET, Images.getCharacterSheetIcon(true));
		mLibraryButton = createButton(newPanel, MSG_NEW_LIBRARY, Images.getLibraryIcon(true));
		mTemplateButton = createButton(newPanel, MSG_NEW_TEMPLATE, Images.getTemplateIcon(true));
		UIUtilities.adjustToSameSize(newPanel.getComponents());
		content.add(newPanel);

		FlexColumn column = new FlexColumn();
		column.add(mSheetButton);
		column.add(mTemplateButton);
		column.add(mLibraryButton);
		column.add(new FlexSpacer(0, 0, false, true));
		column.apply(newPanel);

		JPanel openPanel = new JPanel();
		openPanel.setBorder(new TitledBorder(MSG_OPEN));
		mRecentFilesList = new JList(RecentFilesMenu.getRecents().toArray());
		mRecentFilesList.setCellRenderer(new DefaultListCellRenderer() {
			@Override public Component getListCellRendererComponent(JList list, Object value, int index, boolean isSelected, boolean cellHasFocus) {
				super.getListCellRendererComponent(list, value, index, isSelected, cellHasFocus);
				setText(DataMenu.filterTitle(Path.getLeafName(((File) value).getName(), false)));
				setIcon(new ImageIcon(App.getIconForFile((File) value)));
				return this;
			}
		});
		mRecentFilesList.setSize(mRecentFilesList.getPreferredSize());
		mRecentFilesList.addMouseListener(this);
		mRecentFilesList.addKeyListener(this);
		JScrollPane scrollPane = new JScrollPane(mRecentFilesList);
		scrollPane.setMinimumSize(new Dimension(150, 75));
		openPanel.add(scrollPane);
		mOpenButton = new JButton(MSG_OTHER);
		mOpenButton.addActionListener(this);
		openPanel.add(mOpenButton);
		content.add(openPanel);

		column = new FlexColumn();
		column.add(scrollPane);
		column.add(new FlexComponent(mOpenButton, Alignment.CENTER, Alignment.CENTER));
		column.apply(openPanel);

		FlexRow row = new FlexRow();
		row.setInsets(new Insets(10, 10, 10, 10));
		row.setFill(true);
		row.add(newPanel);
		row.add(openPanel);
		row.apply(content);

		setResizable(true);
		setDefaultCloseOperation(DISPOSE_ON_CLOSE);
		new WindowSizeEnforcer(this);
		addWindowFocusListener(this);
		pack();
		setLocationRelativeTo(null);
	}

	private JButton createButton(Container panel, String title, BufferedImage image) {
		JButton button = new JButton(title, new ImageIcon(image));
		button.setHorizontalTextPosition(SwingConstants.CENTER);
		button.setVerticalTextPosition(SwingConstants.BOTTOM);
		button.setHorizontalAlignment(SwingConstants.CENTER);
		button.addActionListener(this);
		panel.add(button);
		return button;
	}

	public void windowGainedFocus(WindowEvent event) {
		mRecentFilesList.requestFocus();
		removeWindowFocusListener(this);
	}

	public void windowLostFocus(WindowEvent event) {
		// Not used.
	}

	public void actionPerformed(ActionEvent event) {
		Object obj = event.getSource();
		setCursor(Cursor.getPredefinedCursor(Cursor.WAIT_CURSOR));
		if (obj == mSheetButton) {
			NewCharacterSheetCommand.INSTANCE.newSheet();
		} else if (obj == mLibraryButton) {
			NewLibraryCommand.INSTANCE.newLibrary();
		} else if (obj == mTemplateButton) {
			NewCharacterTemplateCommand.INSTANCE.newTemplate();
		} else if (obj == mOpenButton) {
			OpenCommand.INSTANCE.open();
		}
		if (AppWindow.getAllWindows().isEmpty()) {
			setCursor(Cursor.getPredefinedCursor(Cursor.DEFAULT_CURSOR));
		} else {
			dispose();
		}
	}

	public void mouseClicked(MouseEvent event) {
		if (event.getClickCount() > 1 && mRecentFilesList.locationToIndex(event.getPoint()) != -1) {
			openSelectedRecents();
		}
	}

	private void openSelectedRecents() {
		setCursor(Cursor.getPredefinedCursor(Cursor.WAIT_CURSOR));
		for (Object obj : mRecentFilesList.getSelectedValues()) {
			OpenCommand.INSTANCE.open((File) obj);
		}
		if (AppWindow.getAllWindows().isEmpty()) {
			setCursor(Cursor.getPredefinedCursor(Cursor.DEFAULT_CURSOR));
		} else {
			dispose();
		}
	}

	public void mouseEntered(MouseEvent event) {
		// Not used.
	}

	public void mouseExited(MouseEvent event) {
		// Not used.
	}

	public void mousePressed(MouseEvent event) {
		// Not used.
	}

	public void mouseReleased(MouseEvent event) {
		// Not used.
	}

	public void keyPressed(KeyEvent event) {
		char key = event.getKeyChar();
		if (key == '\n' || key == '\r') {
			openSelectedRecents();
		}
	}

	public void keyReleased(KeyEvent event) {
		// Not used.
	}

	public void keyTyped(KeyEvent event) {
		// Not used.
	}
}
