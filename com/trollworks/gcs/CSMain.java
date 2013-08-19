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

package com.trollworks.gcs;

import com.trollworks.gcs.ui.common.CSFont;
import com.trollworks.toolkit.io.TKImage;
import com.trollworks.toolkit.io.TKLaunchProxy;
import com.trollworks.toolkit.io.TKPreferences;
import com.trollworks.toolkit.io.cmdline.TKCmdLine;
import com.trollworks.toolkit.io.cmdline.TKCmdLineOption;
import com.trollworks.toolkit.utility.TKApp;
import com.trollworks.toolkit.utility.TKGraphics;

import java.io.File;
import java.util.ArrayList;

/** The main entry point for the character sheet. */
public class CSMain {
	/** The command-line option for generating PDF's from the sheets. */
	public static final TKCmdLineOption	PDF_OPTION				= new TKCmdLineOption(Msgs.PDF_OPTION, null, "pdf");						//$NON-NLS-1$
	/** The command-line option for generating HTML from the sheets. */
	public static final TKCmdLineOption	HTML_OPTION				= new TKCmdLineOption(Msgs.HTML_OPTION, null, "html");						//$NON-NLS-1$
	/** The command-line option for specifying the HTML template to use. */
	public static final TKCmdLineOption	HTML_TEMPLATE_OPTION	= new TKCmdLineOption(Msgs.HTML_TEMPLATE_OPTION, Msgs.HTML_TEMPLATE_ARG, "html_template");	//$NON-NLS-1$
	/** The command-line option for generating PNG's from the sheets. */
	public static final TKCmdLineOption	PNG_OPTION				= new TKCmdLineOption(Msgs.PNG_OPTION, null, "png");						//$NON-NLS-1$
	/** The command-line option for forcing a particular paper size. */
	public static final TKCmdLineOption	SIZE_OPTION				= new TKCmdLineOption(Msgs.SIZE_OPTION, "SIZE", "paper");					//$NON-NLS-1$ //$NON-NLS-2$
	/** The command-line option for forcing particular paper margins. */
	public static final TKCmdLineOption	MARGIN_OPTION			= new TKCmdLineOption(Msgs.MARGIN_OPTION, "MARGINS", "margins");			//$NON-NLS-1$ //$NON-NLS-2$

	/**
	 * The main entry point for the character sheet.
	 * 
	 * @param args Arguments to the program. None used at the moment.
	 */
	public static void main(String[] args) {
		TKApp.setName(Msgs.APP_NAME);
		TKApp.setVersion(Msgs.APP_VERSION);
		TKApp.setCopyrightYears(Msgs.APP_COPYRIGHT_YEARS);
		TKApp.setCopyrightOwner(Msgs.APP_COPYRIGHT_OWNER);

		ArrayList<TKCmdLineOption> options = new ArrayList<TKCmdLineOption>();
		options.add(HTML_OPTION);
		options.add(HTML_TEMPLATE_OPTION);
		options.add(PDF_OPTION);
		options.add(PNG_OPTION);
		options.add(SIZE_OPTION);
		options.add(MARGIN_OPTION);

		TKCmdLine cmdLine = new TKCmdLine(args, options);
		if (cmdLine.isOptionUsed(HTML_OPTION) || cmdLine.isOptionUsed(PDF_OPTION) || cmdLine.isOptionUsed(PNG_OPTION)) {
			System.setProperty("java.awt.headless", "true"); //$NON-NLS-1$ //$NON-NLS-2$
			initialize();
			new CSBatchApplication(cmdLine);
		} else {
			TKLaunchProxy.configure("GCSLaunchProxy", 1, cmdLine.getArgumentsAsFiles().toArray(new File[0])); //$NON-NLS-1$
			if (TKGraphics.areGraphicsSafeToUse()) {
				initialize();
				new CSApplication(cmdLine);
			} else {
				System.err.println(TKGraphics.getReasonForUnsafeGraphics());
				System.exit(1);
			}
		}
	}

	private static void initialize() {
		TKImage.addLocation(CSMain.class.getResource("ui/common/images/")); //$NON-NLS-1$
		TKPreferences.setPreferenceFile("gcs.pref"); //$NON-NLS-1$
		CSFont.initialize();
	}
}
