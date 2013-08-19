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

package com.trollworks.toolkit.window;

import com.trollworks.toolkit.io.TKFileFilter;
import com.trollworks.toolkit.io.TKImage;
import com.trollworks.toolkit.io.TKPath;
import com.trollworks.toolkit.io.TKPreferences;
import com.trollworks.toolkit.text.TKDocument;
import com.trollworks.toolkit.text.TKDocumentListener;
import com.trollworks.toolkit.utility.TKAlignment;
import com.trollworks.toolkit.utility.TKFont;
import com.trollworks.toolkit.utility.TKGraphics;
import com.trollworks.toolkit.widget.TKDefaultItemRenderer;
import com.trollworks.toolkit.widget.TKDivider;
import com.trollworks.toolkit.widget.TKItemList;
import com.trollworks.toolkit.widget.TKKeyEventFilter;
import com.trollworks.toolkit.widget.TKLabel;
import com.trollworks.toolkit.widget.TKPanel;
import com.trollworks.toolkit.widget.TKPopupMenu;
import com.trollworks.toolkit.widget.TKTextField;
import com.trollworks.toolkit.widget.border.TKEmptyBorder;
import com.trollworks.toolkit.widget.border.TKLineBorder;
import com.trollworks.toolkit.widget.button.TKButton;
import com.trollworks.toolkit.widget.layout.TKColumnLayout;
import com.trollworks.toolkit.widget.layout.TKCompassPosition;
import com.trollworks.toolkit.widget.menu.TKMenu;
import com.trollworks.toolkit.widget.menu.TKMenuItem;
import com.trollworks.toolkit.widget.scroll.TKScrollPanel;

import java.awt.Color;
import java.awt.Dialog;
import java.awt.Dimension;
import java.awt.Frame;
import java.awt.event.ActionEvent;
import java.awt.event.ActionListener;
import java.awt.event.KeyEvent;
import java.awt.image.BufferedImage;
import java.io.File;
import java.text.MessageFormat;
import java.util.ArrayList;
import java.util.Arrays;
import java.util.Comparator;
import java.util.HashMap;
import java.util.HashSet;
import java.util.StringTokenizer;

/** Provides a file dialog for opening or saving files. */
public class TKFileDialog extends TKDialog implements ActionListener, Comparator<File>, TKDocumentListener, TKKeyEventFilter {
	private static final String						SEMICOLON					= ";";						//$NON-NLS-1$
	private static final String						EMPTY						= "";						//$NON-NLS-1$
	private static final String						DOT							= ".";						//$NON-NLS-1$
	/** The button state modifier action command. */
	protected static final String					BSM_SELECTION_CHANGED		= "BSMSelChanged";			//$NON-NLS-1$
	private static final int						CURRENT_VERSION				= 1;
	private static final String						MODULE						= "FileDialog";			//$NON-NLS-1$
	private static final String						LAST_LOCATION_KEY			= "LastLocation";			//$NON-NLS-1$
	private static final String						FAVORITES_KEY				= "Favorites";				//$NON-NLS-1$
	private static final String						RECENT_FOLDERS_KEY			= "RecentFolders";			//$NON-NLS-1$
	private static final int						MAX_RECENT_FOLDERS			= 5;
	private static final int						INDENT_AMOUNT				= 4;
	private static final String						ADD_TO_FAVORITES			= "AddToFavorites";		//$NON-NLS-1$
	private static final String						REMOVE_FROM_FAVORITES		= "RemoveFromFavorites";	//$NON-NLS-1$
	private static final String						OPEN_SAVE					= "OpenSave";				//$NON-NLS-1$
	private static final String						CANCEL_CMD					= "Cancel";				//$NON-NLS-1$
	private static final String						UP							= "Up";					//$NON-NLS-1$
	private static final String						HOME						= "Home";					//$NON-NLS-1$
	private static final String						NEW_FOLDER					= "NewFolder";				//$NON-NLS-1$
	private static final String						HIERARCHY					= "Hierarchy";				//$NON-NLS-1$
	private static final String						FILTER						= "Filter";				//$NON-NLS-1$
	private static String							LAST_LOCATION_KEY_PREFIX	= null;
	private static ArrayList<File>					FAVORITES					= null;
	private static ArrayList<File>					RECENT_FOLDERS				= null;
	private static HashMap<String, BufferedImage>	FILE_ICON_MAP				= null;
	private boolean									mOpenDialog;
	private boolean									mFolderSelectionAllowed;
	private TKItemList<File>						mList;
	private TKTextField								mFileNameField;
	private TKLabel									mFileNameLabel;
	private TKPopupMenu								mHierarchyPopup;
	private TKMenu									mHierarchyMenu;
	private TKPopupMenu								mFilterPopup;
	private TKMenu									mFilterMenu;
	private TKButton								mNewFolderButton;
	private TKButton								mUpButton;
	private TKButton								mHomeButton;
	private File									mLocation;
	private boolean									mTransitioning;
	private TKPanel									mAccessoryPanel;

	/**
	 * Creates a new file dialog.
	 * 
	 * @param openFile Pass in <code>true</code> for an open file dialog, <code>false</code> for
	 *            a save file dialog.
	 */
	public TKFileDialog(boolean openFile) {
		this(openFile, true);
	}

