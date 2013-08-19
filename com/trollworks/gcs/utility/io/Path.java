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

package com.trollworks.gcs.utility.io;

import java.io.File;
import java.util.StringTokenizer;

/** Provides standard file path manipulation facilities. */
public class Path {
	/**
	 * Ensures that the passed in string has the specified extension on it.
	 * 
	 * @param name The name to process.
	 * @param extension The desired extension.
	 * @return A new string with the specified extension.
	 */
	public static final String enforceExtension(String name, String extension) {
		return enforceExtension(name, extension, false);
	}

	/**
	 * Ensures that the passed in string has an extension on it.
	 * 
	 * @param name The name to process.
	 * @param extension The desired extension.
	 * @param onlyIfNoExtension Pass <code>true</code> if extensions other than the one passed in
	 *            are acceptable.
	 * @return A new string with the specified extension.
	 */
	public static final String enforceExtension(String name, String extension, boolean onlyIfNoExtension) {
		name = name.replace('\\', '/');
		if (extension.charAt(0) != '.') {
			extension = '.' + extension;
		}
		if (!name.endsWith(extension)) {
			int lastDot = name.lastIndexOf('.');

			if (name.lastIndexOf('/') > lastDot) {
				lastDot = -1;
			}
			if (lastDot == -1) {
				return name + extension;
			}
			if (lastDot == name.length() - 1) {
				return name + extension.substring(1);
			}
			if (!onlyIfNoExtension) {
				return name.substring(0, lastDot) + extension;
			}
		}
		return name;
	}

	/**
	 * @param fullPathOne The first full path.
	 * @param fullPathTwo The second full path.
	 * @return The common root path of two full paths, or <code>null</code> if there is none.
	 */
	public static final String getCommonRoot(String fullPathOne, String fullPathTwo) {
		int i;
		int len1;
		int len2;
		int max;

		if (fullPathOne == null) {
			len1 = 0;
		} else {
			fullPathOne = fullPathOne.replace('\\', '/');
			len1 = fullPathOne.length();
		}

		if (fullPathTwo == null) {
			len2 = 0;
		} else {
			fullPathTwo = fullPathTwo.replace('\\', '/');
			len2 = fullPathTwo.length();
		}

		max = len1 > len2 ? len2 : len1;

		for (i = 0; i < max; i++) {
			if (fullPathOne.charAt(i) != fullPathTwo.charAt(i)) {
				i--;
				break;
			}
		}

		if (i == max) {
			i = max - 1;
		}
		while (i >= 0 && fullPathOne.charAt(i) != '/') {
			i--;
		}
		if (i < 0) {
			return null;
		}
		return fullPathOne.substring(0, i + 1);
	}

	/**
	 * @param path The path to operate on.
	 * @return The extension of the path name, including the initial ".".
	 */
	public static final String getExtension(String path) {
		int index;

		if (path == null) {
			return null;
		}

		path = path.replace('\\', '/');
		index = path.lastIndexOf('.');
		if (index == -1 || index == path.length() - 1 || index < path.lastIndexOf('/')) {
			return ""; //$NON-NLS-1$
		}

		return path.substring(index);
	}

	/**
	 * @param path The path to operate on.
	 * @return A file based on the specified path.
	 */
	public static final File getFile(String path) {
		return new File(getPathForPlatform(path));
	}

	/**
	 * @param baseFullPath The base path.
	 * @param relativePath The path relative to the base path.
	 * @return A file based on the specified paths. If the relative path passed in is actually a
	 *         full path, the original relative path is used.
	 */
	public static final File getFile(String baseFullPath, String relativePath) {
		return new File(getPathForPlatform(getFullPath(baseFullPath, relativePath)));
	}

