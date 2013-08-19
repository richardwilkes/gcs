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

package com.trollworks.gcs;

import com.apple.eawt.ApplicationEvent;

import com.trollworks.gcs.ui.common.CSAboutPanel;
import com.trollworks.gcs.ui.common.CSFileOpener;
import com.trollworks.gcs.ui.common.CSImage;
import com.trollworks.gcs.ui.common.CSListCollectionThread;
import com.trollworks.gcs.ui.common.CSListOpener;
import com.trollworks.gcs.ui.common.CSOpenAccessoryPanel;
import com.trollworks.gcs.ui.preferences.CSGeneralPreferences;
import com.trollworks.gcs.ui.preferences.CSPreferencesWindow;
import com.trollworks.gcs.ui.sheet.CSSheetOpener;
import com.trollworks.gcs.ui.template.CSTemplateOpener;
import com.trollworks.toolkit.io.TKFileFilter;
import com.trollworks.toolkit.io.TKLaunchProxy;
import com.trollworks.toolkit.io.TKUpdateChecker;
import com.trollworks.toolkit.io.TKWindowsRegistry;
import com.trollworks.toolkit.io.cmdline.TKCmdLine;
import com.trollworks.toolkit.utility.TKAppWithUI;
import com.trollworks.toolkit.utility.TKFont;
import com.trollworks.toolkit.window.TKAboutWindow;
import com.trollworks.toolkit.window.TKDialog;
import com.trollworks.toolkit.window.TKFileDialog;
import com.trollworks.toolkit.window.TKOpenManager;
import com.trollworks.toolkit.window.TKSplashWindow;
import com.trollworks.toolkit.window.TKWindow;

import java.awt.EventQueue;
import java.io.File;
import java.util.ArrayList;
import java.util.HashMap;

/** The GCS application object. */
public class CSApplication extends TKAppWithUI implements Runnable {
	private boolean			mOpenersInstalled;
	private ArrayList<File>	mFilesToOpen;

	/**
	 * Creates a new {@link CSApplication}.
	 * 
	 * @param cmdLine The command line.
	 */
	public CSApplication(TKCmdLine cmdLine) {
		super();

		setSplashGraphic(CSImage.getSplash());
		TKWindow.setDefaultWindowIcon(CSImage.getDefaultWindowIcon());

		boolean displaySplash = CSGeneralPreferences.shouldDisplaySplash();
		if (displaySplash) {
			(new TKSplashWindow(new CSAboutPanel())).display(this);
		}

		startAsynchronousTasks();

		CSPreferencesWindow.initialize();
		setEnabledPreferencesMenu(true);
		TKLaunchProxy.getInstance().setReady(true);

		for (File file : cmdLine.getArgumentsAsFiles()) {
			openFile(file);
		}

		if (!displaySplash) {
			EventQueue.invokeLater(this);
		}
	}

	private void startAsynchronousTasks() {
		HashMap<String, String> map = new HashMap<String, String>();
		map.put(CSSheetOpener.EXTENSION.substring(1), Msgs.SHEET_DESCRIPTION);
		map.put(CSTemplateOpener.EXTENSION.substring(1), Msgs.TEMPLATE_DESCRIPTION);
		map.put(CSListOpener.ADVANTAGE_EXTENSION.substring(1), Msgs.TRAITS_DESCRIPTION);
		map.put(CSListOpener.EQUIPMENT_EXTENSION.substring(1), Msgs.EQUIPMENT_DESCRIPTION);
		map.put(CSListOpener.SKILL_EXTENSION.substring(1), Msgs.SKILLS_DESCRIPTION);
		map.put(CSListOpener.SPELL_EXTENSION.substring(1), Msgs.SPELLS_DESCRIPTION);
		File appDir = new File(System.getProperty("app.home", ".")); //$NON-NLS-1$ //$NON-NLS-2$
		TKWindowsRegistry.register("GCS", map, new File(appDir, "GURPS Character Sheet.exe"), new File(appDir, "icons")); //$NON-NLS-1$ //$NON-NLS-2$ //$NON-NLS-3$

		TKUpdateChecker.check("gcs", "http://gcs.trollworks.com/current.txt", "http://gcs.trollworks.com"); //$NON-NLS-1$ //$NON-NLS-2$ //$NON-NLS-3$

		CSListCollectionThread.get();
	}

	public synchronized void run() {
		TKOpenManager.register(CSSheetOpener.getInstance());
		TKOpenManager.register(CSTemplateOpener.getInstance());
		TKOpenManager.register(CSListOpener.getInstance());
		TKOpenManager.register(CSFileOpener.getInstance());
		mOpenersInstalled = true;
		if (mFilesToOpen != null) {
			TKOpenManager.openFiles(null, mFilesToOpen.toArray(new File[0]), false);
			mFilesToOpen = null;
		}

		if (TKWindow.getAllWindows().isEmpty()) {
			TKFileDialog fileDialog = new TKFileDialog(true);
			TKFileFilter[] filters = TKOpenManager.getFileFilters();

			for (TKFileFilter element : filters) {
				fileDialog.addFileFilter(element);
			}
			fileDialog.setActiveFileFilter(CSFileOpener.getPreferredFileFilter(filters));
			new CSOpenAccessoryPanel(fileDialog);

			if (fileDialog.doModal() == TKDialog.OK) {
				TKOpenManager.openFiles(null, fileDialog.getSelectedItems(), false);
			}
			if (TKWindow.getAllWindows().isEmpty()) {
				System.exit(0);
			}
		}
		setNotificationAllowed(true);
	}

	private synchronized void openFile(File file) {
		if (mOpenersInstalled) {
			TKOpenManager.openFiles(null, new File[] { file }, true);
		} else {
			if (mFilesToOpen == null) {
				mFilesToOpen = new ArrayList<File>();
			}
			mFilesToOpen.add(file);
		}
	}

	@Override public void handleOpenFile(ApplicationEvent event) {
		if (event != null) {
			openFile(new File(event.getFilename()));
			event.setHandled(true);
		}
	}

	@Override public void handlePreferences(ApplicationEvent event) {
		CSPreferencesWindow.display();
		if (event != null) {
			event.setHandled(true);
		}
	}

	@Override protected TKAboutWindow createAbout() {
		return CSAboutPanel.createAboutWindow();
	}

	@Override public void handleQuit(ApplicationEvent event) {
		TKFont.saveToPreferences();
		super.handleQuit(event);
	}
}
