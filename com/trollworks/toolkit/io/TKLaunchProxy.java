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

import com.trollworks.toolkit.io.conduit.TKConduit;
import com.trollworks.toolkit.io.conduit.TKConduitMessage;
import com.trollworks.toolkit.io.conduit.TKConduitReceiver;
import com.trollworks.toolkit.utility.TKDebug;
import com.trollworks.toolkit.utility.TKGraphics;
import com.trollworks.toolkit.window.TKOpenManager;

import java.awt.EventQueue;
import java.io.File;
import java.util.ArrayList;
import java.util.StringTokenizer;

/**
 * Provides the ability for an app to be launched from different terminals and still use only a
 * single instance.
 */
public class TKLaunchProxy implements TKConduitReceiver {
	private static final String		AT					= "@";				//$NON-NLS-1$
	private static final String		SPACE				= " ";				//$NON-NLS-1$
	private static final String		COMMA				= ",";				//$NON-NLS-1$
	private static final String		AT_MARKER			= "@!";			//$NON-NLS-1$
	private static final String		SPACE_MARKER		= "@%";			//$NON-NLS-1$
	private static final String		COMMA_MARKER		= "@#";			//$NON-NLS-1$
	private static final String		LAUNCH_ID			= "Launched";		//$NON-NLS-1$
	private static final String		TOOK_OVER_FOR_ID	= "TookOverFor";	//$NON-NLS-1$
	private static TKLaunchProxy	INSTANCE			= null;
	private TKConduit				mConduit;
	private String					mConduitID;
	private int						mProductID;
	private long					mTimeStamp;
	private boolean					mReady;
	private ArrayList<File>			mFiles;

	/** @return The single instance of the app launch proxy. */
	public synchronized static TKLaunchProxy getInstance() {
		assert INSTANCE != null;
		return INSTANCE;
	}

	/**
	 * Configures the one and only instance that may exist. It is an error to call this method more
	 * than once.
	 * 
	 * @param brokerID The ID to use when creating the communications broker.
	 * @param productID The product ID to use for filtering messages.
	 * @param files The files, if any, that should be passed on to another instance of the app that
	 *            may already be running.
	 */
	public synchronized static void configure(String brokerID, int productID, File[] files) {
		if (INSTANCE == null) {
			INSTANCE = new TKLaunchProxy(brokerID, productID, files);
			try {
				// Give it a chance to terminate this run...
				Thread.sleep(1500);
			} catch (Exception exception) {
				assert false : TKDebug.throwableToString(exception);
			}
		} else {
			assert false : "Can only call configure once."; //$NON-NLS-1$
		}
	}

	private TKLaunchProxy(String brokerID, int productID, File[] files) {
		StringBuilder buffer = new StringBuilder();

		mConduitID = brokerID;
		mProductID = productID;
		mFiles = new ArrayList<File>();
		mTimeStamp = System.currentTimeMillis();
		buffer.append(LAUNCH_ID);
		buffer.append(' ');
		buffer.append(mProductID);
		buffer.append(' ');
		buffer.append(mTimeStamp);
		buffer.append(' ');
		if (files != null) {
			for (int i = 0; i < files.length; i++) {
				if (i != 0) {
					buffer.append(',');
				}
				buffer.append(files[i].getAbsolutePath().replaceAll(AT, AT_MARKER).replaceAll(SPACE, SPACE_MARKER).replaceAll(COMMA, COMMA_MARKER));
			}
		}
		mConduit = new TKConduit(this, false);
		mConduit.send(new TKConduitMessage(mConduitID, buffer.toString()));
	}

	/**
	 * Sets whether this application is ready to take over responsibility for other copies being
	 * launched.
	 * 
	 * @param ready Whether the application is ready or not.
	 */
	public void setReady(boolean ready) {
		mReady = ready;
	}

	public void conduitMessageReceived(TKConduitMessage msg) {
		StringTokenizer tokenizer = new StringTokenizer(msg.getMessage(), SPACE);

		if (tokenizer.hasMoreTokens()) {
			String token = tokenizer.nextToken();

			if (mReady && LAUNCH_ID.equals(token)) {
				if (tokenizer.hasMoreTokens() && getInteger(tokenizer) == mProductID) {
					if (tokenizer.hasMoreTokens()) {
						long timeStamp = getLong(tokenizer);

						if (timeStamp != mTimeStamp) {
							mConduit.send(new TKConduitMessage(mConduitID, TOOK_OVER_FOR_ID + ' ' + mProductID + ' ' + timeStamp));
							TKGraphics.forceAppToFront();
							if (tokenizer.hasMoreTokens()) {
								tokenizer = new StringTokenizer(tokenizer.nextToken(), COMMA);
								while (tokenizer.hasMoreTokens()) {
									synchronized (mFiles) {
										mFiles.add(new File(tokenizer.nextToken().replaceAll(COMMA_MARKER, COMMA).replaceAll(SPACE_MARKER, SPACE).replaceAll(AT_MARKER, AT)));
									}
								}
								synchronized (mFiles) {
									if (!mFiles.isEmpty()) {
										TKOpenManager.openFiles(null, mFiles.toArray(new File[0]), true);
										mFiles.clear();
									}
								}
							} else {
								EventQueue.invokeLater(new Runnable() {
									public void run() {
										TKOpenManager.promptForOpen();
									}
								});
							}
						}
					}
				}
			} else if (TOOK_OVER_FOR_ID.equals(token)) {
				if (tokenizer.hasMoreTokens() && getInteger(tokenizer) == mProductID) {
					if (tokenizer.hasMoreTokens()) {
						if (getLong(tokenizer) == mTimeStamp) {
							System.exit(0);
						}
					}
				}
			}
		}
	}

	private int getInteger(StringTokenizer tokenizer) {
		try {
			return Integer.parseInt(tokenizer.nextToken().trim());
		} catch (Exception exception) {
			return -1;
		}
	}

	private long getLong(StringTokenizer tokenizer) {
		try {
			return Long.parseLong(tokenizer.nextToken().trim());
		} catch (Exception exception) {
			return -1;
		}
	}

	public String getConduitMessageIDFilter() {
		return mConduitID;
	}

	public String getConduitMessageUserFilter() {
		return System.getProperty("user.name"); //$NON-NLS-1$
	}
}
