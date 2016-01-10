/*
 * Copyright (c) 1998-2016 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public License,
 * version 2.0. If a copy of the MPL was not distributed with this file, You
 * can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as defined
 * by the Mozilla Public License, version 2.0.
 */

package com.trollworks.gcs.preferences;

import com.trollworks.toolkit.annotation.Localize;
import com.trollworks.toolkit.ui.UIUtilities;
import com.trollworks.toolkit.ui.layout.ColumnLayout;
import com.trollworks.toolkit.ui.preferences.PreferencePanel;
import com.trollworks.toolkit.ui.preferences.PreferencesWindow;
import com.trollworks.toolkit.ui.widget.BandedPanel;
import com.trollworks.toolkit.utility.Localization;
import com.trollworks.toolkit.utility.Preferences;

import java.awt.BorderLayout;
import java.awt.Color;
import java.awt.Component;
import java.awt.Dimension;
import java.io.File;
import java.util.ArrayList;
import java.util.Collections;
import java.util.List;

import javax.swing.JButton;
import javax.swing.JLabel;
import javax.swing.JScrollPane;
import javax.swing.SwingConstants;
import javax.swing.border.CompoundBorder;
import javax.swing.border.EmptyBorder;
import javax.swing.border.LineBorder;

/** The page reference lookup preferences panel. */
public class ReferenceLookupPreferences extends PreferencePanel {
	@Localize("Page References")
	private static String	TITLE;
	@Localize("Remove")
	private static String	REMOVE;

	static {
		Localization.initialize();
	}

	private static final String	MODULE		= "PageReferences";	//$NON-NLS-1$
	private BandedPanel			mPanel;

	public static synchronized File getPdfLocation(String id) {
		String path = Preferences.getInstance().getStringValue(MODULE, id);
		if (path != null) {
			File file = new File(path);
			if (file.exists()) {
				return file;
			}
		}
		return null;
	}

	public static synchronized void setPdfLocation(String id, File location) {
		Preferences prefs = Preferences.getInstance();
		if (location != null) {
			prefs.setValue(MODULE, id, location.getAbsolutePath());
		} else {
			prefs.removePreference(MODULE, id);
		}
	}

	/**
	 * Creates a new {@link ReferenceLookupPreferences}.
	 *
	 * @param owner The owning {@link PreferencesWindow}.
	 */
	public ReferenceLookupPreferences(PreferencesWindow owner) {
		super(TITLE, owner);
		setLayout(new BorderLayout());
		mPanel = new BandedPanel(TITLE);
		mPanel.setLayout(new ColumnLayout(3, 5, 0));
		mPanel.setBorder(new EmptyBorder(2, 5, 2, 5));
		mPanel.setOpaque(true);
		mPanel.setBackground(Color.WHITE);
		List<String> ids = new ArrayList<>(Preferences.getInstance().getModuleKeys(MODULE));
		Collections.sort(ids);
		for (String id : ids) {
			JButton button = new JButton(REMOVE);
			UIUtilities.setOnlySize(button, button.getPreferredSize());
			button.addActionListener(event -> {
				setPdfLocation(id, null);
				Component[] children = mPanel.getComponents();
				for (int i = 0; i < children.length; i++) {
					if (children[i] == button) {
						mPanel.remove(i + 2);
						mPanel.remove(i + 1);
						mPanel.remove(i);
						mPanel.setSize(mPanel.getPreferredSize());
						break;
					}
				}
			});
			mPanel.add(button);
			JLabel idLabel = new JLabel(id, SwingConstants.CENTER);
			idLabel.setBorder(new CompoundBorder(new LineBorder(Color.BLACK), new EmptyBorder(1, 4, 1, 4)));
			idLabel.setOpaque(true);
			idLabel.setBackground(Color.YELLOW);
			mPanel.add(idLabel);
			mPanel.add(new JLabel(getPdfLocation(id).getAbsolutePath()));
		}
		mPanel.setSize(mPanel.getPreferredSize());
		JScrollPane scroller = new JScrollPane(mPanel);
		Dimension preferredSize = scroller.getPreferredSize();
		if (preferredSize.height > 200) {
			preferredSize.height = 200;
		}
		scroller.setPreferredSize(preferredSize);
		add(scroller);
	}

	@Override
	public boolean isSetToDefaults() {
		return Preferences.getInstance().getModuleKeys(MODULE).isEmpty();
	}

	@Override
	public void reset() {
		Preferences.getInstance().removePreferences(MODULE);
		mPanel.removeAll();
		mPanel.setSize(mPanel.getPreferredSize());
	}
}
