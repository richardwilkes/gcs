/*
 * Copyright (c) 1998-2014 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public License,
 * version 2.0. If a copy of the MPL was not distributed with this file, You
 * can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as defined
 * by the Mozilla Public License, version 2.0.
 */

package com.trollworks.gcs.app;

import com.trollworks.toolkit.io.Log;
import com.trollworks.toolkit.ui.image.IconSet;
import com.trollworks.toolkit.ui.image.Images;
import com.trollworks.toolkit.ui.image.ToolkitIcon;

import java.io.BufferedOutputStream;
import java.io.File;
import java.io.FileOutputStream;
import java.io.IOException;
import java.io.InputStream;
import java.util.ArrayList;
import java.util.HashSet;
import java.util.List;
import java.util.Set;

/** Provides standardized image access. */
@SuppressWarnings("nls")
public class GCSImages {
	private static final Set<String>	ICON_SETS_ATTEMPTED	= new HashSet<>();
	private static final String			SMALL				= "Small";
	private static final String			LARGE				= "Large";
	private static final String			SINGLE				= "Single";
	private static final String			FILE				= "";

	static {
		Images.addLocation(GCSImages.class.getResource("images/"));
	}

	private static final IconSet getOrLoadIconSet(String name) {
		IconSet iconSet = IconSet.get(name);
		if (iconSet == null) {
			if (ICON_SETS_ATTEMPTED.add(name)) {
				try (InputStream in = GCSImages.class.getResource("iconsets/" + name + ".icns").openStream()) {
					iconSet = IconSet.loadIcns(name, in);
				} catch (IOException ioe) {
					Log.error(ioe);
				}
			}
		}
		return iconSet;
	}

	/** @return The exotic type icon. */
	public static final ToolkitIcon getExoticTypeIcon() {
		return Images.get("ExoticType");
	}

	/** @return The selected exotic type icon. */
	public static final ToolkitIcon getExoticTypeSelectedIcon() {
		return Images.get("ExoticTypeSelected");
	}

	/** @return The mental type icon. */
	public static final ToolkitIcon getMentalTypeIcon() {
		return Images.get("MentalType");
	}

	/** @return The selected mental type icon. */
	public static final ToolkitIcon getMentalTypeSelectedIcon() {
		return Images.get("MentalTypeSelected");
	}

	/** @return The physical type icon. */
	public static final ToolkitIcon getPhysicalTypeIcon() {
		return Images.get("PhysicalType");
	}

	/** @return The selected physical type icon. */
	public static final ToolkitIcon getPhysicalTypeSelectedIcon() {
		return Images.get("PhysicalTypeSelected");
	}

	/** @return The social type icon. */
	public static final ToolkitIcon getSocialTypeIcon() {
		return Images.get("SocialType");
	}

	/** @return The selected social type icon. */
	public static final ToolkitIcon getSocialTypeSelectedIcon() {
		return Images.get("SocialTypeSelected");
	}

	/** @return The supernatural type icon. */
	public static final ToolkitIcon getSupernaturalTypeIcon() {
		return Images.get("SupernaturalType");
	}

	/** @return The selected supernatural type icon. */
	public static final ToolkitIcon getSupernaturalTypeSelectedIcon() {
		return Images.get("SupernaturalTypeSelected");
	}

	/** @return The default portrait. */
	public static final ToolkitIcon getDefaultPortrait() {
		return Images.get("DefaultPortrait");
	}

	/** @return The 'about' image. */
	public static final ToolkitIcon getAbout() {
		return Images.get("About");
	}

	/** @return The application icons. */
	public static final IconSet getAppIcons() {
		return getOrLoadIconSet("app");
	}

	/** @return The character sheet icons. */
	public static final IconSet getCharacterSheetIcons() {
		return getOrLoadIconSet("gcs");
	}

	/** @return The character template icons. */
	public static final IconSet getTemplateIcons() {
		return getOrLoadIconSet("gct");
	}

	/** @return The advantages icons. */
	public static final IconSet getAdvantagesIcons() {
		return getOrLoadIconSet("adq");
	}

	/** @return The skills icons. */
	public static final IconSet getSkillsIcons() {
		return getOrLoadIconSet("skl");
	}

