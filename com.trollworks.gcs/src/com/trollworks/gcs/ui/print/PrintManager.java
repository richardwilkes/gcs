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

import com.trollworks.gcs.preferences.Preferences;
import com.trollworks.gcs.ui.UIUtilities;
import com.trollworks.gcs.ui.widget.WindowUtils;
import com.trollworks.gcs.utility.Fixed6;
import com.trollworks.gcs.utility.I18n;
import com.trollworks.gcs.utility.Log;
import com.trollworks.gcs.utility.PrintProxy;
import com.trollworks.gcs.utility.json.Json;
import com.trollworks.gcs.utility.json.JsonMap;
import com.trollworks.gcs.utility.json.JsonWriter;
import com.trollworks.gcs.utility.text.Enums;
import com.trollworks.gcs.utility.text.Numbers;
import com.trollworks.gcs.utility.units.LengthUnits;
import com.trollworks.gcs.utility.xml.XMLNodeType;
import com.trollworks.gcs.utility.xml.XMLReader;

import java.awt.print.PageFormat;
import java.awt.print.Paper;
import java.awt.print.PrinterException;
import java.awt.print.PrinterJob;
import java.io.ByteArrayInputStream;
import java.io.ByteArrayOutputStream;
import java.io.IOException;
import java.io.OutputStreamWriter;
import java.nio.charset.StandardCharsets;
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
    /** The XML root tag for {@link PrintManager}. */
    public static final  String                       TAG_ROOT          = "print_settings";
    private static final String                       ATTRIBUTE_UNITS   = "units";
    private static final String                       ATTRIBUTE_PRINTER = "printer";
    private static final String                       TAG_ORIENTATION   = "orientation";
    private static final String                       TAG_WIDTH         = "width";
    private static final String                       TAG_HEIGHT        = "height";
    private static final String                       TAG_TOP_MARGIN    = "top_margin";
    private static final String                       TAG_BOTTOM_MARGIN = "bottom_margin";
    private static final String                       TAG_LEFT_MARGIN   = "left_margin";
    private static final String                       TAG_RIGHT_MARGIN  = "right_margin";
    private static final String                       TAG_CHROMATICITY  = "ink_chromaticity";
    private static final String                       TAG_SIDES         = "sides";
    private static final String                       TAG_NUMBER_UP     = "number_up";
    private static final String                       TAG_QUALITY       = "quality";
    private static final String                       TAG_RESOLUTION    = "resolution";
    private              PrinterJob                   mJob;
    private              HashPrintRequestAttributeSet mSet;

    /** Creates a new {@link PrintManager} object. */
    public PrintManager() {
        mJob = PrinterJob.getPrinterJob();
        mSet = new HashPrintRequestAttributeSet();
    }

    /**
     * Creates a new {@link PrintManager} object.
     *
     * @param margins The page margins.
     * @param units   The type of units being used for the margins.
     */
    public PrintManager(double margins, LengthUnits units) {
        this();
        setPageMargins(new double[]{margins, margins, margins, margins}, units);
    }

    /**
     * Creates a new {@link PrintManager} object.
     *
     * @param margins The page margins.
     * @param units   The type of units being used for the margins.
     */
    public PrintManager(double[] margins, LengthUnits units) {
        this();
        setPageMargins(margins, units);
    }

    /**
     * Creates a new {@link PrintManager} object.
     *
     * @param orientation The page orientation.
     * @param margins     The page margins.
     * @param units       The type of units being used for the margins.
     */
    public PrintManager(PageOrientation orientation, double margins, LengthUnits units) {
        this();
        setPageOrientation(orientation);
        setPageMargins(new double[]{margins, margins, margins, margins}, units);
    }

    /**
     * Creates a new {@link PrintManager} object.
     *
     * @param orientation The page orientation.
     * @param margins     The page margins.
     * @param units       The type of units being used for the margins.
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
     */
    public PrintManager(XMLReader reader) throws IOException {
        this();
        load(reader);
    }

    public PrintManager(JsonMap m) {
        this();
        double[]    size    = {8.5, 11.0};
        double[]    margins = {0, 0, 0, 0};
        LengthUnits units   = Enums.extract(m.getString(ATTRIBUTE_UNITS), LengthUnits.values(), LengthUnits.IN);
        setPrintServiceForPrinter(m.getString(ATTRIBUTE_PRINTER));
        setPageOrientation(Enums.extract(m.getString(TAG_ORIENTATION), PageOrientation.values(), PageOrientation.PORTRAIT));
        size[0] = getNumberForJSON(m, TAG_WIDTH, units, 8.5);
        size[1] = getNumberForJSON(m, TAG_HEIGHT, units, 11.0);
        margins[0] = getNumberForJSON(m, TAG_TOP_MARGIN, units, 0.0);
        margins[1] = getNumberForJSON(m, TAG_LEFT_MARGIN, units, 0.0);
        margins[2] = getNumberForJSON(m, TAG_BOTTOM_MARGIN, units, 0.0);
        margins[3] = getNumberForJSON(m, TAG_RIGHT_MARGIN, units, 0.0);
        setChromaticity(Enums.extract(m.getString(TAG_CHROMATICITY), InkChromaticity.values(), InkChromaticity.COLOR));
        setSides(Enums.extract(m.getString(TAG_SIDES), PageSides.values(), PageSides.SINGLE));
        setNumberUp(m.getIntWithDefault(TAG_NUMBER_UP, 1));
        setPrintQuality(Enums.extract(m.getString(TAG_QUALITY), Quality.values(), Quality.NORMAL));
        setResolution(extractFromResolutionString(m.getString(TAG_RESOLUTION)));
        setPaperSize(size, units);
        setPaperMargins(margins, units);
    }

    public PrintManager(PrintManager other) {
        this(other.toJSONMap(LengthUnits.IN));
    }

    private void setPrintServiceForPrinter(String printer) {
        if (printer != null && !printer.isEmpty()) {
            try {
                for (PrintService one : PrinterJob.lookupPrintServices()) {
                    if (one.getName().equalsIgnoreCase(printer)) {
                        mJob.setPrintService(one);
                        break;
                    }
                }
            } catch (Exception exception) {
                Log.warn(exception);
            }
        }
    }

    private static double getNumberForJSON(JsonMap m, String key, LengthUnits units, double defInches) {
        return m.getDoubleWithDefault(key, units.convert(LengthUnits.IN, new Fixed6(defInches)).asDouble());
    }

    /**
     * @param reader The XML reader to load from.
     */
    public void load(XMLReader reader) throws IOException {
        String      marker  = reader.getMarker();
        LengthUnits units   = Enums.extract(reader.getAttribute(ATTRIBUTE_UNITS), LengthUnits.values(), LengthUnits.IN);
        double[]    size    = {8.5, 11.0};
        double[]    margins = {0, 0, 0, 0};
        setPrintServiceForPrinter(reader.getAttribute(ATTRIBUTE_PRINTER));
        do {
            if (reader.next() == XMLNodeType.START_TAG) {
                String name = reader.getName();
                try {
                    if (TAG_ORIENTATION.equals(name)) {
                        setPageOrientation(Enums.extract(reader.readText(), PageOrientation.values(), PageOrientation.PORTRAIT));
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
                        setChromaticity(Enums.extract(reader.readText(), InkChromaticity.values(), InkChromaticity.COLOR));
                    } else if (TAG_SIDES.equals(name)) {
                        setSides(Enums.extract(reader.readText(), PageSides.values(), PageSides.SINGLE));
                    } else if (TAG_NUMBER_UP.equals(name)) {
                        setNumberUp(reader.readInteger(1));
                    } else if (TAG_QUALITY.equals(name)) {
                        setPrintQuality(Enums.extract(reader.readText(), Quality.values(), Quality.NORMAL));
                    } else if (TAG_RESOLUTION.equals(name)) {
                        setResolution(extractFromResolutionString(reader.readText()));
                    } else {
                        reader.skipTag(name);
                    }
                } catch (Exception exception) {
                    Log.warn(exception);
                }
            }
        } while (reader.withinMarker(marker));
        setPaperSize(size, units);
        setPaperMargins(margins, units);
    }

    private static double getNumber(XMLReader reader, LengthUnits units, double defInches) throws IOException {
        return reader.readDouble(units.convert(LengthUnits.IN, new Fixed6(defInches)).asDouble());
    }

    /**
     * Presents the page setup dialog and allows the user to change the settings.
     *
     * @param proxy The {@link PrintProxy} representing the information that would be printed.
     * @return Whether the user canceled (or an error occurred).
     */
    public boolean pageSetup(PrintProxy proxy) {
        if (Preferences.getInstance().useNativePrintDialogs()) {
            PageFormat format = mJob.pageDialog(createPageFormat());
            if (format != null) {
                adjustSettingsToPageFormat(format);
                proxy.adjustToPageSetupChanges(false);
                return true;
            }
        } else {
            PageSetupPanel panel = new PageSetupPanel(getPrintService(), mSet);
            if (WindowUtils.showOptionDialog(UIUtilities.getComponentForDialog(proxy), panel, I18n.Text("Page Setup"), false, JOptionPane.OK_CANCEL_OPTION, JOptionPane.PLAIN_MESSAGE, null, null, null) == JOptionPane.OK_OPTION) {
                try {
                    PrintService service = panel.accept(mSet);
                    if (service != null) {
                        mJob.setPrintService(service);
                    }
                    proxy.adjustToPageSetupChanges(false);
                    return true;
                } catch (PrinterException exception) {
                    WindowUtils.showError(UIUtilities.getComponentForDialog(proxy), I18n.Text("Unable to switch printers!"));
                }
            }
        }
        return false;
    }

    /**
     * Presents the print dialog and allows the user to print.
     *
     * @param proxy The {@link PrintProxy} representing the information being printed.
     */
    public void print(PrintProxy proxy) {
        if (proxy != null) {
            PrintService service = getPrintService();
            if (service != null) {
                if (Preferences.getInstance().useNativePrintDialogs()) {
                    mJob.setJobName(proxy.getPrintJobTitle());
                    if (mJob.printDialog()) {
                        try {
                            proxy.adjustToPageSetupChanges(true);
                            proxy.setPrinting(true);
                            mJob.setPrintable(proxy, createPageFormat());
                            mJob.print();
                        } catch (PrinterException exception) {
                            WindowUtils.showError(UIUtilities.getComponentForDialog(proxy), I18n.Text("Printing failed!"));
                        } finally {
                            proxy.setPrinting(false);
                        }
                    }
                } else {
                    mSet.add(new JobName(proxy.getPrintJobTitle(), null));
                    PrintPanel panel = new PrintPanel(getPrintService(), mSet);
                    if (WindowUtils.showOptionDialog(UIUtilities.getComponentForDialog(proxy), panel, I18n.Text("Print"), false, JOptionPane.OK_CANCEL_OPTION, JOptionPane.PLAIN_MESSAGE, null, null, null) == JOptionPane.OK_OPTION) {
                        try {
                            mJob.setPrintService(panel.accept(mSet));
                            try {
                                proxy.adjustToPageSetupChanges(true);
                                proxy.setPrinting(true);
                                mJob.setPrintable(proxy, createPageFormat());
                                mJob.print(mSet);
                            } catch (PrinterException exception) {
                                WindowUtils.showError(UIUtilities.getComponentForDialog(proxy), I18n.Text("Printing failed!"));
                            } finally {
                                proxy.setPrinting(false);
                            }
                        } catch (PrinterException exception) {
                            WindowUtils.showError(UIUtilities.getComponentForDialog(proxy), I18n.Text("Unable to switch printers!"));
                        }
                    }
                }
            } else {
                WindowUtils.showError(UIUtilities.getComponentForDialog(proxy), I18n.Text("No printer is available!"));
            }
        }
    }

    /** @return A newly created {@link PageFormat} with the current settings. */
    public PageFormat createPageFormat() {
        PageFormat      format      = new PageFormat();
        PageOrientation orientation = getPageOrientation();
        double[]        size        = getPaperSize(LengthUnits.PT);
        double[]        margins     = getPaperMargins(LengthUnits.PT);
        Paper           paper       = new Paper();

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
        setPaperSize(new double[]{paper.getWidth(), paper.getHeight()}, LengthUnits.PT);
        setPaperMargins(new double[]{paper.getImageableY(), paper.getImageableX(), paper.getHeight() - (paper.getImageableY() + paper.getImageableHeight()), paper.getWidth() - (paper.getImageableX() + paper.getImageableWidth())}, LengthUnits.PT);
    }

    public String toString() {
        return toJSONMap(LengthUnits.IN).toString(true);
    }

    public void toJSON(JsonWriter w, LengthUnits units) throws IOException {
        double[]     size    = getPaperSize(units);
        double[]     margins = getPaperMargins(units);
        PrintService service = getPrintService();
        w.startMap();
        if (service != null) {
            w.keyValue(ATTRIBUTE_PRINTER, service.getName());
        }
        w.keyValue(ATTRIBUTE_UNITS, Enums.toId(units));
        w.keyValue(TAG_ORIENTATION, Enums.toId(getPageOrientation()));
        w.keyValue(TAG_WIDTH, size[0]);
        w.keyValue(TAG_HEIGHT, size[1]);
        w.keyValue(TAG_TOP_MARGIN, margins[0]);
        w.keyValue(TAG_LEFT_MARGIN, margins[1]);
        w.keyValue(TAG_BOTTOM_MARGIN, margins[2]);
        w.keyValue(TAG_RIGHT_MARGIN, margins[3]);
        w.keyValue(TAG_CHROMATICITY, Enums.toId(getChromaticity(false)));
        w.keyValue(TAG_SIDES, Enums.toId(getSides()));
        w.keyValueNot(TAG_NUMBER_UP, getNumberUp().getValue(), 1);
        w.keyValue(TAG_QUALITY, Enums.toId(getPrintQuality(false)));
        w.keyValueNot(TAG_RESOLUTION, createResolutionString(getResolution(false)), null);
        w.endMap();
    }

    public JsonMap toJSONMap(LengthUnits units) {
        try (ByteArrayOutputStream buffer = new ByteArrayOutputStream()) {
            try (JsonWriter w = new JsonWriter(new OutputStreamWriter(buffer, StandardCharsets.UTF_8), "\t")) {
                toJSON(w, LengthUnits.IN);
            }
            return Json.asMap(Json.parse(new ByteArrayInputStream(buffer.toByteArray())));
        } catch (Exception exception) {
            Log.error(exception);
            return new JsonMap();
        }
    }

    public void save(JsonWriter w, LengthUnits units) throws IOException {
        toJSON(w, units);
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

    private static PrinterResolution extractFromResolutionString(String buffer) {
        if (buffer != null && !buffer.isEmpty()) {
            int sep = buffer.indexOf('x');
            int x;
            int y;

            if (sep != -1 && sep < buffer.length() - 1) {
                x = Numbers.extractInteger(buffer.substring(0, sep), 0, false);
                y = Numbers.extractInteger(buffer.substring(sep + 1), 0, false);
            } else {
                x = Numbers.extractInteger(buffer, 0, false);
                y = x;
            }
            if (x < 1 || y < 1) {
                return null;
            }
            return new PrinterResolution(x, y, 1);
        }
        return null;
    }

    private static String createResolutionString(PrinterResolution res) {
        if (res != null) {
            StringBuilder buffer = new StringBuilder();
            int           x      = res.getCrossFeedResolution(1);
            int           y      = res.getFeedResolution(1);
            buffer.append(Integer.toString(x));
            if (x != y) {
                buffer.append("x");
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
     * @param units   The type of units used.
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
     * @param units   The type of units used.
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
     * @param size  The width and height of the paper.
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
     * @param size  The width and height of the page.
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
     *                         settings.
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
        NumberUp numUp = PrintUtilities.getNumberUp(getPrintService(), mSet);
        return numUp != null ? numUp : new NumberUp(1);
    }

    /** @param numberUp The new number up. */
    public void setNumberUp(NumberUp numberUp) {
        PrintUtilities.setNumberUp(mSet, numberUp);
    }

    /** @param numberUp The new number up. */
    public void setNumberUp(int numberUp) {
        PrintService service = getPrintService();
        if (service != null) {
            NumberUp[] values = (NumberUp[]) service.getSupportedAttributeValues(NumberUp.class, DocFlavor.SERVICE_FORMATTED.PRINTABLE, null);
            if (values != null) {
                for (NumberUp one : values) {
                    if (one.getValue() == numberUp) {
                        setNumberUp(one);
                        return;
                    }
                }
            }
        }
    }

    /**
     * @param tryDefaultIfNull Whether to ask for the default if no value is present in the
     *                         settings.
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
     *                         settings.
     * @return The print resolution.
     */
    public PrinterResolution getResolution(boolean tryDefaultIfNull) {
        return PrintUtilities.getResolution(getPrintService(), mSet, tryDefaultIfNull);
    }

    /** @param resolution The new print resolution. */
    public void setResolution(PrinterResolution resolution) {
        PrintUtilities.setResolution(mSet, resolution);
    }
}
