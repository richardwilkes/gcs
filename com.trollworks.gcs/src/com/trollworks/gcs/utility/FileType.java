/*
 * Copyright ©1998-2022 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, version 2.0. If a copy of the MPL was not distributed with
 * this file, You can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as
 * defined by the Mozilla Public License, version 2.0.
 */

package com.trollworks.gcs.utility;

import com.trollworks.gcs.ui.FontAwesome;
import com.trollworks.gcs.ui.FontIcon;
import com.trollworks.gcs.ui.Fonts;
import com.trollworks.gcs.ui.RPGAwesome;
import com.trollworks.gcs.ui.ThemeFont;

import java.util.ArrayList;
import java.util.Arrays;
import java.util.List;
import javax.swing.Icon;
import javax.swing.filechooser.FileNameExtensionFilter;

/** Describes a file. */
public class FileType {
    public static final List<FileType> OPENABLE     = new ArrayList<>();
    public static final List<FileType> ALL_OPENABLE = new ArrayList<>();

    public static final FileType SHEET              = new FileType(I18n.text("Character Sheet"), RPGAwesome.PLAYER, Fonts.FONT_ICON_FILE_RPG, true, true, "gcs");
    public static final FileType TEMPLATE           = new FileType(I18n.text("Character Template"), RPGAwesome.DOUBLE_TEAM, Fonts.FONT_ICON_FILE_RPG, true, true, "gct");
    public static final FileType ADVANTAGE          = new FileType(I18n.text("Advantages Library"), RPGAwesome.ANKH, Fonts.FONT_ICON_FILE_RPG, true, true, "adq");
    public static final FileType ADVANTAGE_MODIFIER = new FileType(I18n.text("Advantage Modifiers Library"), RPGAwesome.RADIAL_BALANCE, Fonts.FONT_ICON_FILE_RPG, true, true, "adm");
    public static final FileType EQUIPMENT          = new FileType(I18n.text("Equipment Library"), RPGAwesome.AXE, Fonts.FONT_ICON_FILE_RPG, true, true, "eqp");
    public static final FileType EQUIPMENT_MODIFIER = new FileType(I18n.text("Equipment Modifiers Library"), RPGAwesome.GEAR_HAMMER, Fonts.FONT_ICON_FILE_RPG, true, true, "eqm");
    public static final FileType SKILL              = new FileType(I18n.text("Skills Library"), RPGAwesome.FORGING, Fonts.FONT_ICON_FILE_RPG, true, true, "skl");
    public static final FileType SPELL              = new FileType(I18n.text("Spells Library"), RPGAwesome.SCROLL_UNFURLED, Fonts.FONT_ICON_FILE_RPG, true, true, "spl");
    public static final FileType NOTE               = new FileType(I18n.text("Notes Library"), RPGAwesome.WOODEN_SIGN, Fonts.FONT_ICON_FILE_RPG, true, true, "not");

    public static final FileType ATTRIBUTE_SETTINGS      = new FileType(I18n.text("Attribute Settings"), FontAwesome.FILE, Fonts.FONT_ICON_FILE_FA, true, false, "attributes", "gas");
    public static final FileType BODY_SETTINGS           = new FileType(I18n.text("Body Type Settings"), FontAwesome.FILE, Fonts.FONT_ICON_FILE_FA, true, false, "body", "ghl");
    public static final FileType COLOR_SETTINGS          = new FileType(I18n.text("Color Settings"), FontAwesome.FILE, Fonts.FONT_ICON_FILE_FA, true, false, "colors");
    public static final FileType FONT_SETTINGS           = new FileType(I18n.text("Font Settings"), FontAwesome.FILE, Fonts.FONT_ICON_FILE_FA, true, false, "fonts");
    public static final FileType GENERAL_SETTINGS        = new FileType(I18n.text("General Settings"), FontAwesome.FILE, Fonts.FONT_ICON_FILE_FA, true, false, "general");
    public static final FileType KEY_SETTINGS            = new FileType(I18n.text("Key Binding Settings"), FontAwesome.FILE, Fonts.FONT_ICON_FILE_FA, true, false, "keys");
    public static final FileType PAGE_REF_SETTINGS       = new FileType(I18n.text("Page Reference Settings"), FontAwesome.FILE, Fonts.FONT_ICON_FILE_FA, true, false, "refs");
    public static final FileType SHEET_SETTINGS          = new FileType(I18n.text("Sheet Settings"), FontAwesome.FILE, Fonts.FONT_ICON_FILE_FA, true, false, "sheet");
    public static final FileType ANCESTRY_SETTINGS       = new FileType(I18n.text("Ancestry Settings"), FontAwesome.FILE, Fonts.FONT_ICON_FILE_FA, true, false, "ancestry");
    public static final FileType NAME_GENERATOR_SETTINGS = new FileType(I18n.text("Name Generator Settings"), FontAwesome.FILE, Fonts.FONT_ICON_FILE_FA, true, false, "names");
    public static final FileType CALENDAR_SETTINGS       = new FileType(I18n.text("Calendar Settings"), FontAwesome.FILE, Fonts.FONT_ICON_FILE_FA, true, false, "calendar");