	/**
	 * Creates a new file dialog.
	 * 
	 * @param openFile Pass in <code>true</code> for an open file dialog, <code>false</code> for
	 *            a save file dialog.
	 * @param addAllFilesFilter Pass in <code>true</code> to automatically add the "All Files"
	 *            filter. Ignored if <code>openFile</code> is <code>false</code>.
	 */
	public TKFileDialog(boolean openFile, boolean addAllFilesFilter) {
		this(TKGraphics.getAnyFrame(), openFile, addAllFilesFilter);
	}

	/**
	 * Creates a new file dialog.
	 * 
	 * @param openFile Pass in <code>true</code> for an open file dialog, <code>false</code> for
	 *            a save file dialog.
	 * @param addAllFilesFilter Pass in <code>true</code> to automatically add the "All Files"
	 *            filter. Ignored if <code>openFile</code> is <code>false</code>.
	 * @param message A message displayed to the user at the top of the dialog.
	 */
	public TKFileDialog(boolean openFile, boolean addAllFilesFilter, String message) {
		this(TKGraphics.getAnyFrame(), openFile, addAllFilesFilter, message);
	}

	/**
	 * Creates a new file dialog.
	 * 
	 * @param owner The owning window.
	 * @param openFile Pass in <code>true</code> for an open file dialog, <code>false</code> for
	 *            a save file dialog.
	 */
	public TKFileDialog(Frame owner, boolean openFile) {
		this(owner, openFile, true);
	}

	/**
	 * Creates a new file dialog.
	 * 
	 * @param owner The owning window.
	 * @param openFile Pass in <code>true</code> for an open file dialog, <code>false</code> for
	 *            a save file dialog.
	 * @param addAllFilesFilter Pass in <code>true</code> to automatically add the "All Files"
	 *            filter. Ignored if <code>openFile</code> is <code>false</code>.
	 */
	public TKFileDialog(Frame owner, boolean openFile, boolean addAllFilesFilter) {
		this(owner, openFile, addAllFilesFilter, null);
	}

	/**
	 * Creates a new file dialog.
	 * 
	 * @param owner The owning window.
	 * @param openFile Pass in <code>true</code> for an open file dialog, <code>false</code> for
	 *            a save file dialog.
	 * @param addAllFilesFilter Pass in <code>true</code> to automatically add the "All Files"
	 *            filter. Ignored if <code>openFile</code> is <code>false</code>.
	 * @param message A message displayed to the user at the top of the dialog.
	 */
	public TKFileDialog(Frame owner, boolean openFile, boolean addAllFilesFilter, String message) {
		super(owner, openFile ? Msgs.OPEN_DIALOG_TITLE : Msgs.SAVE_DIALOG_TITLE, true);
		initialize(openFile, addAllFilesFilter, message);
	}

	/**
	 * Creates a new file dialog.
	 * 
	 * @param owner The owning dialog.
	 * @param openFile Pass in <code>true</code> for an open file dialog, <code>false</code> for
	 *            a save file dialog.
	 */
	public TKFileDialog(Dialog owner, boolean openFile) {
		this(owner, openFile, true);
	}

	/**
	 * Creates a new file dialog.
	 * 
	 * @param owner The owning dialog.
	 * @param openFile Pass in <code>true</code> for an open file dialog, <code>false</code> for
	 *            a save file dialog.
	 * @param addAllFilesFilter Pass in <code>true</code> to automatically add the "All Files"
	 *            filter. Ignored if <code>openFile</code> is <code>false</code>.
	 */
	public TKFileDialog(Dialog owner, boolean openFile, boolean addAllFilesFilter) {
		this(owner, openFile, addAllFilesFilter, null);
	}

	/**
	 * Creates a new file dialog.
	 * 
	 * @param owner The owning dialog.
	 * @param openFile Pass in <code>true</code> for an open file dialog, <code>false</code> for
	 *            a save file dialog.
	 * @param addAllFilesFilter Pass in <code>true</code> to automatically add the "All Files"
	 *            filter. Ignored if <code>openFile</code> is <code>false</code>.
	 * @param message A message displayed to the user at the top of the dialog.
	 */
	public TKFileDialog(Dialog owner, boolean openFile, boolean addAllFilesFilter, String message) {
		super(owner, openFile ? Msgs.OPEN_DIALOG_TITLE : Msgs.SAVE_DIALOG_TITLE, true);
		initialize(openFile, addAllFilesFilter, message);
	}

