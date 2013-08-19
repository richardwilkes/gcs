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
import java.io.IOException;
import java.util.HashMap;
import java.util.Map;

/**
 * Provides transactional file writing. By using this class to wrap all file updates, you can
 * guarantee that all files will be updated as a logical unit, or, on failure, none will be
 * modified.
 */
public class TKSafeFileUpdater {
	private HashMap<File, File>	mFiles;
	private int					mStarted;

	/** Creates a new transaction. */
	public TKSafeFileUpdater() {
		mStarted = 0;
		mFiles = new HashMap<File, File>();
	}

	/**
	 * Aborts the transaction, leaving the files existing before the transaction was started as they
	 * were.
	 */
	public void abort() {
		mStarted = 0;
		for (File file : mFiles.values()) {
			file.delete();
		}
		mFiles.clear();
	}

	/** Call to begin a transaction. */
	public void begin() {
		mStarted++;
	}

	/**
	 * Commits the transaction. If a transactional file was removed, the corresponding real file
	 * will be removed as well.
	 * 
	 * @throws IOException if a failure occurs. In this case, no files will be altered, as if
	 *             <code>abort()</code> had been called instead.
	 */
	public void commit() throws IOException {
		if (mStarted == 0) {
			throw new IllegalStateException(Msgs.NO_TRANSACTION_IN_PROGRESS);
		}

		if (--mStarted == 0) {
			HashMap<File, File> renameMap = new HashMap<File, File>();
			File destFile;
			File tmpFile;

			// Attempt to swap all the necessary files
			try {
				for (Map.Entry<File, File> entry1 : mFiles.entrySet()) {
					destFile = entry1.getKey();
					tmpFile = entry1.getValue();

					if (destFile.exists()) {
						File tmpRenameFile = File.createTempFile("ren", null, destFile.getParentFile()); //$NON-NLS-1$

						if (tmpRenameFile.delete() && destFile.renameTo(tmpRenameFile)) {
							renameMap.put(destFile, tmpRenameFile);
						} else {
							throw new IOException(Msgs.FILE_SWAP_FAILED);
						}
					}

					if (tmpFile.exists() && !tmpFile.renameTo(destFile)) {
						throw new IOException(Msgs.FILE_SWAP_FAILED);
					}
				}
			} catch (IOException ioe) {
				for (Map.Entry<File, File> entry2 : renameMap.entrySet()) {
					destFile = entry2.getKey();
					tmpFile = entry2.getValue();

					destFile.delete();
					tmpFile.renameTo(destFile);
				}
				abort();
				throw ioe;
			}

			// Clean up our temporary files
			for (File delFile1 : renameMap.values()) {
				delFile1.delete();
			}
			for (File delFile2 : mFiles.values()) {
				delFile2.delete();
			}
			mFiles.clear();
		}
	}

	/**
	 * When <code>commit</code> is called, the transactional file obtained from this call will be
	 * swapped with the original.
	 * 
	 * @param file The file to be created/modified.
	 * @return A <code>File</code> object that can be written to as if it were the file specified.
	 *         It initially points to an empty, zero-byte file.
	 * @throws IOException if the transactional file cannot be created.
	 */
	public File getTransactionFile(File file) throws IOException {
		if (mStarted == 0) {
			throw new IllegalStateException(Msgs.NO_TRANSACTION_IN_PROGRESS);
		} else if (file == null) {
			throw new IllegalArgumentException(Msgs.MAY_NOT_BE_NULL);
		} else if (file.isDirectory()) {
			throw new IllegalArgumentException(Msgs.MAY_NOT_BE_DIRECTORY);
		}

		File transFile = mFiles.get(file);
		if (transFile == null) {
			transFile = File.createTempFile(".trn", null, file.getParentFile()); //$NON-NLS-1$
			mFiles.put(file, transFile);
		}
		return transFile;
	}
}
