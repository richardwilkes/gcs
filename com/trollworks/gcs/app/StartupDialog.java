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

package com.trollworks.gcs.app;

import com.trollworks.gcs.menu.data.DataMenu;
import com.trollworks.gcs.menu.file.NewCharacterSheetCommand;
import com.trollworks.gcs.menu.file.NewCharacterTemplateCommand;
import com.trollworks.gcs.menu.file.NewLibraryCommand;
import com.trollworks.ttk.layout.PrecisionLayout;
import com.trollworks.ttk.menu.file.FileType;
import com.trollworks.ttk.menu.file.OpenCommand;
import com.trollworks.ttk.menu.file.RecentFilesMenu;
import com.trollworks.ttk.utility.LocalizedMessages;
import com.trollworks.ttk.utility.Path;
import com.trollworks.ttk.utility.UIUtilities;
import com.trollworks.ttk.utility.WindowSizeEnforcer;
import com.trollworks.ttk.widgets.AppWindow;

import java.awt.Component;
import java.awt.Container;
import java.awt.Cursor;
import java.awt.Dimension;
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
import javax.swing.event.ListSelectionEvent;
import javax.swing.event.ListSelectionListener;

/** The initial dialog user's see upon launch if no files have been specified. */
public class StartupDialog extends JDialog implements WindowFocusListener, ActionListener, MouseListener, KeyListener, ListSelectionListener {
	private static String	MSG_TITLE;
	private static String	MSG_CREATE;
	private static String	MSG_RECENT_FILES;
	private static String	MSG_NEW_CHARACTER_SHEET;
	private static String	MSG_NEW_LIBRARY;
	private static String	MSG_NEW_TEMPLATE;
	private static String	MSG_CHOOSE_OTHER;
	private static String	MSG_OPEN;
	private JButton			mSheetButton;
	private JButton			mLibraryButton;
	private JButton			mTemplateButton;
	private JButton			mChooseOtherButton;
	private JButton			mOpenButton;
	private JList<File>		mRecentFilesList;

	static {
		LocalizedMessages.initialize(StartupDialog.class);
	}

	/** Creates a new {@link StartupDialog}. */
	public StartupDialog() {
		super(JOptionPane.getRootFrame(), MSG_TITLE, true);
		Container content = getContentPane();
		content.setLayout(new PrecisionLayout("margins:10 columns:2")); //$NON-NLS-1$
		content.add(createCreatePanel(), "vGrab:yes vAlign:fill hAlign:fill"); //$NON-NLS-1$
		content.add(createRecentFilesPanel(), "vGrab:yes hGrab:yes vAlign:fill hAlign:fill"); //$NON-NLS-1$
		setResizable(true);
		setDefaultCloseOperation(DISPOSE_ON_CLOSE);
		WindowSizeEnforcer.monitor(this);
		addWindowFocusListener(this);
		pack();
		setLocationRelativeTo(null);
	}

	private JPanel createCreatePanel() {
		JPanel panel = new JPanel(new PrecisionLayout());
		panel.setBorder(new TitledBorder(MSG_CREATE));
		mSheetButton = createButton(panel, MSG_NEW_CHARACTER_SHEET, GCSImages.getCharacterSheetIcon(true));
		mLibraryButton = createButton(panel, MSG_NEW_LIBRARY, GCSImages.getLibraryIcon(true));
		mTemplateButton = createButton(panel, MSG_NEW_TEMPLATE, GCSImages.getTemplateIcon(true));
		UIUtilities.adjustToSameSize(panel.getComponents());
		return panel;
	}

