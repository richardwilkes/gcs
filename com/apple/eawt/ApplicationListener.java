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

package com.apple.eawt;

import java.util.EventListener;

/**
 * This is a stub implementation that allows non-Apple platforms to compile and run. Other than this
 * paragraph, all remaining comments in this file were copied from Apple's online Java documentation
 * for this class.
 * <p>
 * The listener interface for receiving ApplicationEvents. Classes needing to process Application
 * Events implement this interface. Multiple listeners are allowed for each event. When an event is
 * fired off, an ApplicationEvent is sent to each listener in turn until one of the them sets
 * <code>isHandled</code> to <code>true</code>. Before being able to receive ApplicationEvents,
 * the listener must register with the Application object's addListener method. A default
 * implementation of this interface is provided in the ApplicationAdapter class.
 */
public interface ApplicationListener extends EventListener {
	/**
	 * Called when the user selects the About item in the application menu. If <code>event</code>
	 * is not handled, by setting <code>isHandled(true),</code> a default About window consisting
	 * of the application's name and icon is displayed. To display a custom About window, designate
	 * the event as being handled and display the appropriate About window.
	 * 
	 * @param event An ApplicationEvent initiated by the user choosing About in the application menu
	 */
	public void handleAbout(ApplicationEvent event);

	/**
	 * Called when the application receives an Open Application event from the Finder or another
	 * application. Usually this will come from the Finder when a user double-clicks your
	 * application icon. If there is any special code that you want to run when you user launches
	 * your application from the Finder or by sending an Open Application event from another
	 * application, include that code as part of this handler. The Open Application event is sent
	 * after AWT has been loaded.
	 * 
	 * @param event The Open Application event
	 */
	public void handleOpenApplication(ApplicationEvent event);

	/**
	 * Called when the application receives an Open Document event from the Finder or another
	 * application. This event is generated when a user double-clicks a document in the Finder. If
	 * the document is registered as belonging to your application, this event is sent to your
	 * application. Documents are bound to a particular application based primarily on their suffix.
	 * In the Finder, a user selects a document and then from the File Menu chooses Get Info. The
	 * Info window allows users to bind a document to a particular application. These events are
	 * sent only if the bound application has file types listed in the Info.plist entries Document
	 * Types or CFBundleDocumentTypes. The ApplicationEvent sent to this handler holds a reference
	 * to the file being opened.
	 * 
	 * @param event An Open Document event with reference to the file to be opened
	 */
	public void handleOpenFile(ApplicationEvent event);

	/**
	 * Called when the Preference item in the application menu is selected. Native Mac OS X
	 * applications make their Preferences window available through the application menu. Java
	 * applications are automatically given an application menu in Mac OS X. By default, the
	 * Preferences item is disabled in that menu. If you are deploying an application on Mac OS X,
	 * you should enable the preferences item with <code>setEnabledPreferencesMenu(true)</code> in
	 * the Application object and then display your Preferences window in this handler.
	 * 
	 * @param event Triggered when the user selects Preferences from the application menu
	 */
	public void handlePreferences(ApplicationEvent event);

	/**
	 * Called when the application is sent a request to print a particular file or files. You can
	 * allow other applications to print files with your application by implementing this handler.
	 * If another application sends a Print Event along with the name of a file that your
	 * application knows how to process, you can use this handler to determine what to do with that
	 * request. You might open your entire application, or just invoke your printing classes. These
	 * events are sent only if the bound application has file types listed in the Info.plist entries
	 * Document Types or CFBundleDocumentTypes. The ApplicationEvent sent to this handler holds a
	 * reference to the file being opened.
	 * 
	 * @param event A Print Document event with a reference to the file(s) to be printed
	 */
	public void handlePrintFile(ApplicationEvent event);

	/**
	 * Called when the application is sent the Quit event. This event is generated when the user
	 * selects Quit from the application menu, when the user types Command-Q, or when the user
	 * control clicks on your application icon in the Dock and chooses Quit. You can either accept
	 * or reject the request to quit. You might want to reject the request to quit if the user has
	 * unsaved work. Reject the request, move into your code to save changes, then quit your
	 * application. To accept the request to quit, and terminate the application, set
	 * <code>isHandled(true)</code> for the event. To reject the quit, set
	 * <code>isHandled(false)</code>.
	 * 
	 * @param event A Quit Application event
	 */
	public void handleQuit(ApplicationEvent event);

	/**
	 * Called when the application receives a Reopen Application event from the Finder or another
	 * application. Usually this will come when a user clicks on your application icon in the Dock.
	 * If there is any special code that needs to run when your user clicks on your application icon
	 * in the Dock or when a Reopen Application event is sent from another application, include that
	 * code as part of this handler.
	 * 
	 * @param event The Reopen Application event
	 */
	public void handleReOpenApplication(ApplicationEvent event);
}
