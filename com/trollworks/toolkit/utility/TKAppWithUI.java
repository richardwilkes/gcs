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

package com.trollworks.toolkit.utility;

import com.apple.eawt.Application;
import com.apple.eawt.ApplicationEvent;
import com.apple.eawt.ApplicationListener;

import com.trollworks.toolkit.io.TKPreferences;
import com.trollworks.toolkit.window.TKAboutWindow;
import com.trollworks.toolkit.window.TKBaseWindow;
import com.trollworks.toolkit.window.TKBaseWindowMonitor;
import com.trollworks.toolkit.window.TKDialog;
import com.trollworks.toolkit.window.TKOptionDialog;
import com.trollworks.toolkit.window.TKSplashWindow;
import com.trollworks.toolkit.window.TKWindow;

import java.awt.image.BufferedImage;
import java.util.Iterator;
import java.util.List;

/** Provides common information and performs various tasks for UI management. */
public class TKAppWithUI extends TKApp implements ApplicationListener, TKBaseWindowMonitor {
	/** The preference key for confirming whether or not to quit. */
	public static final String		CONFIRM_QUIT_TAG	= "Application.ConfirmQuit";	//$NON-NLS-1$
	private static BufferedImage	SPLASH_GRAPHIC;
	private boolean					mOKToQuit;
	private boolean					mIsNotificationAllowed;
	private Application				mMacApp;

	/** @return The last instance created. */
	public static TKAppWithUI getInstanceWithUI() {
		return (TKAppWithUI) INSTANCE;
	}

	/**
	 * Creates a central application object. On Mac OS X, this object gets called for a variety of
	 * standard actions. This object will register as a window monitor.
	 */
	public TKAppWithUI() {
		super();
		INSTANCE = this;
		mOKToQuit = true;
		mMacApp = new Application();
		mMacApp.addApplicationListener(this);
		TKWindow.addBaseWindowMonitor(this);
	}

	/** @return The default splash graphic for this application. */
	public static BufferedImage getSplashGraphic() {
		return SPLASH_GRAPHIC;
	}

	/** @param img The default splash graphic for this application. */
	public static void setSplashGraphic(BufferedImage img) {
		SPLASH_GRAPHIC = img;
	}

	/** @return A new about box. */
	protected TKAboutWindow createAbout() {
		return new TKAboutWindow();
	}

	/** @return <code>true</code> if the Mac OS X preference menu is enabled. */
	public boolean getEnabledPreferencesMenu() {
		return mMacApp.getEnabledPreferencesMenu();
	}

	/** @param enabled Whether the Mac OS X preference menu is enabled. */
	public void setEnabledPreferencesMenu(boolean enabled) {
		mMacApp.setEnabledPreferencesMenu(enabled);
	}

	public void handleAbout(ApplicationEvent event) {
		Iterator<TKAboutWindow> iterator = TKWindow.getWindows(TKAboutWindow.class).iterator();

		if (iterator.hasNext()) {
			iterator.next().setVisible(true);
		} else {
			createAbout().display();
		}
		if (event != null) {
			event.setHandled(true);
		}
	}

	public void handlePreferences(ApplicationEvent event) {
		// Nothing to do...
	}

	public void handleOpenApplication(ApplicationEvent event) {
		// Nothing to do...
	}

	public void handleReOpenApplication(ApplicationEvent event) {
		// Nothing to do...
	}

	public void handleOpenFile(ApplicationEvent event) {
		// Nothing to do...
	}

	public void handlePrintFile(ApplicationEvent event) {
		// Nothing to do...
	}

	/**
	 * Implemented to attempt to close all {@link TKWindow} objects. If successful, it then exits
	 * the Java VM with a result of <code>0</code>. {@inheritDoc}
	 */
	public void handleQuit(ApplicationEvent event) {
		if (isOKToQuit()) {
			List<TKWindow> allWindows = TKWindow.getAllWindows();

			if (!allWindows.isEmpty() && TKOptionDialog.confirm(Msgs.CONFIRM_QUIT, TKOptionDialog.TYPE_OK_CANCEL, CONFIRM_QUIT_TAG) != TKDialog.OK) {
				return;
			}

			for (TKWindow window : allWindows) {
				try {
					if (!window.attemptClose()) {
						return;
					}
				} catch (Exception exception) {
					assert false : TKDebug.throwableToString(exception);
				}
			}

			try {
				TKPreferences.getInstance().save();
			} catch (Exception exception) {
				// Ignore, since preferences may not have been initialized...
			}
			System.exit(0);
		}
	}

	/** @return Whether or not quitting is permitted. */
	public boolean isOKToQuit() {
		return mOKToQuit;
	}

	/**
	 * Sets whether it is OK to quit or not.
	 * 
	 * @param okToQuit Whether it is OK to quit or not.
	 */
	public void setOKToQuit(boolean okToQuit) {
		mOKToQuit = okToQuit;
	}

	/**
	 * Implemented to call <code>handleQuit()</code> if no other {@link TKWindow} objects are
	 * open. {@inheritDoc}
	 */
	public void windowClosed(TKBaseWindow window) {
		if (window instanceof TKWindow && !(window instanceof TKSplashWindow) && TKWindow.getAllWindows().isEmpty()) {
			handleQuit(null);
		}
	}

	/** @return Whether it is OK to put up a notification dialog yet. */
	public boolean isNotificationAllowed() {
		return mIsNotificationAllowed;
	}

	/** @param allowed Whether it is OK to put up a notification dialog yet. */
	public void setNotificationAllowed(boolean allowed) {
		mIsNotificationAllowed = allowed;
	}
}
