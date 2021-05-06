/*
 * Copyright ©1998-2021 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, version 2.0. If a copy of the MPL was not distributed with
 * this file, You can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as
 * defined by the Mozilla Public License, version 2.0.
 */

package com.trollworks.gcs.page;

import com.trollworks.gcs.utility.I18n;

import javax.print.attribute.standard.MediaSize;
import javax.print.attribute.standard.MediaSizeName;

public class PaperSize {
    private static PaperSize   DEFAULT;
    private static PaperSize[] SIZES;
    private        MediaSize   mMediaSize;
    private        String      mTitle;

    public static final synchronized PaperSize getDefaultPaperSize() {
        if (DEFAULT == null) {
            DEFAULT = new PaperSize(MediaSize.NA.LETTER, I18n.Text("Letter"));
        }
        return DEFAULT;
    }

    public static final synchronized PaperSize[] getPaperSizes() {
        if (SIZES == null) {
            SIZES = new PaperSize[]{
                    getDefaultPaperSize(),
                    new PaperSize(MediaSize.NA.LEGAL, I18n.Text("Legal")),
                    new PaperSize(MediaSize.Other.TABLOID, I18n.Text("Tabloid")),
                    new PaperSize(MediaSize.ISO.A0, I18n.Text("A0")),
                    new PaperSize(MediaSize.ISO.A1, I18n.Text("A1")),
                    new PaperSize(MediaSize.ISO.A2, I18n.Text("A2")),
                    new PaperSize(MediaSize.ISO.A3, I18n.Text("A3")),
                    new PaperSize(MediaSize.ISO.A4, I18n.Text("A4")),
                    new PaperSize(MediaSize.ISO.A5, I18n.Text("A5")),
                    new PaperSize(MediaSize.ISO.A6, I18n.Text("A6")),
            };
        }
        return SIZES;
    }

    public PaperSize(MediaSize size, String title) {
        mMediaSize = size;
        mTitle = title;
    }

    public MediaSize getMediaSize() {
        return mMediaSize;
    }

    public MediaSizeName getMediaSizeName() {
        return mMediaSize.getMediaSizeName();
    }

    public String getKey() {
        return getMediaSizeName().toString();
    }

    @Override
    public String toString() {
        return mTitle;
    }
}
