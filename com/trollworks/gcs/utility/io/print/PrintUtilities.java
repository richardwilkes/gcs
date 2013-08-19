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

package com.trollworks.gcs.utility.io.print;

import com.trollworks.gcs.utility.units.LengthUnits;

import javax.print.PrintService;
import javax.print.attribute.Attribute;
import javax.print.attribute.PrintRequestAttributeSet;
import javax.print.attribute.Size2DSyntax;
import javax.print.attribute.standard.Chromaticity;
import javax.print.attribute.standard.Copies;
import javax.print.attribute.standard.Media;
import javax.print.attribute.standard.MediaPrintableArea;
import javax.print.attribute.standard.MediaSize;
import javax.print.attribute.standard.MediaSizeName;
import javax.print.attribute.standard.NumberUp;
import javax.print.attribute.standard.OrientationRequested;
import javax.print.attribute.standard.PageRanges;
import javax.print.attribute.standard.PrintQuality;
import javax.print.attribute.standard.PrinterResolution;
import javax.print.attribute.standard.Sides;

/** Provides access to various print settings. */
public class PrintUtilities {
	/**
	 * Extracts a setting from the specified {@link PrintRequestAttributeSet} or looks up a default
	 * value from the specified {@link PrintService} if the set doesn't contain it.
	 * 
	 * @param <T> The class type to return.
	 * @param service The {@link PrintService} to use.
	 * @param set The {@link PrintRequestAttributeSet} to use.
	 * @param type The setting type to extract.
	 * @param tryDefaultIfNull Whether to ask for the default if no value is present in the set.
	 * @return The value for the setting, or <code>null</code> if neither the set or the service
	 *         has a value for it.
	 */
	@SuppressWarnings("unchecked") public static <T extends Attribute> T getSetting(PrintService service, PrintRequestAttributeSet set, Class<T> type, boolean tryDefaultIfNull) {
		T current = (T) set.get(type);

		if (tryDefaultIfNull && current == null && service != null) {
			current = (T) service.getDefaultAttributeValue(type);
		}
		return current;
	}

	/**
	 * @param service The {@link PrintService} to use.
	 * @param set The {@link PrintRequestAttributeSet} to use.
	 * @return The page orientation.
	 */
	public static PageOrientation getPageOrientation(PrintService service, PrintRequestAttributeSet set) {
		return PageOrientation.get(getSetting(service, set, OrientationRequested.class, true));
	}

	/**
	 * Sets the page orientation.
	 * 
	 * @param set The {@link PrintRequestAttributeSet} to use.
	 * @param orientation The new page orientation.
	 */
	public static void setPageOrientation(PrintRequestAttributeSet set, PageOrientation orientation) {
		set.add(orientation.getOrientationRequested());
	}

	/**
	 * @param service The {@link PrintService} to use.
	 * @param set The {@link PrintRequestAttributeSet} to use.
	 * @param units The units to return.
	 * @return The margins of the paper (top, left, bottom, right).
	 */
	public static double[] getPaperMargins(PrintService service, PrintRequestAttributeSet set, LengthUnits units) {
		double[] size = getPaperSize(service, set, LengthUnits.INCHES);
		MediaPrintableArea current = getSetting(service, set, MediaPrintableArea.class, true);
		double x;
		double y;

		if (current == null) {
			current = new MediaPrintableArea(0.5f, 0.5f, (float) (size[0] - 1.0), (float) (size[1] - 1.0), MediaPrintableArea.INCH);
		}
		x = current.getX(MediaPrintableArea.INCH);
		y = current.getY(MediaPrintableArea.INCH);
		return new double[] { units.convert(LengthUnits.INCHES, y), units.convert(LengthUnits.INCHES, x), units.convert(LengthUnits.INCHES, size[1] - (y + current.getHeight(MediaPrintableArea.INCH))), units.convert(LengthUnits.INCHES, size[0] - (x + current.getWidth(MediaPrintableArea.INCH))) };
	}

	/**
	 * Sets the paper margins.
	 * 
	 * @param service The {@link PrintService} to use.
	 * @param set The {@link PrintRequestAttributeSet} to use.
	 * @param margins The margins of the paper (top, left, bottom, right).
	 * @param units The type of units being used.
	 */
	public static void setPaperMargins(PrintService service, PrintRequestAttributeSet set, double[] margins, LengthUnits units) {
		double[] size = getPaperSize(service, set, units);

		set.add(new MediaPrintableArea((float) LengthUnits.INCHES.convert(units, margins[1]), (float) LengthUnits.INCHES.convert(units, margins[0]), (float) LengthUnits.INCHES.convert(units, size[0] - (margins[1] + margins[3])), (float) LengthUnits.INCHES.convert(units, size[1] - (margins[0] + margins[2])), Size2DSyntax.INCH));
	}

