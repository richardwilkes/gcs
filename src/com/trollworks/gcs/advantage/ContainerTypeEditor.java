/*
 * Copyright (c) 1998-2018 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public License,
 * version 2.0. If a copy of the MPL was not distributed with this file, You
 * can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as defined
 * by the Mozilla Public License, version 2.0.
 */

package com.trollworks.gcs.advantage;

interface ContainerTypeEditor {
    AdvantageContainer getAdvantageContainer();
}

class SummativeContainerTypeEditor implements ContainerTypeEditor {
    static SummativeContainerTypeEditor containerTypeEditor = null;

    static SummativeContainerTypeEditor getInstance() {
        if (containerTypeEditor == null) {
            containerTypeEditor = new SummativeContainerTypeEditor();
        }
        return containerTypeEditor;
    }

    @Override
    public AdvantageContainer getAdvantageContainer() {
        return SummativeAdvantageContainer.getInstance();
    }
}

class AlternativeAbilitiesContainerTypeEditor implements ContainerTypeEditor {
    @Override
    public AdvantageContainer getAdvantageContainer() {
        return new AlternativeAbilitiesAdvantageContainer();
    }
}