	private void initialize(boolean openFile, boolean addAllFilesFilter, String message) {
		File[] files;

		mOpenDialog = openFile;

		TKPanel content = getContent();

		TKButton openSaveButton = new TKButton(openFile ? Msgs.OPEN_TITLE : Msgs.SAVE_TITLE);
		openSaveButton.setActionCommand(OPEN_SAVE);
		openSaveButton.addActionListener(this);

		TKButton cancelButton = new TKButton(Msgs.CANCEL_TITLE);
		cancelButton.setActionCommand(CANCEL_CMD);
		cancelButton.addActionListener(this);

		mUpButton = new TKButton(TKImage.getUpFolderIcon());
		mUpButton.setActionCommand(UP);
		mUpButton.addActionListener(this);
		mUpButton.setToolTipText(Msgs.UP_FOLDER_TOOLTIP);

		mHomeButton = new TKButton(TKImage.getHomeIcon());
		mHomeButton.setActionCommand(HOME);
		mHomeButton.addActionListener(this);
		mHomeButton.setToolTipText(Msgs.HOME_TOOLTIP);

		mNewFolderButton = new TKButton(TKImage.getNewFolderIcon());
		mNewFolderButton.setActionCommand(NEW_FOLDER);
		mNewFolderButton.addActionListener(this);
		mNewFolderButton.setToolTipText(Msgs.NEW_FOLDER_TOOLTIP);

		mLocation = getLastLocationUsed();
		files = mLocation.isDirectory() ? mLocation.listFiles() : null;
		if (files == null) {
			mLocation = TKPath.getFile(TKPath.getFullPath(new File(DOT)));
			files = mLocation.listFiles();
			if (files == null) {
				files = new File[0];
			}
		}
		Arrays.sort(files, this);

		mHierarchyMenu = new TKMenu(EMPTY);
		mHierarchyPopup = new TKPopupMenu(mHierarchyMenu, adjustHierarchyMenu());
		mHierarchyPopup.setActionCommand(HIERARCHY);
		mHierarchyPopup.addActionListener(this);

		TKPanel panel = new TKPanel(new TKColumnLayout(4));
		panel.add(mHierarchyPopup);
		panel.add(mUpButton);
		panel.add(mHomeButton);
		panel.add(mNewFolderButton);
		panel.setBorder(new TKEmptyBorder(10, 10, 5, 10));

		if (message == null) {
			content.add(panel, TKCompassPosition.NORTH);
		} else {
			TKPanel headerPanel = new TKPanel(new TKColumnLayout(1));
			TKLabel messageLabel = new TKLabel(message);
			messageLabel.setBorder(new TKEmptyBorder(7, 5, 1, 0));
			headerPanel.add(messageLabel);
			headerPanel.add(panel);
			content.add(headerPanel, TKCompassPosition.NORTH);
		}

		mList = new TKItemList<File>(files);
		mList.addActionListener(this);
		mList.setItemRenderer(new FileItemRenderer(!openFile));
		TKScrollPanel scrollPane = new TKScrollPanel(mList);
		scrollPane.getContentView().setMinimumSize(new Dimension(300, 250));
		scrollPane.getContentBorderView().setBorder(new TKLineBorder(Color.black, 1, TKLineBorder.LEFT_EDGE | TKLineBorder.TOP_EDGE));
		scrollPane.setBorder(new TKEmptyBorder(5, 10, 5, 10));
		content.add(scrollPane, TKCompassPosition.CENTER);

		mFilterMenu = new TKMenu(EMPTY);
		if (openFile && addAllFilesFilter) {
			addFileFilter(new TKFileFilter());
		}
		mFilterPopup = new TKPopupMenu(mFilterMenu, 0);
		mFilterPopup.setActionCommand(FILTER);
		mFilterPopup.addActionListener(this);

		int width = 0;
		int height = 0;
		Dimension size = openSaveButton.getPreferredSize();
		if (size.width > width) {
			width = size.width;
		}
		if (size.height > height) {
			height = size.height;
		}
		size = cancelButton.getPreferredSize();
		if (size.width > width) {
			width = size.width;
		}
		if (size.height > height) {
			height = size.height;
		}
		size.width = width;
		size.height = height;
		openSaveButton.setMinimumSize(size);
		openSaveButton.setMaximumSize(size);
		openSaveButton.setPreferredSize(size);
		cancelButton.setMinimumSize(size);
		cancelButton.setMaximumSize(size);
		cancelButton.setPreferredSize(size);

		mAccessoryPanel = new TKPanel(new TKColumnLayout(1, 0, 0));
		if (openFile) {
			panel = new TKPanel(new TKColumnLayout(4));
			panel.add(new TKLabel(Msgs.FILTER_TITLE, null, TKAlignment.RIGHT, false, TKFont.TEXT_FONT_KEY));
			panel.add(mFilterPopup);
			panel.add(cancelButton);
			panel.add(openSaveButton);
		} else {
			panel = new TKPanel(new TKColumnLayout(3));
			mFileNameField = new TKTextField();
			mFileNameField.setKeyEventFilter(this);
			mFileNameField.addDocumentListener(this);
			mFileNameLabel = new TKLabel(Msgs.FILENAME_TITLE, null, TKAlignment.RIGHT, false, TKFont.TEXT_FONT_KEY);
			panel.add(mFileNameLabel);
			panel.add(mFileNameField);
			panel.add(openSaveButton);
			panel.add(new TKLabel(Msgs.FILTER_TITLE, null, TKAlignment.RIGHT, false, TKFont.TEXT_FONT_KEY));
			panel.add(mFilterPopup);
			panel.add(cancelButton);
		}
		mAccessoryPanel.add(panel);
		mAccessoryPanel.setBorder(new TKEmptyBorder(5, 10, 10, 10));
		content.add(mAccessoryPanel, TKCompassPosition.SOUTH);

		setFolderSelectionAllowed(false);
		setMultipleSelectionAllowed(openFile);
		setDefaultButton(openSaveButton);
		setCancelButton(cancelButton);
		if (openFile) {
			mList.requestFocus();
		} else {
			mFileNameField.requestFocus();
		}
	}

