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

package com.trollworks.toolkit.io.cmdline;

import com.trollworks.toolkit.text.TKTextUtility;
import com.trollworks.toolkit.utility.TKApp;
import com.trollworks.toolkit.utility.TKPlatform;

import java.io.File;
import java.text.MessageFormat;
import java.util.ArrayList;
import java.util.Collection;
import java.util.Collections;
import java.util.HashMap;
import java.util.HashSet;

/** Provides standardized command-line argument parsing. */
public class TKCmdLine {
	private static final TKCmdLineOption	HELP_OPTION		= new TKCmdLineOption(Msgs.HELP_DESCRIPTION, null, "h", "?", "help");	//$NON-NLS-1$ //$NON-NLS-2$ //$NON-NLS-3$
	private static final TKCmdLineOption	VERSION_OPTION	= new TKCmdLineOption(Msgs.VERSION_DESCRIPTION, null, "v", "version");	//$NON-NLS-1$ //$NON-NLS-2$
	private ArrayList<TKCmdLineData>		mData;
	private HashSet<TKCmdLineOption>		mUsedOptions;

	/**
	 * Creates a new {@link TKCmdLine}. If the command line arguments are invalid, or the version
	 * number or help is requested, then this constructor will cause the program to exit.
	 * 
	 * @param args The arguments passed to <code>main</code>.
	 * @param options Valid options.
	 */
	public TKCmdLine(String[] args, Collection<TKCmdLineOption> options) {
		this(args, options, null);
	}

	/**
	 * Creates a new {@link TKCmdLine}. If the command line arguments are invalid, or the version
	 * number or help is requested, then this constructor will cause the program to exit.
	 * 
	 * @param args The arguments passed to <code>main</code>.
	 * @param options Valid options.
	 * @param extraHelp Any text that you would like appended to the end of the help output.
	 */
	public TKCmdLine(String[] args, Collection<TKCmdLineOption> options, String extraHelp) {
		HashMap<String, TKCmdLineOption> map = new HashMap<String, TKCmdLineOption>();
		ArrayList<TKCmdLineOption> all = new ArrayList<TKCmdLineOption>(options);
		ArrayList<String> msgs = new ArrayList<String>();

		all.add(HELP_OPTION);
		all.add(VERSION_OPTION);

		for (TKCmdLineOption option : all) {
			for (String name : option.getNames()) {
				map.put(name, option);
			}
		}

		mData = new ArrayList<TKCmdLineData>();
		mUsedOptions = new HashSet<TKCmdLineOption>();

		for (int i = 0; i < args.length; i++) {
			String one = args[i];

			if (i == 0 && TKPlatform.isMacintosh() && args[i].startsWith("-psn_")) { //$NON-NLS-1$
				continue;
			}

			if (hasOptionPrefix(one)) {
				String part = one.substring(one.startsWith("--") ? 2 : 1); //$NON-NLS-1$
				String name = part.toLowerCase();
				int index = name.indexOf('=');
				String arg;
				TKCmdLineOption option;

				if (index != -1) {
					arg = part.substring(index + 1);
					name = name.substring(0, index);
				} else {
					arg = null;
				}

				option = map.get(name);
				if (option != null) {
					if (option.takesArgument()) {
						if (arg != null) {
							mData.add(new TKCmdLineData(option, arg));
							mUsedOptions.add(option);
						} else {
							if (++i < args.length) {
								arg = args[i];
								if (hasOptionPrefix(arg)) {
									msgs.add(MessageFormat.format(Msgs.MISSING_OPTION_ARGUMENT, one));
									i--;
								} else {
									mData.add(new TKCmdLineData(option, arg));
									mUsedOptions.add(option);
								}
							} else {
								msgs.add(MessageFormat.format(Msgs.MISSING_OPTION_ARGUMENT, one));
							}
						}
					} else if (arg != null) {
						msgs.add(MessageFormat.format(Msgs.UNEXPECTED_OPTION_ARGUMENT, one));
					} else {
						mData.add(new TKCmdLineData(option));
						mUsedOptions.add(option);
					}
				} else {
					msgs.add(MessageFormat.format(Msgs.UNEXPECTED_OPTION, one));
				}
			} else {
				mData.add(new TKCmdLineData(one));
			}
		}

		if (mUsedOptions.contains(HELP_OPTION)) {
			showHelp(map, extraHelp);
		}

		if (mUsedOptions.contains(VERSION_OPTION)) {
			System.out.println();
			System.out.println(TKApp.getVersionBanner(false));
			System.out.println();
			System.exit(0);
		}

		if (!msgs.isEmpty()) {
			for (String msg : msgs) {
				System.out.println(msg);
			}
			System.exit(1);
		}
	}

