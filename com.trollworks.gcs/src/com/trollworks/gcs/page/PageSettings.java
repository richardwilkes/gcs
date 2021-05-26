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
import com.trollworks.gcs.utility.units.LengthValue;

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

    private static final LengthValue     DEFAULT_MARGIN      = new LengthValue(new Fixed6(0.25), LengthUnits.IN);
    private static final PageOrientation DEFAULT_ORIENTATION = PageOrientation.PORTRAIT;

    private ChangeNotifier  mNotifier;
    private PaperSize       mPaperSize;
    private LengthValue     mTopMargin;
    private LengthValue     mLeftMargin;
    private LengthValue     mBottomMargin;
    private LengthValue     mRightMargin;
    private PageOrientation mOrientation;

    public PageSettings(ChangeNotifier notifier) {
        mNotifier = notifier;
        reset();
    }

    public PageSettings(ChangeNotifier notifier, PageSettings other) {
        mNotifier = notifier;
        copy(other);
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
        mTopMargin = LengthValue.extract(m.getString(TOP_MARGIN), false);
        mLeftMargin = LengthValue.extract(m.getString(LEFT_MARGIN), false);
        mBottomMargin = LengthValue.extract(m.getString(BOTTOM_MARGIN), false);
        mRightMargin = LengthValue.extract(m.getString(RIGHT_MARGIN), false);
        mOrientation = Enums.extract(m.getString(ORIENTATION), PageOrientation.values(), DEFAULT_ORIENTATION);
    }

    public void toJSON(JsonWriter w) throws IOException {
        w.startMap();
        w.keyValue(PAPER_SIZE, mPaperSize.getKey());
        w.keyValue(TOP_MARGIN, mTopMargin.toString());
        w.keyValue(LEFT_MARGIN, mLeftMargin.toString());
        w.keyValue(BOTTOM_MARGIN, mBottomMargin.toString());
        w.keyValue(RIGHT_MARGIN, mRightMargin.toString());
        w.keyValue(ORIENTATION, Enums.toId(mOrientation));
        w.endMap();
    }

    public void copy(PageSettings other) {
        mPaperSize = other.mPaperSize;
        mTopMargin = new LengthValue(other.mTopMargin);
        mLeftMargin = new LengthValue(other.mLeftMargin);
        mBottomMargin = new LengthValue(other.mBottomMargin);
        mRightMargin = new LengthValue(other.mRightMargin);
        mOrientation = other.mOrientation;
    }

    public void reset() {
        mPaperSize = PaperSize.getDefaultPaperSize();
        mTopMargin = DEFAULT_MARGIN;
        mLeftMargin = DEFAULT_MARGIN;
        mBottomMargin = DEFAULT_MARGIN;
        mRightMargin = DEFAULT_MARGIN;
        mOrientation = DEFAULT_ORIENTATION;
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

    public LengthValue getTopMargin() {
        return mTopMargin;
    }

    public int getTopMarginPts() {
        return (int) LengthUnits.PT.convert(mTopMargin.getUnits(), mTopMargin.getValue()).asLong();
    }

    public void setTopMargin(LengthValue topMargin) {
        if (!mTopMargin.equals(topMargin)) {
            mTopMargin = new LengthValue(topMargin);
            mNotifier.notifyOfChange();
        }
    }

    public LengthValue getLeftMargin() {
        return mLeftMargin;
    }

    public int getLeftMarginPts() {
        return (int) LengthUnits.PT.convert(mLeftMargin.getUnits(), mLeftMargin.getValue()).asLong();
    }

    public void setLeftMargin(LengthValue leftMargin) {
        if (!mLeftMargin.equals(leftMargin)) {
            mLeftMargin = new LengthValue(leftMargin);
            mNotifier.notifyOfChange();
        }
    }

    public LengthValue getBottomMargin() {
        return mBottomMargin;
    }

    public int getBottomMarginPts() {
        return (int) LengthUnits.PT.convert(mBottomMargin.getUnits(), mBottomMargin.getValue()).asLong();
    }

    public void setBottomMargin(LengthValue bottomMargin) {
        if (!mBottomMargin.equals(bottomMargin)) {
            mBottomMargin = new LengthValue(bottomMargin);
            mNotifier.notifyOfChange();
        }
    }

    public LengthValue getRightMargin() {
        return mRightMargin;
    }

    public int getRightMarginPts() {
        return (int) LengthUnits.PT.convert(mRightMargin.getUnits(), mRightMargin.getValue()).asLong();
    }

    public void setRightMargin(LengthValue rightMargin) {
        if (!mRightMargin.equals(rightMargin)) {
            mRightMargin = new LengthValue(rightMargin);
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
        return mOrientation == that.mOrientation;
    }

    @Override
    public int hashCode() {
        int result = mPaperSize.hashCode();
        result = 31 * result + mTopMargin.hashCode();
        result = 31 * result + mLeftMargin.hashCode();
        result = 31 * result + mBottomMargin.hashCode();
        result = 31 * result + mRightMargin.hashCode();
        result = 31 * result + mOrientation.hashCode();
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