	/**
	 * @param service The {@link PrintService} to use.
	 * @param set The {@link PrintRequestAttributeSet} to use.
	 * @param units The units to return.
	 * @return The margins of the page (top, left, bottom, right).
	 */
	public static double[] getPageMargins(PrintService service, PrintRequestAttributeSet set, LengthUnits units) {
		PageOrientation orientation = getPageOrientation(service, set);
		double[] margins = getPaperMargins(service, set, units);

		if (orientation == PageOrientation.LANDSCAPE) {
			return new double[] { margins[1], margins[2], margins[3], margins[0] };
		}
		if (orientation == PageOrientation.REVERSE_PORTRAIT) {
			return new double[] { margins[2], margins[3], margins[0], margins[1] };
		}
		if (orientation == PageOrientation.REVERSE_LANDSCAPE) {
			return new double[] { margins[3], margins[0], margins[1], margins[2] };
		}
		return margins;
	}

	/**
	 * @param service The {@link PrintService} to use.
	 * @param set The {@link PrintRequestAttributeSet} to use.
	 * @param margins The margins of the page (top, left, bottom, right).
	 * @param units The type of units being used.
	 */
	public static void setPageMargins(PrintService service, PrintRequestAttributeSet set, double[] margins, LengthUnits units) {
		PageOrientation orientation = getPageOrientation(service, set);

		if (orientation == PageOrientation.LANDSCAPE) {
			setPaperMargins(service, set, new double[] { margins[3], margins[0], margins[1], margins[2] }, units);
		} else if (orientation == PageOrientation.REVERSE_PORTRAIT) {
			setPaperMargins(service, set, new double[] { margins[2], margins[3], margins[0], margins[1] }, units);
		} else if (orientation == PageOrientation.REVERSE_LANDSCAPE) {
			setPaperMargins(service, set, new double[] { margins[1], margins[2], margins[3], margins[0] }, units);
		} else {
			setPaperMargins(service, set, margins, units);
		}
	}

	/**
	 * @param service The {@link PrintService} to use.
	 * @param set The {@link PrintRequestAttributeSet} to use.
	 * @param units The units to return.
	 * @return The width and height of the paper.
	 */
	public static double[] getPaperSize(PrintService service, PrintRequestAttributeSet set, LengthUnits units) {
		return getMediaDimensions(getSetting(service, set, Media.class, true), units);
	}

	/**
	 * @param media The {@link Media} to extract dimensions for.
	 * @param units The units to return.
	 * @return The dimensions of the specified {@link Media}.
	 */
	public static double[] getMediaDimensions(Media media, LengthUnits units) {
		MediaSize size = media instanceof MediaSizeName ? MediaSize.getMediaSizeForName((MediaSizeName) media) : null;

		if (size == null) {
			size = MediaSize.NA.LETTER;
		}
		return new double[] { units.convert(LengthUnits.INCHES, size.getX(Size2DSyntax.INCH)), units.convert(LengthUnits.INCHES, size.getY(Size2DSyntax.INCH)) };
	}

	/**
	 * Sets the paper size.
	 * 
	 * @param service The {@link PrintService} to use.
	 * @param set The {@link PrintRequestAttributeSet} to use.
	 * @param size The size of the paper.
	 * @param units The type of units being used.
	 */
	public static void setPaperSize(PrintService service, PrintRequestAttributeSet set, double[] size, LengthUnits units) {
		double[] margins = getPaperMargins(service, set, units);
		MediaSizeName mediaSizeName = MediaSize.findMedia((float) LengthUnits.INCHES.convert(units, size[0]), (float) LengthUnits.INCHES.convert(units, size[1]), Size2DSyntax.INCH);

		if (mediaSizeName == null) {
			mediaSizeName = MediaSizeName.NA_LETTER;
		}
		set.add(mediaSizeName);
		setPaperMargins(service, set, margins, units);
	}

	/**
	 * @param service The {@link PrintService} to use.
	 * @param set The {@link PrintRequestAttributeSet} to use.
	 * @param units The units to return.
	 * @return The width and height of the page.
	 */
	public static double[] getPageSize(PrintService service, PrintRequestAttributeSet set, LengthUnits units) {
		PageOrientation orientation = getPageOrientation(service, set);
		double[] size = getPaperSize(service, set, units);

		if (orientation == PageOrientation.LANDSCAPE || orientation == PageOrientation.REVERSE_LANDSCAPE) {
			double tmp = size[0];

			size[0] = size[1];
			size[1] = tmp;
		}

		return size;
	}

	/**
	 * Sets the paper size.
	 * 
	 * @param service The {@link PrintService} to use.
	 * @param set The {@link PrintRequestAttributeSet} to use.
	 * @param size The size of the page.
	 * @param units The type of units being used.
	 */
	public static void setPageSize(PrintService service, PrintRequestAttributeSet set, double[] size, LengthUnits units) {
		PageOrientation orientation = getPageOrientation(service, set);

		if (orientation == PageOrientation.LANDSCAPE || orientation == PageOrientation.REVERSE_LANDSCAPE) {
			size = new double[] { size[1], size[0] };
		}
		setPaperSize(service, set, size, units);
	}

