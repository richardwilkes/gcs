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

import java.awt.EventQueue;
import java.io.File;
import java.text.MessageFormat;
import java.util.ArrayList;
import java.util.Arrays;
import java.util.Collection;
import java.util.HashSet;
import java.util.List;

/** Provides window opening services. */
public class TKOpenManager implements Runnable {
	private static ArrayList<TKWindowOpener>	OPENERS				= null;
	private static TKWindowOpener				DEFAULT_OPENER		= null;
	private static TKFileFilter					DEFAULT_FILE_FILTER	= null;
	private TKWindow							mWindow;
	private File[]								mFiles;

	private TKOpenManager(TKWindow window, File[] files) {
		mWindow = window;
		mFiles = files;
	}

	public void run() {
		StringBuilder buffer = new StringBuilder();

		try {
			if (mFiles.length > 1 && mWindow != null) {
				mWindow.startProgress(mFiles.length);
			}
			for (File element : mFiles) {
				ArrayList<String> msgs = new ArrayList<String>();

				if (openWindow(element, true, msgs) == null) {
					if (msgs.isEmpty()) {
						msgs.add(MessageFormat.format(Msgs.OPEN_FAILURE, element.getAbsolutePath()));
					}
					for (String msg : msgs) {
						msg = msg.trim();
						if (msg.length() > 0) {
							if (buffer.length() > 0) {
								buffer.append('\n');
							}
							buffer.append(msg);
						}
					}
				}
				if (mWindow != null) {
					mWindow.incrementProgress();
				}
			}
		} finally {
			if (mWindow != null) {
				mWindow.finishProgress();
			}
		}

		if (buffer.length() > 0) {
			TKOptionDialog.error(mWindow, buffer.toString());
		}
	}

	/**
	 * Opens the specified files, using the specified window's progress bar, if any. This action
	 * will take place on the event queue.
	 * 
	 * @param window The base window. May be <code>null</code>.
	 * @param files The files to open.
	 * @param onEventQueue Whether the opening should occur on the event queue.
	 */
	public static void openFiles(TKWindow window, File[] files, boolean onEventQueue) {
		TKOpenManager openMgr = new TKOpenManager(window, files);

		if (onEventQueue) {
			EventQueue.invokeLater(openMgr);
		} else {
			openMgr.run();
		}
	}

	/**
	 * Registers the specified opener.
	 * 
	 * @param opener The opener to register.
	 */
	public static void register(TKWindowOpener opener) {
		if (OPENERS == null) {
			OPENERS = new ArrayList<TKWindowOpener>(1);
		}
		if (!OPENERS.contains(opener)) {
			OPENERS.add(opener);
		}
	}

	/**
	 * Removes the specified opener from the registered openers.
	 * 
	 * @param opener The opener to remove.
	 */
	public static void deregister(TKWindowOpener opener) {
		if (OPENERS != null) {
			OPENERS.remove(opener);
			if (OPENERS.isEmpty()) {
				OPENERS = null;
			}
		}
	}

	/** @return The currently registered openers. */
	public static TKWindowOpener[] getOpeners() {
		TKWindowOpener[] opener = new TKWindowOpener[0];

		if (OPENERS == null) {
			return opener;
		}
		return OPENERS.toArray(opener);
	}

	/** @return The file filters from the currently registered openers. */
	public static TKFileFilter[] getFileFilters() {
		TKFileFilter[] filter = new TKFileFilter[0];

		if (OPENERS == null) {
			return filter;
		}

		HashSet<TKFileFilter> set = new HashSet<TKFileFilter>();

		for (TKWindowOpener opener : OPENERS) {
			TKFileFilter[] filters = opener.getFileFilters();

			for (TKFileFilter element : filters) {
				set.add(element);
			}
		}

		filter = set.toArray(filter);
		Arrays.sort(filter);
		return filter;
	}

	/**
	 * Sets the default file filter used when {@link #promptForOpen()} is called.
	 * 
	 * @param filter The filter to use.
	 */
	public static void setDefaultFileFilter(TKFileFilter filter) {
		DEFAULT_FILE_FILTER = filter;
	}

	/**
	 * @return The default file filter to use.
	 */
	public static TKFileFilter getDefaultFileFilter() {
		return DEFAULT_FILE_FILTER;
	}

	/** Prompts the user to open one or more files. */
	public static void promptForOpen() {
		TKFileDialog dialog = new TKFileDialog(true, true);
		TKFileFilter[] filter = getFileFilters();
		TKFileFilter defFilter = getDefaultFileFilter();
		int i;

		for (i = 0; i < filter.length; i++) {
			dialog.addFileFilter(filter[i]);
			if (filter[i] == defFilter) {
				dialog.setActiveFileFilter(filter[i]);
			}
		}
		if (dialog.doModal() == TKDialog.OK) {
			openFiles(null, dialog.getSelectedItems(), false);
		}
	}

	/**
	 * Attempts to open a window for the specified object. Each registered opener is given a chance
	 * to open a window for the object. Once an opener is successful, no further openers are given a
	 * chance.
	 * 
	 * @param obj The object to open.
	 * @param msgs A list of error and/or warning messages associated with opening this object.
	 * @return The {@link TKWindow} associated with the object that was opened.
	 */
	public static TKWindow openWindow(Object obj, List<String> msgs) {
		return openWindow(obj, true, msgs);
	}

	/**
	 * Attempts to open a window for the specified object. Each registered opener is given a chance
	 * to open a window for the object. Once an opener is successful, no further openers are given a
	 * chance.
	 * 
	 * @param obj The object to open.
	 * @param show Pass in <code>true</code> if the window should be made visible.
	 * @param msgs A list of error and/or warning messages associated with opening this object.
	 * @return The {@link TKWindow} associated with the object that was opened.
	 */
	public static TKWindow openWindow(Object obj, boolean show, List<String> msgs) {
		TKWindow window = null;

		if (OPENERS != null) {
			for (TKWindowOpener opener : OPENERS) {
				window = opener.openWindow(obj, show, false, msgs);
				if (window != null) {
					break;
				}
			}
		}

		if (window == null && DEFAULT_OPENER != null) {
			window = DEFAULT_OPENER.openWindow(obj, show, true, msgs);
		}
		return window;
	}

	/**
	 * Attempts to open a window for the specified object. Each registered opener is given a chance
	 * to open a window for the object. Once an opener is successful, no further openers are given a
	 * chance. After the window has been opened, if it implements the {@link TKSelectionManager}
	 * interface, the <code>selectInfo</code> object will be passed to it.
	 * 
	 * @param obj The object to open.
	 * @param selectInfo The selection information.
	 * @param msgs A list of error and/or warning messages associated with opening this object.
	 * @return The {@link TKWindow} associated with the object that was opened.
	 */
	public static TKWindow openWindowAndSelect(Object obj, Collection<?> selectInfo, List<String> msgs) {
		TKWindow window = openWindow(obj, msgs);

		if (window != null && window instanceof TKSelectionManager) {
			((TKSelectionManager) window).setSelection(selectInfo);
		}
		return window;
	}

	/**
	 * @return The default opener. This opener is used if no other opener is able to create a window
	 *         for the object.
	 */
	public static TKWindowOpener getDefaultOpener() {
		return DEFAULT_OPENER;
	}

	/**
	 * Sets the default opener. This opener is used if no other opener is able to create a window
	 * for the object.
	 * 
	 * @param defaultOpener The opener to set.
	 */
	public static void setDefaultOpener(TKWindowOpener defaultOpener) {
		DEFAULT_OPENER = defaultOpener;
	}
}