	private boolean hasOptionPrefix(String arg) {
		return arg.startsWith("-") || TKPlatform.isWindows() && arg.startsWith("/"); //$NON-NLS-1$ //$NON-NLS-2$
	}

	private void showHelp(HashMap<String, TKCmdLineOption> options, String extraHelp) {
		ArrayList<String> names = new ArrayList<String>(options.keySet());
		int cmdWidth = 0;

		Collections.sort(names);

		System.out.println();
		System.out.println(TKApp.getVersionBanner(false));
		System.out.println();
		System.out.println(Msgs.AVAILABLE_OPTIONS);
		System.out.println();

		for (String name : names) {
			TKCmdLineOption option = options.get(name);
			int width = 5 + name.length();

			if (option.takesArgument()) {
				width += 1 + option.getArgumentLabel().length();
			}
			if (width > cmdWidth) {
				cmdWidth = width;
			}
		}

		for (String name : names) {
			TKCmdLineOption option = options.get(name);

			StringBuilder builder = new StringBuilder();
			String[] allNames = option.getNames();
			String prefix = TKPlatform.isWindows() ? "/" : "-"; //$NON-NLS-1$ //$NON-NLS-2$
			String description;

			if (allNames[allNames.length - 1].equals(name)) {
				description = option.getDescription();
			} else {
				description = MessageFormat.format(Msgs.REFERENCE_DESCRIPTION, prefix, allNames[allNames.length - 1]);
			}
			builder.append("  "); //$NON-NLS-1$
			builder.append(prefix);
			builder.append(name);
			if (option.takesArgument()) {
				builder.append('=');
				builder.append(option.getArgumentLabel());
			}
			builder.append(TKTextUtility.makeFiller(cmdWidth - builder.length(), ' '));
			System.out.print(TKTextUtility.makeNote(builder.toString(), TKTextUtility.wrapToCharacterCount(description, 75 - cmdWidth)));
		}

		if (extraHelp != null) {
			System.out.println();
			System.out.println(extraHelp);
		}
		System.out.println();
		System.exit(0);
	}

	/**
	 * @param option The option to check for.
	 * @return Whether the option was present on the command line or not.
	 */
	public boolean isOptionUsed(TKCmdLineOption option) {
		return mUsedOptions.contains(option);
	}

	/**
	 * @param option The option to return the argument for.
	 * @return The option's argument.
	 */
	public String getOptionArgument(TKCmdLineOption option) {
		if (isOptionUsed(option)) {
			for (TKCmdLineData one : mData) {
				if (one.isOption() && one.getOption() == option) {
					return one.getArgument();
				}
			}
		}
		return null;
	}

	/**
	 * @param option The option to return the arguments for.
	 * @return The option's arguments.
	 */
	public ArrayList<String> getOptionArguments(TKCmdLineOption option) {
		ArrayList<String> list = new ArrayList<String>();

		if (isOptionUsed(option)) {
			for (TKCmdLineData one : mData) {
				if (one.isOption() && one.getOption() == option) {
					list.add(one.getArgument());
				}
			}
		}
		return list;
	}

	/** @return The arguments that were not options. */
	public ArrayList<String> getArguments() {
		ArrayList<String> arguments = new ArrayList<String>();

		for (TKCmdLineData one : mData) {
			if (!one.isOption()) {
				arguments.add(one.getArgument());
			}
		}
		return arguments;
	}

	/** @return The arguments that were not options. */
	public ArrayList<File> getArgumentsAsFiles() {
		ArrayList<File> arguments = new ArrayList<File>();

		for (TKCmdLineData one : mData) {
			if (!one.isOption()) {
				arguments.add(new File(one.getArgument()));
			}
		}
		return arguments;
	}
}
