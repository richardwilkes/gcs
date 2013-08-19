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

import com.trollworks.gcs.utility.io.LocalizedMessages;

import javax.print.DocFlavor;
import javax.print.DocPrintJob;
import javax.print.PrintService;
import javax.print.ServiceUIFactory;
import javax.print.attribute.Attribute;
import javax.print.attribute.AttributeSet;
import javax.print.attribute.PrintServiceAttribute;
import javax.print.attribute.PrintServiceAttributeSet;
import javax.print.attribute.standard.Media;
import javax.print.attribute.standard.MediaSizeName;
import javax.print.attribute.standard.OrientationRequested;
import javax.print.event.PrintServiceAttributeListener;

class DummyPrintService implements PrintService {
	private static String	MSG_NO_PRINTER_AVAILABLE;

	static {
		LocalizedMessages.initialize(DummyPrintService.class);
	}

	public void addPrintServiceAttributeListener(PrintServiceAttributeListener listener) {
		// Not used.
	}

	public DocPrintJob createPrintJob() {
		return null;
	}

	public <T extends PrintServiceAttribute> T getAttribute(Class<T> category) {
		return null;
	}

	public PrintServiceAttributeSet getAttributes() {
		return null;
	}

	public Object getDefaultAttributeValue(Class<? extends Attribute> category) {
		if (category == Media.class) {
			return MediaSizeName.NA_LETTER;
		}
		if (category == OrientationRequested.class) {
			return OrientationRequested.PORTRAIT;
		}
		return null;
	}

	public String getName() {
		return MSG_NO_PRINTER_AVAILABLE;
	}

	public ServiceUIFactory getServiceUIFactory() {
		return null;
	}

	public Class<?>[] getSupportedAttributeCategories() {
		return null;
	}

	public Object getSupportedAttributeValues(Class<? extends Attribute> category, DocFlavor flavor, AttributeSet attributes) {
		if (category == Media.class) {
			return new Media[] { MediaSizeName.NA_LETTER, MediaSizeName.NA_LEGAL, MediaSizeName.ISO_A4 };
		}
		if (category == OrientationRequested.class) {
			return new OrientationRequested[] { OrientationRequested.PORTRAIT, OrientationRequested.LANDSCAPE };
		}
		return null;
	}

	public DocFlavor[] getSupportedDocFlavors() {
		return null;
	}

	public AttributeSet getUnsupportedAttributes(DocFlavor flavor, AttributeSet attributes) {
		return null;
	}

	public boolean isAttributeCategorySupported(Class<? extends Attribute> category) {
		return false;
	}

	public boolean isAttributeValueSupported(Attribute attrval, DocFlavor flavor, AttributeSet attributes) {
		return false;
	}

	public boolean isDocFlavorSupported(DocFlavor flavor) {
		return false;
	}

	public void removePrintServiceAttributeListener(PrintServiceAttributeListener listener) {
		// Not used.
	}
}