	public void actionPerformed(ActionEvent event) {
		String command = event.getActionCommand();

		if (OPEN_SAVE.equals(command)) {
			commit();
		} else if (CANCEL_CMD.equals(command)) {
			setResult(CANCEL);
			attemptClose();
		} else if (HOME.equals(command)) {
			setLocation(new File(System.getProperty("user.home", DOT))); //$NON-NLS-1$
		} else if (UP.equals(command)) {
			setLocation(mLocation.getParentFile());
		} else if (NEW_FOLDER.equals(command)) {
			if (mLocation != null) {
				String name = Msgs.DEFAULT_NEW_FOLDER_NAME;
				File dirTry = new File(mLocation, name);
				int count = 2;

				while (dirTry.exists()) {
					name = Msgs.DEFAULT_NEW_FOLDER_NAME + count++;
					dirTry = new File(mLocation, name);
				}

				name = TKOptionDialog.response(this, Msgs.NEW_FOLDER_PROMPT, name, 200);
				if (name != null) {
					boolean success = false;

					dirTry = new File(mLocation, name);
					if (!dirTry.exists()) {
						success = dirTry.mkdir();
					}
					if (success) {
						applyFilter();
						adjustForSelection();
					} else {
						TKOptionDialog.error(this, MessageFormat.format(Msgs.UNABLE_TO_CREATE_NEW_FOLDER, name));
					}
				}
			}
		} else if (HIERARCHY.equals(command)) {
			Object obj = mHierarchyPopup.getSelectedItem().getUserObject();

			if (obj instanceof File) {
				if (!mTransitioning) {
					mLocation = null; // Force the change
				}
				setLocation((File) obj);
			} else if (obj instanceof String) {
				File oldLocation = mLocation;

				command = (String) obj;
				if (ADD_TO_FAVORITES.equals(command)) {
					addToFavorites(mLocation);
				} else if (REMOVE_FROM_FAVORITES.equals(command)) {
					removeFromFavorites();
				}
				mLocation = null;
				setLocation(oldLocation);
			}
		} else if (FILTER.equals(command)) {
			applyFilter();
			adjustForSelection();
		} else if (TKItemList.CMD_OPEN_SELECTION.equals(command)) {
			if (mList.getSelectionCount() == 1) {
				File file = mList.getSelectedItem();

				if (file.isDirectory()) {
					setLocation(file);
				} else if (mOpenDialog) {
					commit();
				} else {
					setFileName(file.getName());
				}
			}
		} else if (TKItemList.CMD_SELECTION_CHANGED.equals(command)) {
			adjustForSelection();
		}
	}

	/**
	 * Adds one or more file filters.
	 * 
	 * @param filters The filters to add.
	 */
	public void addFileFilter(TKFileFilter... filters) {
		for (TKFileFilter filter : filters) {
			TKMenuItem filterItem = new TKMenuItem(filter.getDescription());

			filterItem.setUserObject(filter);
			mFilterMenu.add(filterItem);
		}
	}

	private void addToFavorites(File favorite) {
		loadFavoritesAndRecentFolders();
		for (File file : FAVORITES) {
			if (file.equals(favorite)) {
				return;
			}
		}
		FAVORITES.add(favorite);
		saveFavorites();
	}

	/** Called to adjust controls in the panel for the current selection. */
	protected void adjustForSelection() {
		boolean enabled = false;
		int i;

		if (mOpenDialog) {
			enabled = mList.getSelectionCount() > 0;
			if (enabled && !isFolderSelectionAllowed()) {
				for (File file : mList.getSelectedItems()) {
					if (file.isDirectory()) {
						enabled = false;
						break;
					}
				}
			}
		} else {
			int length = mFileNameField.getLength();

			if (length > 0 && length < 240) {
				String name = mFileNameField.getText();

				for (i = 0; i < length; i++) {
					if (name.charAt(i) != ' ') {
						enabled = true;
						break;
					}
				}
			}
		}

		getDefaultButton().setEnabled(enabled && isSaveEnabled());
	}

	/** @return <code>true</code> if the "Save" button can be enabled. */
	protected boolean isSaveEnabled() {
		return true;
	}

