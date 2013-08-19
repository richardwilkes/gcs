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

import com.trollworks.toolkit.utility.TKDebug;

import java.io.BufferedReader;
import java.io.File;
import java.io.IOException;
import java.io.InputStream;
import java.io.InputStreamReader;
import java.io.OutputStream;
import java.io.OutputStreamWriter;

/** A wrapper for controlling another process. */
public class TKProcess implements Runnable {
	/** The result value if the process watching thread was interrupted. */
	public static final int		INTERRUPTED		= Integer.MIN_VALUE;
	/** The result value if the process watching thread has not completed. */
	public static final int		NOT_FINISHED	= INTERRUPTED + 1;
	private Process				mProcess;
	private int					mResult;
	private ProcessInputThread	mOutputThread;
	private ProcessInputThread	mErrorThread;
	private OutputStreamWriter	mInputStream;

	/**
	 * Creates a new process instance.
	 * <p>
	 * The underlying process is created using {@link Runtime#exec(String)}.
	 * 
	 * @param command A system command.
	 * @throws IOException if an i/o error occurs.
	 */
	public TKProcess(String command) throws IOException {
		mProcess = Runtime.getRuntime().exec(command);
		init();
	}

	/**
	 * Creates a new process instance.
	 * <p>
	 * The underlying process is created using {@link Runtime#exec(String,String[])}.
	 * 
	 * @param command A system command.
	 * @param envArray An array of strings, each element of which has environment variable settings
	 *            in the format <code>name=value</code>.
	 * @throws IOException if an i/o error occurs.
	 */
	public TKProcess(String command, String[] envArray) throws IOException {
		mProcess = Runtime.getRuntime().exec(command, envArray);
		init();
	}

	/**
	 * Creates a new process instance.
	 * <p>
	 * The underlying process is created using {@link Runtime#exec(String,String[],File)}.
	 * 
	 * @param command A system command.
	 * @param envArray An array of strings, each element of which has environment variable settings
	 *            in the format <code>name=value</code>.
	 * @param workingDir The working directory of the subprocess, or <code>null</code> if the
	 *            subprocess should inherit the working directory of the current process.
	 * @throws IOException if an i/o error occurs.
	 */
	public TKProcess(String command, String[] envArray, File workingDir) throws IOException {
		mProcess = Runtime.getRuntime().exec(command, envArray, workingDir);
		init();
	}

	/**
	 * Creates a new process instance.
	 * <p>
	 * The underlying process is created using {@link Runtime#exec(String,String[],File)}.
	 * 
	 * @param command A system command.
	 * @param workingDir The working directory of the subprocess, or <code>null</code> if the
	 *            subprocess should inherit the working directory of the current process.
	 * @throws IOException if an i/o error occurs.
	 */
	public TKProcess(String command, File workingDir) throws IOException {
		mProcess = Runtime.getRuntime().exec(command, null, workingDir);
		init();
	}

	/**
	 * Creates a new process instance.
	 * <p>
	 * The underlying process is created using {@link Runtime#exec(String[])}.
	 * 
	 * @param cmdArray The array containing the command to call and its arguments.
	 * @throws IOException if an i/o error occurs.
	 */
	public TKProcess(String[] cmdArray) throws IOException {
		mProcess = Runtime.getRuntime().exec(cmdArray);
		init();
	}

	/**
	 * Creates a new process instance.
	 * <p>
	 * The underlying process is created using {@link Runtime#exec(String[],String[])}.
	 * 
	 * @param cmdArray The array containing the command to call and its arguments.
	 * @param envArray An array of strings, each element of which has environment variable settings
	 *            in the format <code>name=value</code>.
	 * @throws IOException if an i/o error occurs.
	 */
	public TKProcess(String[] cmdArray, String[] envArray) throws IOException {
		mProcess = Runtime.getRuntime().exec(cmdArray, envArray);
		init();
	}

	/**
	 * Creates a new process instance.
	 * <p>
	 * The underlying process is created using {@link Runtime#exec(String[],String[],File)}.
	 * 
	 * @param cmdArray The array containing the command to call and its arguments.
	 * @param workingDir The working directory of the subprocess, or <code>null</code> if the
	 *            subprocess should inherit the working directory of the current process.
	 * @throws IOException if an i/o error occurs.
	 */
	public TKProcess(String[] cmdArray, File workingDir) throws IOException {
		mProcess = Runtime.getRuntime().exec(cmdArray, null, workingDir);
		init();
	}

