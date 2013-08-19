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

package com.trollworks.toolkit.io;

import com.trollworks.toolkit.utility.TKColor;
import com.trollworks.toolkit.utility.TKDebug;
import com.trollworks.toolkit.utility.TKGraphics;

import java.awt.Color;
import java.awt.Graphics2D;
import java.awt.GraphicsConfiguration;
import java.awt.Image;
import java.awt.Toolkit;
import java.awt.Transparency;
import java.awt.image.BufferedImage;
import java.awt.image.ImageObserver;
import java.io.ByteArrayInputStream;
import java.io.File;
import java.io.FileOutputStream;
import java.io.OutputStream;
import java.net.URL;
import java.util.HashMap;
import java.util.HashSet;

import javax.imageio.IIOImage;
import javax.imageio.ImageIO;
import javax.imageio.ImageTypeSpecifier;
import javax.imageio.ImageWriter;
import javax.imageio.metadata.IIOMetadata;
import javax.imageio.metadata.IIOMetadataNode;
import javax.imageio.stream.ImageOutputStream;

import org.w3c.dom.Node;

/** Provides standardized image access. */
public class TKImage {
	/** The extensions used for image files. */
	public static final String[]						EXTENSIONS			= { ".png", ".gif", ".jpg" };				//$NON-NLS-1$ //$NON-NLS-2$ //$NON-NLS-3$
	private static final String							COLORIZED_POSTFIX	= new String(new char[] { ':', 'C', 22 });
	private static final String							FADED_POSTFIX		= new String(new char[] { ':', 'F', 22 });
	private static final HashSet<URL>					LOCATIONS			= new HashSet<URL>();
	private static final HashMap<String, BufferedImage>	MAP					= new HashMap<String, BufferedImage>();
	private static final HashMap<BufferedImage, String>	REVERSE_MAP			= new HashMap<BufferedImage, String>();
	private static final HashSet<String>				FAILED_LOOKUPS		= new HashSet<String>();

	static {
		LOCATIONS.add(TKImage.class.getResource("images/")); //$NON-NLS-1$
	}

	/**
	 * Adds a location to search for images.
	 * 
	 * @param file A directory to be be searched.
	 */
	public static final synchronized void addLocation(File file) {
		try {
			if (file.isDirectory()) {
				addLocation(file.toURI().toURL());
			}
		} catch (Exception exception) {
			assert false : TKDebug.throwableToString(exception);
		}
	}

	/**
	 * Adds a location to search for images.
	 * 
	 * @param url The location this URL points to will be searched.
	 */
	public static final synchronized void addLocation(URL url) {
		if (!LOCATIONS.contains(url)) {
			LOCATIONS.add(url);
			FAILED_LOOKUPS.clear();
		}
	}

	/**
	 * @param name The name to search for.
	 * @return The image for the specified name.
	 */
	public static final synchronized BufferedImage get(String name) {
		return get(name, true);
	}

	/**
	 * @param name The name to search for.
	 * @param cache Whether or not to cache the image. If <code>cache</code> is <code>false</code>,
	 *            then a new copy of the image will be generated, even if it was in the cache
	 *            already.
	 * @return The image for the specified name.
	 */
	public static final synchronized BufferedImage get(String name, boolean cache) {
		BufferedImage img = cache ? MAP.get(name) : null;

		if (img == null && !FAILED_LOOKUPS.contains(name)) {
			for (URL url : LOCATIONS) {
				for (int i = 0; i < EXTENSIONS.length && img == null; i++) {
					String filename = name + EXTENSIONS[i];

					try {
						img = loadImage(new URL(url, filename));
					} catch (Exception exception) {
						// Ignore...
					}
				}
				if (img != null) {
					break;
				}
			}

			if (img != null) {
				if (cache) {
					MAP.put(name, img);
					REVERSE_MAP.put(img, name);
				}
			} else {
				FAILED_LOOKUPS.add(name);
			}
		}
		return img;
	}

	/**
	 * Manually adds the specified image to the cache. If an image was already mapped to the
	 * specified name, it will be replaced.
	 * 
	 * @param name The name to map the image to.
	 * @param image The image itself.
	 */
	public static final synchronized void add(String name, BufferedImage image) {
		MAP.put(name, image);
		REVERSE_MAP.put(image, name);
		FAILED_LOOKUPS.remove(name);
	}