	private int adjustHierarchyMenu() {
		BufferedImage openFolderIcon = TKImage.getOpenFolderIcon();
		BufferedImage folderIcon = TKImage.getFolderIcon();
		File location = mLocation;
		File[] roots = File.listRoots();
		ArrayList<File> list = new ArrayList<File>();
		int level = 0;
		int select = 0;
		boolean needAddFavorite = true;
		HashMap<String, File> leafNames = new HashMap<String, File>();
		HashMap<File, String> invLeafName = new HashMap<File, String>();
		HashSet<String> dupLeafNames = new HashSet<String>();
		int i;
		int count;
		int realCount;
		boolean isRoot;
		TKMenuItem item;
		String name;

		mHierarchyMenu.removeAll();

		while (location != null) {
			list.add(location);
			location = location.getParentFile();
		}

		if (list.size() > 0) {
			ArrayList<File> newList = new ArrayList<File>(list.size() + roots.length);

			count = list.size() - 1;
			location = list.get(count);
			for (i = 0; i < roots.length; i++) {
				newList.add(roots[i]);
				if (roots[i].equals(location)) {
					for (int j = count - 1; j >= 0; j--) {
						newList.add(list.get(j));
					}
				}
			}
			list = newList;
		}

		count = list.size();
		for (i = 0; i < count; i++) {
			File folder = list.get(i);

			name = folder.getName();
			isRoot = name.length() == 0;
			item = new TKMenuItem(isRoot ? folder.toString() : name);

			if (folder.equals(mLocation)) {
				select = i;
			}
			item.setUserObject(folder);
			if (isRoot) {
				level = 0;
			} else {
				level++;
			}
			item.setIndent(level * INDENT_AMOUNT);
			item.setIcon(openFolderIcon);
			mHierarchyMenu.add(item);
		}

		loadFavoritesAndRecentFolders();

		// Disambiguate duplicate leaf names
		for (File favorite : FAVORITES) {
			name = favorite.getName();
			if (name.length() != 0) {
				if (leafNames.containsKey(name)) {
					if (!favorite.equals(leafNames.get(name))) {
						dupLeafNames.add(name);
					}
				} else {
					leafNames.put(name, favorite);
				}
			}
		}

		for (File recent : RECENT_FOLDERS) {
			name = recent.getName();
			if (name.length() != 0) {
				if (leafNames.containsKey(name)) {
					if (!recent.equals(leafNames.get(name))) {
						dupLeafNames.add(name);
					}
				} else {
					leafNames.put(name, recent);
				}
			}
		}

		for (String dupName : dupLeafNames) {
			int min = dupName.length() + 1;
			int len;

			list.clear();
			for (File favorite : FAVORITES) {
				name = favorite.getName();
				if (name.length() != 0 && name.equals(dupName)) {
					list.add(favorite);
				}
			}
			for (File recent : RECENT_FOLDERS) {
				name = recent.getName();
				if (name.length() != 0 && name.equals(dupName)) {
					list.add(recent);
				}
			}

			count = list.size();
			for (i = 0; i < count - 1; i++) {
				File folder = list.get(i);

				name = folder.toString();
				for (int j = i + 1; j < count; j++) {
					String name2 = list.get(j).toString();
					int k = name.length() - 1;
					int k2 = name2.length() - 1;

					len = 1;
					for (; k >= 0 && k2 >= 0; k--, k2--) {
						if (name.charAt(k) == name2.charAt(k2)) {
							len++;
						} else {
							break;
						}
					}
					if (len > min) {
						min = len;
					}
				}
			}

			for (File folder : list) {
				name = folder.toString();
				len = name.length();
				len -= min;
				if (len > 4) {
					char ch = name.charAt(len);

					while (len > 4 && ch != '/' && ch != '\\') {
						len--;
						ch = name.charAt(len);
					}
					if (len > 4) {
						name = "\u2026" + name.substring(len); //$NON-NLS-1$
					}
				}
				invLeafName.put(folder, name);
			}
		}

		// Add "Favorites"
		mHierarchyMenu.addSeparator();
		item = new TKMenuItem(Msgs.FAVORITES);
		item.setIcon(openFolderIcon);
		item.setEnabled(false);
		mHierarchyMenu.add(item);
		realCount = FAVORITES.size();
		for (File favorite : FAVORITES) {
			if (favorite.isDirectory()) {
				name = favorite.getName();
				if (name.length() == 0) {
					name = favorite.toString();
				} else if (dupLeafNames.contains(name)) {
					name = invLeafName.get(favorite);
				}
				item = new TKMenuItem(name);
				item.setUserObject(favorite);
				item.setIndent(INDENT_AMOUNT);
				item.setIcon(folderIcon);
				mHierarchyMenu.add(item);
				if (favorite.equals(mLocation)) {
					needAddFavorite = false;
				}
			} else {
				realCount--;
			}
		}
		item = new TKMenuItem(Msgs.ADD_TO_FAVORITES);
		item.setUserObject(ADD_TO_FAVORITES);
		item.setEnabled(needAddFavorite);
		mHierarchyMenu.add(item);
		item = new TKMenuItem(Msgs.REMOVE_FROM_FAVORITES);
		item.setUserObject(REMOVE_FROM_FAVORITES);
		item.setEnabled(realCount > 0);
		mHierarchyMenu.add(item);

		// Add "Recent Folders"
		mHierarchyMenu.addSeparator();
		item = new TKMenuItem(Msgs.RECENT_FOLDERS);
		item.setIcon(openFolderIcon);
		item.setEnabled(false);
		mHierarchyMenu.add(item);
		for (File recent : RECENT_FOLDERS) {
			if (recent.isDirectory()) {
				name = recent.getName();
				if (name.length() == 0) {
					name = recent.toString();
				} else if (dupLeafNames.contains(name)) {
					name = invLeafName.get(recent);
				}
				item = new TKMenuItem(name);
				item.setUserObject(recent);
				item.setIndent(INDENT_AMOUNT);
				item.setIcon(folderIcon);
				mHierarchyMenu.add(item);
			}
		}

		return select;
	}

	/**
	 * @return The active filter.
	 */
	public TKFileFilter getActiveFilter() {
		return (TKFileFilter) mFilterPopup.getSelectedItem().getUserObject();
	}

