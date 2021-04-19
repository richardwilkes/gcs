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

package com.trollworks.gcs.character.panels;

import com.trollworks.gcs.character.CharacterSheet;
import com.trollworks.gcs.character.FieldFactory;
import com.trollworks.gcs.character.GURPSCharacter;
import com.trollworks.gcs.page.DropPanel;
import com.trollworks.gcs.page.PageField;
import com.trollworks.gcs.page.PageLabel;
import com.trollworks.gcs.page.PagePoints;
import com.trollworks.gcs.pointpool.PointPool;
import com.trollworks.gcs.pointpool.PointPoolDef;
import com.trollworks.gcs.pointpool.PoolThreshold;
import com.trollworks.gcs.ui.Fonts;
import com.trollworks.gcs.ui.ThemeColor;
import com.trollworks.gcs.ui.layout.PrecisionLayout;
import com.trollworks.gcs.ui.layout.PrecisionLayoutAlignment;
import com.trollworks.gcs.ui.layout.PrecisionLayoutData;
import com.trollworks.gcs.ui.widget.Label;
import com.trollworks.gcs.utility.I18n;

import java.util.Map;
import java.util.Objects;
import javax.swing.SwingConstants;
import javax.swing.UIManager;

public class PointPoolsPanel extends DropPanel {
    public PointPoolsPanel(CharacterSheet sheet) {
        super(new PrecisionLayout().setColumns(5).setMargins(0).setSpacing(2, 0).setFillAlignment(), I18n.Text("Point Pools"));
        GURPSCharacter         gch = sheet.getCharacter();
        Map<String, PointPool> pp  = gch.getPointPools();
        for (PointPoolDef poolDef : PointPoolDef.getOrderedPools(gch.getSettings().getPointPools())) {
            PointPool pool = pp.get(poolDef.getID());
            if (pool != null) {
                addPool(sheet, gch, poolDef, pool);
            }
        }
    }

    private void addPool(CharacterSheet sheet, GURPSCharacter gch, PointPoolDef poolDef, PointPool pool) {
        String id   = pool.getPoolID();
        String name = poolDef.getName();
        add(new PagePoints(pool.getPointCost(gch)), new PrecisionLayoutData().setHorizontalAlignment(PrecisionLayoutAlignment.END));
        add(new PageField(FieldFactory.INT7, Integer.valueOf(pool.getCurrent(gch)), (c, v) -> pool.setDamage(gch, -Math.min(((Integer) v).intValue() - pool.getMaximum(gch), 0)), sheet, id + ".current", SwingConstants.RIGHT, true, String.format(I18n.Text("Current %s"), name), ThemeColor.ON_PAGE), new PrecisionLayoutData().setGrabHorizontalSpace(true).setHorizontalAlignment(PrecisionLayoutAlignment.FILL));
        add(new PageLabel(I18n.Text("of"), null));
        add(new PageField(FieldFactory.POSINT6, Integer.valueOf(pool.getMaximum(gch)), (c, v) -> pool.setMaximum(gch, ((Integer) v).intValue()), sheet, id + ".maximum", SwingConstants.RIGHT, true, String.format(I18n.Text("Maximum %s"), name), ThemeColor.ON_PAGE), new PrecisionLayoutData().setGrabHorizontalSpace(true).setHorizontalAlignment(PrecisionLayoutAlignment.FILL));
        PageLabel label = new PageLabel(name, null);
        label.setToolTipText(poolDef.getDescription());
        add(label);
        Label mState = new Label("");
        setFont(UIManager.getFont(Fonts.KEY_LABEL_SECONDARY));
        setForeground(ThemeColor.ON_PAGE);
        PoolThreshold threshold = pool.getCurrentThreshold(gch);
        if (threshold != null) {
            mState.setText(threshold.getState());
            String explanation = threshold.getExplanation();
            if (explanation.isEmpty()) {
                explanation = null;
            }
            if (!Objects.equals(explanation, mState.getToolTipText())) {
                mState.setToolTipText(explanation);
            }
        }
        add(mState, new PrecisionLayoutData().setHorizontalAlignment(PrecisionLayoutAlignment.MIDDLE).setHorizontalSpan(5));
    }
}
