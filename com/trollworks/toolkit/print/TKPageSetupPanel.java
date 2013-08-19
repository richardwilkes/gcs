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

package com.trollworks.toolkit.print;

import com.trollworks.toolkit.text.TKNumberFilter;
import com.trollworks.toolkit.utility.TKAlignment;
import com.trollworks.toolkit.utility.TKNumberUtils;
import com.trollworks.toolkit.utility.TKFont;
import com.trollworks.toolkit.utility.TKGraphics;
import com.trollworks.toolkit.utility.units.TKLengthUnits;
import com.trollworks.toolkit.widget.TKLabel;
import com.trollworks.toolkit.widget.TKPopupMenu;
import com.trollworks.toolkit.widget.TKPanel;
import com.trollworks.toolkit.widget.TKTextField;
import com.trollworks.toolkit.widget.TKWidgetBorderPanel;
import com.trollworks.toolkit.widget.border.TKCompoundBorder;
import com.trollworks.toolkit.widget.border.TKEmptyBorder;
import com.trollworks.toolkit.widget.border.TKTitledBorder;
import com.trollworks.toolkit.widget.layout.TKColumnLayout;
import com.trollworks.toolkit.widget.menu.TKMenu;
import com.trollworks.toolkit.widget.menu.TKMenuItem;

import java.awt.Window;
import java.awt.event.ActionEvent;
import java.awt.event.ActionListener;
import java.awt.print.PrinterJob;
import java.text.NumberFormat;
import java.util.HashSet;
import java.util.StringTokenizer;

import javax.print.DocFlavor;
import javax.print.PrintService;
import javax.print.attribute.HashPrintRequestAttributeSet;
import javax.print.attribute.PrintRequestAttributeSet;
import javax.print.attribute.ResolutionSyntax;
import javax.print.attribute.standard.Chromaticity;
import javax.print.attribute.standard.Media;
import javax.print.attribute.standard.MediaSizeName;
import javax.print.attribute.standard.NumberUp;
import javax.print.attribute.standard.OrientationRequested;
import javax.print.attribute.standard.PrintQuality;
import javax.print.attribute.standard.PrinterResolution;
import javax.print.attribute.standard.Sides;

/** Provides the basic page setup panel. */
public class TKPageSetupPanel extends TKWidgetBorderPanel implements ActionListener {
	private TKPopupMenu		mServices;
	private TKPopupMenu		mOrientation;
	private TKPopupMenu		mPaperType;
	private TKTextField		mTopMargin;
	private TKTextField		mLeftMargin;
	private TKTextField		mRightMargin;
	private TKTextField		mBottomMargin;
	private TKPopupMenu		mChromaticity;
	private TKPopupMenu		mSides;
	private TKPopupMenu		mNumberUp;
	private TKPopupMenu		mPrintQuality;
	private TKPopupMenu		mResolution;
	private NumberFormat	mInchFormatter;

	/**
	 * Creates a new page setup panel.
	 * 
	 * @param service The {@link PrintService} to use.
	 * @param set The {@link PrintRequestAttributeSet} to use.
	 */
	public TKPageSetupPanel(PrintService service, PrintRequestAttributeSet set) {
		super();
		mInchFormatter = NumberFormat.getNumberInstance();
		mInchFormatter.setGroupingUsed(false);
		mInchFormatter.setMaximumFractionDigits(3);
		rebuild(service, set);
	}

	private void rebuild(PrintService service, PrintRequestAttributeSet set) {
		Window window = getBaseWindowAsWindow();
		TKPanel content = new TKPanel(new TKColumnLayout());

		removeAll();
		content.setBorder(new TKEmptyBorder(5));
		setContentPanel(content);
		rebuildSelf(addPrinterMenu(service), set);
		revalidate();
		if (window != null) {
			window.setSize(window.getPreferredSize());
			TKGraphics.forceOnScreen(window);
		}
	}

	/**
	 * Called to rebuild the panel.
	 * 
	 * @param service The {@link PrintService} to use.
	 * @param set The {@link PrintRequestAttributeSet} to use.
	 */
	protected void rebuildSelf(PrintService service, PrintRequestAttributeSet set) {
		TKPanel content = getContentPanel();
		TKPanel panel = new TKPanel(new TKColumnLayout(2));

		panel.setBorder(new TKCompoundBorder(new TKTitledBorder(Msgs.PAGE_SETTINGS, TKFont.lookup(TKFont.TEXT_FONT_KEY)), new TKEmptyBorder(5)));
		addPaperTypeMenu(service, set, panel);
		addOrientationMenu(service, set, panel);
		addSidesMenu(service, set, panel);
		addNumberUpMenu(service, set, panel);
		addMarginsPanel(service, set, panel);
		content.add(panel);

		panel = new TKPanel(new TKColumnLayout(2));
		panel.setBorder(new TKCompoundBorder(new TKTitledBorder(Msgs.QUALITY_SETTINGS, TKFont.lookup(TKFont.TEXT_FONT_KEY)), new TKEmptyBorder(5)));
		addChromaticityMenu(service, set, panel);
		addPrintQualityMenu(service, set, panel);
		addResolutionMenu(service, set, panel);
		content.add(panel);
	}