	private void applyFilter() {
		File[] files = mLocation != null ? mLocation.listFiles(getActiveFilter()) : null;

		if (files != null) {
			Arrays.sort(files, this);
			mList.replaceItems(files);
		} else {
			mList.removeAllItems();
		}
	}

	private void commit() {
		if (!mOpenDialog) {
			File file = getSelectedItem();

			if (file.exists()) {
				if (file.isDirectory()) {
					TKOptionDialog.error(this, MessageFormat.format(Msgs.IS_FOLDER_ERROR, file));
					return;
				}
				if (!file.canWrite()) {
					TKOptionDialog.error(this, MessageFormat.format(Msgs.WRITE_PERMISSION_ERROR, file));
					return;
				}

				TKOptionDialog dialog = new TKOptionDialog(this, Msgs.OVERWRITE_TITLE, TKOptionDialog.TYPE_YES_NO);

				dialog.setDefaultButton(dialog.getNoButton());
				if (dialog.doModal(TKImage.getMessageIcon(), new TKLabel(MessageFormat.format(Msgs.OVERWRITE_CONFIRMATION, file), null, TKAlignment.LEFT, true, TKFont.CONTROL_FONT_KEY)) == TKOptionDialog.NO) {
					return;
				}
			}
		}
		setResult(OK);
		setLastLocationUsed(mLocation);
		attemptClose();
	}

	public int compare(File first, File second) {
		if (first != second) {
			boolean firstIsDir = first.isDirectory();

			if (firstIsDir != second.isDirectory()) {
				return firstIsDir ? -1 : 1;
			}
			return first.compareTo(second);
		}
		return 0;
	}

	public void documentChanged(TKDocument document) {
		adjustForSelection();
	}

	@Override public int doModal() {
		if (mFilterMenu.getMenuItemCount() == 0) {
			TKFileFilter all = new TKFileFilter();

			addFileFilter(all);
			setActiveFileFilter(all);
		}
		adjustForSelection();
		return super.doModal();
	}

	public boolean filterKeyEvent(TKPanel owner, KeyEvent event, boolean isReal) {
		if (event.getID() == KeyEvent.KEY_TYPED) {
			char ch = event.getKeyChar();

			if (ch == '\n' || ch == '\r' || ch == File.separatorChar || ch == File.pathSeparatorChar) {
				return true;
			}
		}
		return false;
	}

	/** @return The current file icon map. */
	public static HashMap<String, BufferedImage> getFileIconMap() {
		if (FILE_ICON_MAP == null) {
			FILE_ICON_MAP = new HashMap<String, BufferedImage>();
		}
		return FILE_ICON_MAP;
	}

	/**
	 * @param path The path to return an icon for.
	 * @return The icon for the specified file.
	 */
	public static BufferedImage getIconForFile(String path) {
		return getIconForFileExtension(TKPath.getExtension(path));
	}

	/**
	 * @param file The file to return an icon for.
	 * @return The icon for the specified file.
	 */
	public static BufferedImage getIconForFile(File file) {
		return getIconForFile(file != null && file.isFile() ? file.getName() : null);
	}

	/**
	 * @param extension The extension to return an icon for.
	 * @return The icon for the specified file extension.
	 */
	public static BufferedImage getIconForFileExtension(String extension) {
		if (extension != null) {
			BufferedImage icon;

			if (!extension.startsWith(DOT)) {
				extension = DOT + extension;
			}

			icon = getFileIconMap().get(extension);
			if (icon != null) {
				return icon;
			}

			return TKImage.getFileIcon();
		}
		return TKImage.getFolderIcon();
	}

	/**
	 * Sets the icon for the specified file extension.
	 * 
	 * @param extension The file extension.
	 * @param icon The icon to map to an extension.
	 */
	public static void setIconForFileExtension(String extension, BufferedImage icon) {
		if (extension != null && icon != null) {
			if (!extension.startsWith(DOT)) {
				extension = DOT + extension;
			}
			getFileIconMap().put(extension, icon);
		}
	}

	/** @return The location a new file dialog will open to. */
	public static File getLastLocationUsed() {
		TKPreferences prefs = TKPreferences.getInstance();

		prefs.resetIfVersionMisMatch(MODULE, CURRENT_VERSION);

		return TKPath.getFile(TKPath.getFullPath(TKPath.getFile(prefs.getStringValue(MODULE, getLastLocationKey(), DOT))));
	}

	/** @return The first selected item. */
	public File getSelectedItem() {
		if (mOpenDialog) {
			return mList.getSelectedItem();
		}

		return getActiveFilter().enforceExtension(new File(mLocation, mFileNameField.getText()), false);
	}

	/** @return An array of selected items. */
	public File[] getSelectedItems() {
		if (mOpenDialog) {
			return mList.getSelectedItems().toArray(new File[0]);
		}

		return new File[] { getSelectedItem() };
	}

	/** @return The number of items selected. */
	public int getSelectionCount() {
		if (mOpenDialog) {
			return mList.getSelectionCount();
		}

		return 1;
	}

	/** @return <code>true</code> if folders may be selected. */
	public boolean isFolderSelectionAllowed() {
		return mFolderSelectionAllowed;
	}