	/**
	 * @param service The {@link PrintService} to use.
	 * @param set The {@link PrintRequestAttributeSet} to use.
	 * @param tryDefaultIfNull Whether to ask for the default if no value is present in the set.
	 * @return The chromaticity.
	 */
	public static InkChromaticity getChromaticity(PrintService service, PrintRequestAttributeSet set, boolean tryDefaultIfNull) {
		return InkChromaticity.get(getSetting(service, set, Chromaticity.class, tryDefaultIfNull));
	}

	/**
	 * Sets the chromaticity.
	 * 
	 * @param set The {@link PrintRequestAttributeSet} to use.
	 * @param chromaticity The new chromaticity.
	 */
	public static void setChromaticity(PrintRequestAttributeSet set, InkChromaticity chromaticity) {
		set.add(chromaticity.getChromaticity());
	}

	/**
	 * @param service The {@link PrintService} to use.
	 * @param set The {@link PrintRequestAttributeSet} to use.
	 * @return The sides.
	 */
	public static PageSides getSides(PrintService service, PrintRequestAttributeSet set) {
		return PageSides.get(getSetting(service, set, Sides.class, true));
	}

	/**
	 * Sets the sides.
	 * 
	 * @param set The {@link PrintRequestAttributeSet} to use.
	 * @param sides The new sides.
	 */
	public static void setSides(PrintRequestAttributeSet set, PageSides sides) {
		set.add(sides.getSides());
	}

	/**
	 * @param service The {@link PrintService} to use.
	 * @param set The {@link PrintRequestAttributeSet} to use.
	 * @param tryDefaultIfNull Whether to ask for the default if no value is present in the set.
	 * @return The print quality.
	 */
	public static Quality getPrintQuality(PrintService service, PrintRequestAttributeSet set, boolean tryDefaultIfNull) {
		return Quality.get(getSetting(service, set, PrintQuality.class, tryDefaultIfNull));
	}

	/**
	 * Sets the print quality.
	 * 
	 * @param set The {@link PrintRequestAttributeSet} to use.
	 * @param quality The new print quality.
	 */
	public static void setPrintQuality(PrintRequestAttributeSet set, Quality quality) {
		set.add(quality.getPrintQuality());
	}

	/**
	 * @param service The {@link PrintService} to use.
	 * @param set The {@link PrintRequestAttributeSet} to use.
	 * @return The copies to print.
	 */
	public static int getCopies(PrintService service, PrintRequestAttributeSet set) {
		Copies copies = PrintUtilities.getSetting(service, set, Copies.class, false);

		return copies == null ? 1 : copies.getValue();
	}

	/**
	 * Sets the copies to print.
	 * 
	 * @param set The {@link PrintRequestAttributeSet} to use.
	 * @param copies The new copies to print.
	 */
	public static void setCopies(PrintRequestAttributeSet set, int copies) {
		set.add(new Copies(Math.min(Math.max(copies, 1), 999)));
	}

	/**
	 * @param service The {@link PrintService} to use.
	 * @param set The {@link PrintRequestAttributeSet} to use.
	 * @return The number up to print.
	 */
	public static NumberUp getNumberUp(PrintService service, PrintRequestAttributeSet set) {
		return PrintUtilities.getSetting(service, set, NumberUp.class, false);
	}

	/**
	 * Sets the number up to print.
	 * 
	 * @param set The {@link PrintRequestAttributeSet} to use.
	 * @param numberUp The new number up to print.
	 */
	public static void setNumberUp(PrintRequestAttributeSet set, NumberUp numberUp) {
		set.add(numberUp);
	}

	/**
	 * @param service The {@link PrintService} to use.
	 * @param set The {@link PrintRequestAttributeSet} to use.
	 * @return The page ranges to print.
	 */
	public static PageRanges getPageRanges(PrintService service, PrintRequestAttributeSet set) {
		return PrintUtilities.getSetting(service, set, PageRanges.class, false);
	}

	/**
	 * Sets the page ranges to print.
	 * 
	 * @param set The {@link PrintRequestAttributeSet} to use.
	 * @param ranges The new page ranges to print.
	 */
	public static void setPageRanges(PrintRequestAttributeSet set, PageRanges ranges) {
		if (ranges != null) {
			set.add(ranges);
		} else {
			set.remove(PageRanges.class);
		}
	}

	/**
	 * @param service The {@link PrintService} to use.
	 * @param set The {@link PrintRequestAttributeSet} to use.
	 * @param tryDefaultIfNull Whether to ask for the default if no value is present in the set.
	 * @return The print resolution.
	 */
	public static PrinterResolution getResolution(PrintService service, PrintRequestAttributeSet set, boolean tryDefaultIfNull) {
		return getSetting(service, set, PrinterResolution.class, tryDefaultIfNull);
	}

	/**
	 * Sets the print resolution.
	 * 
	 * @param set The {@link PrintRequestAttributeSet} to use.
	 * @param resolution The new print resolution.
	 */
	public static void setResolution(PrintRequestAttributeSet set, PrinterResolution resolution) {
		if (resolution != null) {
			set.add(resolution);
		} else {
			set.remove(PrinterResolution.class);
		}
	}
}