	/** @return The spells icons. */
	public static final IconSet getSpellsIcons() {
		return getOrLoadIconSet("spl");
	}

	/** @return The equipment icons. */
	public static final IconSet getEquipmentIcons() {
		return getOrLoadIconSet("eqp");
	}

	/**
	 * @param large The large (32x32) or the small (16x16) version.
	 * @param single The single or file version.
	 * @return The advantage icon.
	 */
	public static final ToolkitIcon getAdvantageIcon(boolean large, boolean single) {
		return Images.get("Advantage" + (large ? LARGE : SMALL) + (single ? SINGLE : FILE));
	}

	/**
	 * @param large The large (32x32) or the small (16x16) version.
	 * @param single The single or file version.
	 * @return The skill icon.
	 */
	public static final ToolkitIcon getSkillIcon(boolean large, boolean single) {
		return Images.get("Skill" + (large ? LARGE : SMALL) + (single ? SINGLE : FILE));
	}

	/**
	 * @param large The large (32x32) or the small (16x16) version.
	 * @param single The single or file version.
	 * @return The spell icon.
	 */
	public static final ToolkitIcon getSpellIcon(boolean large, boolean single) {
		return Images.get("Spell" + (large ? LARGE : SMALL) + (single ? SINGLE : FILE));
	}

	/**
	 * @param large The large (32x32) or the small (16x16) version.
	 * @param single The single or file version.
	 * @return The equipment icon.
	 */
	public static final ToolkitIcon getEquipmentIcon(boolean large, boolean single) {
		return Images.get("Equipment" + (large ? LARGE : SMALL) + (single ? SINGLE : FILE));
	}

	/** Utility for creating GCS's icon sets. */
	public static void main(String[] args) {
		try {
			ToolkitIcon base1024 = Images.loadImage(new File("graphics/iconset_parts/document_1024.png"));
			ToolkitIcon base64 = Images.loadImage(new File("graphics/iconset_parts/document_64.png"));
			ToolkitIcon base32 = Images.loadImage(new File("graphics/iconset_parts/document_32.png"));
			ToolkitIcon base16 = Images.loadImage(new File("graphics/iconset_parts/document_16.png"));
			for (String one : new String[] { "adq", "eqp", "gcs", "gct", "skl", "spl" }) {
				createIcns(one, base1024, base64, base32, base16);
			}
			List<ToolkitIcon> images = new ArrayList<>();
			ToolkitIcon icon = Images.loadImage(new File("graphics/iconset_parts/app_1024.png"));
			images.add(icon);
			images.add(Images.scale(icon, 512));
			images.add(Images.scale(icon, 256));
			images.add(Images.scale(icon, 128));
			images.add(Images.scale(icon, 64));
			images.add(Images.scale(icon, 32));
			images.add(Images.scale(icon, 16));
			IconSet set = new IconSet("app", images);
			try (BufferedOutputStream out = new BufferedOutputStream(new FileOutputStream("src/com/trollworks/gcs/app/iconsets/app.icns"))) {
				set.saveAsIcns(out);
			}
		} catch (Exception exception) {
			exception.printStackTrace(System.err);
		}
	}

	private static void createIcns(String name, ToolkitIcon base1024, ToolkitIcon base64, ToolkitIcon base32, ToolkitIcon base16) throws IOException {
		List<ToolkitIcon> images = new ArrayList<>();
		ToolkitIcon icon = Images.loadImage(new File("graphics/iconset_parts/" + name + "_1024.png"));
		images.add(Images.superimpose(base1024, icon));
		images.add(Images.superimpose(Images.scale(base1024, 512), Images.scale(icon, 512)));
		images.add(Images.superimpose(Images.scale(base1024, 256), Images.scale(icon, 256)));
		images.add(Images.superimpose(Images.scale(base1024, 128), Images.scale(icon, 128)));
		images.add(Images.superimpose(base64, Images.scale(icon, 64)));
		images.add(Images.superimpose(base32, Images.scale(icon, 32)));
		images.add(Images.superimpose(base16, Images.scale(icon, 16)));
		IconSet set = new IconSet(name, images);
		try (BufferedOutputStream out = new BufferedOutputStream(new FileOutputStream("src/com/trollworks/gcs/app/iconsets/" + name + ".icns"))) {
			set.saveAsIcns(out);
		}
	}
}