	/** @return <code>true</code> if more than one file may be selected. */
	public boolean isMultipleSelectionAllowed() {
		return mList.isMultipleSelectionAllowed();
	}

	private static void loadFavoritesAndRecentFolders() {
		if (FAVORITES == null) {
			TKPreferences prefs = TKPreferences.getInstance();
			StringTokenizer tokenizer;

			prefs.resetIfVersionMisMatch(MODULE, CURRENT_VERSION);

			FAVORITES = new ArrayList<File>();
			tokenizer = new StringTokenizer(prefs.getStringValueForced(MODULE, FAVORITES_KEY), SEMICOLON);
			while (tokenizer.hasMoreTokens()) {
				File favorite = TKPath.getFile(tokenizer.nextToken());

				if (favorite.isDirectory()) {
					FAVORITES.add(favorite);
				}
			}

			RECENT_FOLDERS = new ArrayList<File>();
			tokenizer = new StringTokenizer(prefs.getStringValueForced(MODULE, RECENT_FOLDERS_KEY), SEMICOLON);
			while (tokenizer.hasMoreTokens()) {
				File recent = TKPath.getFile(tokenizer.nextToken());

				if (recent.isDirectory()) {
					RECENT_FOLDERS.add(recent);
				}
			}
		}
	}

	/** Removes any previously set accessory panel. */
	public void removeAccessoryPanel() {
		int count = mAccessoryPanel.getComponentCount();

		while (count > 1) {
			mAccessoryPanel.remove(--count);
		}
	}

	private void removeFromFavorites() {
		TKOptionDialog dialog = new TKOptionDialog(this, Msgs.REMOVAL_DIALOG_TITLE, TKOptionDialog.TYPE_OK_CANCEL);
		TKPanel panel = new TKPanel(new TKColumnLayout());
		TKItemList<File> list = new TKItemList<File>(FAVORITES);
		TKScrollPanel scroller = new TKScrollPanel(TKScrollPanel.BOTH, list);
		int result;

		scroller.setMinimumSize(new Dimension(200, 100));

		scroller.getContentBorderView().setBorder(new TKLineBorder(Color.black, 1, TKLineBorder.LEFT_EDGE | TKLineBorder.TOP_EDGE | TKLineBorder.BOTTOM_EDGE));

		panel.add(new TKLabel(Msgs.REMOVAL_TITLE, null, TKAlignment.LEFT, true, TKFont.CONTROL_FONT_KEY));
		panel.add(scroller);

		list.setActionCommand(TKOptionDialog.CMD_OK);
		list.addActionListener(dialog);

		new ButtonStateModifier(dialog, list);
		dialog.getOKButton().setEnabled(false);

		result = dialog.doModal(TKImage.getMessageIcon(), panel);

		if (result == TKDialog.OK) {
			for (File file : list.getSelectedItems()) {
				FAVORITES.remove(file);
			}
			saveFavorites();
		}
	}

	private void saveFavorites() {
		StringBuilder buffer = new StringBuilder();

		loadFavoritesAndRecentFolders();

		for (File file : FAVORITES) {
			if (buffer.length() > 0) {
				buffer.append(';');
			}
			buffer.append(TKPath.getFullPath(file));
		}

		TKPreferences.getInstance().setValue(MODULE, FAVORITES_KEY, buffer.toString());
	}

	/**
	 * Sets an accessory panel on the dialog.
	 * 
	 * @param panel The panel to set.
	 */
	public void setAccessoryPanel(TKPanel panel) {
		removeAccessoryPanel();
		if (panel != null) {
			TKDivider divider = new TKDivider(false);

			divider.setBorder(new TKEmptyBorder(5, 0, 5, 0));
			mAccessoryPanel.add(divider);
			mAccessoryPanel.add(panel);
		}
	}

	/**
	 * Sets the active file filter.
	 * 
	 * @param filter The filter to set as active.
	 */
	public void setActiveFileFilter(TKFileFilter filter) {
		int count = mFilterMenu.getMenuItemCount();

		for (int i = 0; i < count; i++) {
			TKMenuItem item = mFilterMenu.getMenuItem(i);

			if (item.getUserObject() == filter) {
				mFilterPopup.setSelectedItem(item);
				break;
			}
		}
	}

	/**
	 * Sets the file name to use.
	 * 
	 * @param name The file name.
	 */
	public void setFileName(String name) {
		mFileNameField.setText(name);
		adjustForSelection();
	}

	/**
	 * Sets whether folders may be selected.
	 * 
	 * @param allowed Whether folder selection is allowed.
	 */
	public void setFolderSelectionAllowed(boolean allowed) {
		mFolderSelectionAllowed = allowed;
	}

	/**
	 * Sets the prefix pre-pended to {@link #LAST_LOCATION_KEY}.
	 * 
	 * @param prefix The prefix to use, or <code>null</code>.
	 */
	public static void setLastLocationKeyPrefix(String prefix) {
		LAST_LOCATION_KEY_PREFIX = prefix;
	}

	/** @return The prefix pre-pended to {@link #LAST_LOCATION_KEY}. */
	public static String getLastLocationKeyPrefix() {
		return LAST_LOCATION_KEY_PREFIX;
	}

