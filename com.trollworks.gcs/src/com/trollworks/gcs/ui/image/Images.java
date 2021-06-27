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

package com.trollworks.gcs.ui.image;

import com.trollworks.gcs.character.Profile;
import com.trollworks.gcs.ui.RetinaIcon;
import com.trollworks.gcs.utility.Log;

import java.awt.Transparency;
import java.io.InputStream;
import java.util.Arrays;
import java.util.List;

public final class Images {
    public static final List<Img>  APP_ICON_LIST    = Arrays.asList(
            get("app_1024"),
            get("app_512"),
            get("app_256"),
            get("app_128"),
            get("app_64"),
            get("app_32"),
            get("app_16"));
    public static final RetinaIcon ABOUT            = getRetina("about");
    public static final RetinaIcon DEFAULT_PORTRAIT = Profile.createPortrait(get("default_portrait"));

    private Images() {
    }

    static synchronized Img get(String name) {
        name += ".png";
        try (InputStream in = Img.class.getModule().getResourceAsStream("/images/" + name)) {
            return Img.create(in);
        } catch (Exception exception) {
            Log.error("unable to load image for: " + name);
            return Img.create(1, 1, Transparency.TRANSLUCENT);
        }
    }

    static RetinaIcon getRetina(String name) {
        return new RetinaIcon(get(name), get(name + "@2x"));
    }
}
