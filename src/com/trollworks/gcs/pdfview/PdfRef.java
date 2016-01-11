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

package com.trollworks.gcs.pdfview;

import com.trollworks.toolkit.utility.Preferences;
import com.trollworks.toolkit.utility.text.NumericComparator;

import java.io.File;
import java.util.ArrayList;
import java.util.Collections;
import java.util.List;

/** Tracks data for opening and navigating PDFs. */
public class PdfRef implements Comparable<PdfRef> {
	private static final String	MODULE	= "PageReferences";	//$NON-NLS-1$
	private static final int	VERSION	= 2;

	/** @return <code>true</code> if the {@link PdfRef} preferences are equal to their defaults. */
	public static synchronized boolean isSetToDefaults() {
		Preferences prefs = Preferences.getInstance();
		prefs.resetIfVersionMisMatch(MODULE, VERSION);
		List<String> keys = prefs.getModuleKeys(MODULE);
		return keys.size() == 1 ? Preferences.VERSION_KEY.equals(keys.get(0)) : keys.isEmpty();
	}

	/** Reset the {@link PdfRef} preferences. */
	public static synchronized void reset() {
		Preferences.getInstance().removePreferences(MODULE);
	}

	/**
	 * Get a list of all known {@link PdfRef}s.
	 *
	 * @param requireExistence <code>true</code> if only references that refer to an existing file
	 *            should be returned.
	 * @return The list of {@link PdfRef}s.
	 */
	public static synchronized List<PdfRef> getKnown(boolean requireExistence) {
		List<PdfRef> list = new ArrayList<>();
		Preferences prefs = Preferences.getInstance();
		prefs.resetIfVersionMisMatch(MODULE, VERSION);
		for (String id : prefs.getModuleKeys(MODULE)) {
			PdfRef ref = lookup(id, requireExistence);
			if (ref != null) {
				list.add(ref);
			}
		}
		Collections.sort(list);
		return list;
	}

	/**
	 * Attempts to locate an existing {@link PdfRef}.
	 *
	 * @param id The id to lookup.
	 * @param requireExistence <code>true</code> if only a reference that refers to an existing file
	 *            should be returned.
	 * @return The {@link PdfRef}, or <code>null</code>.
	 */
	public static synchronized PdfRef lookup(String id, boolean requireExistence) {
		Preferences prefs = Preferences.getInstance();
		prefs.resetIfVersionMisMatch(MODULE, VERSION);
		String data = prefs.getStringValue(MODULE, id);
		if (data != null) {
			int colon = data.indexOf(':');
			if (colon != -1 && colon < data.length() - 1) {
				File file = new File(data.substring(colon + 1));
				if (!requireExistence || file.exists()) {
					try {
						return new PdfRef(id, file, Integer.parseInt(data.substring(0, colon)));
					} catch (NumberFormatException nfex) {
						// Ignore
					}
				}
			}
		}
		return null;
	}

	private String	mId;
	private File	mFile;
	private int		mPageToIndexOffset;

	/**
	 * Creates a new {@link PdfRef}.
	 *
	 * @param id The id to use. Pass in <code>null</code> or an empty string to create a
	 *            {@link PdfRef} that won't update preferences.
	 * @param file The file that the <code>id</code> refers to.
	 * @param offset The amount to add to a symbolic page number to find the actual index.
	 */
	public PdfRef(String id, File file, int offset) {
		mId = id == null ? "" : id; //$NON-NLS-1$
		mFile = file;
		mPageToIndexOffset = offset;
	}

	/** @return The id. */
	public String getId() {
		return mId;
	}

	/** @return The file. */
	public File getFile() {
		return mFile;
	}

	/** @return The amount to add to a symbolic page number to find the actual index. */
	public int getPageToIndexOffset() {
		return mPageToIndexOffset;
	}

	/** @param offset The amount to add to a symbolic page number to find the actual index. */
	public void setPageToIndexOffset(int offset) {
		if (mPageToIndexOffset != offset) {
			mPageToIndexOffset = offset;
			save();
		}
	}

	@Override
	public int compareTo(PdfRef other) {
		return NumericComparator.caselessCompareStrings(mId, other.mId);
	}

	/** Removes the {@link PdfRef} from preferences. */
	public void remove() {
		if (!mId.isEmpty()) {
			synchronized (PdfRef.class) {
				Preferences prefs = Preferences.getInstance();
				prefs.resetIfVersionMisMatch(MODULE, VERSION);
				prefs.removePreference(MODULE, mId);
			}
			mId = ""; //$NON-NLS-1$
		}
	}

	/** Saves the {@link PdfRef} to preferences if it has a valid id. */
	public void save() {
		if (!mId.isEmpty()) {
			synchronized (PdfRef.class) {
				Preferences prefs = Preferences.getInstance();
				prefs.resetIfVersionMisMatch(MODULE, VERSION);
				prefs.setValue(MODULE, mId, mPageToIndexOffset + ":" + mFile.getAbsolutePath()); //$NON-NLS-1$
			}
		}
	}
}
