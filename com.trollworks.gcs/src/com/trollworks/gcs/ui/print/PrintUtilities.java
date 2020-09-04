/*
 * Copyright ©1998-2020 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, version 2.0. If a copy of the MPL was not distributed with
 * this file, You can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as
 * defined by the Mozilla Public License, version 2.0.
 */

package com.trollworks.gcs.ui.print;

import com.trollworks.gcs.ui.UIUtilities;
import com.trollworks.gcs.utility.Fixed6;
import com.trollworks.gcs.utility.PrintProxy;
import com.trollworks.gcs.utility.units.LengthUnits;

import java.awt.Component;
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
public final class PrintUtilities {
    private PrintUtilities() {
    }

    /**
     * @param component The {@link Component} to check.
     * @return If the {@link Component} or one of its ancestors is currently printing.
     */
    public static boolean isPrinting(Component component) {
        PrintProxy proxy = UIUtilities.getSelfOrAncestorOfType(component, PrintProxy.class);
        return proxy != null && proxy.isPrinting();
    }

    /**
     * Extracts a setting from the specified {@link PrintRequestAttributeSet} or looks up a default
     * value from the specified {@link PrintService} if the set doesn't contain it.
     *
     * @param service          The {@link PrintService} to use.
     * @param set              The {@link PrintRequestAttributeSet} to use.
     * @param type             The setting type to extract.
     * @param tryDefaultIfNull Whether to ask for the default if no value is present in the set.
     * @return The value for the setting, or {@code null} if neither the set or the service has a
     *         value for it.
     */
    public static Attribute getSetting(PrintService service, PrintRequestAttributeSet set, Class<? extends Attribute> type, boolean tryDefaultIfNull) {
        Attribute attribute = set.get(type);
        if (tryDefaultIfNull && attribute == null && service != null) {
            attribute = (Attribute) service.getDefaultAttributeValue(type);
        }
        return attribute;
    }

    /**
     * @param service The {@link PrintService} to use.
     * @param set     The {@link PrintRequestAttributeSet} to use.
     * @return The page orientation.
     */
    public static PageOrientation getPageOrientation(PrintService service, PrintRequestAttributeSet set) {
        return PageOrientation.get((OrientationRequested) getSetting(service, set, OrientationRequested.class, true));
    }

    /**
     * Sets the page orientation.
     *
     * @param set         The {@link PrintRequestAttributeSet} to use.
     * @param orientation The new page orientation.
     */
    public static void setPageOrientation(PrintRequestAttributeSet set, PageOrientation orientation) {
        set.add(orientation.getOrientationRequested());
    }

    /**
     * @param service The {@link PrintService} to use.
     * @param set     The {@link PrintRequestAttributeSet} to use.
     * @param units   The units to return.
     * @return The margins of the paper (top, left, bottom, right).
     */
    public static double[] getPaperMargins(PrintService service, PrintRequestAttributeSet set, LengthUnits units) {
        double[]           size    = getPaperSize(service, set, LengthUnits.IN);
        MediaPrintableArea current = (MediaPrintableArea) getSetting(service, set, MediaPrintableArea.class, true);
        if (current == null) {
            current = new MediaPrintableArea(0.5f, 0.5f, (float) (size[0] - 1.0), (float) (size[1] - 1.0), MediaPrintableArea.INCH);
        }
        double x = current.getX(MediaPrintableArea.INCH);
        double y = current.getY(MediaPrintableArea.INCH);
        return new double[]{units.convert(LengthUnits.IN, new Fixed6(y)).asDouble(), units.convert(LengthUnits.IN, new Fixed6(x)).asDouble(), units.convert(LengthUnits.IN, new Fixed6(size[1] - (y + current.getHeight(MediaPrintableArea.INCH)))).asDouble(), units.convert(LengthUnits.IN, new Fixed6(size[0] - (x + current.getWidth(MediaPrintableArea.INCH)))).asDouble()};
    }

    /**
     * Sets the paper margins.
     *
     * @param service The {@link PrintService} to use.
     * @param set     The {@link PrintRequestAttributeSet} to use.
     * @param margins The margins of the paper (top, left, bottom, right).
     * @param units   The type of units being used.
     */
    public static void setPaperMargins(PrintService service, PrintRequestAttributeSet set, double[] margins, LengthUnits units) {
        double[] size = getPaperSize(service, set, units);

        set.add(new MediaPrintableArea((float) LengthUnits.IN.convert(units, new Fixed6(margins[1])).asDouble(), (float) LengthUnits.IN.convert(units, new Fixed6(margins[0])).asDouble(), (float) LengthUnits.IN.convert(units, new Fixed6(size[0] - (margins[1] + margins[3]))).asDouble(), (float) LengthUnits.IN.convert(units, new Fixed6(size[1] - (margins[0] + margins[2]))).asDouble(), Size2DSyntax.INCH));
    }

