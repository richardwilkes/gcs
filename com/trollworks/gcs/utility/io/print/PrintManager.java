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

package com.trollworks.gcs.utility.io.print;

import com.trollworks.gcs.utility.collections.EnumExtractor;
import com.trollworks.gcs.utility.io.LocalizedMessages;
import com.trollworks.gcs.utility.io.Preferences;
import com.trollworks.gcs.utility.io.xml.XMLNodeType;
import com.trollworks.gcs.utility.io.xml.XMLReader;
import com.trollworks.gcs.utility.io.xml.XMLWriter;
import com.trollworks.gcs.utility.text.NumberUtils;
import com.trollworks.gcs.utility.units.LengthUnits;
import com.trollworks.gcs.widgets.AppWindow;
import com.trollworks.gcs.widgets.WindowUtils;

import java.awt.print.PageFormat;
import java.awt.print.Paper;
import java.awt.print.Printable;
import java.awt.print.PrinterException;
import java.awt.print.PrinterJob;
import java.io.IOException;

import javax.print.DocFlavor;
import javax.print.PrintService;
import javax.print.attribute.Attribute;
import javax.print.attribute.HashPrintRequestAttributeSet;
import javax.print.attribute.standard.JobName;
import javax.print.attribute.standard.NumberUp;
import javax.print.attribute.standard.PageRanges;
import javax.print.attribute.standard.PrinterResolution;
import javax.swing.JOptionPane;

/** Manages printing. */
public class PrintManager {
	private static String					MSG_PRINTING_FAILED;
	private static String					MSG_NO_PRINTER_AVAILABLE;
	private static String					MSG_UNABLE_TO_SWITCH_PRINTERS;
	private static String					MSG_PAGE_SETUP_TITLE;
	private static String					MSG_PRINT_TITLE;
	private static final String				MODULE						= "PrintManager";		//$NON-NLS-1$
	private static final int				MODULE_VERSION				= 1;
	private static final String				NATIVE_DIALOGS_ENABLED_KEY	= "UseNativeDialogs";	//$NON-NLS-1$
	/** The XML root tag for {@link PrintManager}. */
	public static final String				TAG_ROOT					= "print_settings";	//$NON-NLS-1$
	private static final String				ATTRIBUTE_UNITS				= "units";				//$NON-NLS-1$
	private static final String				ATTRIBUTE_PRINTER			= "printer";			//$NON-NLS-1$
	private static final String				TAG_ORIENTATION				= "orientation";		//$NON-NLS-1$
	private static final String				TAG_WIDTH					= "width";				//$NON-NLS-1$
	private static final String				TAG_HEIGHT					= "height";			//$NON-NLS-1$
	private static final String				TAG_TOP_MARGIN				= "top_margin";		//$NON-NLS-1$
	private static final String				TAG_BOTTOM_MARGIN			= "bottom_margin";		//$NON-NLS-1$
	private static final String				TAG_LEFT_MARGIN				= "left_margin";		//$NON-NLS-1$
	private static final String				TAG_RIGHT_MARGIN			= "right_margin";		//$NON-NLS-1$
	private static final String				TAG_CHROMATICITY			= "ink_chromaticity";	//$NON-NLS-1$
	private static final String				TAG_SIDES					= "sides";				//$NON-NLS-1$
	private static final String				TAG_NUMBER_UP				= "number_up";			//$NON-NLS-1$
	private static final String				TAG_QUALITY					= "quality";			//$NON-NLS-1$
	private static final String				TAG_RESOLUTION				= "resolution";		//$NON-NLS-1$
	private PrinterJob						mJob;
	private HashPrintRequestAttributeSet	mSet;

	static {
		LocalizedMessages.initialize(PrintManager.class);
		Preferences.getInstance().resetIfVersionMisMatch(MODULE, MODULE_VERSION);
	}

	/** Creates a new {@link PrintManager} object. */
	public PrintManager() {
		mJob = PrinterJob.getPrinterJob();
		mSet = new HashPrintRequestAttributeSet();
	}

	/**
	 * Creates a new {@link PrintManager} object.
	 * 
	 * @param margins The page margins.
	 * @param units The type of units being used for the margins.
	 */
	public PrintManager(double margins, LengthUnits units) {
		this();
		setPageMargins(new double[] { margins, margins, margins, margins }, units);
	}

	/**
	 * Creates a new {@link PrintManager} object.
	 * 
	 * @param margins The page margins.
	 * @param units The type of units being used for the margins.
	 */
	public PrintManager(double[] margins, LengthUnits units) {
		this();
		setPageMargins(margins, units);
	}

