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

package com.trollworks.toolkit.io.xml;

/** The various node types that the {@link TKXMLReader} generates. */
public enum TKXMLNodeType {
	/** The start of the XML document. */
	START_DOCUMENT,
	/** The end of XML document. */
	END_DOCUMENT,
	/** A start tag. */
	START_TAG,
	/** An end tag. */
	END_TAG,
	/** A text block. */
	TEXT,
	/** A data section. */
	DATA,
	/** An entity reference. */
	ENTITY_REF,
	/** Other... */
	OTHER;
}