	private PrintService addPrinterMenu(PrintService service) {
		TKMenu menu = new TKMenu();
		TKPanel wrapper = new TKPanel(new TKColumnLayout(2));

		for (PrintService one : PrinterJob.lookupPrintServices()) {
			TKMenuItem item = new TKMenuItem(one.getName());

			item.setUserObject(one);
			menu.add(item);
		}
		mServices = new TKPopupMenu(menu);
		mServices.setSelectedUserObject(service);
		mServices.setOnlySize(mServices.getPreferredSize());
		mServices.addActionListener(this);
		wrapper.add(new TKLabel(Msgs.PRINTER, TKAlignment.RIGHT));
		wrapper.add(mServices);
		setBorderPanel(wrapper);
		return (PrintService) mServices.getSelectedItemUserObject();
	}

	private void addOrientationMenu(PrintService service, PrintRequestAttributeSet set, TKPanel parent) {
		TKMenu menu = new TKMenu();
		HashSet<OrientationRequested> possible = new HashSet<OrientationRequested>();

		for (OrientationRequested one : (OrientationRequested[]) service.getSupportedAttributeValues(OrientationRequested.class, DocFlavor.SERVICE_FORMATTED.PRINTABLE, null)) {
			possible.add(one);
		}
		for (TKPageOrientation orientation : TKPageOrientation.values()) {
			if (possible.contains(orientation.getOrientationRequested())) {
				TKMenuItem item = new TKMenuItem(orientation.toString());

				item.setUserObject(orientation);
				menu.add(item);
			}
		}
		mOrientation = new TKPopupMenu(menu);
		mOrientation.setSelectedUserObject(TKPrintUtilities.getPageOrientation(service, set));
		parent.add(new TKLabel(Msgs.ORIENTATION, TKAlignment.RIGHT));
		parent.add(mOrientation);
	}

	private void addPaperTypeMenu(PrintService service, PrintRequestAttributeSet set, TKPanel parent) {
		TKMenu menu = new TKMenu();
		Media current = TKPrintUtilities.getSetting(service, set, Media.class, true);

		if (current == null) {
			current = MediaSizeName.NA_LETTER;
		}

		for (Media one : (Media[]) service.getSupportedAttributeValues(Media.class, DocFlavor.SERVICE_FORMATTED.PRINTABLE, null)) {
			if (one instanceof MediaSizeName) {
				TKMenuItem item = new TKMenuItem(cleanUpMediaSizeName((MediaSizeName) one));

				item.setUserObject(one);
				menu.add(item);
			}
		}
		mPaperType = new TKPopupMenu(menu);
		mPaperType.setSelectedUserObject(current);
		parent.add(new TKLabel(Msgs.PAPER_TYPE, TKAlignment.RIGHT));
		parent.add(mPaperType);
	}

	private String cleanUpMediaSizeName(MediaSizeName msn) {
		StringBuilder builder = new StringBuilder();
		StringTokenizer tokenizer = new StringTokenizer(msn.toString(), "- ", true); //$NON-NLS-1$

		while (tokenizer.hasMoreTokens()) {
			String token = tokenizer.nextToken();

			if (token.equalsIgnoreCase("na")) { //$NON-NLS-1$
				builder.append("US"); //$NON-NLS-1$
			} else if (token.equalsIgnoreCase("iso")) { //$NON-NLS-1$
				builder.append("ISO"); //$NON-NLS-1$
			} else if (token.equalsIgnoreCase("jis")) { //$NON-NLS-1$
				builder.append("JIS"); //$NON-NLS-1$
			} else if (token.equals("-")) { //$NON-NLS-1$
				builder.append(" "); //$NON-NLS-1$
			} else if (token.length() > 1) {
				builder.append(Character.toUpperCase(token.charAt(0)));
				builder.append(token.substring(1));
			} else {
				builder.append(token.toUpperCase());
			}
		}

		return builder.toString();
	}

