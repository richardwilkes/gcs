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

package com.trollworks.toolkit.io.conduit;

import java.io.DataInputStream;
import java.io.DataOutputStream;
import java.io.IOException;

/** The message sent between between processes using a conduit. */
public class TKConduitMessage implements Runnable {
	// No repeating characters allowed!
	private static final byte[]	ID	= { '#', 'W', 'i', 'l', 'k', 'e', 's', '!' };
	private String				mUser;
	private String				mID;
	private String				mMessage;
	private TKConduitReceiver	mReceiver;

	/**
	 * Creates a new conduit message.
	 * 
	 * @param id An ID that clients will use to filter reception of messages.
	 * @param message The message.
	 */
	public TKConduitMessage(String id, String message) {
		mUser = System.getProperty("user.name"); //$NON-NLS-1$
		mID = id;
		mMessage = message;
	}

	/**
	 * Creates a new conduit message by reading it in from the specified stream.
	 * 
	 * @param stream The stream to read the message from.
	 * @throws IOException if the underlying data stream throws an exception.
	 */
	public TKConduitMessage(DataInputStream stream) throws IOException {
		int i = 0;

		while (i < ID.length) {
			byte value = stream.readByte();

			if (value == ID[i]) {
				i++;
			} else if (value == ID[0]) {
				i = 1;
			} else {
				i = 0;
			}
		}

		mUser = stream.readUTF();
		mID = stream.readUTF();
		mMessage = stream.readUTF();
	}

	/**
	 * Writes the message to a data output stream.
	 * 
	 * @param stream The stream to write to.
	 * @throws IOException if the stream throws an exception.
	 */
	void send(DataOutputStream stream) throws IOException {
		stream.write(ID);
		stream.writeUTF(getUser());
		stream.writeUTF(getID());
		stream.writeUTF(getMessage());
		stream.flush();
	}

	/** @param receiver The message receiver. */
	void setReceiver(TKConduitReceiver receiver) {
		mReceiver = receiver;
	}

	public void run() {
		mReceiver.conduitMessageReceived(this);
	}

	/** @return The user. */
	public String getUser() {
		return mUser;
	}

	/** @return The message ID. */
	public String getID() {
		return mID;
	}

	/** @return The message. */
	public String getMessage() {
		return mMessage;
	}

	@Override public String toString() {
		return "[" + getUser() + " : " + getID() + "] " + getMessage(); //$NON-NLS-1$ //$NON-NLS-2$ //$NON-NLS-3$
	}
}