	/**
	 * Creates a new {@link PrintManager} object.
	 * 
	 * @param orientation The page orientation.
	 * @param margins The page margins.
	 * @param units The type of units being used for the margins.
	 */
	public PrintManager(PageOrientation orientation, double margins, LengthUnits units) {
		this();
		setPageOrientation(orientation);
		setPageMargins(new double[] { margins, margins, margins, margins }, units);
	}

	/**
	 * Creates a new {@link PrintManager} object.
	 * 
	 * @param orientation The page orientation.
	 * @param margins The page margins.
	 * @param units The type of units being used for the margins.
	 */
	public PrintManager(PageOrientation orientation, double[] margins, LengthUnits units) {
		this();
		setPageOrientation(orientation);
		setPageMargins(margins, units);
	}

	/**
	 * Creates a new {@link PrintManager} object.
	 * 
	 * @param reader The XML reader to load from.
	 * @throws IOException
	 */
	public PrintManager(XMLReader reader) throws IOException {
		load(reader);
	}

	/**
	 * @param reader The XML reader to load from.
	 * @throws IOException
	 */
	public void load(XMLReader reader) throws IOException {
		String marker = reader.getMarker();
		LengthUnits units = (LengthUnits) EnumExtractor.extract(reader.getAttribute(ATTRIBUTE_UNITS), LengthUnits.values(), LengthUnits.INCHES);
		String printer = reader.getAttribute(ATTRIBUTE_PRINTER);
		double[] size = new double[] { 8.5, 11.0 };
		double[] margins = new double[] { 0, 0, 0, 0 };

		if (printer != null && printer.length() > 0) {
			try {
				for (PrintService one : PrinterJob.lookupPrintServices()) {
					if (one.getName().equalsIgnoreCase(printer)) {
						mJob.setPrintService(one);
						break;
					}
				}
			} catch (Exception exception) {
				// Ignore...
			}
		}

		do {
			if (reader.next() == XMLNodeType.START_TAG) {
				String name = reader.getName();

				try {
					if (TAG_ORIENTATION.equals(name)) {
						setPageOrientation((PageOrientation) EnumExtractor.extract(reader.readText(), PageOrientation.values(), PageOrientation.PORTRAIT));
					} else if (TAG_WIDTH.equals(name)) {
						size[0] = getNumber(reader, units, 8.5);
					} else if (TAG_HEIGHT.equals(name)) {
						size[1] = getNumber(reader, units, 11.0);
					} else if (TAG_TOP_MARGIN.equals(name)) {
						margins[0] = getNumber(reader, units, 0.0);
					} else if (TAG_LEFT_MARGIN.equals(name)) {
						margins[1] = getNumber(reader, units, 0.0);
					} else if (TAG_BOTTOM_MARGIN.equals(name)) {
						margins[2] = getNumber(reader, units, 0.0);
					} else if (TAG_RIGHT_MARGIN.equals(name)) {
						margins[3] = getNumber(reader, units, 0.0);
					} else if (TAG_CHROMATICITY.equals(name)) {
						setChromaticity((InkChromaticity) EnumExtractor.extract(reader.readText(), InkChromaticity.values(), InkChromaticity.COLOR));
					} else if (TAG_SIDES.equals(name)) {
						setSides((PageSides) EnumExtractor.extract(reader.readText(), PageSides.values(), PageSides.SINGLE));
					} else if (TAG_NUMBER_UP.equals(name)) {
						setNumberUp(reader.readInteger(1));
					} else if (TAG_QUALITY.equals(name)) {
						setPrintQuality((Quality) EnumExtractor.extract(reader.readText(), Quality.values(), Quality.NORMAL));
					} else if (TAG_RESOLUTION.equals(name)) {
						setResolution(extractFromResolutionString(reader.readText()));
					} else {
						reader.skipTag(name);
					}
				} catch (Exception exception) {
					// Ignore, since we can't really do anything about it.
				}
			}
		} while (reader.withinMarker(marker));
		setPaperSize(size, units);
		setPaperMargins(margins, units);
	}

	private double getNumber(XMLReader reader, LengthUnits units, double defInches) throws IOException {
		return reader.readDouble(units.convert(LengthUnits.INCHES, defInches));
	}