	private void addMarginsPanel(PrintService service, PrintRequestAttributeSet set, TKPanel parent) {
		double[] margins = TKPrintUtilities.getPageMargins(service, set, TKLengthUnits.INCHES);
		TKPanel wrapper = new TKPanel(new TKColumnLayout(3));

		mTopMargin = createMarginField(margins[0]);
		mLeftMargin = createMarginField(margins[1]);
		mBottomMargin = createMarginField(margins[2]);
		mRightMargin = createMarginField(margins[3]);
		wrapper.add(new TKPanel());
		wrapper.add(mTopMargin);
		wrapper.add(new TKPanel());
		wrapper.add(mLeftMargin);
		wrapper.add(new TKPanel());
		wrapper.add(mRightMargin);
		wrapper.add(new TKPanel());
		wrapper.add(mBottomMargin);
		wrapper.add(new TKPanel());
		parent.add(new TKLabel(Msgs.MARGINS, TKAlignment.RIGHT, true));
		parent.add(wrapper);
	}

	private TKTextField createMarginField(double margin) {
		TKTextField field = new TKTextField("999.999", 0, TKAlignment.RIGHT); //$NON-NLS-1$

		field.setKeyEventFilter(new TKNumberFilter(true, false, false, 6));
		field.setOnlySize(field.getPreferredSize());
		field.setText(mInchFormatter.format(margin));
		return field;
	}

	private void addChromaticityMenu(PrintService service, PrintRequestAttributeSet set, TKPanel parent) {
		Chromaticity[] chromacities = (Chromaticity[]) service.getSupportedAttributeValues(Chromaticity.class, DocFlavor.SERVICE_FORMATTED.PRINTABLE, null);

		if (chromacities != null && chromacities.length > 0) {
			TKMenu menu = new TKMenu();
			HashSet<Chromaticity> possible = new HashSet<Chromaticity>();

			for (Chromaticity one : chromacities) {
				possible.add(one);
			}
			for (TKInkChromaticity chromaticity : TKInkChromaticity.values()) {
				if (possible.contains(chromaticity.getChromaticity())) {
					TKMenuItem item = new TKMenuItem(chromaticity.toString());

					item.setUserObject(chromaticity);
					menu.add(item);
				}
			}
			mChromaticity = new TKPopupMenu(menu);
			mChromaticity.setSelectedUserObject(TKPrintUtilities.getChromaticity(service, set, true));
			parent.add(new TKLabel(Msgs.CHROMATICITY, TKAlignment.RIGHT));
			parent.add(mChromaticity);
		}
	}

	private void addSidesMenu(PrintService service, PrintRequestAttributeSet set, TKPanel parent) {
		Sides[] sides = (Sides[]) service.getSupportedAttributeValues(Sides.class, DocFlavor.SERVICE_FORMATTED.PRINTABLE, null);

		if (sides != null && sides.length > 0) {
			TKMenu menu = new TKMenu();
			HashSet<Sides> possible = new HashSet<Sides>();

			for (Sides one : sides) {
				possible.add(one);
			}
			for (TKPageSides side : TKPageSides.values()) {
				if (possible.contains(side.getSides())) {
					TKMenuItem item = new TKMenuItem(side.toString());

					item.setUserObject(side);
					menu.add(item);
				}
			}
			mSides = new TKPopupMenu(menu);
			mSides.setSelectedUserObject(TKPrintUtilities.getSides(service, set));
			parent.add(new TKLabel(Msgs.SIDES, TKAlignment.RIGHT));
			parent.add(mSides);
		}
	}

	private void addNumberUpMenu(PrintService service, PrintRequestAttributeSet set, TKPanel parent) {
		NumberUp[] numUp = (NumberUp[]) service.getSupportedAttributeValues(NumberUp.class, DocFlavor.SERVICE_FORMATTED.PRINTABLE, null);

		if (numUp != null && numUp.length > 0) {
			// Only Mac OS X seems to indicate support for this attribute, yet it is ignored
			// even there... so don't show it for now.
			if (false) {
				TKMenu menu = new TKMenu();

				for (NumberUp one : numUp) {
					TKMenuItem item = new TKMenuItem(one.toString());

					item.setUserObject(one);
					menu.add(item);
				}
				mNumberUp = new TKPopupMenu(menu);
				mNumberUp.setSelectedUserObject(TKPrintUtilities.getNumberUp(service, set));
				parent.add(new TKLabel(Msgs.NUMBER_UP, TKAlignment.RIGHT));
				parent.add(mNumberUp);
			}
		}
	}

