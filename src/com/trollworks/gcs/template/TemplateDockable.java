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

package com.trollworks.gcs.template;

import com.trollworks.gcs.advantage.Advantage;
import com.trollworks.gcs.app.CommonDockable;
import com.trollworks.gcs.preferences.SheetPreferences;
import com.trollworks.toolkit.ui.menu.file.PrintProxy;
import com.trollworks.toolkit.utility.PathUtils;
import com.trollworks.toolkit.utility.Preferences;
import com.trollworks.toolkit.utility.notification.NotifierTarget;

import java.awt.Color;

import javax.swing.JScrollPane;

/** A list of advantages and disadvantages from a library. */
public class TemplateDockable extends CommonDockable implements NotifierTarget {
	/** The extension for templates. */
	public static final String	EXTENSION	= "gct";	//$NON-NLS-1$
	private JScrollPane			mScroller;
	private TemplateSheet		mTemplate;

	/** Creates a new {@link TemplateDockable}. */
	public TemplateDockable(Template template) {
		super(template);
	}

	@Override
	public Template getDataFile() {
		return (Template) super.getDataFile();
	}

	@Override
	public PrintProxy getPrintProxy() {
		return null;
	}

	@Override
	public String getDescriptor() {
		// RAW: Implement
		return null;
	}

	@Override
	public String getTitle() {
		return PathUtils.getLeafName(getCurrentBackingFile().getName(), false);
	}

	@Override
	public JScrollPane getContent() {
		if (mScroller == null) {
			Template dataFile = getDataFile();
			mTemplate = new TemplateSheet(dataFile);
			mScroller = new JScrollPane(mTemplate);
			mScroller.setBorder(null);
			mScroller.getViewport().setBackground(Color.LIGHT_GRAY);
			getUndoManager().discardAllEdits();
			dataFile.setModified(false);
			Preferences.getInstance().getNotifier().add(this, SheetPreferences.OPTIONAL_MODIFIER_RULES_PREF_KEY);
		}
		return mScroller;
	}

	@Override
	public String[] getAllowedExtensions() {
		return new String[] { EXTENSION };
	}

	@Override
	public int getNotificationPriority() {
		return 0;
	}

	@Override
	public void handleNotification(Object producer, String name, Object data) {
		getDataFile().notifySingle(Advantage.ID_LIST_CHANGED, null);
	}
}