	/**
	 * Removes from the image cache any image cached for the specified name.
	 * 
	 * @param name The key for the image to release.
	 */
	public static final synchronized void release(String name) {
		Object img = MAP.remove(name);

		if (img != null) {
			REVERSE_MAP.remove(img);
		}
	}

	/**
	 * @param width The width to scale the image to.
	 * @param height The hight to scale the image to.
	 * @return A new image of the given size.
	 */
	public static final BufferedImage create(int width, int height) {
		Graphics2D g2d = TKGraphics.getGraphics();
		GraphicsConfiguration gc = g2d.getDeviceConfiguration();
		BufferedImage buffer = gc.createCompatibleImage(width, height, Transparency.TRANSLUCENT);

		g2d.dispose();
		g2d = (Graphics2D) buffer.getGraphics();
		g2d.setClip(0, 0, width, height);
		g2d.setBackground(new Color(0, true));
		g2d.clearRect(0, 0, width, height);
		g2d.dispose();
		return buffer;
	}

	/**
	 * @param image The image to scale.
	 * @param scaledWidth The width to scale the image to.
	 * @param scaledHeight The hight to scale the image to.
	 * @return The new image scaled to the given size.
	 */
	public static final BufferedImage scale(BufferedImage image, int scaledWidth, int scaledHeight) {
		int width = image.getWidth();
		int height = image.getHeight();
		Graphics2D g2d = TKGraphics.getGraphics();
		GraphicsConfiguration gc = g2d.getDeviceConfiguration();
		BufferedImage buffer = gc.createCompatibleImage(scaledWidth, scaledHeight, Transparency.TRANSLUCENT);
		Image scaledImg;

		if (width != scaledWidth || height != scaledHeight) {
			double wMult = width / (double) scaledWidth;
			double hMult = height / (double) scaledHeight;

			if (wMult > hMult) {
				scaledImg = loadToolkitImage(image.getScaledInstance(scaledWidth, -1, Image.SCALE_SMOOTH));
			} else {
				scaledImg = loadToolkitImage(image.getScaledInstance(-1, scaledHeight, Image.SCALE_SMOOTH));
			}
		} else {
			scaledImg = image;
		}

		g2d.dispose();
		g2d = (Graphics2D) buffer.getGraphics();
		g2d.setClip(0, 0, scaledWidth, scaledHeight);
		g2d.setBackground(new Color(0, true));
		g2d.clearRect(0, 0, scaledWidth, scaledHeight);
		if (scaledImg != null) {
			g2d.drawImage(scaledImg, (scaledWidth - scaledImg.getWidth(null)) / 2, (scaledHeight - scaledImg.getHeight(null)) / 2, null);
		}
		g2d.dispose();

		return buffer;
	}

	/**
	 * If the image passed in is already a {@link BufferedImage}, it is returned. However, if it is
	 * not, then a new {@link BufferedImage} is created with the contents of the image and returned.
	 * 
	 * @param image The image to work on.
	 * @return A buffered image.
	 */
	public static final BufferedImage getBufferedImage(Image image) {
		if (!(image instanceof BufferedImage)) {
			loadToolkitImage(image);
			return createOptimizedImage(image);
		}
		return (BufferedImage) image;
	}

	private static BufferedImage createOptimizedImage(Image image) {
		Graphics2D g2d = TKGraphics.getGraphics();
		GraphicsConfiguration gc = g2d.getDeviceConfiguration();
		int width = image.getWidth(null);
		int height = image.getHeight(null);
		BufferedImage buffer = gc.createCompatibleImage(width, height, Transparency.TRANSLUCENT);

		g2d.dispose();
		g2d = (Graphics2D) buffer.getGraphics();
		g2d.setClip(0, 0, width, height);
		g2d.setBackground(new Color(0, true));
		g2d.clearRect(0, 0, width, height);
		g2d.drawImage(image, 0, 0, null);
		g2d.dispose();
		return buffer;
	}