	/**
	 * Presents the page setup dialog and allows the user to change the settings.
	 * 
	 * @param window The owning window.
	 * @return Whether the user canceled (or an error occurred).
	 */
	public boolean pageSetup(AppWindow window) {
		if (useNativeDialogs()) {
			PageFormat format = mJob.pageDialog(createPageFormat());
			if (format != null) {
				adjustSettingsToPageFormat(format);
				window.adjustToPageSetupChanges();
				return true;
			}
		} else {
			PageSetupPanel panel = new PageSetupPanel(getPrintService(), mSet);
			if (WindowUtils.showOptionDialog(window, panel, MSG_PAGE_SETUP_TITLE, false, JOptionPane.OK_CANCEL_OPTION, JOptionPane.PLAIN_MESSAGE, null, null, null) == JOptionPane.OK_OPTION) {
				try {
					PrintService service = panel.accept(mSet);
					if (service != null) {
						mJob.setPrintService(service);
					}
					window.adjustToPageSetupChanges();
					return true;
				} catch (PrinterException exception) {
					WindowUtils.showError(window, MSG_UNABLE_TO_SWITCH_PRINTERS);
				}
			}
		}
		return false;
	}

	/**
	 * Presents the print dialog and allows the user to print.
	 * 
	 * @param window The owning window.
	 * @param jobTitle The title for this print job.
	 * @param printable The printable component.
	 */
	public void print(AppWindow window, String jobTitle, Printable printable) {
		PrintService service = getPrintService();

		if (service != null) {
			if (useNativeDialogs()) {
				mJob.setJobName(jobTitle);
				if (mJob.printDialog()) {
					try {
						window.adjustToPageSetupChanges();
						window.setPrinting(true);
						mJob.setPrintable(printable, createPageFormat());
						mJob.print();
					} catch (PrinterException exception) {
						WindowUtils.showError(window, MSG_PRINTING_FAILED);
					} finally {
						window.setPrinting(false);
					}
				}
			} else {
				mSet.add(new JobName(jobTitle, null));
				PrintPanel panel = new PrintPanel(getPrintService(), mSet);
				if (WindowUtils.showOptionDialog(window, panel, MSG_PRINT_TITLE, false, JOptionPane.OK_CANCEL_OPTION, JOptionPane.PLAIN_MESSAGE, null, null, null) == JOptionPane.OK_OPTION) {
					try {
						mJob.setPrintService(panel.accept(mSet));
						try {
							window.adjustToPageSetupChanges();
							window.setPrinting(true);
							mJob.setPrintable(printable, createPageFormat());
							mJob.print(mSet);
						} catch (PrinterException exception) {
							WindowUtils.showError(window, MSG_PRINTING_FAILED);
						} finally {
							window.setPrinting(false);
						}
					} catch (PrinterException exception) {
						WindowUtils.showError(window, MSG_UNABLE_TO_SWITCH_PRINTERS);
					}
				}
			}
		} else {
			WindowUtils.showError(window, MSG_NO_PRINTER_AVAILABLE);
		}
	}

	/** @return A newly created {@link PageFormat} with the current settings. */
	public PageFormat createPageFormat() {
		PageFormat format = new PageFormat();
		PageOrientation orientation = getPageOrientation();
		double[] size = getPaperSize(LengthUnits.POINTS);
		double[] margins = getPaperMargins(LengthUnits.POINTS);
		Paper paper = new Paper();

		if (orientation == PageOrientation.PORTRAIT) {
			format.setOrientation(PageFormat.PORTRAIT);
		} else if (orientation == PageOrientation.LANDSCAPE) {
			format.setOrientation(PageFormat.LANDSCAPE);
		} else if (orientation == PageOrientation.REVERSE_PORTRAIT) {
			// REVERSE_PORTRAIT doesn't exist in the old PageFormat class
			format.setOrientation(PageFormat.PORTRAIT);
		} else if (orientation == PageOrientation.REVERSE_LANDSCAPE) {
			format.setOrientation(PageFormat.REVERSE_LANDSCAPE);
		}
		paper.setSize(size[0], size[1]);
		paper.setImageableArea(margins[1], margins[0], size[0] - (margins[1] + margins[3]), size[1] - (margins[0] + margins[2]));
		format.setPaper(paper);
		return format;
	}

	/** @param format The format to adjust the settings to match. */
	public void adjustSettingsToPageFormat(PageFormat format) {
		Paper paper = format.getPaper();
		setPageOrientation(PageOrientation.get(format));
		setPaperSize(new double[] { paper.getWidth(), paper.getHeight() }, LengthUnits.POINTS);
		setPaperMargins(new double[] { paper.getImageableY(), paper.getImageableX(), paper.getHeight() - (paper.getImageableY() + paper.getImageableHeight()), paper.getWidth() - (paper.getImageableX() + paper.getImageableWidth()) }, LengthUnits.POINTS);
	}

