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

package com.trollworks.gcs.feature;

import com.trollworks.gcs.widgets.outline.ListRow;
import com.trollworks.ttk.utility.LocalizedMessages;
import com.trollworks.ttk.widgets.BandedPanel;

import java.util.ArrayList;
import java.util.List;

/** Displays and edits {@link Feature} objects. */
public class FeaturesPanel extends BandedPanel {
	private static String	MSG_FEATURES;

	static {
		LocalizedMessages.initialize(FeaturesPanel.class);
	}

	/**
	 * Creates a new feature editor.
	 * 
	 * @param row The row these features will belong to.
	 * @param features The initial features to display.
	 */
	public FeaturesPanel(ListRow row, List<Feature> features) {
		super(MSG_FEATURES);
		for (Feature feature : features) {
			add(FeatureEditor.create(row, feature.cloneFeature()));
		}
		if (getComponentCount() == 0) {
			add(new NoFeature(row));
		}
	}

	/** @return The current set of features. */
	public List<Feature> getFeatures() {
		int count = getComponentCount();
		ArrayList<Feature> list = new ArrayList<Feature>(count);

		for (int i = 0; i < count; i++) {
			Feature feature = ((FeatureEditor) getComponent(i)).getFeature();

			if (feature != null) {
				list.add(feature);
			}
		}
		return list;
	}
}
