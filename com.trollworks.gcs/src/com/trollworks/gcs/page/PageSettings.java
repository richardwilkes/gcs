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

import com.trollworks.gcs.datafile.ChangeNotifier;
import com.trollworks.gcs.utility.Fixed6;
import com.trollworks.gcs.utility.json.JsonMap;
import com.trollworks.gcs.utility.json.JsonWriter;
import com.trollworks.gcs.utility.text.Enums;
import com.trollworks.gcs.utility.units.LengthUnits;

import java.awt.print.PageFormat;
import java.awt.print.Paper;
import java.io.IOException;
import javax.print.attribute.Size2DSyntax;

public class PageSettings {
    private static final String PAPER_SIZE    = "paper_size";
    private static final String TOP_MARGIN    = "top_margin";
    private static final String BOTTOM_MARGIN = "bottom_margin";
    private static final String LEFT_MARGIN   = "left_margin";
    private static final String RIGHT_MARGIN  = "right_margin";
    private static final String ORIENTATION   = "orientation";
    private static final String UNITS         = "units";

    private static final Fixed6          DEFAULT_HEIGHT        = new Fixed6(11);
    private static final Fixed6          DEFAULT_TOP_MARGIN    = new Fixed6(0.25);
    private static final Fixed6          DEFAULT_BOTTOM_MARGIN = new Fixed6(0.25);
    private static final Fixed6          DEFAULT_LEFT_MARGIN   = new Fixed6(0.25);
    private static final Fixed6          DEFAULT_RIGHT_MARGIN  = new Fixed6(0.25);
    private static final PageOrientation DEFAULT_ORIENTATION   = PageOrientation.PORTRAIT;
    private static final LengthUnits     DEFAULT_UNITS         = LengthUnits.IN;

    private ChangeNotifier  mNotifier;
    private PaperSize       mPaperSize;
    private Fixed6          mTopMargin;
    private Fixed6          mLeftMargin;
    private Fixed6          mBottomMargin;
    private Fixed6          mRightMargin;
    private PageOrientation mOrientation;
    private LengthUnits     mUnits;

    public PageSettings(ChangeNotifier notifier) {
        mNotifier = notifier;
        reset();
    }

    public PageSettings(ChangeNotifier notifier, PageSettings other) {
        mNotifier = notifier;
        mPaperSize = other.mPaperSize;
        mTopMargin = other.mTopMargin;
        mLeftMargin = other.mLeftMargin;
        mBottomMargin = other.mBottomMargin;
        mRightMargin = other.mRightMargin;
        mOrientation = other.mOrientation;
        mUnits = other.mUnits;
    }

    public void load(JsonMap m) {
        String    paperSize = m.getString(PAPER_SIZE);
        PaperSize found     = null;
        for (PaperSize one : PaperSize.getPaperSizes()) {
            if (paperSize.equalsIgnoreCase(one.getKey())) {
                found = one;
                break;
            }
        }
        mPaperSize = found != null ? found : PaperSize.getDefaultPaperSize();
        mTopMargin = new Fixed6(m.getString(TOP_MARGIN), DEFAULT_TOP_MARGIN, false);
        mLeftMargin = new Fixed6(m.getString(LEFT_MARGIN), DEFAULT_LEFT_MARGIN, false);
        mBottomMargin = new Fixed6(m.getString(BOTTOM_MARGIN), DEFAULT_BOTTOM_MARGIN, false);
        mRightMargin = new Fixed6(m.getString(RIGHT_MARGIN), DEFAULT_RIGHT_MARGIN, false);
        mOrientation = Enums.extract(m.getString(ORIENTATION), PageOrientation.values(), DEFAULT_ORIENTATION);
        mUnits = Enums.extract(m.getString(UNITS), LengthUnits.values(), DEFAULT_UNITS);
    }

    public void save(JsonWriter w) throws IOException {
        w.startMap();
        w.keyValue(PAPER_SIZE, mPaperSize.getKey());
        w.keyValue(TOP_MARGIN, mTopMargin.toString());
        w.keyValue(LEFT_MARGIN, mLeftMargin.toString());
        w.keyValue(BOTTOM_MARGIN, mBottomMargin.toString());
        w.keyValue(RIGHT_MARGIN, mRightMargin.toString());
        w.keyValue(ORIENTATION, Enums.toId(mOrientation));
        w.keyValue(UNITS, Enums.toId(mUnits));
        w.endMap();
    }