	/**
	 * Creates a new process instance.
	 * <p>
	 * The underlying process is created using {@link Runtime#exec(String[],String[],File)}.
	 * 
	 * @param cmdArray The array containing the command to call and its arguments.
	 * @param envArray An array of strings, each element of which has environment variable settings
	 *            in the format <code>name=value</code>.
	 * @param workingDir The working directory of the subprocess, or <code>null</code> if the
	 *            subprocess should inherit the working directory of the current process.
	 * @throws IOException if an i/o error occurs.
	 */
	public TKProcess(String[] cmdArray, String[] envArray, File workingDir) throws IOException {
		mProcess = Runtime.getRuntime().exec(cmdArray, envArray, workingDir);
		init();
	}

	/** Closes the process input stream. */
	public void closeInputStream() {
		try {
			mInputStream.close();
		} catch (Exception exception) {
			assert false : TKDebug.throwableToString(exception);
		}
	}

	/** @return The exit value for this process. */
	public synchronized int getResult() {
		return mResult;
	}

	/** @return <code>true</code> if the process has exited. */
	public boolean hasExited() {
		return getResult() != NOT_FINISHED;
	}

	private void init() {
		Thread thread;

		mResult = NOT_FINISHED;
		mOutputThread = new ProcessInputThread(mProcess.getInputStream());
		mErrorThread = new ProcessInputThread(mProcess.getErrorStream());
		mInputStream = new OutputStreamWriter(mProcess.getOutputStream());
		thread = new Thread(this, "ProcessWatcher"); //$NON-NLS-1$
		thread.setPriority(Thread.NORM_PRIORITY);
		thread.start();
		mOutputThread.start();
		mErrorThread.start();
	}

	/**
	 * Read a line of input.
	 * 
	 * @param useStdErr Pass in <code>true</code> to use the standard error stream for reading,
	 *            <code>false</code> to use the standard output stream for reading.
	 * @return The line of input or <code>null</code> if none was ready.
	 */
	public String read(boolean useStdErr) {
		ProcessInputThread thread = useStdErr ? mErrorThread : mOutputThread;

		if (thread.isLineAvailable()) {
			return thread.getLine();
		}
		return null;
	}

	/**
	 * Read the contents of the stream buffer and clear the buffer.
	 * 
	 * @param useStdErr Pass in <code>true</code> to use the standard error stream for reading,
	 *            <code>false</code> to use the standard output stream for reading.
	 * @return The contents of the stream buffer.
	 */
	public String getBuffer(boolean useStdErr) {
		ProcessInputThread thread = useStdErr ? mErrorThread : mOutputThread;

		return thread.getBuffer();
	}

	/**
	 * @param useStdErr Whether to use <code>stderr</code> or <code>stdout</code>.
	 * @param suffix The suffix to check for.
	 * @return Whether or not the current buffer ends with the specified suffix.
	 */
	public boolean endsWith(boolean useStdErr, String suffix) {
		return (useStdErr ? mErrorThread : mOutputThread).endsWith(suffix);
	}

	/**
	 * Redirects <code>stderr</code> or <code>stdout</code> to the specified output stream.
	 * 
	 * @param useStdErr Pass in <code>true</code> to redirect the standard error stream,
	 *            <code>false</code> to redirect the standard output stream.
	 * @param stream The stream to use for output.
	 */
	public void redirectOutput(boolean useStdErr, OutputStream stream) {
		(useStdErr ? mErrorThread : mOutputThread).redirect(stream);
	}

	/** Calls <code>waitForProcess</code>. */
	public void run() {
		waitForProcess();
	}

	/**
	 * Sets the exit value for this process.
	 * 
	 * @param result The value to set.
	 */
	protected synchronized void setResult(int result) {
		mResult = result;
	}

	/** Waits for the process to die, then sets the exit result. */
	public void waitForProcess() {
		try {
			int result = mProcess.waitFor();

			// Wait for i/o to finish...
			mErrorThread.join();
			mOutputThread.join();

			setResult(result);
		} catch (InterruptedException ie) {
			setResult(INTERRUPTED);
			assert false : TKDebug.throwableToString(ie);
		}
	}

	/**
	 * Write a string to the process's input.
	 * 
	 * @param buffer The string to write.
	 * @return <code>true</code> if the write was successful, <code>false</code> if not.
	 */
	public boolean write(String buffer) {
		if (!hasExited()) {
			try {
				mInputStream.write(buffer);
				mInputStream.flush();
				return true;
			} catch (IOException ioe) {
				assert false : TKDebug.throwableToString(ioe);
			}
		}
		return false;
	}

