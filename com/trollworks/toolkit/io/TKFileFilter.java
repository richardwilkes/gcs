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

package com.trollworks.toolkit.io;

import java.io.File;
import java.io.FileFilter;
import java.text.MessageFormat;
import java.util.ArrayList;
import java.util.StringTokenizer;

/** Provides a standard way to filter files with particular extensions. */
public class TKFileFilter implements Comparable<TKFileFilter>, FileFilter {
	private String			mDescription;
	private String			mExtensions;
	private TKFileFilter[]	mFilters;

	/** Creates a new file filter that allows all files and folders through. */
	public TKFileFilter() {
		this(false);
	}

	/**
	 * Creates a new, standard file filter.
	 * 
	 * @param foldersOnly Pass in <code>true</code> to only allow folders to pass through the
	 *            filter or <code>false</code> to allow everything to pass through the filter.
	 */
	public TKFileFilter(boolean foldersOnly) {
		mDescription = foldersOnly ? Msgs.DIRECTORIES_ONLY : Msgs.ALL_FILES;
		if (foldersOnly) {
			mExtensions = ""; //$NON-NLS-1$
		}
	}

	/**
	 * Creates a new file filter.
	 * 
	 * @param description The description to use.
	 * @param extensions The allowable extensions, separated by whitespace.
	 */
	public TKFileFilter(String description, String extensions) {
		mDescription = description;
		mExtensions = extensions != null ? extensions.toLowerCase() : null;
	}

	/**
	 * Creates a new file filter that is the union of the two file filters.
	 * 
	 * @param one The first filter.
	 * @param two The second filter.
	 */
	public TKFileFilter(TKFileFilter one, TKFileFilter two) {
		mFilters = new TKFileFilter[] { one, two };
	}

	/**
	 * Creates a new file filter that is the union of the two file filters.
	 * 
	 * @param one The first filter.
	 * @param two The second filter.
	 * @param description The description to use.
	 */
	public TKFileFilter(TKFileFilter one, TKFileFilter two, String description) {
		mDescription = description;
		mFilters = new TKFileFilter[] { one, two };
	}

	public boolean accept(File file) {
		if (mFilters != null) {
			return mFilters[0].accept(file) || mFilters[1].accept(file);
		}

		if (mExtensions == null) {
			return true;
		}

		if (file == null) {
			return false;
		}

		String name = file.getName();
		if (name == null) {
			return false;
		}

		if (file.isDirectory()) {
			return !name.startsWith(".") || mExtensions.length() == 0; //$NON-NLS-1$
		}

		StringTokenizer tokenizer = new StringTokenizer(mExtensions);
		name = name.toLowerCase();
		while (tokenizer.hasMoreTokens()) {
			if (name.endsWith(tokenizer.nextToken())) {
				return true;
			}
		}
		return false;
	}

	/**
	 * @param file The <code>File</code> object to process.
	 * @param force Pass <code>true</code> if you want to force the first extension this file
	 *            filter supports.
	 * @return A <code>File</code> object that has an extension in the file name.
	 */
	public File enforceExtension(File file, boolean force) {
		if (mFilters == null) {
			String[] extensions = getExtensions();
			int max = force ? extensions.length > 0 ? 1 : 0 : extensions.length;

			if (max > 0) {
				String name = file.getName();

				for (int i = 0; i < max; i++) {
					if (name.endsWith(extensions[i])) {
						return file;
					}
				}

				return TKPath.getFile(TKPath.enforceExtension(TKPath.getFullPath(file), extensions[0], !force));
			}
			return file;
		}
		return mFilters[0].enforceExtension(file, force);
	}

	/** @return The description of this filter. */
	public String getDescription() {
		return toString();
	}

	/** @return The extensions this filter supports. */
	public String[] getExtensions() {
		if (mExtensions != null) {
			ArrayList<String> list = new ArrayList<String>();
			StringTokenizer tokenizer = new StringTokenizer(mExtensions);

			while (tokenizer.hasMoreTokens()) {
				list.add(tokenizer.nextToken());
			}
			return list.toArray(new String[0]);
		}
		return new String[0];
	}

	@Override public String toString() {
		if (mFilters == null || mDescription != null) {
			return mDescription;
		}
		return MessageFormat.format(Msgs.AND, mFilters[0], mFilters[1]);
	}

	public int compareTo(TKFileFilter other) {
		return toString().compareTo(other.toString());
	}
}