	/**
	 * Writes this object to an XML stream.
	 * 
	 * @param out The XML writer to use.
	 * @param units The type of units to write the data out with.
	 */
	public void save(XMLWriter out, LengthUnits units) {
		double[] size = getPaperSize(units);
		double[] margins = getPaperMargins(units);
		PrintService service = getPrintService();

		out.startTag(TAG_ROOT);
		if (service != null) {
			out.writeAttribute(ATTRIBUTE_PRINTER, service.getName());
		}
		out.writeAttribute(ATTRIBUTE_UNITS, units.toString());
		out.finishTagEOL();
		out.simpleTag(TAG_ORIENTATION, getPageOrientation());
		out.simpleTag(TAG_WIDTH, size[0]);
		out.simpleTag(TAG_HEIGHT, size[1]);
		out.simpleTag(TAG_TOP_MARGIN, margins[0]);
		out.simpleTag(TAG_LEFT_MARGIN, margins[1]);
		out.simpleTag(TAG_BOTTOM_MARGIN, margins[2]);
		out.simpleTag(TAG_RIGHT_MARGIN, margins[3]);
		out.simpleTag(TAG_CHROMATICITY, getChromaticity(false));
		out.simpleTag(TAG_SIDES, getSides());
		out.simpleTag(TAG_NUMBER_UP, getNumberUp());
		out.simpleTag(TAG_QUALITY, getPrintQuality(false));
		out.simpleTagNotEmpty(TAG_RESOLUTION, createResolutionString(getResolution(false)));
		out.endTagEOL(TAG_ROOT, true);
	}

	private PrintService getPrintService() {
		// This method exists since some implementations throw a runtime
		// exception if no printer was ever defined on the system.
		try {
			return mJob.getPrintService();
		} catch (Exception exception) {
			return null;
		}
	}

	private PrinterResolution extractFromResolutionString(String buffer) {
		if (buffer != null && buffer.length() > 0) {
			int sep = buffer.indexOf('x');
			int x;
			int y;

			if (sep != -1 && sep < buffer.length() - 1) {
				x = NumberUtils.getNonLocalizedInteger(buffer.substring(0, sep), 0);
				y = NumberUtils.getNonLocalizedInteger(buffer.substring(sep + 1), 0);
			} else {
				x = NumberUtils.getNonLocalizedInteger(buffer, 0);
				y = x;
			}
			if (x < 1 || y < 1) {
				return null;
			}
			return new PrinterResolution(x, y, 1);
		}
		return null;
	}

	private String createResolutionString(PrinterResolution res) {
		if (res != null) {
			StringBuilder buffer = new StringBuilder();
			int x = res.getCrossFeedResolution(1);
			int y = res.getFeedResolution(1);
			buffer.append(Integer.toString(x));
			if (x != y) {
				buffer.append("x"); //$NON-NLS-1$
				buffer.append(Integer.toString(y));
			}
			return buffer.toString();
		}
		return null;
	}

	/** @return The page orientation. */
	public PageOrientation getPageOrientation() {
		return PrintUtilities.getPageOrientation(getPrintService(), mSet);
	}

	/** @param orientation The new page orientation. */
	public void setPageOrientation(PageOrientation orientation) {
		PrintUtilities.setPageOrientation(mSet, orientation);
	}

	/**
	 * @param units The units to return.
	 * @return The margins of the paper (top, left, bottom, right).
	 */
	public double[] getPaperMargins(LengthUnits units) {
		return PrintUtilities.getPaperMargins(getPrintService(), mSet, units);
	}

	/**
	 * @param margins The margins of the paper (top, left, bottom, right).
	 * @param units The type of units used.
	 */
	public void setPaperMargins(double[] margins, LengthUnits units) {
		PrintUtilities.setPaperMargins(getPrintService(), mSet, margins, units);
	}

	/**
	 * @param units The units to return.
	 * @return The margins of the page (top, left, bottom, right).
	 */
	public double[] getPageMargins(LengthUnits units) {
		return PrintUtilities.getPageMargins(getPrintService(), mSet, units);
	}

	/**
	 * @param margins The margins of the page (top, left, bottom, right).
	 * @param units The type of units used.
	 */
	public void setPageMargins(double[] margins, LengthUnits units) {
		PrintUtilities.setPageMargins(getPrintService(), mSet, margins, units);
	}

	/**
	 * @param units The units to return.
	 * @return The width and height of the paper.
	 */
	public double[] getPaperSize(LengthUnits units) {
		return PrintUtilities.getPaperSize(getPrintService(), mSet, units);
	}