	/**
	 * @param baseFullPath The base path.
	 * @param relativePath The path relative to the base path.
	 * @return A full path based on the specified base path. If the relative path passed in is
	 *         actually a full path, the original relative path is returned.
	 */
	public static final String getFullPath(String baseFullPath, String relativePath) {
		String result;

		if (relativePath != null) {
			if (baseFullPath != null) {
				baseFullPath = baseFullPath.replace('\\', '/');
				relativePath = relativePath.replace('\\', '/');
				if (isFullPath(relativePath)) {
					result = relativePath;
				} else if (baseFullPath.endsWith("/")) { //$NON-NLS-1$
					result = baseFullPath + relativePath;
				} else {
					result = baseFullPath + "/" + relativePath; //$NON-NLS-1$
				}
			} else {
				result = relativePath.replace('\\', '/');
			}
		} else if (baseFullPath != null) {
			result = baseFullPath;
		} else {
			result = "./"; //$NON-NLS-1$
		}

		if (result.startsWith("./") || result.startsWith("../")) { //$NON-NLS-1$ //$NON-NLS-2$
			return getFullPath(getFullPath(new File(".")), result); //$NON-NLS-1$
		}

		return normalizeFullPath(result);
	}

	/**
	 * @param file The file to operate on.
	 * @return A full path from a file.
	 */
	public static final String getFullPath(File file) {
		if (file != null) {
			return normalizeFullPath(file.getAbsolutePath().replace('\\', '/'));
		}
		return null;
	}

	/**
	 * @param path The path to process.
	 * @return The leaf portion of the path name (everything to the right of the last path
	 *         separator).
	 */
	public static final String getLeafName(String path) {
		return getLeafName(path, true);
	}

	/**
	 * @param path The path to process.
	 * @param includeExtension Pass in <code>true</code> to leave the extension on the name or
	 *            <code>false</code> to strip it off.
	 * @return The leaf portion of the path name (everything to the right of the last path
	 *         separator).
	 */
	public static final String getLeafName(String path, boolean includeExtension) {
		if (path != null) {
			int index;

			path = path.replace('\\', '/');
			index = path.lastIndexOf('/');
			if (index != -1) {
				if (index == path.length() - 1) {
					return ""; //$NON-NLS-1$
				}
				path = path.substring(index + 1);
			}

			if (!includeExtension) {
				index = path.lastIndexOf('.');
				if (index != -1) {
					path = path.substring(0, index);
				}
			}
			return path;
		}
		return null;
	}

	/**
	 * @param path The path to operate on.
	 * @return The parent portion of the path name (everything to the left of the last path
	 *         separator, plus the path separator itself).
	 */
	public static final String getParent(String path) {
		return getParent(path, true);
	}

	/**
	 * @param path The path to operate on.
	 * @param includeTrailingSep Whether or not the trailing separator character should be included.
	 * @return The parent portion of the path name (everything to the left of the last path
	 *         separator, plus the path separator itself, if desired).
	 */
	public static final String getParent(String path, boolean includeTrailingSep) {
		int index;

		if (path == null) {
			return null;
		}

		path = path.replace('\\', '/');
		index = path.lastIndexOf('/');
		if (index == -1) {
			return ""; //$NON-NLS-1$
		}
		return path.substring(0, index + (includeTrailingSep ? 1 : 0));
	}

	/**
	 * @param path The path to operate on.
	 * @return A sanitized version of the path suitable for use with the native platform.
	 */
	public static final String getPathForPlatform(String path) {
		return path.replace('\\', '/').replace('/', File.separatorChar);
	}

	/**
	 * @param baseFullPath The base full path.
	 * @param targetFullPath The target full path.
	 * @return A relative path based on a specified full path. If this is not possible, the original
	 *         target full path is returned.
	 */
	public static final String getRelativePath(String baseFullPath, String targetFullPath) {
		if (baseFullPath != null) {
			baseFullPath = baseFullPath.replace('\\', '/');
		}
		if (targetFullPath != null) {
			targetFullPath = targetFullPath.replace('\\', '/');
		}

		String common = getCommonRoot(baseFullPath, targetFullPath);

		if (common != null) {
			if (common.equals(baseFullPath)) {
				return targetFullPath.substring(common.length());
			}

			StringBuilder buffer = new StringBuilder(targetFullPath.length());
			String remainder = baseFullPath.substring(common.length());
			int i = remainder.indexOf('/');

			while (i != -1) {
				buffer.append("../"); //$NON-NLS-1$
				i = remainder.indexOf('/', i + 1);
			}
			buffer.append(targetFullPath.substring(common.length()));
			return buffer.toString();
		}
		return targetFullPath;
	}

