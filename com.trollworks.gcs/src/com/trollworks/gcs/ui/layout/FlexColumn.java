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

package com.trollworks.gcs.ui.layout;

import com.trollworks.gcs.ui.scale.Scale;

import java.awt.Dimension;
import java.awt.Rectangle;
import java.util.Arrays;
import java.util.List;

/** A column within a {@link FlexLayout}. */
public class FlexColumn extends FlexContainer {
    /** Creates a new {@link FlexColumn}. */
    public FlexColumn() {
        setFill(true);
    }

    @Override
    protected void layoutSelf(Scale scale, Rectangle bounds) {
        int   count = getChildCount();
        int   vGap  = scale.scale(getVerticalGap());
        int[] gaps  = new int[count > 0 ? count - 1 : 0];
        Arrays.fill(gaps, vGap);
        int         height    = vGap * (count > 0 ? count - 1 : 0);
        Dimension[] minSizes  = getChildSizes(scale, LayoutSize.MINIMUM);
        Dimension[] prefSizes = getChildSizes(scale, LayoutSize.PREFERRED);
        Dimension[] maxSizes  = getChildSizes(scale, LayoutSize.MAXIMUM);
        for (int i = 0; i < count; i++) {
            height += prefSizes[i].height;
        }
        int extra = bounds.height - height;
        if (extra != 0) {
            int[] values = new int[count];
            int[] limits = new int[count];
            for (int i = 0; i < count; i++) {
                values[i] = prefSizes[i].height;
                limits[i] = extra > 0 ? maxSizes[i].height : minSizes[i].height;
            }
            extra = distribute(extra, values, limits);
            for (int i = 0; i < count; i++) {
                prefSizes[i].height = values[i];
            }
            int length = gaps.length;
            if (extra > 0 && getFillVertical() && length > 0) {
                int amt = extra / length;
                for (int i = 0; i < length; i++) {
                    gaps[i] += amt;
                }
                extra -= amt * length;
                for (int i = 0; i < extra; i++) {
                    gaps[i]++;
                }
                extra = 0;
            }
        }
        List<FlexCell> children    = getChildren();
        Rectangle[]    childBounds = new Rectangle[count];
        for (int i = 0; i < count; i++) {
            childBounds[i] = new Rectangle(prefSizes[i]);
            if (getFillHorizontal()) {
                childBounds[i].width = Math.min(maxSizes[i].width, bounds.width);
            }
            switch (children.get(i).getHorizontalAlignment()) {
            case CENTER -> childBounds[i].x = bounds.x + (bounds.width - childBounds[i].width) / 2;
            case RIGHT_BOTTOM -> childBounds[i].x = bounds.x + bounds.width - childBounds[i].width;
            default -> childBounds[i].x = bounds.x;
            }
        }
        int       y      = bounds.y;
        Alignment vAlign = getVerticalAlignment();
        if (vAlign == Alignment.CENTER) {
            y += extra / 2;
        } else if (vAlign == Alignment.RIGHT_BOTTOM) {
            y += extra;
        }
        for (int i = 0; i < count; i++) {
            childBounds[i].y = y;
            if (i < count - 1) {
                y += prefSizes[i].height;
                y += gaps[i];
            }
        }
        layoutChildren(scale, childBounds);
    }

    @Override
    protected Dimension getSizeSelf(Scale scale, LayoutSize type) {
        Dimension[] sizes = getChildSizes(scale, type);
        Dimension   size  = new Dimension(0, scale.scale(getVerticalGap()) * (sizes.length > 0 ? sizes.length - 1 : 0));
        for (Dimension one : sizes) {
            size.height += one.height;
            if (one.width > size.width) {
                size.width = one.width;
            }
        }
        return size;
    }
}
