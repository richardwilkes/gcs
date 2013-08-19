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

package com.trollworks.gcs.utility.io.conduit;

import com.trollworks.gcs.utility.Debug;

import java.awt.EventQueue;
import java.io.DataInputStream;
import java.io.DataOutputStream;
import java.io.IOException;
import java.net.InetAddress;
import java.net.InetSocketAddress;
import java.net.ServerSocket;
import java.net.Socket;
import java.net.UnknownHostException;
import java.util.ArrayList;

/** Provides a conduit through which messages from external processes can be recieved. */
public class Conduit implements Runnable {
	/** The default port used by the conduit. */
	public static final int		DEFAULT_PORT	= 13321;
	private InetSocketAddress	mSocketAddress;
	private Server				mServer;
	private Socket				mSocket;
	private DataInputStream		mInput;
	private DataOutputStream	mOutput;
	private ConduitReceiver	mReceiver;
	private boolean				mOnEventThread;
	private Thread				mReceptionThread;
	private String				mUserFilter;
	private String				mIDFilter;

	/**
	 * Creates a new conduit with the default port on the loopback address.
	 * 
	 * @param receiver The object that wants the messages from this conduit.
	 * @param onEventThread Pass in <code>true</code> to receive the messages on the event thread.
	 */
	public Conduit(ConduitReceiver receiver, boolean onEventThread) {
		this(null, receiver, onEventThread);
	}

	/**
	 * Creates a new conduit with the specified port on the loopback address.
	 * 
	 * @param port The port to communicate on.
	 * @param receiver The object that wants the messages from this conduit.
	 * @param onEventThread Pass in <code>true</code> to receive the messages on the event thread.
	 */
	public Conduit(int port, ConduitReceiver receiver, boolean onEventThread) {
		this(new InetSocketAddress(getLoopBackAddress(), port), receiver, onEventThread);
	}

	/**
	 * Creates a new conduit with the specified port on the loopback address.
	 * 
	 * @param address The address to communicate on.
	 * @param port The port to communicate on.
	 * @param receiver The object that wants the messages from this conduit.
	 * @param onEventThread Pass in <code>true</code> to receive the messages on the event thread.
	 */
	public Conduit(InetAddress address, int port, ConduitReceiver receiver, boolean onEventThread) {
		this(new InetSocketAddress(address, port), receiver, onEventThread);
	}

	/**
	 * Creates a new conduit with the specified socket address.
	 * 
	 * @param socketAddress The socket address to use.
	 * @param receiver The object that wants the messages from this conduit.
	 * @param onEventThread Pass in <code>true</code> to receive the messages on the event thread.
	 */
	public Conduit(InetSocketAddress socketAddress, ConduitReceiver receiver, boolean onEventThread) {
		if (socketAddress == null) {
			socketAddress = new InetSocketAddress(getLoopBackAddress(), DEFAULT_PORT);
		}
		mSocketAddress = socketAddress;
		mReceiver = receiver;
		mOnEventThread = onEventThread;
		mReceptionThread = new Thread(this, Conduit.class.getSimpleName() + '@' + mSocketAddress);
		mUserFilter = mReceiver.getConduitMessageUserFilter();
		mIDFilter = mReceiver.getConduitMessageIDFilter();
		reconnect();
		mReceptionThread.setPriority(Thread.NORM_PRIORITY);
		mReceptionThread.setDaemon(true);
		mReceptionThread.start();
	}

	private static final InetAddress getLoopBackAddress() {
		try {
			return InetAddress.getByName(null);
		} catch (UnknownHostException uhe) {
			// This can't occur, as the loopback address is always valid.
			assert false : Debug.throwableToString(uhe);
			return null;
		}
	}

	private void reconnect() {
		shutdownSocket();
		if (mServer != null) {
			mServer.shutdown();
			mServer = null;
		}

		while (true) {
			try {
				mServer = new Server(mSocketAddress);
				mServer.setDaemon(true);
				mServer.start();
			} catch (Exception exception) {
				// Someone else is already the server, just start a client.
			}

			mSocket = new Socket();
			try {
				mSocket.connect(mSocketAddress);
				mInput = new DataInputStream(mSocket.getInputStream());
				mOutput = new DataOutputStream(mSocket.getOutputStream());
				return;
			} catch (Exception ex2) {
				// The server is no longer around or hasn't quite started up
				// yet, so loop and try again.
				shutdownSocket();
			}
		}
	}

	/**
	 * Sends a message to all clients connected to the conduit.
	 * 
	 * @param msg The message.
	 */
	public void send(ConduitMessage msg) {
		while (true) {
			try {
				msg.send(mOutput);
				return;
			} catch (Exception exception) {
				reconnect();
			}
		}
	}

