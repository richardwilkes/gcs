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
 * 2005-2008 the Initial Developer. All Rights Reserved.
 *
 * Contributor(s):
 *
 * ***** END LICENSE BLOCK ***** */

package com.trollworks.gcs.menu.file;

import java.io.File;

/**
 * Windows that want to participate in the standard {@link SaveCommand} and {@link SaveAsCommand}
 * processing must implement this interface.
 */
public interface Saveable {
	/** @return Whether the changes have been made that could be saved. */
	boolean isModified();

	/**
	 * @return The file extensions allowed when saving. The first one should be used if the user
	 *         doesn't specify an extension.
	 */
	String[] getAllowedExtensions();

	/** @return The preferred file path to use when saving. */
	String getPreferredSavePath();

	/** @return The backing file object, if any. */
	public File getBackingFile();

	/**
	 * Called to actually save the contents to a file.
	 * 
	 * @param file The file to save to.
	 * @return The file(s) actually written to.
	 */
	File[] saveTo(File file);
}