    /**
     * @param service The {@link PrintService} to use.
     * @param set     The {@link PrintRequestAttributeSet} to use.
     * @param units   The units to return.
     * @return The margins of the page (top, left, bottom, right).
     */
    public static double[] getPageMargins(PrintService service, PrintRequestAttributeSet set, LengthUnits units) {
        PageOrientation orientation = getPageOrientation(service, set);
        double[]        margins     = getPaperMargins(service, set, units);

        if (orientation == PageOrientation.LANDSCAPE) {
            return new double[]{margins[1], margins[2], margins[3], margins[0]};
        }
        if (orientation == PageOrientation.REVERSE_PORTRAIT) {
            return new double[]{margins[2], margins[3], margins[0], margins[1]};
        }
        if (orientation == PageOrientation.REVERSE_LANDSCAPE) {
            return new double[]{margins[3], margins[0], margins[1], margins[2]};
        }
        return margins;
    }

    /**
     * @param service The {@link PrintService} to use.
     * @param set     The {@link PrintRequestAttributeSet} to use.
     * @param margins The margins of the page (top, left, bottom, right).
     * @param units   The type of units being used.
     */
    public static void setPageMargins(PrintService service, PrintRequestAttributeSet set, double[] margins, LengthUnits units) {
        PageOrientation orientation = getPageOrientation(service, set);

        if (orientation == PageOrientation.LANDSCAPE) {
            setPaperMargins(service, set, new double[]{margins[3], margins[0], margins[1], margins[2]}, units);
        } else if (orientation == PageOrientation.REVERSE_PORTRAIT) {
            setPaperMargins(service, set, new double[]{margins[2], margins[3], margins[0], margins[1]}, units);
        } else if (orientation == PageOrientation.REVERSE_LANDSCAPE) {
            setPaperMargins(service, set, new double[]{margins[1], margins[2], margins[3], margins[0]}, units);
        } else {
            setPaperMargins(service, set, margins, units);
        }
    }

    /**
     * @param service The {@link PrintService} to use.
     * @param set     The {@link PrintRequestAttributeSet} to use.
     * @param units   The units to return.
     * @return The width and height of the paper.
     */
    public static double[] getPaperSize(PrintService service, PrintRequestAttributeSet set, LengthUnits units) {
        return getMediaDimensions((Media) getSetting(service, set, Media.class, true), units);
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
        return new double[]{units.convert(LengthUnits.IN, new Fixed6(size.getX(Size2DSyntax.INCH))).asDouble(), units.convert(LengthUnits.IN, new Fixed6(size.getY(Size2DSyntax.INCH))).asDouble()};
    }

    /**
     * Sets the paper size.
     *
     * @param service The {@link PrintService} to use.
     * @param set     The {@link PrintRequestAttributeSet} to use.
     * @param size    The size of the paper.
     * @param units   The type of units being used.
     */
    public static void setPaperSize(PrintService service, PrintRequestAttributeSet set, double[] size, LengthUnits units) {
        double[]      margins       = getPaperMargins(service, set, units);
        MediaSizeName mediaSizeName = MediaSize.findMedia((float) LengthUnits.IN.convert(units, new Fixed6(size[0])).asDouble(), (float) LengthUnits.IN.convert(units, new Fixed6(size[1])).asDouble(), Size2DSyntax.INCH);

        if (mediaSizeName == null) {
            mediaSizeName = MediaSizeName.NA_LETTER;
        }
        set.add(mediaSizeName);
        setPaperMargins(service, set, margins, units);
    }

