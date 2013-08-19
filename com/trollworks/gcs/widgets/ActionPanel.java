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

package com.trollworks.gcs.widgets;

import java.awt.LayoutManager;
import java.awt.event.ActionEvent;
import java.awt.event.ActionListener;
import java.util.ArrayList;

import javax.swing.JPanel;

/** A {@link JPanel} with {@link ActionListener} support. */
public class ActionPanel extends JPanel {
	private ArrayList<ActionListener>	mActionListeners;
	private String						mActionCommand;

	/** Creates a new {@link ActionPanel} with no layout. */
	public ActionPanel() {
		this(null);
	}

	/**
	 * Creates a new {@link ActionPanel} with the specified layout.
	 * 
	 * @param layout The layout manager to use. May be <code>null</code>.
	 */
	public ActionPanel(LayoutManager layout) {
		super(layout);
	}

	/**
	 * Adds an action listener.
	 * 
	 * @param listener The listener to add.
	 */
	public void addActionListener(ActionListener listener) {
		if (mActionListeners == null) {
			mActionListeners = new ArrayList<ActionListener>(1);
		}
		if (!mActionListeners.contains(listener)) {
			mActionListeners.add(listener);
		}
	}

	/**
	 * Removes an action listener.
	 * 
	 * @param listener The listener to remove.
	 */
	public void removeActionListener(ActionListener listener) {
		if (mActionListeners != null) {
			mActionListeners.remove(listener);
			if (mActionListeners.isEmpty()) {
				mActionListeners = null;
			}
		}
	}

	/** Notifies all action listeners with the standard action command. */
	public void notifyActionListeners() {
		notifyActionListeners(new ActionEvent(this, ActionEvent.ACTION_PERFORMED, getActionCommand()));
	}

	/**
	 * Notifies all action listeners.
	 * 
	 * @param event The action event to notify with.
	 */
	public void notifyActionListeners(ActionEvent event) {
		if (mActionListeners != null && event != null) {
			for (ActionListener listener : new ArrayList<ActionListener>(mActionListeners)) {
				listener.actionPerformed(event);
			}
		}
	}

	/** @return The action command. */
	public String getActionCommand() {
		return mActionCommand;
	}

	/** @param command The action command. */
	public void setActionCommand(String command) {
		mActionCommand = command;
	}
}
