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

package com.trollworks.gcs.ui.image;

import com.trollworks.gcs.ui.RetinaIcon;
import com.trollworks.gcs.utility.Log;

import java.awt.Transparency;
import java.io.InputStream;
import java.util.Arrays;
import java.util.List;

public class Images {
    public static final List<Img>  APP_ICON_LIST         = Arrays.asList(get("app_1024"), get("app_512"), get("app_256"), get("app_128"), get("app_64"), get("app_32"), get("app_16"));
    public static final RetinaIcon ABOUT                 = getRetina("about");
    public static final RetinaIcon ACTUAL_SIZE           = getRetina("actual_size");
    public static final RetinaIcon ADD                   = getRetina("add");
    public static final RetinaIcon ADM_FILE              = getRetina("adm_file");
    public static final RetinaIcon ADM_MARKER            = getRetina("adm_marker");
    public static final RetinaIcon ADQ_FILE              = getRetina("adq_file");
    public static final RetinaIcon ADQ_MARKER            = getRetina("adq_marker");
    public static final RetinaIcon COLLAPSE              = getRetina("collapse");
    public static final RetinaIcon DISCLOSURE_DOWN       = getRetina("disclosure_down");
    public static final RetinaIcon DISCLOSURE_DOWN_ROLL  = getRetina("disclosure_down_roll");
    public static final RetinaIcon DISCLOSURE_RIGHT      = getRetina("disclosure_right");
    public static final RetinaIcon DISCLOSURE_RIGHT_ROLL = getRetina("disclosure_right_roll");
    public static final RetinaIcon DOCK_CLOSE            = getRetina("dock_close");
    public static final RetinaIcon DOCK_MAXIMIZE         = getRetina("dock_maximize");
    public static final RetinaIcon DOCK_RESTORE          = getRetina("dock_restore");
    public static final RetinaIcon EQM_FILE              = getRetina("eqm_file");
    public static final RetinaIcon EQM_MARKER            = getRetina("eqm_marker");
    public static final RetinaIcon EQP_FILE              = getRetina("eqp_file");
    public static final RetinaIcon EQP_MARKER            = getRetina("eqp_marker");
    public static final RetinaIcon EXOTIC_TYPE           = getRetina("exotic_type");
    public static final RetinaIcon EXPAND                = getRetina("expand");
    public static final RetinaIcon FILE                  = getRetina("file");
    public static final RetinaIcon FOLDER                = getRetina("folder");
    public static final RetinaIcon GCS_FILE              = getRetina("gcs_file");
    public static final RetinaIcon GCS_MARKER            = getRetina("gcs_marker");
    public static final RetinaIcon GCT_FILE              = getRetina("gct_file");
    public static final RetinaIcon GCT_MARKER            = getRetina("gct_marker");
    public static final RetinaIcon GEAR                  = getRetina("gear");
    public static final RetinaIcon LOCKED                = getRetina("locked");
    public static final RetinaIcon MENTAL_TYPE           = getRetina("mental_type");
    public static final RetinaIcon MORE                  = getRetina("more");
    public static final RetinaIcon NOT_FILE              = getRetina("not_file");
    public static final RetinaIcon NOT_MARKER            = getRetina("not_marker");
    public static final RetinaIcon PAGE_DOWN             = getRetina("page_down");
    public static final RetinaIcon PAGE_UP               = getRetina("page_up");
    public static final RetinaIcon PDF_FILE              = getRetina("pdf_file");
    public static final RetinaIcon PHYSICAL_TYPE         = getRetina("physical_type");
    public static final RetinaIcon REFRESH               = getRetina("refresh");
    public static final RetinaIcon REMOVE                = getRetina("remove");
    public static final RetinaIcon SIZE_TO_FIT           = getRetina("size_to_fit");
    public static final RetinaIcon SKL_FILE              = getRetina("skl_file");
    public static final RetinaIcon SKL_MARKER            = getRetina("skl_marker");
    public static final RetinaIcon SOCIAL_TYPE           = getRetina("social_type");
    public static final RetinaIcon SPL_FILE              = getRetina("spl_file");
    public static final RetinaIcon SPL_MARKER            = getRetina("spl_marker");
    public static final RetinaIcon SUPERNATURAL_TYPE     = getRetina("supernatural_type");
    public static final RetinaIcon TOGGLE_OPEN           = getRetina("toggle_open");
    public static final RetinaIcon UNLOCKED              = getRetina("unlocked");
    public static final RetinaIcon ZOOM_IN               = getRetina("zoom_in");
    public static final RetinaIcon ZOOM_OUT              = getRetina("zoom_out");
    public static final Img        DEFAULT_PORTRAIT      = get("default_portrait");

    static final synchronized Img get(String name) {
        name += ".png";
        try (InputStream in = Img.class.getModule().getResourceAsStream("/images/" + name)) {
            return Img.create(in);
        } catch (Exception exception) {
            Log.error("unable to load image for: " + name);
            return Img.create(1, 1, Transparency.TRANSLUCENT);
        }
    }

    static final RetinaIcon getRetina(String name) {
        return new RetinaIcon(get(name), get(name + "@2x"));
    }
}