    public void reset() {
        mPaperSize = PaperSize.getDefaultPaperSize();
        mTopMargin = DEFAULT_TOP_MARGIN;
        mLeftMargin = DEFAULT_LEFT_MARGIN;
        mBottomMargin = DEFAULT_BOTTOM_MARGIN;
        mRightMargin = DEFAULT_RIGHT_MARGIN;
        mOrientation = DEFAULT_ORIENTATION;
        mUnits = DEFAULT_UNITS;
    }

    public PaperSize getPaperSize() {
        return mPaperSize;
    }

    public int getWidthPts() {
        return (int) (mPaperSize.getMediaSize().getX(Size2DSyntax.INCH) * 72);
    }

    public int getHeightPts() {
        return (int) (mPaperSize.getMediaSize().getY(Size2DSyntax.INCH) * 72);
    }

    public void setPaperSize(PaperSize size) {
        if (mPaperSize != size) {
            mPaperSize = size;
            mNotifier.notifyOfChange();
        }
    }

    public Fixed6 getTopMargin() {
        return mTopMargin;
    }

    public int getTopMarginPts() {
        return (int) LengthUnits.PT.convert(mUnits, mTopMargin).asLong();
    }

    public void setTopMargin(Fixed6 topMargin) {
        if (mTopMargin != topMargin) {
            mTopMargin = topMargin;
            mNotifier.notifyOfChange();
        }
    }

    public Fixed6 getLeftMargin() {
        return mLeftMargin;
    }

    public int getLeftMarginPts() {
        return (int) LengthUnits.PT.convert(mUnits, mLeftMargin).asLong();
    }

    public void setLeftMargin(Fixed6 leftMargin) {
        if (mLeftMargin != leftMargin) {
            mLeftMargin = leftMargin;
            mNotifier.notifyOfChange();
        }
    }

    public Fixed6 getBottomMargin() {
        return mBottomMargin;
    }

    public int getBottomMarginPts() {
        return (int) LengthUnits.PT.convert(mUnits, mBottomMargin).asLong();
    }

    public void setBottomMargin(Fixed6 bottomMargin) {
        if (mBottomMargin != bottomMargin) {
            mBottomMargin = bottomMargin;
            mNotifier.notifyOfChange();
        }
    }

    public Fixed6 getRightMargin() {
        return mRightMargin;
    }

    public int getRightMarginPts() {
        return (int) LengthUnits.PT.convert(mUnits, mRightMargin).asLong();
    }

    public void setRightMargin(Fixed6 rightMargin) {
        if (mRightMargin != rightMargin) {
            mRightMargin = rightMargin;
            mNotifier.notifyOfChange();
        }
    }

    public PageOrientation getPageOrientation() {
        return mOrientation;
    }

    public void setPageOrientation(PageOrientation orientation) {
        if (mOrientation != orientation) {
            mOrientation = orientation;
            mNotifier.notifyOfChange();
        }
    }

    public LengthUnits getUnits() {
        return mUnits;
    }

    public void setUnits(LengthUnits units) {
        if (mUnits != units) {
            mUnits = units;
            mNotifier.notifyOfChange();
        }
    }

    @Override
    public boolean equals(Object other) {
        if (this == other) {
            return true;
        }
        if (!(other instanceof PageSettings)) {
            return false;
        }
        PageSettings that = (PageSettings) other;
        if (!mPaperSize.equals(that.mPaperSize)) {
            return false;
        }
        if (!mTopMargin.equals(that.mTopMargin)) {
            return false;
        }
        if (!mLeftMargin.equals(that.mLeftMargin)) {
            return false;
        }
        if (!mBottomMargin.equals(that.mBottomMargin)) {
            return false;
        }
        if (!mRightMargin.equals(that.mRightMargin)) {
            return false;
        }
        if (mOrientation != that.mOrientation) {
            return false;
        }
        return mUnits == that.mUnits;
    }

    @Override
    public int hashCode() {
        int result = mPaperSize.hashCode();
        result = 31 * result + mTopMargin.hashCode();
        result = 31 * result + mLeftMargin.hashCode();
        result = 31 * result + mBottomMargin.hashCode();
        result = 31 * result + mRightMargin.hashCode();
        result = 31 * result + mOrientation.hashCode();
        result = 31 * result + mUnits.hashCode();
        return result;
    }

    public PageFormat createPageFormat() {
        Paper paper = new Paper();
        paper.setSize(getWidthPts(), getHeightPts());
        paper.setImageableArea(getLeftMarginPts(), getTopMarginPts(),
                getWidthPts() - (getLeftMarginPts() + getRightMarginPts()),
                getHeightPts() - (getTopMarginPts() + getBottomMarginPts()));
        PageFormat format = new PageFormat();
        format.setOrientation(mOrientation.getOrientationForPageFormat());
        format.setPaper(paper);
        return format;
    }
}