	/**
	 * Reads the specified input stream until its end and writes it to the process input stream.
	 * 
	 * @param in The stream to pipe through.
	 * @return <code>true</code> if the write was successful, <code>false</code> if not.
	 */
	public boolean writeUntilEnd(InputStream in) {
		try {
			BufferedReader reader = new BufferedReader(new InputStreamReader(in));
			String lineEnding = System.getProperty("line.separator"); //$NON-NLS-1$
			String line;

			while ((line = reader.readLine()) != null) {
				if (!write(line + lineEnding)) {
					return false;
				}
			}

			if (!hasExited()) {
				mInputStream.close();
				return true;
			}
		} catch (IOException ioe) {
			assert false : TKDebug.throwableToString(ioe);
		}
		return false;
	}

	/**
	 * Processes the input stream of a process.
	 */
	private class ProcessInputThread extends Thread {
		private InputStream		mStream;
		private OutputStream	mRedirect;
		private StringBuilder	mBuffer;

		/**
		 * Creates a new input stream processor.
		 * 
		 * @param stream The input stream to process.
		 */
		ProcessInputThread(InputStream stream) {
			super("ProcessOutput"); //$NON-NLS-1$
			setPriority(NORM_PRIORITY);
			mStream = stream;
			mRedirect = null;
			mBuffer = new StringBuilder(1024);
		}

		/** @return A line of input and removes it from the underlying buffer. */
		String getLine() {
			String buffer = ""; //$NON-NLS-1$

			synchronized (mBuffer) {
				int length = mBuffer.length();

				for (int i = 0; i < length; i++) {
					char ch = mBuffer.charAt(i);

					if (ch == '\n' || ch == '\r') {
						length = i;
						break;
					}
				}

				buffer = mBuffer.substring(0, length);
				mBuffer.delete(0, length);
				if (mBuffer.length() > 0) {
					mBuffer.deleteCharAt(0);
				}
			}
			return buffer;
		}

		/**
		 * @return The entire contents of the underlying buffer. The underlying will be empty after
		 *         this call.
		 */
		String getBuffer() {
			String buffer = ""; //$NON-NLS-1$

			synchronized (mBuffer) {
				buffer = mBuffer.toString();
				mBuffer.setLength(0);
			}
			return buffer;
		}

		/** @return Whether or not a line of data is available. */
		boolean isLineAvailable() {
			boolean avail = false;

			synchronized (mBuffer) {
				int length = mBuffer.length();

				for (int i = 0; i < length; i++) {
					char ch = mBuffer.charAt(i);

					if (ch == '\n' || ch == '\r') {
						avail = true;
						break;
					}
				}

				if (!avail) {
					if (length > 0 && hasExited()) {
						avail = true;
					}
				}
			}
			return avail;
		}

		/**
		 * @param suffix The suffix to check for.
		 * @return Whether or not the underlying buffer contents ends with the specified suffix.
		 */
		boolean endsWith(String suffix) {
			boolean avail = false;

			synchronized (mBuffer) {
				int length = mBuffer.length();
				int suffixLength = suffix.length();

				if (length >= suffixLength) {
					int bufferPos = length - suffixLength;
					avail = true;
					for (int i = 0; i < suffixLength; i++, bufferPos++) {

						if (mBuffer.charAt(bufferPos) != suffix.charAt(i)) {
							avail = false;
							break;
						}
					}
				}
			}

			return avail;
		}

		/**
		 * Redirects all input to the specified output stream.
		 * 
		 * @param stream The stream to write the input to.
		 */
		void redirect(OutputStream stream) {
			synchronized (mBuffer) {
				mRedirect = stream;
				if (stream != null && mBuffer.length() > 0) {
					try {
						mRedirect.write(mBuffer.toString().getBytes());
					} catch (Exception exception) {
						assert false : TKDebug.throwableToString(exception);
					}
					mBuffer.setLength(0);
				}
			}
		}

		@Override public void run() {
			try {
				while (true) {
					int value = mStream.read();

					if (value == -1) {
						break;
					}

					synchronized (mBuffer) {
						if (mRedirect == null) {
							mBuffer.append((char) value);
						} else {
							mRedirect.write((char) value);
						}
					}
				}
			} catch (IOException ioe) {
				assert false : TKDebug.throwableToString(ioe);
			}
		}
	}
}