	/**
	 * @param image The image to work on.
	 * @return A new image rotated 90 degrees.
	 */
	public static final BufferedImage rotate90(BufferedImage image) {
		Graphics2D g2d = TKGraphics.getGraphics();
		GraphicsConfiguration gc = g2d.getDeviceConfiguration();
		int width = image.getWidth();
		int height = image.getHeight();
		BufferedImage buffer = gc.createCompatibleImage(height, width, Transparency.TRANSLUCENT);

		g2d.dispose();
		g2d = (Graphics2D) buffer.getGraphics();
		g2d.setClip(0, 0, height, width);
		g2d.setBackground(new Color(0, true));
		g2d.clearRect(0, 0, height, width);
		g2d.rotate(Math.toRadians(90));
		g2d.drawImage(image, 0, -width, null);
		g2d.dispose();
		return buffer;
	}

	/**
	 * @param image The image to work on.
	 * @return A new image rotated 180 degrees.
	 */
	public static final BufferedImage rotate180(BufferedImage image) {
		Graphics2D g2d = TKGraphics.getGraphics();
		GraphicsConfiguration gc = g2d.getDeviceConfiguration();
		int width = image.getWidth();
		int height = image.getHeight();
		BufferedImage buffer = gc.createCompatibleImage(height, width, Transparency.TRANSLUCENT);

		g2d.dispose();
		g2d = (Graphics2D) buffer.getGraphics();
		g2d.setClip(0, 0, width, height);
		g2d.setBackground(new Color(0, true));
		g2d.clearRect(0, 0, width, height);
		g2d.rotate(Math.toRadians(180));
		g2d.drawImage(image, -width, -height, null);
		g2d.dispose();
		return buffer;
	}

	/**
	 * @param image The image to work on.
	 * @return A new image rotated 270 degrees.
	 */
	public static final BufferedImage rotate270(BufferedImage image) {
		Graphics2D g2d = TKGraphics.getGraphics();
		GraphicsConfiguration gc = g2d.getDeviceConfiguration();
		int width = image.getWidth();
		int height = image.getHeight();
		BufferedImage buffer = gc.createCompatibleImage(height, width, Transparency.TRANSLUCENT);

		g2d.dispose();
		g2d = (Graphics2D) buffer.getGraphics();
		g2d.setClip(0, 0, height, width);
		g2d.setBackground(new Color(0, true));
		g2d.clearRect(0, 0, height, width);
		g2d.rotate(Math.toRadians(270));
		g2d.drawImage(image, -height, 0, null);
		g2d.dispose();
		return buffer;
	}

	/**
	 * Creates a new image by superimposing the specified image centered on top of another.
	 * 
	 * @param baseImage The base image.
	 * @param image The image to superimpose.
	 * @return The new image.
	 */
	public static final BufferedImage superimpose(BufferedImage baseImage, BufferedImage image) {
		return superimpose(baseImage, image, (baseImage.getWidth() - image.getWidth()) / 2, (baseImage.getHeight() - image.getHeight()) / 2);
	}

	/**
	 * Creates a new image by superimposing the specified image on top of another.
	 * 
	 * @param baseImage The base image.
	 * @param image The image to superimpose.
	 * @param x The x-coordinate to draw the top image at.
	 * @param y The y-coordinate to draw the top image at.
	 * @return The new image.
	 */
	public static final BufferedImage superimpose(BufferedImage baseImage, BufferedImage image, int x, int y) {
		int width = baseImage.getWidth();
		int height = baseImage.getHeight();
		int tWidth = image.getWidth();
		int tHeight = image.getHeight();
		Graphics2D g2d = TKGraphics.getGraphics();
		GraphicsConfiguration gc = g2d.getDeviceConfiguration();
		BufferedImage buffer;

		if (x + tWidth > width) {
			width = x + tWidth;
		}
		if (y + tHeight > height) {
			height = y + tHeight;
		}

		buffer = gc.createCompatibleImage(width, height, Transparency.TRANSLUCENT);
		g2d.dispose();
		g2d = (Graphics2D) buffer.getGraphics();
		g2d.setClip(0, 0, width, height);
		g2d.setBackground(new Color(0, true));
		g2d.clearRect(0, 0, width, height);
		g2d.drawImage(baseImage, 0, 0, null);
		g2d.drawImage(image, x, y, null);
		g2d.dispose();
		return buffer;
	}

	/**
	 * Creates a new image by superimposing the specified image centered on top of another. In
	 * addition, the newly created image is added as a named image.
	 * 
	 * @param baseImage The base image.
	 * @param image The image to superimpose.
	 * @param name The name to use.
	 * @return The new image.
	 */
	public static final BufferedImage superimposeAndName(BufferedImage baseImage, BufferedImage image, String name) {
		BufferedImage img = superimpose(baseImage, image);

		add(name, img);
		return img;
	}

