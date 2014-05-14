/*
 * Copyright (c) 1998-2014 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public License,
 * version 2.0. If a copy of the MPL was not distributed with this file, You
 * can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as defined
 * by the Mozilla Public License, version 2.0.
 */

package com.trollworks.gcs.character;

import com.trollworks.gcs.app.GCSImages;
import com.trollworks.gcs.widgets.LibraryDockable;
import com.trollworks.toolkit.annotation.Localize;
import com.trollworks.toolkit.ui.UIUtilities;
import com.trollworks.toolkit.ui.menu.file.SaveCommand;
import com.trollworks.toolkit.ui.menu.file.Saveable;
import com.trollworks.toolkit.ui.widget.DataModifiedListener;
import com.trollworks.toolkit.ui.widget.WindowUtils;
import com.trollworks.toolkit.ui.widget.dock.DockCloseable;
import com.trollworks.toolkit.ui.widget.dock.Dockable;
import com.trollworks.toolkit.utility.Localization;
import com.trollworks.toolkit.utility.PathUtils;

import java.awt.Color;
import java.io.File;
import java.util.ArrayList;

import javax.swing.Icon;
import javax.swing.JOptionPane;
import javax.swing.JScrollPane;

/** A list of advantages and disadvantages from a library. */
public class SheetDockable implements Dockable, DockCloseable, Saveable {
	@Localize("Save")
	private static String		SAVE;
	@Localize("Save changes to \"%s\"?")
	private static String		SAVE_CHANGES;
	@Localize("An error occurred while trying to save the sheet as a PNG.")
	private static String		SAVE_AS_PNG_ERROR;
	@Localize("An error occurred while trying to save the sheet as a PDF.")
	private static String		SAVE_AS_PDF_ERROR;
	@Localize("An error occurred while trying to save the sheet as HTML.")
	private static String		SAVE_AS_HTML_ERROR;
	@Localize("An error occurred while trying to save the sheet.")
	private static String		SAVE_ERROR;

	static {
		Localization.initialize();
	}

	/** The extension for character sheets. */
	public static final String	SHEET_EXTENSION	= "gcs";	//$NON-NLS-1$
	/** The PNG extension. */
	public static final String	PNG_EXTENSION	= "png";	//$NON-NLS-1$
	/** The PDF extension. */
	public static final String	PDF_EXTENSION	= "pdf";	//$NON-NLS-1$
	/** The HTML extension. */
	public static final String	HTML_EXTENSION	= "html";	//$NON-NLS-1$
	private GURPSCharacter		mCharacter;
	private JScrollPane			mScroller;
	private CharacterSheet		mSheet;
	private PrerequisitesThread	mPrereqThread;

	/** Creates a new {@link SheetDockable}. */
	public SheetDockable(GURPSCharacter character) {
		mCharacter = character;
	}

	@Override
	public String getDescriptor() {
		// RAW: Implement
		return null;
	}

	@Override
	public Icon getTitleIcon() {
		return GCSImages.getCharacterSheetIcons().getIcon(16);
	}

	@Override
	public String getTitle() {
		return PathUtils.getLeafName(mCharacter.getFile().getName(), false);
	}

	@Override
	public String getTitleTooltip() {
		return LibraryDockable.createTooltip(getTitle(), mCharacter.getFile());
	}

	@Override
	public JScrollPane getContent() {
		if (mScroller == null) {
			mSheet = new CharacterSheet(mCharacter);
			mScroller = new JScrollPane(mSheet);
			mScroller.setBorder(null);
			mScroller.getViewport().setBackground(Color.LIGHT_GRAY);
			mSheet.rebuild();
			mScroller.getViewport().addChangeListener(mSheet);
			mPrereqThread = new PrerequisitesThread(mSheet);
			mPrereqThread.start();
			PrerequisitesThread.waitForProcessingToFinish(mCharacter);
			mCharacter.setModified(false);
		}
		return mScroller;
	}

	@Override
	public boolean attemptClose() {
		UIUtilities.forceFocusToAccept();
		if (mCharacter.isModified()) {
			switch (JOptionPane.showConfirmDialog(mScroller, String.format(SAVE_CHANGES, getTitle()), SAVE, JOptionPane.YES_NO_CANCEL_OPTION)) {
				case JOptionPane.CANCEL_OPTION:
				case JOptionPane.CLOSED_OPTION:
					return false;
				case JOptionPane.YES_OPTION:
					SaveCommand.save(this);
					return !mCharacter.isModified();
				default:
					return true;
			}
		}
		return true;
	}

	@Override
	public File getBackingFile() {
		return mCharacter.getFile();
	}

	@Override
	public boolean isModified() {
		return mCharacter.isModified();
	}

	@Override
	public String[] getAllowedExtensions() {
		return new String[] { SHEET_EXTENSION, PDF_EXTENSION, HTML_EXTENSION, PNG_EXTENSION };
	}

	@Override
	public String getPreferredSavePath() {
		String name = mCharacter.getDescription().getName();
		if (name.length() == 0) {
			name = getTitle();
		}
		return PathUtils.getFullPath(PathUtils.getParent(PathUtils.getFullPath(getBackingFile())), name);
	}

	@Override
	public File[] saveTo(File file) {
		ArrayList<File> result = new ArrayList<>();
		String extension = PathUtils.getExtension(file.getName());
		if (HTML_EXTENSION.equals(extension)) {
			if (mSheet.saveAsHTML(file, null, null)) {
				result.add(file);
			} else {
				WindowUtils.showError(mScroller, SAVE_AS_HTML_ERROR);
			}
		} else if (PNG_EXTENSION.equals(extension)) {
			if (!mSheet.saveAsPNG(file, result)) {
				WindowUtils.showError(mScroller, SAVE_AS_PNG_ERROR);
			}
		} else if (PDF_EXTENSION.equals(extension)) {
			if (mSheet.saveAsPDF(file)) {
				result.add(file);
			} else {
				WindowUtils.showError(mScroller, SAVE_AS_PDF_ERROR);
			}
		} else {
			if (mCharacter.save(file)) {
				result.add(file);
				mCharacter.setFile(file);
				getDockContainer().updateTitle(this);
			} else {
				WindowUtils.showError(mScroller, SAVE_ERROR);
			}
		}
		return result.toArray(new File[result.size()]);
	}

	@Override
	public void addDataModifiedListener(DataModifiedListener listener) {
		mCharacter.addDataModifiedListener(listener);
	}

	@Override
	public void removeDataModifiedListener(DataModifiedListener listener) {
		mCharacter.removeDataModifiedListener(listener);
	}
}
