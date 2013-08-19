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

package com.trollworks.gcs.app;

import com.apple.eawt.Application;
import com.apple.eawt.ApplicationEvent;
import com.apple.eawt.ApplicationListener;

import com.trollworks.gcs.advantage.Advantage;
import com.trollworks.gcs.character.SheetWindow;
import com.trollworks.gcs.common.ListCollectionThread;
import com.trollworks.gcs.equipment.Equipment;
import com.trollworks.gcs.library.LibraryFile;
import com.trollworks.gcs.menu.Command;
import com.trollworks.gcs.menu.data.OpenDataFileCommand;
import com.trollworks.gcs.menu.edit.PreferencesCommand;
import com.trollworks.gcs.menu.file.NewCharacterSheetCommand;
import com.trollworks.gcs.menu.file.PrintCommand;
import com.trollworks.gcs.menu.file.QuitCommand;
import com.trollworks.gcs.menu.help.AboutCommand;
import com.trollworks.gcs.preferences.GeneralPreferences;
import com.trollworks.gcs.preferences.PreferencesWindow;
import com.trollworks.gcs.skill.Skill;
import com.trollworks.gcs.spell.Spell;
import com.trollworks.gcs.template.TemplateWindow;
import com.trollworks.gcs.utility.DelayedTask;
import com.trollworks.gcs.utility.Platform;
import com.trollworks.gcs.utility.io.Images;
import com.trollworks.gcs.utility.io.LaunchProxy;
import com.trollworks.gcs.utility.io.LocalizedMessages;
import com.trollworks.gcs.utility.io.Path;
import com.trollworks.gcs.utility.io.UpdateChecker;
import com.trollworks.gcs.utility.io.WindowsRegistry;
import com.trollworks.gcs.utility.io.cmdline.CmdLine;
import com.trollworks.gcs.widgets.AppWindow;
import com.trollworks.gcs.widgets.GraphicsUtilities;

import java.awt.BorderLayout;
import java.awt.Container;
import java.awt.Dimension;
import java.awt.EventQueue;
import java.awt.KeyEventDispatcher;
import java.awt.KeyboardFocusManager;
import java.awt.Rectangle;
import java.awt.event.KeyAdapter;
import java.awt.event.KeyEvent;
import java.awt.event.MouseAdapter;
import java.awt.event.MouseEvent;
import java.awt.image.BufferedImage;
import java.io.File;
import java.util.ArrayList;
import java.util.BitSet;
import java.util.HashMap;

import javax.swing.JComponent;
import javax.swing.SwingUtilities;
import javax.swing.border.LineBorder;

/** The main application user interface. */
public class App implements ApplicationListener, Runnable, KeyEventDispatcher {
	private static String							MSG_SHEET_DESCRIPTION;
	private static String							MSG_LIBRARY_DESCRIPTION;
	private static String							MSG_TEMPLATE_DESCRIPTION;
	private static String							MSG_TRAITS_DESCRIPTION;
	private static String							MSG_EQUIPMENT_DESCRIPTION;
	private static String							MSG_SKILLS_DESCRIPTION;
	private static String							MSG_SPELLS_DESCRIPTION;

	static {
		LocalizedMessages.initialize(App.class);
	}

	/** The one and only instance of this class. */
	public static final App							INSTANCE	= new App();
	private static final String						DOT			= ".";									//$NON-NLS-1$
	private static HashMap<String, BufferedImage>	ICON_MAP	= new HashMap<String, BufferedImage>();
	private static BitSet							KEY_STATE	= new BitSet();
	private Application								mApp;
	private boolean									mNotificationAllowed;
	private boolean									mOpenersInstalled;
	private ArrayList<File>							mFilesToOpen;

	private App() {
		if (Platform.isMacintosh()) {
			mApp = new Application();
			mApp.addApplicationListener(this);
		}
		AppWindow.setDefaultWindowIcon(Images.getDefaultWindowIcon());
	}