	/**
	 * Creates a new image by superimposing the specified image on top of another. In addition, the
	 * newly created image is added as a named image.
	 * 
	 * @param baseImage The base image.
	 * @param image The image to superimpose.
	 * @param x The x-coordinate to draw the top image at.
	 * @param y The y-coordinate to draw the top image at.
	 * @param name The name to use.
	 * @return The new image.
	 */
	public static final BufferedImage superimposeAndName(BufferedImage baseImage, BufferedImage image, int x, int y, String name) {
		BufferedImage img = superimpose(baseImage, image, x, y);

		add(name, img);
		return img;
	}

	/**
	 * Creates a colorized version of an image.
	 * 
	 * @param image The image to work on.
	 * @param color The color to apply.
	 * @return The colorized image.
	 */
	public static final synchronized BufferedImage createColorizedImage(BufferedImage image, Color color) {
		String name = REVERSE_MAP.get(image);
		BufferedImage img = null;

		if (name != null) {
			name = name + TKColor.getStringValue(color) + COLORIZED_POSTFIX;
			img = get(name);
		}
		if (img == null) {
			img = TKColorFilter.createColorizedImage(image, color);
			if (img != null && name != null) {
				MAP.put(name, img);
				REVERSE_MAP.put(img, name);
			}
		}
		return img;
	}

	/**
	 * Creates a faded version of an image.
	 * 
	 * @param image The image to work on.
	 * @param percentage The percentage of black or white to use.
	 * @param useWhite Whether to use black or white.
	 * @return The faded image.
	 */
	public static final synchronized BufferedImage createFadedImage(BufferedImage image, int percentage, boolean useWhite) {
		String name = REVERSE_MAP.get(image);
		BufferedImage img = null;

		if (name != null) {
			name = name + percentage + useWhite + FADED_POSTFIX;
			img = get(name);
		}
		if (img == null) {
			img = TKFadeFilter.createFadedImage(image, percentage, useWhite);
			if (img != null && name != null) {
				MAP.put(name, img);
				REVERSE_MAP.put(img, name);
			}
		}
		return img;
	}

	/**
	 * Creates a disabled version of an image.
	 * 
	 * @param image The image to work on.
	 * @return The disabled image.
	 */
	public static final BufferedImage createDisabledImage(BufferedImage image) {
		return createFadedImage(createColorizedImage(image, Color.white), 50, true);
	}

	/**
	 * Forces the image to be loaded.
	 * 
	 * @param image The image to load.
	 * @return The fully loaded image, or <code>null</code> if it can't be loaded.
	 */
	public static final Image loadToolkitImage(Image image) {
		Toolkit tk = Toolkit.getDefaultToolkit();

		if (!tk.prepareImage(image, -1, -1, null)) {
			while (true) {
				int result = tk.checkImage(image, -1, -1, null);

				if ((result & (ImageObserver.ERROR | ImageObserver.ABORT)) != 0) {
					return null;
				}
				if ((result & (ImageObserver.ALLBITS | ImageObserver.SOMEBITS | ImageObserver.FRAMEBITS)) != 0) {
					break;
				}
				// Try to allow the image loading thread a chance to run.
				// At least on Windows, not sleeping seems to be a problem.
				try {
					Thread.sleep(1);
				} catch (InterruptedException ie) {
					return null;
				}
			}
		}
		return image;
	}

	/**
	 * Loads an optimized, buffered image from the specified file.
	 * 
	 * @param file The file to load from.
	 * @return The image, or <code>null</code> if it cannot be loaded.
	 */
	public static final BufferedImage loadImage(File file) {
		try {
			return loadImage(file.toURI().toURL());
		} catch (Exception exception) {
			return null;
		}
	}

	/**
	 * Loads an optimized, buffered image from the specified URL.
	 * 
	 * @param url The URL to load from.
	 * @return The image, or <code>null</code> if it cannot be loaded.
	 */
	public static final BufferedImage loadImage(URL url) {
		try {
			return createOptimizedImage(ImageIO.read(url));
		} catch (Exception exception) {
			return null;
		}
	}