	private JPanel createRecentFilesPanel() {
		JPanel panel = new JPanel(new PrecisionLayout("columns:2 equalColumns:yes")); //$NON-NLS-1$
		panel.setBorder(new TitledBorder(MSG_RECENT_FILES));
		mRecentFilesList = new JList<>(RecentFilesMenu.getRecents().toArray(new File[0]));
		mRecentFilesList.setCellRenderer(new DefaultListCellRenderer() {
			@Override
			public Component getListCellRendererComponent(JList<?> list, Object value, int index, boolean isSelected, boolean cellHasFocus) {
				super.getListCellRendererComponent(list, value, index, isSelected, cellHasFocus);
				setText(DataMenu.filterTitle(Path.getLeafName(((File) value).getName(), false)));
				setIcon(new ImageIcon(FileType.getIconForFile((File) value)));
				return this;
			}
		});
		mRecentFilesList.setSize(mRecentFilesList.getPreferredSize());
		mRecentFilesList.addMouseListener(this);
		mRecentFilesList.addKeyListener(this);
		JScrollPane scrollPane = new JScrollPane(mRecentFilesList);
		scrollPane.setMinimumSize(new Dimension(150, 75));
		panel.add(scrollPane, "hSpan:2 vGrab:yes hGrab:yes vAlign:fill hAlign:fill"); //$NON-NLS-1$
		mChooseOtherButton = new JButton(MSG_CHOOSE_OTHER);
		mChooseOtherButton.addActionListener(this);
		panel.add(mChooseOtherButton, "hAlign:middle"); //$NON-NLS-1$
		mOpenButton = new JButton(MSG_OPEN);
		mOpenButton.setEnabled(false);
		mOpenButton.addActionListener(this);
		mOpenButton.setDefaultCapable(true);
		getRootPane().setDefaultButton(mOpenButton);
		mRecentFilesList.addListSelectionListener(this);
		panel.add(mOpenButton, "hAlign:middle"); //$NON-NLS-1$
		UIUtilities.adjustToSameSize(mChooseOtherButton, mOpenButton);
		return panel;
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

	@Override
	public void windowGainedFocus(WindowEvent event) {
		mRecentFilesList.requestFocus();
		removeWindowFocusListener(this);
	}

	@Override
	public void windowLostFocus(WindowEvent event) {
		// Not used.
	}

	@Override
	public void actionPerformed(ActionEvent event) {
		Object obj = event.getSource();
		setCursor(Cursor.getPredefinedCursor(Cursor.WAIT_CURSOR));
		if (obj == mSheetButton) {
			NewCharacterSheetCommand.newSheet();
		} else if (obj == mLibraryButton) {
			NewLibraryCommand.newLibrary();
		} else if (obj == mTemplateButton) {
			NewCharacterTemplateCommand.newTemplate();
		} else if (obj == mChooseOtherButton) {
			OpenCommand.open();
		} else if (obj == mOpenButton) {
			openSelectedRecents(false);
		}
		if (AppWindow.getAllWindows().isEmpty()) {
			setCursor(Cursor.getPredefinedCursor(Cursor.DEFAULT_CURSOR));
		} else {
			dispose();
		}
	}

	@Override
	public void mouseClicked(MouseEvent event) {
		if (event.getClickCount() > 1 && mRecentFilesList.locationToIndex(event.getPoint()) != -1) {
			openSelectedRecents(true);
		}
	}

	private void openSelectedRecents(boolean performPostAmble) {
		setCursor(Cursor.getPredefinedCursor(Cursor.WAIT_CURSOR));
		for (File obj : mRecentFilesList.getSelectedValuesList()) {
			OpenCommand.open(obj);
		}
		if (performPostAmble) {
			if (AppWindow.getAllWindows().isEmpty()) {
				setCursor(Cursor.getPredefinedCursor(Cursor.DEFAULT_CURSOR));
			} else {
				dispose();
			}
		}
	}

	@Override
	public void mouseEntered(MouseEvent event) {
		// Not used.
	}

	@Override
	public void mouseExited(MouseEvent event) {
		// Not used.
	}

	@Override
	public void mousePressed(MouseEvent event) {
		// Not used.
	}

	@Override
	public void mouseReleased(MouseEvent event) {
		// Not used.
	}

	@Override
	public void keyPressed(KeyEvent event) {
		char key = event.getKeyChar();
		if (key == '\n' || key == '\r') {
			openSelectedRecents(true);
		}
	}

	@Override
	public void keyReleased(KeyEvent event) {
		// Not used.
	}

	@Override
	public void keyTyped(KeyEvent event) {
		// Not used.
	}

	@Override
	public void valueChanged(ListSelectionEvent event) {
		mOpenButton.setEnabled(!mRecentFilesList.isSelectionEmpty());
	}
}