	private void addPrintQualityMenu(PrintService service, PrintRequestAttributeSet set, TKPanel parent) {
		PrintQuality[] qualities = (PrintQuality[]) service.getSupportedAttributeValues(PrintQuality.class, DocFlavor.SERVICE_FORMATTED.PRINTABLE, null);

		if (qualities != null && qualities.length > 0) {
			TKMenu menu = new TKMenu();
			HashSet<PrintQuality> possible = new HashSet<PrintQuality>();

			for (PrintQuality one : qualities) {
				possible.add(one);
			}
			for (TKQuality quality : TKQuality.values()) {
				if (possible.contains(quality.getPrintQuality())) {
					TKMenuItem item = new TKMenuItem(quality.toString());

					item.setUserObject(quality);
					menu.add(item);
				}
			}
			mPrintQuality = new TKPopupMenu(menu);
			mPrintQuality.setSelectedUserObject(TKPrintUtilities.getPrintQuality(service, set, true));
			parent.add(new TKLabel(Msgs.QUALITY, TKAlignment.RIGHT));
			parent.add(mPrintQuality);
		}
	}

	private void addResolutionMenu(PrintService service, PrintRequestAttributeSet set, TKPanel parent) {
		PrinterResolution[] resolutions = (PrinterResolution[]) service.getSupportedAttributeValues(PrinterResolution.class, DocFlavor.SERVICE_FORMATTED.PRINTABLE, null);

		if (resolutions != null && resolutions.length > 0) {
			TKMenu menu = new TKMenu();

			for (PrinterResolution one : resolutions) {
				TKMenuItem item = new TKMenuItem(generateResolutionTitle(one));

				item.setUserObject(one);
				menu.add(item);
			}
			mResolution = new TKPopupMenu(menu);
			mResolution.setSelectedUserObject(TKPrintUtilities.getResolution(service, set, true));
			parent.add(new TKLabel(Msgs.RESOLUTION, TKAlignment.RIGHT));
			parent.add(mResolution);
		}
	}

	private String generateResolutionTitle(PrinterResolution res) {
		StringBuilder buffer = new StringBuilder();
		int x = res.getCrossFeedResolution(ResolutionSyntax.DPI);
		int y = res.getFeedResolution(ResolutionSyntax.DPI);

		buffer.append(Integer.toString(x));
		if (x != y) {
			buffer.append(" x "); //$NON-NLS-1$
			buffer.append(Integer.toString(y));
		}
		buffer.append(Msgs.DPI);
		return buffer.toString();
	}

	/**
	 * Accepts the changes made by the user and incorporates them into the specified attribute set.
	 * 
	 * @param set The set to modify.
	 * @return The {@link PrintService} selected by the user.
	 */
	public PrintService accept(PrintRequestAttributeSet set) {
		PrintService service = (PrintService) mServices.getSelectedItemUserObject();

		TKPrintUtilities.setPageOrientation(set, (TKPageOrientation) mOrientation.getSelectedItemUserObject());
		TKPrintUtilities.setPaperSize(service, set, TKPrintUtilities.getMediaDimensions((MediaSizeName) mPaperType.getSelectedItemUserObject(), TKLengthUnits.INCHES), TKLengthUnits.INCHES);
		TKPrintUtilities.setPageMargins(service, set, new double[] { TKNumberUtils.getDouble(mTopMargin.getText(), 0.0), TKNumberUtils.getDouble(mLeftMargin.getText(), 0.0), TKNumberUtils.getDouble(mBottomMargin.getText(), 0.0), TKNumberUtils.getDouble(mRightMargin.getText(), 0.0) }, TKLengthUnits.INCHES);
		if (mChromaticity != null) {
			TKPrintUtilities.setChromaticity(set, ((TKInkChromaticity) mChromaticity.getSelectedItemUserObject()));
		}
		if (mSides != null) {
			TKPrintUtilities.setSides(set, ((TKPageSides) mSides.getSelectedItemUserObject()));
		}
		if (mNumberUp != null) {
			TKPrintUtilities.setNumberUp(set, ((NumberUp) mNumberUp.getSelectedItemUserObject()));
		}
		if (mPrintQuality != null) {
			TKPrintUtilities.setPrintQuality(set, ((TKQuality) mPrintQuality.getSelectedItemUserObject()));
		}
		if (mResolution != null) {
			TKPrintUtilities.setResolution(set, ((PrinterResolution) mResolution.getSelectedItemUserObject()));
		}
		return service;
	}

	public void actionPerformed(ActionEvent event) {
		Object src = event.getSource();

		if (src == mServices) {
			HashPrintRequestAttributeSet set = new HashPrintRequestAttributeSet();

			rebuild(accept(set), set);
		}
	}
}