	/**
	 * Must be called as early as possible and only once.
	 * 
	 * @param cmdLine The command-line arguments.
	 */
	public void startup(CmdLine cmdLine) {
		KeyboardFocusManager.getCurrentKeyboardFocusManager().addKeyEventDispatcher(this);

		boolean displaySplash = GeneralPreferences.shouldDisplaySplash();
		if (displaySplash) {
			displaySplash();
		}
		startAsynchronousTasks();

		PreferencesWindow.initialize();
		LaunchProxy.getInstance().setReady(true);
		if (Platform.isMacintosh()) {
			mApp.setEnabledAboutMenu(true);
			mApp.setEnabledPreferencesMenu(true);
		}

		for (File file : cmdLine.getArgumentsAsFiles()) {
			open(file);
		}

		if (!displaySplash) {
			EventQueue.invokeLater(this);
		}
	}

	private void displaySplash() {
		final AppWindow window = new AppWindow(Main.getName(), null, null, null, true);
		Container content = window.getContentPane();
		if (!Platform.isMacintosh()) {
			((JComponent) content).setBorder(LineBorder.createBlackLineBorder());
		}
		content.setLayout(new BorderLayout());
		content.add(new AboutPanel(), BorderLayout.CENTER);
		final Runnable runnable = new Runnable() {
			public void run() {
				if (!window.isClosed()) {
					window.dispose();
					EventQueue.invokeLater(App.this);
				}
			}
		};
		window.addMouseListener(new MouseAdapter() {
			@Override public void mouseClicked(MouseEvent event) {
				EventQueue.invokeLater(runnable);
			}
		});
		window.addKeyListener(new KeyAdapter() {
			@Override public void keyPressed(KeyEvent event) {
				EventQueue.invokeLater(runnable);
			}
		});
		window.pack();
		Dimension size = window.getSize();
		Rectangle bounds = window.getGraphicsConfiguration().getBounds();
		window.setLocation((bounds.width - size.width) / 2, (bounds.height - size.height) / 3);
		GraphicsUtilities.forceOnScreen(window);
		window.setVisible(true);
		DelayedTask.schedule(runnable, 5000);
	}

	private void startAsynchronousTasks() {
		HashMap<String, String> map = new HashMap<String, String>();
		map.put(SheetWindow.SHEET_EXTENSION.substring(1), MSG_SHEET_DESCRIPTION);
		map.put(LibraryFile.EXTENSION.substring(1), MSG_LIBRARY_DESCRIPTION);
		map.put(TemplateWindow.EXTENSION.substring(1), MSG_TEMPLATE_DESCRIPTION);
		map.put(Advantage.OLD_ADVANTAGE_EXTENSION.substring(1), MSG_TRAITS_DESCRIPTION);
		map.put(Equipment.OLD_EQUIPMENT_EXTENSION.substring(1), MSG_EQUIPMENT_DESCRIPTION);
		map.put(Skill.OLD_SKILL_EXTENSION.substring(1), MSG_SKILLS_DESCRIPTION);
		map.put(Spell.OLD_SPELL_EXTENSION.substring(1), MSG_SPELLS_DESCRIPTION);
		File appDir = new File(System.getProperty("app.home", ".")); //$NON-NLS-1$ //$NON-NLS-2$
		WindowsRegistry.register("GCS", map, new File(appDir, "GURPS Character Sheet.exe"), new File(appDir, "icons")); //$NON-NLS-1$ //$NON-NLS-2$ //$NON-NLS-3$

		UpdateChecker.check("gcs", "http://gcs.trollworks.com/current.txt", "http://gcs.trollworks.com"); //$NON-NLS-1$ //$NON-NLS-2$ //$NON-NLS-3$

		ListCollectionThread.get();
	}

	public synchronized void run() {
		setIconForFileExtension(SheetWindow.SHEET_EXTENSION, Images.getCharacterSheetIcon(false));
		setIconForFileExtension(LibraryFile.EXTENSION, Images.getLibraryIcon(false));
		setIconForFileExtension(TemplateWindow.EXTENSION, Images.getTemplateIcon(false));
		setIconForFileExtension(Advantage.OLD_ADVANTAGE_EXTENSION, Images.getAdvantageIcon(false, false));
		setIconForFileExtension(Equipment.OLD_EQUIPMENT_EXTENSION, Images.getEquipmentIcon(false, false));
		setIconForFileExtension(Skill.OLD_SKILL_EXTENSION, Images.getSkillIcon(false, false));
		setIconForFileExtension(Spell.OLD_SPELL_EXTENSION, Images.getSpellIcon(false, false));
		mOpenersInstalled = true;
		if (mFilesToOpen != null) {
			for (File file : mFilesToOpen) {
				open(file);
			}
			mFilesToOpen = null;
		}

		if (AppWindow.getAllWindows().isEmpty()) {
			StartupDialog sd = new StartupDialog();
			sd.setVisible(true);
			if (AppWindow.getAllWindows().isEmpty()) {
				NewCharacterSheetCommand.INSTANCE.newSheet();
			}
		}
		setNotificationAllowed(true);
	}

