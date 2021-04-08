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

package com.trollworks.gcs.feature;

import com.trollworks.gcs.ui.widget.BandedPanel;
import com.trollworks.gcs.ui.widget.outline.ListRow;
import com.trollworks.gcs.utility.I18n;

import java.util.ArrayList;
import java.util.List;

/** Displays and edits {@link Feature} objects. */
public class FeaturesPanel extends BandedPanel {
    /**
     * Creates a new feature editor.
     *
     * @param row      The row these features will belong to.
     * @param features The initial features to display.
     */
    public FeaturesPanel(ListRow row, List<Feature> features) {
        super(I18n.Text("Features"));
        for (Feature feature : features) {
            add(FeatureEditor.create(row, feature.cloneFeature()));
        }
        if (getComponentCount() == 0) {
            add(new NoFeature(row));
        }
    }

    /** @return The current set of features. */
    public List<Feature> getFeatures() {
        int           count = getComponentCount();
        List<Feature> list  = new ArrayList<>(count);

        for (int i = 0; i < count; i++) {
            Feature feature = ((FeatureEditor) getComponent(i)).getFeature();

            if (feature != null) {
                list.add(feature);
            }
        }
        return list;
    }
}