	/**
	 * Loads an optimized, buffered image from the specified byte array.
	 * 
	 * @param data The byte array to load from.
	 * @return The image, or <code>null</code> if it cannot be loaded.
	 */
	public static final BufferedImage loadImage(byte[] data) {
		try {
			return createOptimizedImage(ImageIO.read(new ByteArrayInputStream(data)));
		} catch (Exception exception) {
			return null;
		}
	}

	/**
	 * Writes a PNG to a file, using the specified DPI.
	 * 
	 * @param file The file to write to.
	 * @param image The image to use.
	 * @param dpi The DPI to use.
	 * @return <code>true</code> on success.
	 */
	public static final boolean writePNG(File file, BufferedImage image, int dpi) {
		boolean result;

		try {
			FileOutputStream os = new FileOutputStream(file);

			result = writePNG(os, image, dpi);
			os.close();
		} catch (Exception exception) {
			result = false;
		}
		return result;
	}

	/**
	 * Writes a PNG to a stream, using the specified DPI.
	 * 
	 * @param os The stream to write to.
	 * @param image The image to use.
	 * @param dpi The DPI to use.
	 * @return <code>true</code> on success.
	 */
	public static final boolean writePNG(OutputStream os, BufferedImage image, int dpi) {
		ImageWriter writer = null;

		try {
			ImageOutputStream stream = ImageIO.createImageOutputStream(os);
			ImageTypeSpecifier type = ImageTypeSpecifier.createFromRenderedImage(image);
			IIOMetadata metaData;

			writer = ImageIO.getImageWriters(type, "png").next(); //$NON-NLS-1$
			metaData = writer.getDefaultImageMetadata(type, null);
			try {
				Node root = metaData.getAsTree("javax_imageio_png_1.0"); //$NON-NLS-1$
				IIOMetadataNode pHYs_node = new IIOMetadataNode("pHYs"); //$NON-NLS-1$
				String ppu = Integer.toString((int) (dpi / 0.0254));

				pHYs_node.setAttribute("pixelsPerUnitXAxis", ppu); //$NON-NLS-1$
				pHYs_node.setAttribute("pixelsPerUnitYAxis", ppu); //$NON-NLS-1$
				pHYs_node.setAttribute("unitSpecifier", "meter"); //$NON-NLS-1$ //$NON-NLS-2$
				root.appendChild(pHYs_node);
				metaData.setFromTree("javax_imageio_png_1.0", root); //$NON-NLS-1$
			} catch (Exception exception) {
				assert false : TKDebug.throwableToString(exception);
			}
			writer.setOutput(stream);
			writer.write(new IIOImage(image, null, metaData));
			stream.flush();
			writer.dispose();
			return true;
		} catch (Exception exception) {
			if (writer != null) {
				writer.dispose();
			}
			return false;
		}
	}

	/** @return The error dialog icon. */
	public static final BufferedImage getErrorIcon() {
		return get("TKError"); //$NON-NLS-1$
	}

	/** @return The warning dialog icon. */
	public static final BufferedImage getWarningIcon() {
		return get("TKWarning"); //$NON-NLS-1$
	}

	/** @return The mini warning dialog icon. */
	public static final BufferedImage getMiniWarningIcon() {
		return get("TKMiniWarning"); //$NON-NLS-1$
	}

	/** @return The message dialog icon. */
	public static final BufferedImage getMessageIcon() {
		return get("TKMessage"); //$NON-NLS-1$
	}

	/** @return The file icon. */
	public static final BufferedImage getFileIcon() {
		return get("TKFile"); //$NON-NLS-1$
	}

	/** @return The up folder icon. */
	public static final BufferedImage getUpFolderIcon() {
		return get("TKUpFolder"); //$NON-NLS-1$
	}

	/** @return The home icon. */
	public static final BufferedImage getHomeIcon() {
		return get("TKHome"); //$NON-NLS-1$
	}

	/** @return The folder icon. */
	public static final BufferedImage getFolderIcon() {
		return get("TKFolder"); //$NON-NLS-1$
	}

	/** @return The open folder icon. */
	public static final BufferedImage getOpenFolderIcon() {
		return get("TKOpenFolder"); //$NON-NLS-1$
	}

	/** @return The new folder icon. */
	public static final BufferedImage getNewFolderIcon() {
		return get("TKNewFolder"); //$NON-NLS-1$
	}