    /**
     * @param service The {@link PrintService} to use.
     * @param set     The {@link PrintRequestAttributeSet} to use.
     * @param units   The units to return.
     * @return The width and height of the page.
     */
    public static double[] getPageSize(PrintService service, PrintRequestAttributeSet set, LengthUnits units) {
        PageOrientation orientation = getPageOrientation(service, set);
        double[]        size        = getPaperSize(service, set, units);

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
     * @param set     The {@link PrintRequestAttributeSet} to use.
     * @param size    The size of the page.
     * @param units   The type of units being used.
     */
    public static void setPageSize(PrintService service, PrintRequestAttributeSet set, double[] size, LengthUnits units) {
        PageOrientation orientation = getPageOrientation(service, set);

        if (orientation == PageOrientation.LANDSCAPE || orientation == PageOrientation.REVERSE_LANDSCAPE) {
            size = new double[]{size[1], size[0]};
        }
        setPaperSize(service, set, size, units);
    }

    /**
     * @param service          The {@link PrintService} to use.
     * @param set              The {@link PrintRequestAttributeSet} to use.
     * @param tryDefaultIfNull Whether to ask for the default if no value is present in the set.
     * @return The chromaticity.
     */
    public static InkChromaticity getChromaticity(PrintService service, PrintRequestAttributeSet set, boolean tryDefaultIfNull) {
        return InkChromaticity.get((Chromaticity) getSetting(service, set, Chromaticity.class, tryDefaultIfNull));
    }

    /**
     * Sets the chromaticity.
     *
     * @param set          The {@link PrintRequestAttributeSet} to use.
     * @param chromaticity The new chromaticity.
     */
    public static void setChromaticity(PrintRequestAttributeSet set, InkChromaticity chromaticity) {
        set.add(chromaticity.getChromaticity());
    }

    /**
     * @param service The {@link PrintService} to use.
     * @param set     The {@link PrintRequestAttributeSet} to use.
     * @return The sides.
     */
    public static PageSides getSides(PrintService service, PrintRequestAttributeSet set) {
        return PageSides.get((Sides) getSetting(service, set, Sides.class, true));
    }

    /**
     * Sets the sides.
     *
     * @param set   The {@link PrintRequestAttributeSet} to use.
     * @param sides The new sides.
     */
    public static void setSides(PrintRequestAttributeSet set, PageSides sides) {
        set.add(sides.getSides());
    }

    /**
     * @param service          The {@link PrintService} to use.
     * @param set              The {@link PrintRequestAttributeSet} to use.
     * @param tryDefaultIfNull Whether to ask for the default if no value is present in the set.
     * @return The print quality.
     */
    public static Quality getPrintQuality(PrintService service, PrintRequestAttributeSet set, boolean tryDefaultIfNull) {
        return Quality.get((PrintQuality) getSetting(service, set, PrintQuality.class, tryDefaultIfNull));
    }

    /**
     * Sets the print quality.
     *
     * @param set     The {@link PrintRequestAttributeSet} to use.
     * @param quality The new print quality.
     */
    public static void setPrintQuality(PrintRequestAttributeSet set, Quality quality) {
        set.add(quality.getPrintQuality());
    }

    /**
     * @param service The {@link PrintService} to use.
     * @param set     The {@link PrintRequestAttributeSet} to use.
     * @return The copies to print.
     */
    public static int getCopies(PrintService service, PrintRequestAttributeSet set) {
        Copies copies = (Copies) getSetting(service, set, Copies.class, false);
        return copies == null ? 1 : copies.getValue();
    }

    /**
     * Sets the copies to print.
     *
     * @param set    The {@link PrintRequestAttributeSet} to use.
     * @param copies The new copies to print.
     */
    public static void setCopies(PrintRequestAttributeSet set, int copies) {
        set.add(new Copies(Math.min(Math.max(copies, 1), 999)));
    }

    /**
     * @param service The {@link PrintService} to use.
     * @param set     The {@link PrintRequestAttributeSet} to use.
     * @return The number up to print.
     */
    public static NumberUp getNumberUp(PrintService service, PrintRequestAttributeSet set) {
        return (NumberUp) getSetting(service, set, NumberUp.class, false);
    }

    /**
     * Sets the number up to print.
     *
     * @param set      The {@link PrintRequestAttributeSet} to use.
     * @param numberUp The new number up to print.
     */
    public static void setNumberUp(PrintRequestAttributeSet set, NumberUp numberUp) {
        set.add(numberUp);
    }

    /**
     * @param service The {@link PrintService} to use.
     * @param set     The {@link PrintRequestAttributeSet} to use.
     * @return The page ranges to print.
     */
    public static PageRanges getPageRanges(PrintService service, PrintRequestAttributeSet set) {
        return (PageRanges) getSetting(service, set, PageRanges.class, false);
    }

    /**
     * Sets the page ranges to print.
     *
     * @param set    The {@link PrintRequestAttributeSet} to use.
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
     * @param service          The {@link PrintService} to use.
     * @param set              The {@link PrintRequestAttributeSet} to use.
     * @param tryDefaultIfNull Whether to ask for the default if no value is present in the set.
     * @return The print resolution.
     */
    public static PrinterResolution getResolution(PrintService service, PrintRequestAttributeSet set, boolean tryDefaultIfNull) {
        return (PrinterResolution) getSetting(service, set, PrinterResolution.class, tryDefaultIfNull);
    }

    /**
     * Sets the print resolution.
     *
     * @param set        The {@link PrintRequestAttributeSet} to use.
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