	public void run() {
		while (true) {
			try {
				ConduitMessage msg = new ConduitMessage(mInput);

				if ((mUserFilter == null || mUserFilter.equals(msg.getUser())) && (mIDFilter == null || mIDFilter.equals(msg.getID()))) {
					if (mOnEventThread) {
						msg.setReceiver(mReceiver);
						EventQueue.invokeLater(msg);
					} else {
						mReceiver.conduitMessageReceived(msg);
					}
				}
			} catch (Exception exception) {
				reconnect();
			}
		}
	}

	private void shutdownSocket() {
		if (mSocket != null) {
			try {
				mSocket.close();
			} catch (Exception exception) {
				// Ignore.
			}
			mSocket = null;
			mInput = null;
			mOutput = null;
		}
	}

	private class Server extends Thread {
		private ServerSocket		mServerSocket;
		private ArrayList<Client>	mClients;
		private Object				mSendLock;
		private int					mClientCounter;

		/**
		 * Creates a new conduit message server.
		 * 
		 * @param socketAddress The socket address to attach to.
		 * @throws IOException if the server socket cannot be created.
		 */
		Server(InetSocketAddress socketAddress) throws IOException {
			super(Conduit.class.getSimpleName() + '$' + Server.class.getSimpleName() + '@' + socketAddress);
			setPriority(NORM_PRIORITY);
			setDaemon(true);
			mServerSocket = new ServerSocket(socketAddress.getPort(), 0, socketAddress.getAddress());
			mClients = new ArrayList<Client>();
			mSendLock = new Object();
		}

		/** @return The next client counter. */
		int getNextClientCounter() {
			return ++mClientCounter;
		}

		/** @return The server socket. */
		ServerSocket getServerSocket() {
			return mServerSocket;
		}

		/**
		 * Handles accepting new incoming connections and starting client processors for each.
		 */
		@Override public void run() {
			Client[] clients;

			try {
				while (true) {
					Socket socket = mServerSocket.accept();

					try {
						Client client = new Client(socket);

						client.setDaemon(true);
						client.start();
						synchronized (mClients) {
							mClients.add(client);
						}
					} catch (IOException ioe) {
						// The client died an early death... ignore it.
						socket.close();
					}
				}
			} catch (Exception exception) {
				shutdown();
			}

			synchronized (mClients) {
				clients = mClients.toArray(new Client[0]);
				mClients.clear();
			}

			for (Client element : clients) {
				element.shutdown();
			}
		}

		/**
		 * Removes a client from the list of clients being served.
		 * 
		 * @param client The client to remove.
		 */
		void remove(Client client) {
			synchronized (mClients) {
				mClients.remove(client);
			}
		}

		/**
		 * Sends a message to all connected clients.
		 * 
		 * @param msg The message to send.
		 */
		void send(ConduitMessage msg) {
			Client[] clients;

			synchronized (mClients) {
				clients = mClients.toArray(new Client[0]);
			}

			synchronized (mSendLock) {
				for (Client element : clients) {
					element.send(msg);
				}
			}
		}

		/** Shuts down this communication server. */
		synchronized void shutdown() {
			try {
				mServerSocket.close();
			} catch (Exception exception) {
				assert false : Debug.throwableToString(exception);
			}
		}

		private class Client extends Thread {
			private Socket				mClientSocket;
			private DataInputStream		mClientInput;
			private DataOutputStream	mClientOutput;

			/**
			 * Creates a new client processor for the server.
			 * 
			 * @param socket The socket containing the client connection.
			 * @throws IOException if the socket's i/o streams cannot be retrieved.
			 */
			Client(Socket socket) throws IOException {
				super(Conduit.class.getSimpleName() + '$' + Client.class.getSimpleName() + '#' + getNextClientCounter() + '@' + getServerSocket().getLocalSocketAddress());
				setPriority(NORM_PRIORITY);
				setDaemon(true);
				mClientSocket = socket;
				mClientInput = new DataInputStream(socket.getInputStream());
				mClientOutput = new DataOutputStream(socket.getOutputStream());
			}

			@Override public void run() {
				try {
					while (true) {
						Server.this.send(new ConduitMessage(mClientInput));
					}
				} catch (Exception exception) {
					// An exception here can be ignored, as its just the connection
					// going away, which we deal with later.
				}

				shutdown();
			}

			/**
			 * Sends a message to the client.
			 * 
			 * @param msg The message.
			 */
			void send(ConduitMessage msg) {
				try {
					msg.send(mClientOutput);
				} catch (Exception exception) {
					shutdown();
				}
			}

			/** Shuts down this client processor. */
			synchronized void shutdown() {
				try {
					mClientSocket.close();
				} catch (Exception exception) {
					assert false : Debug.throwableToString(exception);
				}
				remove(this);
			}
		}
	}
}