	/**
	 * @param size The width and height of the paper.
	 * @param units The type of units used.
	 */
	public void setPaperSize(double[] size, LengthUnits units) {
		PrintUtilities.setPaperSize(getPrintService(), mSet, size, units);
	}

	/**
	 * @param units The units to return.
	 * @return The width and height of the page.
	 */
	public double[] getPageSize(LengthUnits units) {
		return PrintUtilities.getPageSize(getPrintService(), mSet, units);
	}

	/**
	 * @param size The width and height of the page.
	 * @param units The type of units used.
	 */
	public void setPageSize(double[] size, LengthUnits units) {
		PrintUtilities.setPageSize(getPrintService(), mSet, size, units);
	}

	/** @return The printer job. */
	public PrinterJob getJob() {
		return mJob;
	}

	/** @param attribute The attribute to set. */
	public void setAttribute(Attribute attribute) {
		mSet.add(attribute);
	}

	/**
	 * @param tryDefaultIfNull Whether to ask for the default if no value is present in the
	 *            settings.
	 * @return The chromaticity.
	 */
	public InkChromaticity getChromaticity(boolean tryDefaultIfNull) {
		return PrintUtilities.getChromaticity(getPrintService(), mSet, tryDefaultIfNull);
	}

	/** @param chromaticity The new chromaticity. */
	public void setChromaticity(InkChromaticity chromaticity) {
		PrintUtilities.setChromaticity(mSet, chromaticity);
	}

	/** @return The sides. */
	public PageSides getSides() {
		return PrintUtilities.getSides(getPrintService(), mSet);
	}

	/** @param sides The new sides. */
	public void setSides(PageSides sides) {
		PrintUtilities.setSides(mSet, sides);
	}

	/** @return The number up. */
	public NumberUp getNumberUp() {
		return PrintUtilities.getNumberUp(getPrintService(), mSet);
	}

	/** @param numberUp The new number up. */
	public void setNumberUp(NumberUp numberUp) {
		PrintUtilities.setNumberUp(mSet, numberUp);
	}

	/** @param numberUp The new number up. */
	public void setNumberUp(int numberUp) {
		PrintService service = getPrintService();

		if (service != null) {
			for (NumberUp one : (NumberUp[]) service.getSupportedAttributeValues(NumberUp.class, DocFlavor.SERVICE_FORMATTED.PRINTABLE, null)) {
				if (one.getValue() == numberUp) {
					setNumberUp(one);
					return;
				}
			}
		}
	}

	/**
	 * @param tryDefaultIfNull Whether to ask for the default if no value is present in the
	 *            settings.
	 * @return The print quality.
	 */
	public Quality getPrintQuality(boolean tryDefaultIfNull) {
		return PrintUtilities.getPrintQuality(getPrintService(), mSet, tryDefaultIfNull);
	}

	/** @param quality The new print quality. */
	public void setPrintQuality(Quality quality) {
		PrintUtilities.setPrintQuality(mSet, quality);
	}

	/** @return The copies to print. */
	public int getCopies() {
		return PrintUtilities.getCopies(getPrintService(), mSet);
	}

	/** @param copies The new copies to print. */
	public void setCopies(int copies) {
		PrintUtilities.setCopies(mSet, copies);
	}

	/** @return The page ranges to print. */
	public PageRanges getPageRanges() {
		return PrintUtilities.getPageRanges(getPrintService(), mSet);
	}

	/** @param ranges The new page ranges to print. */
	public void setPageRanges(PageRanges ranges) {
		PrintUtilities.setPageRanges(mSet, ranges);
	}

	/**
	 * @param tryDefaultIfNull Whether to ask for the default if no value is present in the
	 *            settings.
	 * @return The print resolution.
	 */
	public PrinterResolution getResolution(boolean tryDefaultIfNull) {
		return PrintUtilities.getResolution(getPrintService(), mSet, tryDefaultIfNull);
	}

	/** @param resolution The new print resolution. */
	public void setResolution(PrinterResolution resolution) {
		PrintUtilities.setResolution(mSet, resolution);
	}

	/** @return Whether native print dialogs will be used. */
	public static boolean useNativeDialogs() {
		return Preferences.getInstance().getBooleanValue(MODULE, NATIVE_DIALOGS_ENABLED_KEY, false);
	}

	/** @param useNative Whether native print dialogs will be used. */
	public static void useNativeDialogs(boolean useNative) {
		Preferences.getInstance().setValue(MODULE, NATIVE_DIALOGS_ENABLED_KEY, useNative);
	}
}