	/** @return The up arrow icon. */
	public static final BufferedImage getUpArrowIcon() {
		return get("TKUpArrow"); //$NON-NLS-1$
	}

	/** @return The down arrow icon. */
	public static final BufferedImage getDownArrowIcon() {
		return get("TKDownArrow"); //$NON-NLS-1$
	}

	/** @return The hierarchical menu arrow icon. */
	public static final BufferedImage getHierarchicalMenuArrowIcon() {
		return get("TKHierarchicalMenuArrow"); //$NON-NLS-1$
	}

	/** @return The check mark icon. */
	public static final BufferedImage getCheckMarkIcon() {
		return get("TKCheckMark"); //$NON-NLS-1$
	}

	/** @return The toggle open icon. */
	public static final BufferedImage getToggleOpenIcon() {
		return get("TKToggleOpen"); //$NON-NLS-1$
	}

	/** @return The size-to-fit icon. */
	public static final BufferedImage getSizeToFitIcon() {
		return get("TKSizeToFit"); //$NON-NLS-1$
	}

	/** @return The right triangle icon. */
	public static final BufferedImage getRightTriangleIcon() {
		return get("TKRightTriangle"); //$NON-NLS-1$
	}

	/** @return The right triangle roll icon. */
	public static final BufferedImage getRightTriangleRollIcon() {
		return get("TKRightTriangleRoll"); //$NON-NLS-1$
	}

	/** @return The down triangle icon. */
	public static final BufferedImage getDownTriangleIcon() {
		return get("TKDownTriangle"); //$NON-NLS-1$
	}

	/** @return The down triangle roll icon. */
	public static final BufferedImage getDownTriangleRollIcon() {
		return get("TKDownTriangleRoll"); //$NON-NLS-1$
	}

	/** @return The radio button empty icon. */
	public static final BufferedImage getRadioButtonEmptyIcon() {
		return get("TKRadioButtonEmpty"); //$NON-NLS-1$
	}

	/** @return The radio button selected icon. */
	public static final BufferedImage getRadioButtonSelectedIcon() {
		return get("TKRadioButtonSelected"); //$NON-NLS-1$
	}

	/** @return The radio button pressed empty icon. */
	public static final BufferedImage getRadioButtonPressedEmptyIcon() {
		return get("TKRadioButtonPressedEmpty"); //$NON-NLS-1$
	}

	/** @return The radio button pressed selected icon. */
	public static final BufferedImage getRadioButtonPressedSelectedIcon() {
		return get("TKRadioButtonPressedSelected"); //$NON-NLS-1$
	}

	/** @return The quick search icon. */
	public static final BufferedImage getQuickSearchIcon() {
		return get("TKQuickSearch"); //$NON-NLS-1$
	}

	/** @return The preferences icon. */
	public static final BufferedImage getPreferencesIcon() {
		return get("TKPreferences"); //$NON-NLS-1$
	}

	/**
	 * @param which The specific state (0-7) to return.
	 * @return The busy icon.
	 */
	public static final BufferedImage getBusyIcon(int which) {
		return get("TKBusyIndicator" + (which % 8)); //$NON-NLS-1$
	}

	/** @return The checkbox checked icon. */
	public static final BufferedImage getCheckboxCheckedIcon() {
		return get("TKCheckboxChecked"); //$NON-NLS-1$
	}

	/** @return The checkbox empty icon. */
	public static final BufferedImage getCheckboxEmptyIcon() {
		return get("TKCheckboxEmpty"); //$NON-NLS-1$
	}

	/** @return The checkbox mixed icon. */
	public static final BufferedImage getCheckboxMixedIcon() {
		return get("TKCheckboxMixed"); //$NON-NLS-1$
	}

	/** @return The checkbox pressed checked icon. */
	public static final BufferedImage getCheckboxPressedCheckedIcon() {
		return get("TKCheckboxPressedChecked"); //$NON-NLS-1$
	}

	/** @return The checkbox pressed empty icon. */
	public static final BufferedImage getCheckboxPressedEmptyIcon() {
		return get("TKCheckboxPressedEmpty"); //$NON-NLS-1$
	}

	/** @return The checkbox pressed mixed icon. */
	public static final BufferedImage getCheckboxPressedMixedIcon() {
		return get("TKCheckboxPressedMixed"); //$NON-NLS-1$
	}
}