	public void handleAbout(ApplicationEvent event) {
		handleCommand(event, AboutCommand.INSTANCE);
	}

	public void handleOpenFile(ApplicationEvent event) {
		if (event != null) {
			open(new File(event.getFilename()));
			event.setHandled(true);
		}
	}

	public void handlePreferences(ApplicationEvent event) {
		handleCommand(event, PreferencesCommand.INSTANCE);
	}

	public void handlePrintFile(ApplicationEvent event) {
		if (event != null) {
			String filename = event.getFilename();
			if (filename.toLowerCase().endsWith(SheetWindow.SHEET_EXTENSION)) {
				final File file = new File(filename);
				open(file);
				EventQueue.invokeLater(new Runnable() {
					public void run() {
						SheetWindow sheet = SheetWindow.findSheetWindow(file);
						if (sheet != null) {
							PrintCommand.INSTANCE.print(sheet);
						} else {
							EventQueue.invokeLater(this);
						}
					}
				});
			}
			event.setHandled(true);
		}
	}

	public void handleQuit(ApplicationEvent event) {
		handleCommand(event, QuitCommand.INSTANCE);
	}

	public void handleOpenApplication(ApplicationEvent event) {
		// Ignored.
	}

	public void handleReOpenApplication(ApplicationEvent event) {
		// Ignored.
	}

	private void handleCommand(ApplicationEvent event, Command cmd) {
		cmd.adjustForMenu(null);
		if (cmd.isEnabled()) {
			cmd.actionPerformed(null);
		}
		if (event != null) {
			event.setHandled(true);
		}
	}

	/** @param file The file to open. */
	public synchronized void open(File file) {
		if (mOpenersInstalled) {
			OpenDataFileCommand opener = new OpenDataFileCommand(file);
			if (!SwingUtilities.isEventDispatchThread()) {
				EventQueue.invokeLater(opener);
			} else {
				opener.run();
			}
		} else {
			if (mFilesToOpen == null) {
				mFilesToOpen = new ArrayList<File>();
			}
			mFilesToOpen.add(file);
		}
	}

	/** @return Whether it is OK to put up a notification dialog yet. */
	public boolean isNotificationAllowed() {
		return mNotificationAllowed;
	}

	/** @param allowed Whether it is OK to put up a notification dialog yet. */
	public void setNotificationAllowed(boolean allowed) {
		mNotificationAllowed = allowed;
	}

	/**
	 * @param path The path to return an icon for.
	 * @return The icon for the specified file.
	 */
	public static BufferedImage getIconForFile(String path) {
		return getIconForFileExtension(Path.getExtension(path));
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

			icon = ICON_MAP.get(extension);
			if (icon != null) {
				return icon;
			}

			return Images.getFileIcon();
		}
		return Images.getFolderIcon();
	}

	private static void setIconForFileExtension(String extension, BufferedImage icon) {
		if (extension != null && icon != null) {
			if (!extension.startsWith(DOT)) {
				extension = DOT + extension;
			}
			ICON_MAP.put(extension, icon);
		}
	}

	/**
	 * @param keyCode The key code to check for.
	 * @return Whether the specified key code is currently pressed.
	 */
	public static boolean isKeyPressed(int keyCode) {
		return KEY_STATE.get(keyCode);
	}

	public boolean dispatchKeyEvent(KeyEvent event) {
		if (event.getID() == KeyEvent.KEY_PRESSED) {
			KEY_STATE.set(event.getKeyCode());
		} else if (event.getID() == KeyEvent.KEY_RELEASED) {
			KEY_STATE.clear(event.getKeyCode());
		}
		return false;
	}
}