	/**
	 * @param baseFullPath The base full path.
	 * @param file The target file.
	 * @return A relative path from a file based on a specified full path. If this is not possible,
	 *         the original target full path is returned.
	 */
	public static final String getRelativePath(String baseFullPath, File file) {
		return getRelativePath(baseFullPath, getFullPath(file));
	}

	/**
	 * @param path The path to operate on.
	 * @return <code>true</code> if the specified path is an absolute path.
	 */
	public static final boolean isFullPath(String path) {
		boolean isFullPath = false;

		if (path != null) {
			int length = path.length();

			if (length > 0) {
				char ch;

				path = path.replace('\\', '/');
				ch = path.charAt(0);

				if (ch == '/' || path.startsWith("//") || length > 1 && path.charAt(1) == ':' && (ch >= 'a' && ch <= 'z' || ch >= 'A' && ch <= 'Z')) { //$NON-NLS-1$
					isFullPath = true;
				}
			}
		}
		return isFullPath;
	}

	/**
	 * Normalizes full path names by resolving . and .. path portions.
	 * 
	 * @param path The path to operate on.
	 * @return The normalized path.
	 */
	public static final String normalizeFullPath(String path) {
		if (path != null) {
			int index;
			StringBuilder buffer;
			char ch;

			path = path.replace('\\', '/');

			do {
				index = path.indexOf("/./"); //$NON-NLS-1$
				if (index != -1) {
					buffer = new StringBuilder(path);
					buffer.delete(index, index + 2);
					path = buffer.toString();
				}
			} while (index != -1);

			do {
				index = path.indexOf("/../"); //$NON-NLS-1$
				if (index != -1) {
					int length = 3;

					buffer = new StringBuilder(path);

					while (index > 0) {
						ch = buffer.charAt(--index);
						length++;
						if (ch == '/') {
							break;
						} else if (ch == ':') {
							index++;
							length--;
							break;
						}
					}

					buffer.delete(index, index + length);
					path = buffer.toString();
				}
			} while (index != -1);

			if (path.endsWith("/.")) { //$NON-NLS-1$
				path = path.substring(0, path.length() - 2);
			}

			if (path.endsWith("/..")) { //$NON-NLS-1$
				index = path.length() - 3;

				while (index > 0) {
					ch = path.charAt(--index);
					if (ch == '/' || ch == ':') {
						break;
					}
				}

				path = path.substring(0, index);
			}

			if (path.length() > 1 && path.charAt(1) == ':') {
				path = Character.toUpperCase(path.charAt(0)) + ":" + path.substring(2); //$NON-NLS-1$
			}
		}

		return path;
	}

	/**
	 * Changes ".\src\foo.c" to "src/foo.c". Changes ".\foo.c" to "foo.c". Changes "./foo.c" to
	 * "foo.c". Changes "foo.c" to "foo.c". You get the idea.
	 * 
	 * @param path
	 * @return the simplified path
	 */
	public static String simplifyRelativePath(String path) {
		path = path.replace('\\', '/');
		if (path.startsWith("./")) { //$NON-NLS-1$
			path = path.substring(2);
		}
		return path;
	}

	/**
	 * Determines whether the specified path is viable as a command.
	 * 
	 * @param path The path to check.
	 * @param extraBinDir A directory to check for the executable, in addition to the standard
	 *            locations.
	 * @return The {@link File} representing the full path to the executable, or <code>null</code>.
	 */
	public static File isCommandPathViable(String path, File extraBinDir) {
		if (path != null) {
			File file = new File(path);

			if (file.isFile()) {
				return file;
			}

			if (path.indexOf(File.separatorChar) == -1) {
				if (extraBinDir != null) {
					file = new File(extraBinDir, path);
					if (file.isFile()) {
						return file;
					}
				}

				String cmdPath = System.getenv("PATH"); //$NON-NLS-1$
				if (cmdPath != null) {
					StringTokenizer tokenizer = new StringTokenizer(cmdPath, File.pathSeparator);

					while (tokenizer.hasMoreTokens()) {
						file = new File(tokenizer.nextToken(), path);
						if (file.isFile()) {
							return file;
						}
					}
				}
			}
		}
		return null;
	}
}