	private static String getLastLocationKey() {
		return LAST_LOCATION_KEY_PREFIX == null ? LAST_LOCATION_KEY : LAST_LOCATION_KEY_PREFIX + LAST_LOCATION_KEY;
	}

	/**
	 * Sets the location a new file dialog will open to.
	 * 
	 * @param location The file location.
	 */
	public static void setLastLocationUsed(File location) {
		TKPreferences prefs = TKPreferences.getInstance();

		if (location == null) {
			prefs.removePreference(MODULE, getLastLocationKey());
		} else {
			if (location.isFile()) {
				location = location.getParentFile();
			}
			prefs.setValue(MODULE, getLastLocationKey(), TKPath.getFullPath(location));
			updateRecentFolders(location);
		}
	}

	/** @return The current location of the file list. */
	protected File getCurrentLocation() {
		return mLocation;
	}

	/**
	 * Sets the location of the current file display
	 * 
	 * @param location The file location.
	 * @return <code>true</code> if the location was actually changed.
	 */
	public boolean setLocation(File location) {
		if (!mTransitioning && location != null && !location.equals(mLocation)) {
			mLocation = location;
			mTransitioning = true;
			mHierarchyPopup.setSelectedItem(adjustHierarchyMenu());
			mTransitioning = false;
			applyFilter();
			adjustForSelection();
			return true;
		}

		return false;
	}

	/**
	 * Sets the location of the current file display, selecting the actual file or folder.
	 * 
	 * @param locationFile The file location.
	 */
	public void setLocationAndFile(File locationFile) {
		setLocation(locationFile.getParentFile());
		mList.setSelection(locationFile);
	}

	/**
	 * Sets whether multiple files may be selected.
	 * 
	 * @param allowed Whether multiple files may be selected.
	 */
	public void setMultipleSelectionAllowed(boolean allowed) {
		mList.setMultipleSelectionAllowed(allowed);
	}

	private static void updateRecentFolders(File mostRecent) {
		loadFavoritesAndRecentFolders();

		int count = RECENT_FOLDERS.size();
		boolean needAdd = true;
		StringBuilder buffer = new StringBuilder();
		int i;

		for (i = 0; i < count; i++) {
			if (RECENT_FOLDERS.get(i).equals(mostRecent)) {
				if (i != 0) {
					RECENT_FOLDERS.remove(i);
					RECENT_FOLDERS.add(0, mostRecent);
				}
				needAdd = false;
				break;
			}
		}

		if (needAdd) {
			RECENT_FOLDERS.add(0, mostRecent);
			if (count >= MAX_RECENT_FOLDERS) {
				RECENT_FOLDERS.remove(MAX_RECENT_FOLDERS - 1);
			}
		}

		for (File file : RECENT_FOLDERS) {
			if (buffer.length() > 0) {
				buffer.append(';');
			}
			buffer.append(TKPath.getFullPath(file));
		}

		TKPreferences.getInstance().setValue(MODULE, RECENT_FOLDERS_KEY, buffer.toString());
	}

	/**
	 * @return The file name label.
	 */
	protected TKLabel getFileNameLabel() {
		return mFileNameLabel;
	}

	/**
	 * @return The file name field.
	 */
	protected TKTextField getFileNameField() {
		return mFileNameField;
	}

	private class FileItemRenderer extends TKDefaultItemRenderer {
		private boolean	mDimFiles;

		/**
		 * Creates a new file item renderer.
		 * 
		 * @param dimFiles Whether files should be drawn "greyed" out.
		 */
		FileItemRenderer(boolean dimFiles) {
			mDimFiles = dimFiles;
			setWrapText(false);
		}

		@Override public Color getForegroundForItem(Object item, int index, boolean selected, boolean active) {
			if (item instanceof File) {
				File file = (File) item;

				if (!file.isDirectory() && mDimFiles) {
					return Color.lightGray;
				}
			}
			return super.getForegroundForItem(item, index, selected, active);
		}

		@Override public BufferedImage getImageForItem(Object item, int index) {
			if (item instanceof File) {
				File file = (File) item;
				BufferedImage icon = getIconForFile(file);

				if (mDimFiles && file.isFile()) {
					icon = TKImage.createDisabledImage(icon);
				}
				return icon;
			}
			return super.getImageForItem(item, index);
		}

		@Override public String getStringForItem(Object item, int index) {
			if (item instanceof File) {
				File file = (File) item;

				return file.getName();
			}
			return super.getStringForItem(item, index);
		}
	}

	private class ButtonStateModifier implements ActionListener {
		private TKOptionDialog		mBSMDialog;
		private TKItemList<File>	mBSMList;

		/**
		 * Creates a new button state modifier.
		 * 
		 * @param dialog The dialog we're attached to.
		 * @param list The list to monitor.
		 */
		ButtonStateModifier(TKOptionDialog dialog, TKItemList<File> list) {
			mBSMDialog = dialog;
			mBSMList = list;
			list.setSelectionChangedActionCommand(BSM_SELECTION_CHANGED);
			list.addActionListener(this);
		}

		public void actionPerformed(ActionEvent event) {
			if (BSM_SELECTION_CHANGED.equals(event.getActionCommand())) {
				mBSMDialog.getOKButton().setEnabled(mBSMList.getSelectionCount() > 0);
			}
		}
	}
}