    public static final FileType PDF  = new FileType(I18n.text("PDF Files"), FontAwesome.FILE_PDF, Fonts.FONT_ICON_FILE_FA, false, true, "pdf");
    public static final FileType PNG  = new FileType(I18n.text("PNG Files"), FontAwesome.FILE_IMAGE, Fonts.FONT_ICON_FILE_FA, false, false, "png");
    public static final FileType JPEG = new FileType(I18n.text("JPEG Files"), FontAwesome.FILE_IMAGE, Fonts.FONT_ICON_FILE_FA, false, false, "jpg", "jpeg");
    public static final FileType GIF  = new FileType(I18n.text("GIF Files"), FontAwesome.FILE_IMAGE, Fonts.FONT_ICON_FILE_FA, false, false, "gif");

    public static final FileNameExtensionFilter[] IMAGE_FILTERS = createFileFilters(I18n.text("Image Files"), PNG, JPEG, GIF);

    private String                  mIconName;
    private ThemeFont               mIconFont;
    private String                  mPrimaryExtension;
    private FileNameExtensionFilter mFilter;

    public FileType(String description, String iconName, ThemeFont font, boolean owned, boolean openable, String... extension) {
        mIconName = iconName;
        mIconFont = font;
        mPrimaryExtension = extension[0];
        mFilter = new FileNameExtensionFilter(description, extension);
        if (openable) {
            if (owned) {
                OPENABLE.add(this);
            }
            ALL_OPENABLE.add(this);
        }
    }

    public Icon getIcon() {
        return new FontIcon(mIconName, mIconFont);
    }

    public String getUntitledDefaultFileName() {
        return addExtension(I18n.text("untitled"));
    }

    public String addExtension(String filename) {
        return filename + "." + mPrimaryExtension;
    }

    /** @return The primary extension of the file. */
    public String getExtension() {
        return mPrimaryExtension;
    }

    /**
     * @param extension The extension to check.
     * @return {@code true} if the extension matches one for this file type.
     */
    public boolean matchExtension(String extension) {
        for (String ext : mFilter.getExtensions()) {
            if (ext.equalsIgnoreCase(extension)) {
                return true;
            }
        }
        return false;
    }

    /** @return The file filter. */
    public FileNameExtensionFilter getFilter() {
        return mFilter;
    }

    /**
     * @param nameForAggregate The name to use for an aggregate filter that accepts all of the file
     *                         types in the list. Pass in {@code null} for no aggregate filter.
     * @param fileTypes        The set of file types to return filters for.
     * @return A set of file filters for the specified file types.
     */
    public static FileNameExtensionFilter[] createFileFilters(String nameForAggregate, FileType... fileTypes) {
        List<String>                  extensions      = new ArrayList<>();
        List<FileNameExtensionFilter> filters         = new ArrayList<>();
        boolean                       createAggregate = nameForAggregate != null && !nameForAggregate.isBlank();
        for (FileType fileType : fileTypes) {
            if (createAggregate) {
                extensions.addAll(Arrays.asList(fileType.getFilter().getExtensions()));
            }
            filters.add(fileType.getFilter());
        }
        if (createAggregate) {
            filters.add(0, new FileNameExtensionFilter(nameForAggregate, extensions.toArray(new String[0])));
        }
        return filters.toArray(new FileNameExtensionFilter[0]);
    }

    /**
     * @param name The name to return an icon for.
     * @return The icon for the specified file.
     */
    public static Icon getIconForFileName(String name) {
        String extension = PathUtils.getExtension(name);
        for (FileType one : ALL_OPENABLE) {
            if (one.matchExtension(extension)) {
                return one.getIcon();
            }
        }
        return new FontIcon(FontAwesome.FILE, Fonts.FONT_ICON_LABEL_PRIMARY);
    }
}
