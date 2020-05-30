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

package com.trollworks.gcs.utility.xml;

/** The various node types that the {@link XMLReader} generates. */
public enum XMLNodeType {
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
    OTHER
}
