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

package com.trollworks.gcs.ui.editor.feature;

import com.trollworks.gcs.model.CMRow;
import com.trollworks.gcs.model.feature.CMFeature;
import com.trollworks.gcs.ui.editor.CSBandedPanel;

import java.util.ArrayList;
import java.util.List;

/** Displays and edits {@link CMFeature} objects. */
public class CSFeatures extends CSBandedPanel {
	/**
	 * Creates a new feature editor.
	 * 
	 * @param row The row these features will belong to.
	 * @param features The initial features to display.
	 */
	public CSFeatures(CMRow row, List<CMFeature> features) {
		super(Msgs.FEATURES);
		for (CMFeature feature : features) {
			add(CSBaseFeature.create(row, feature.cloneFeature()));
		}
		if (getComponentCount() == 0) {
			add(new CSNoFeature(row));
		}
	}

	/** @return The current set of features. */
	public List<CMFeature> getFeatures() {
		int count = getComponentCount();
		ArrayList<CMFeature> list = new ArrayList<CMFeature>(count);

		for (int i = 0; i < count; i++) {
			CMFeature feature = ((CSBaseFeature) getComponent(i)).getFeature();

			if (feature != null) {
				list.add(feature);
			}
		}
		return list;
	}
}
