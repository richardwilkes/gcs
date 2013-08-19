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

package com.trollworks.toolkit.utility;

import java.awt.Rectangle;
import java.awt.geom.Rectangle2D;
import java.util.ArrayList;
import java.util.Arrays;
import java.util.Comparator;

/** Various rectangle utility routines. */
public class TKRectUtils {
	private static Comparator<Rectangle>	SORTER	= null;

	/**
	 * Returns a value measuring the area that would be wasted if the two bounds were merged into
	 * one.
	 * 
	 * @param first The first bounds.
	 * @param second The second bounds.
	 * @return A value representing the percentage area wasted.
	 */
	public static int areaWasted(Rectangle first, Rectangle second) {
		if (first.width > 0 && first.height > 0 && second.width > 0 && second.height > 0) {
			long combinedArea = (long) first.width * (long) first.height + (long) second.width * (long) second.height;
			Rectangle union = union(first, second);
			long unionArea = (long) union.height * (long) union.width;
			Rectangle overlap = intersection(first, second);
			long overlapArea = (long) overlap.height * (long) overlap.width;

			return (int) ((unionArea - (combinedArea - overlapArea)) * 100L / combinedArea);
		}
		return Integer.MAX_VALUE / 2;
	}

	/**
	 * Returns a value measuring the area that would be wasted if the two bounds were merged into
	 * one.
	 * 
	 * @param first The first bounds.
	 * @param second The second bounds.
	 * @return A value representing the percentage area wasted.
	 */
	public static double areaWasted(Rectangle2D first, Rectangle2D second) {
		double firstWidth = first.getWidth();
		double firstHeight = first.getHeight();
		double secondWidth = second.getWidth();
		double secondHeight = second.getHeight();

		if (firstWidth > 0.0 && firstHeight > 0.0 && secondWidth > 0.0 && secondHeight > 0.0) {
			double combinedArea = firstWidth * firstHeight + secondWidth * secondHeight;
			Rectangle2D union = union(first, second);
			double unionArea = union.getHeight() * union.getWidth();
			Rectangle2D overlap = intersection(first, second);
			double overlapArea = overlap.getHeight() * overlap.getWidth();

			return (unionArea - (combinedArea - overlapArea)) * 100.0 / combinedArea;
		}
		return Double.MAX_VALUE / 2;
	}

	/**
	 * @param first The first rectangle.
	 * @param second The second rectangle.
	 * @return An array of rectangles representing the regions within the first rectangle that do
	 *         not overlap with the second Rectangle. If the two rectangles do not overlap, the
	 *         first rectangle is returned.
	 */
	public static Rectangle[] computeDifference(Rectangle first, Rectangle second) {
		if (first == null) {
			return new Rectangle[0];
		}

		if (second == null || !first.intersects(second)) {
			return new Rectangle[] { new Rectangle(first) };
		}

		if (contains(second, first)) {
			return new Rectangle[0];
		}

		Rectangle tmp = new Rectangle();
		Rectangle ra = null;
		Rectangle rb = null;
		Rectangle rc = null;
		Rectangle rd = null;
		int count = 0;

		if (contains(first, second)) {
			tmp.x = first.x;
			tmp.y = first.y;
			tmp.width = second.x - first.x;
			tmp.height = first.height;
			if (tmp.width > 0 && tmp.height > 0) {
				ra = new Rectangle(tmp);
				count++;
			}

			tmp.x = second.x;
			tmp.width = second.width;
			tmp.height = second.y - first.y;
			if (tmp.width > 0 && tmp.height > 0) {
				rb = new Rectangle(tmp);
				count++;
			}

			tmp.y = second.y + second.height;
			tmp.height = first.y + first.height - (second.y + second.height);
			if (tmp.width > 0 && tmp.height > 0) {
				rc = new Rectangle(tmp);
				count++;
			}

			tmp.x = second.x + second.width;
			tmp.y = first.y;
			tmp.width = first.x + first.width - (second.x + second.width);
			tmp.height = first.height;
			if (tmp.width > 0 && tmp.height > 0) {
				rd = new Rectangle(tmp);
				count++;
			}
		} else {
			if (second.x <= first.x && second.y <= first.y) {
				if (second.x + second.width > first.x + first.width) {
					tmp.x = first.x;
					tmp.y = second.y + second.height;
					tmp.width = first.width;
					tmp.height = first.y + first.height - (second.y + second.height);
					if (tmp.width > 0 && tmp.height > 0) {
						ra = tmp;
						count++;
					}
				} else if (second.y + second.height > first.y + first.height) {
					tmp.x = second.x + second.width;
					tmp.y = first.y;
					tmp.width = first.x + first.width - (second.x + second.width);
					tmp.height = first.height;
					if (tmp.width > 0 && tmp.height > 0) {
						ra = tmp;
						count++;
					}
				} else {
					tmp.x = second.x + second.width;
					tmp.y = first.y;
					tmp.width = first.x + first.width - (second.x + second.width);
					tmp.height = second.y + second.height - first.y;
					if (tmp.width > 0 && tmp.height > 0) {
						ra = new Rectangle(tmp);
						count++;
					}

					tmp.x = first.x;
					tmp.y = second.y + second.height;
					tmp.width = first.width;
					tmp.height = first.y + first.height - (second.y + second.height);
					if (tmp.width > 0 && tmp.height > 0) {
						rb = new Rectangle(tmp);
						count++;
					}
				}
			} else if (second.x <= first.x && second.y + second.height >= first.y + first.height) {
				tmp.x = first.x;
				tmp.y = first.y;
				tmp.width = first.width;
				tmp.height = second.y - first.y;

				if (second.x + second.width > first.x + first.width) {
					if (tmp.width > 0 && tmp.height > 0) {
						ra = tmp;
						count++;
					}
				} else {
					if (tmp.width > 0 && tmp.height > 0) {
						ra = new Rectangle(tmp);
						count++;
					}

					tmp.x = second.x + second.width;
					tmp.y = second.y;
					tmp.width = first.x + first.width - (second.x + second.width);
					tmp.height = first.y + first.height - second.y;
					if (tmp.width > 0 && tmp.height > 0) {
						rb = new Rectangle(tmp);
						count++;
					}
				}
			} else if (second.x <= first.x) {
				tmp.x = first.x;
				tmp.y = first.y;
				tmp.width = first.width;
				tmp.height = second.y - first.y;

				if (second.x + second.width >= first.x + first.width) {
					if (tmp.width > 0 && tmp.height > 0) {
						ra = new Rectangle(tmp);
						count++;
					}

					tmp.y = second.y + second.height;
					tmp.height = first.y + first.height - (second.y + second.height);
					if (tmp.width > 0 && tmp.height > 0) {
						rb = new Rectangle(tmp);
						count++;
					}
				} else {
					if (tmp.width > 0 && tmp.height > 0) {
						ra = new Rectangle(tmp);
						count++;
					}

					tmp.x = second.x + second.width;
					tmp.y = second.y;
					tmp.width = first.x + first.width - (second.x + second.width);
					tmp.height = second.height;
					if (tmp.width > 0 && tmp.height > 0) {
						rb = new Rectangle(tmp);
						count++;
					}

					tmp.x = first.x;
					tmp.y = second.y + second.height;
					tmp.width = first.width;
					tmp.height = first.y + first.height - (second.y + second.height);
					if (tmp.width > 0 && tmp.height > 0) {
						rc = new Rectangle(tmp);
						count++;
					}
				}
			} else if (second.x <= first.x + first.width && second.x + second.width > first.x + first.width) {
				tmp.x = first.x;
				tmp.y = first.y;

				if (second.y <= first.y && second.y + second.height > first.y + first.height) {
					tmp.width = second.x - first.x;
					tmp.height = first.height;
					if (tmp.width > 0 && tmp.height > 0) {
						ra = tmp;
						count++;
					}
				} else if (second.y <= first.y) {
					tmp.width = second.x - first.x;
					tmp.height = second.y + second.height - first.y;
					if (tmp.width > 0 && tmp.height > 0) {
						ra = new Rectangle(tmp);
						count++;
					}

					tmp.y = second.y + second.height;
					tmp.width = first.width;
					tmp.height = first.y + first.height - (second.y + second.height);
					if (tmp.width > 0 && tmp.height > 0) {
						rb = new Rectangle(tmp);
						count++;
					}
				} else if (second.y + second.height > first.y + first.height) {
					tmp.width = first.width;
					tmp.height = second.y - first.y;
					if (tmp.width > 0 && tmp.height > 0) {
						ra = new Rectangle(tmp);
						count++;
					}

					tmp.y = second.y;
					tmp.width = second.x - first.x;
					tmp.height = first.y + first.height - second.y;
					if (tmp.width > 0 && tmp.height > 0) {
						rb = new Rectangle(tmp);
						count++;
					}
				} else {
					tmp.width = first.width;
					tmp.height = second.y - first.y;
					if (tmp.width > 0 && tmp.height > 0) {
						ra = new Rectangle(tmp);
						count++;
					}

					tmp.y = second.y;
					tmp.width = second.x - first.x;
					tmp.height = second.height;
					if (tmp.width > 0 && tmp.height > 0) {
						rb = new Rectangle(tmp);
						count++;
					}

					tmp.y = second.y + second.height;
					tmp.width = first.width;
					tmp.height = first.y + first.height - (second.y + second.height);
					if (tmp.width > 0 && tmp.height > 0) {
						rc = new Rectangle(tmp);
						count++;
					}
				}
			} else if (second.x >= first.x && second.x + second.width <= first.x + first.width) {
				tmp.x = first.x;
				tmp.y = first.y;
				tmp.width = second.x - first.x;
				tmp.height = first.height;
				if (second.y <= first.y && second.y + second.height > first.y + first.height) {
					if (tmp.width > 0 && tmp.height > 0) {
						ra = new Rectangle(tmp);
						count++;
					}

					tmp.x = second.x + second.width;
					tmp.width = first.x + first.width - (second.x + second.width);
					if (tmp.width > 0 && tmp.height > 0) {
						rb = new Rectangle(tmp);
						count++;
					}
				} else if (second.y <= first.y) {
					if (tmp.width > 0 && tmp.height > 0) {
						ra = new Rectangle(tmp);
						count++;
					}

					tmp.x = second.x;
					tmp.y = second.y + second.height;
					tmp.width = second.width;
					tmp.height = first.y + first.height - (second.y + second.height);
					if (tmp.width > 0 && tmp.height > 0) {
						rb = new Rectangle(tmp);
						count++;
					}

					tmp.x = second.x + second.width;
					tmp.y = first.y;
					tmp.width = first.x + first.width - (second.x + second.width);
					tmp.height = first.height;
					if (tmp.width > 0 && tmp.height > 0) {
						rc = new Rectangle(tmp);
						count++;
					}
				} else {
					if (tmp.width > 0 && tmp.height > 0) {
						ra = new Rectangle(tmp);
						count++;
					}

					tmp.x = second.x;
					tmp.width = second.width;
					tmp.height = second.y - first.y;
					if (tmp.width > 0 && tmp.height > 0) {
						rb = new Rectangle(tmp);
						count++;
					}

					tmp.x = second.x + second.width;
					tmp.y = first.y;
					tmp.width = first.x + first.width - (second.x + second.width);
					tmp.height = first.height;
					if (tmp.width > 0 && tmp.height > 0) {
						rc = new Rectangle(tmp);
						count++;
					}
				}
			}
		}

		Rectangle result[] = new Rectangle[count];

		count = 0;
		if (ra != null) {
			result[count++] = ra;
		}
		if (rb != null) {
			result[count++] = rb;
		}
		if (rc != null) {
			result[count++] = rc;
		}
		if (rd != null) {
			result[count] = rd;
		}
		return result;
	}

	/**
	 * @param first The first rectangle.
	 * @param second The second rectangle.
	 * @return <code>true</code> if the first rectangle contains the second.
	 */
	public static boolean contains(Rectangle first, Rectangle second) {
		return second.x >= first.x && second.x + second.width <= first.x + first.width && second.y >= first.y && second.y + second.height <= first.y + first.height;
	}

	/** @return A comparator that will sort rectangles, first by y/height, then by x/width. */
	public static Comparator<Rectangle> getSorter() {
		if (SORTER == null) {
			SORTER = new Comparator<Rectangle>() {
				public int compare(Rectangle r1, Rectangle r2) {
					if (r1.y < r2.y) {
						return -1;
					}
					if (r1.y > r2.y) {
						return 1;
					}
					if (r1.height < r2.height) {
						return 1;
					}
					if (r1.height > r2.height) {
						return -1;
					}
					if (r1.x < r2.x) {
						return -1;
					}
					if (r1.x > r2.x) {
						return 1;
					}
					if (r1.width < r2.width) {
						return 1;
					}
					if (r1.width > r2.width) {
						return -1;
					}
					return 0;
				}
			};
		}
		return SORTER;
	}

	/**
	 * @param r The original rectangle.
	 * @param x The horizontal inset.
	 * @param y The vertical inset.
	 * @return A new rectangle created by inseting the passed in rectangle.
	 */
	public static Rectangle insetRectangle(Rectangle r, int x, int y) {
		return new Rectangle(r.x + x, r.y + y, r.width - x * 2, r.height - y * 2);
	}

	/**
	 * Intersects two rectangles, producing a third. Unlike the
	 * {@link Rectangle#intersection(Rectangle)} method, the resulting rectangle's width & height
	 * will not be set to less than zero when there is no overlap.
	 * 
	 * @param first The first rectangle.
	 * @param second The second rectangle.
	 * @return The intersection of the two rectangles.
	 */
	public static Rectangle intersection(Rectangle first, Rectangle second) {
		if (first.width < 1 || first.height < 1 || second.width < 1 || second.height < 1) {
			return new Rectangle();
		}

		int x = Math.max(first.x, second.x);
		int y = Math.max(first.y, second.y);
		int w = Math.min(first.x + first.width, second.x + second.width) - x;
		int h = Math.min(first.y + first.height, second.y + second.height) - y;

		if (w < 0 || h < 0) {
			return new Rectangle();
		}

		return new Rectangle(x, y, w, h);
	}

	/**
	 * Intersects two rectangles, producing a third. Unlike the
	 * {@link Rectangle#intersection(Rectangle)} method, the resulting rectangle's width & height
	 * will not be set to less than zero when there is no overlap.
	 * 
	 * @param first The first rectangle.
	 * @param second The second rectangle.
	 * @return The intersection of the two rectangles.
	 */
	public static Rectangle2D intersection(Rectangle2D first, Rectangle2D second) {
		double x1 = first.getX();
		double y1 = first.getY();
		double w1 = first.getWidth();
		double h1 = first.getHeight();
		double x2 = second.getX();
		double y2 = second.getY();
		double w2 = second.getWidth();
		double h2 = second.getHeight();

		if (!(w1 > 0.0 && h1 > 0.0 && w2 > 0.0 && h2 > 0.0)) {
			return new Rectangle.Double();
		}

		double x = Math.max(x1, x2);
		double y = Math.max(y1, y2);
		double w = Math.min(x1 + w1, x2 + w2) - x;
		double h = Math.min(y1 + h1, y2 + h2) - y;

		if (w < 0.0 || h < 0.0) {
			return new Rectangle();
		}

		return new Rectangle2D.Double(x, y, w, h);
	}

	/**
	 * @param array The array to intersect with <code>bounds</code>.
	 * @param bounds The intersecting rectangle.
	 * @return An array of rectangles from the passed in <code>array</code> that intersect with
	 *         <code>bounds</code>. The rectangles will be added to the new list after creating
	 *         the intersection between them and <code>bounds</code>.
	 */
	public static Rectangle[] intersection(Rectangle[] array, Rectangle bounds) {
		ArrayList<Rectangle> list = new ArrayList<Rectangle>(array.length);

		for (Rectangle element : array) {
			if (bounds.intersects(element)) {
				list.add(intersection(element, bounds));
			}
		}
		return list.toArray(new Rectangle[0]);
	}

	/**
	 * @param array The array to intersect with <code>bounds</code>.
	 * @param bounds The intersecting rectangle.
	 * @return <code>true</code> if <code>bounds</code> intersects any of the rectangles in
	 *         <code>array</code>.
	 */
	public static boolean intersects(Rectangle[] array, Rectangle bounds) {
		for (Rectangle element : array) {
			if (bounds.intersects(element)) {
				return true;
			}
		}
		return false;
	}

	/**
	 * @param array The array to intersect with <code>bounds</code>.
	 * @param bounds The intersecting rectangle.
	 * @return <code>true</code> if <code>bounds</code> intersects any of the rectangles in
	 *         <code>array</code>.
	 */
	public static boolean intersects(Rectangle2D[] array, Rectangle2D bounds) {
		for (Rectangle2D element : array) {
			if (bounds.intersects(element)) {
				return true;
			}
		}
		return false;
	}

	/**
	 * Merges the specified rectangles into an array of rectangles.
	 * 
	 * @param array The array to merge into.
	 * @param bounds The rectangles to merge.
	 * @param allowWastage Pass in <code>true</code> to allow rectangles to be merged together
	 *            even when it might result in some extra space being added to the overall area.
	 * @return The resulting rectangles.
	 */
	public static Rectangle[] mergeInto(Rectangle[] array, Rectangle[] bounds, boolean allowWastage) {
		for (Rectangle element : bounds) {
			array = mergeInto(array, element, allowWastage);
		}
		return array;
	}

	/**
	 * Merges the specified rectangle into an array of rectangles.
	 * 
	 * @param array The array to merge into.
	 * @param bounds The rectangle to merge.
	 * @param allowWastage Pass in <code>true</code> to allow rectangles to be merged together
	 *            even when it might result in some extra space being added to the overall area.
	 * @return The resulting rectangles.
	 */
	public static Rectangle[] mergeInto(Rectangle[] array, Rectangle bounds, boolean allowWastage) {
		ArrayList<Rectangle> list = new ArrayList<Rectangle>(array.length * 2);
		boolean[] used = new boolean[array.length];
		boolean[] touched = new boolean[array.length];
		int top = Integer.MAX_VALUE / 4;
		int bottom = Integer.MIN_VALUE / 4;
		Rectangle clip = new Rectangle();
		int i;
		int tmp;

		bounds = new Rectangle(bounds); // For safety...

		Arrays.sort(array, TKRectUtils.getSorter());

		for (i = 0; i < array.length; i++) {
			used[i] = true;
			touched[i] = false;

			if (allowWastage) {
				long totalArea = (long) array[i].width * (long) array[i].height + (long) bounds.width * (long) bounds.height;
				int maxWastageAllowed = 75 - (int) (totalArea / 363L);

				if (maxWastageAllowed > 75) {
					maxWastageAllowed = 75;
				} else if (maxWastageAllowed < 20) {
					maxWastageAllowed = 20;
				}

				if (areaWasted(array[i], bounds) < maxWastageAllowed) {
					bounds = union(array[i], bounds);
					used[i] = false;
				}
			}
			if (used[i] && bounds.intersects(array[i])) {
				if (bounds.contains(array[i])) {
					used[i] = false;
				} else if (array[i].contains(bounds)) {
					return array;
				} else {
					tmp = array[i].y;
					touched[i] = true;
					if (tmp < top) {
						top = tmp;
					}
					tmp += array[i].height;
					if (tmp > bottom) {
						bottom = tmp;
					}
				}
			}
			if (used[i] && !touched[i]) {
				list.add(array[i]);
			}
		}

		if (top != Integer.MAX_VALUE / 4) {
			if (top < bounds.y) {
				clip.x = Integer.MIN_VALUE / 4;
				clip.y = top;
				clip.width = Integer.MAX_VALUE / 2;
				clip.height = bounds.y - top;
				for (i = 0; i < array.length; i++) {
					if (touched[i] && clip.intersects(array[i])) {
						list.add(intersection(clip, array[i]));
						array[i].height -= bounds.y - array[i].y;
						if (array[i].height <= 0) {
							touched[i] = false;
						} else {
							array[i].y = bounds.y;
						}
					}
				}
			}

			tmp = bounds.y + bounds.height;
			if (bottom >= tmp) {
				clip.x = Integer.MIN_VALUE / 4;
				clip.y = tmp;
				clip.width = Integer.MAX_VALUE / 2;
				clip.height = bottom - tmp;
				for (i = 0; i < array.length; i++) {
					if (touched[i] && clip.intersects(array[i])) {
						list.add(intersection(clip, array[i]));
						array[i].height = tmp - array[i].y;
						if (array[i].height <= 0) {
							touched[i] = false;
						}
					}
				}
			}

			for (i = 0; i < array.length; i++) {
				if (touched[i]) {
					Rectangle overlap;
					Rectangle[] remains;

					clip.x = Integer.MIN_VALUE / 4;
					clip.y = Math.max(bounds.y, array[i].y);
					clip.width = Integer.MAX_VALUE / 2;
					clip.height = Math.min(bounds.y + bounds.height, array[i].y + array[i].height) - clip.y;
					overlap = intersection(clip, bounds);
					for (int j = i; j < array.length; j++) {
						if (touched[j] && clip.intersects(array[j])) {
							overlap.add(intersection(clip, array[j]));
							tmp = clip.y + clip.height;
							array[j].height -= tmp - array[j].y;
							if (array[j].height <= 0) {
								touched[j] = false;
							} else {
								array[j].y = tmp;
							}
						}
					}
					remains = computeDifference(bounds, overlap);
					if (remains.length == 1) {
						bounds = remains[0];
					} else if (remains.length > 1) {
						if (remains[0].y < remains[1].y) {
							bounds = remains[1];
							list.add(remains[0]);
						} else {
							bounds = remains[0];
							list.add(remains[1]);
						}
					} else {
						bounds.height = 0;
					}
					list.add(overlap);
				}
			}
		}

		if (bounds.height > 0) {
			list.add(bounds);
		}

		return list.toArray(new Rectangle[0]);
	}

	/**
	 * @param r The original rectangle.
	 * @param x The horizontal offset.
	 * @param y The vertical offset.
	 * @return A new rectangle created by offseting the passed in rectangle.
	 */
	public static Rectangle2D.Double offsetRectangle(Rectangle2D r, double x, double y) {
		return new Rectangle2D.Double(r.getX() + x, r.getY() + y, r.getWidth(), r.getHeight());
	}

	/**
	 * Offsets each of the rectangles in the array.
	 * 
	 * @param array The array of rectangles to operate on.
	 * @param x The horizontal offset.
	 * @param y The vertical offset.
	 */
	public static void offsetRectangles(Rectangle[] array, int x, int y) {
		for (Rectangle element : array) {
			element.x += x;
			element.y += y;
		}
	}

	/**
	 * Removes the specified rectangle from an array of rectangles.
	 * 
	 * @param array The array of rectangles to operate on.
	 * @param bounds The rectangle to subtract.
	 * @return A new array of rectangles that don't overlap <code>bounds</code>.
	 */
	public static Rectangle[] removeFrom(Rectangle[] array, Rectangle bounds) {
		ArrayList<Rectangle> list = new ArrayList<Rectangle>(array.length + 10);

		for (Rectangle element : array) {
			Rectangle[] result = computeDifference(element, bounds);

			for (Rectangle element0 : result) {
				list.add(element0);
			}
		}
		return list.toArray(new Rectangle[0]);
	}

	/**
	 * Removes the specified array of rectangles from an array of rectangles.
	 * 
	 * @param array The array of rectangles to operate on.
	 * @param rmArray The array of rectangles to subtract.
	 * @return A new array of rectangles that don't overlap <code>rmArray</code>.
	 */
	public static Rectangle[] removeFrom(Rectangle[] array, Rectangle[] rmArray) {
		ArrayList<Rectangle> list = new ArrayList<Rectangle>(array.length + 10);
		int count = array.length;
		int i;

		for (i = 0; i < count; i++) {
			list.add(array[i]);
		}

		for (i = 0; i < rmArray.length; i++) {
			for (int j = count - 1; j >= 0; j--) {
				Rectangle[] result = computeDifference(list.get(j), rmArray[i]);

				list.remove(j);
				count--;
				for (Rectangle element : result) {
					list.add(element);
					count++;
				}
			}
		}

		return list.toArray(new Rectangle[0]);
	}

	/**
	 * Unions two rectangles, producing a third. Unlike the {@link Rectangle#union(Rectangle)}
	 * method, an empty rectangle will not cause the rectangle's boundary to extend to the 0,0
	 * point.
	 * 
	 * @param first The first rectangle.
	 * @param second The second rectangle.
	 * @return The resulting rectangle.
	 */
	public static Rectangle union(Rectangle first, Rectangle second) {
		boolean firstEmpty = first.width < 1 || first.height < 1;
		boolean secondEmpty = second.width < 1 || second.height < 1;

		if (firstEmpty && secondEmpty) {
			return new Rectangle();
		}
		if (firstEmpty) {
			return new Rectangle(second);
		}
		if (secondEmpty) {
			return new Rectangle(first);
		}
		return first.union(second);
	}

	/**
	 * Unions two rectangles, producing a third. Unlike the {@link Rectangle#union(Rectangle)}
	 * method, an empty rectangle will not cause the rectangle's boundary to extend to the 0,0
	 * point.
	 * 
	 * @param first The first rectangle.
	 * @param second The second rectangle.
	 * @return The resulting rectangle.
	 */
	public static Rectangle2D union(Rectangle2D first, Rectangle2D second) {
		double w1 = first.getWidth();
		double w2 = second.getWidth();
		double h1 = first.getHeight();
		double h2 = second.getHeight();
		boolean firstEmpty = !(w1 > 0.0 && h1 > 0.0);
		boolean secondEmpty = !(w2 > 0.0 && h2 > 0.0);

		if (firstEmpty && secondEmpty) {
			return new Rectangle2D.Double();
		}
		if (firstEmpty) {
			return new Rectangle2D.Double(second.getX(), second.getY(), w2, h2);
		}
		if (secondEmpty) {
			return new Rectangle2D.Double(first.getX(), first.getY(), w1, h1);
		}
		return first.createUnion(second);
	}

	/**
	 * @param array An array of rectangles.
	 * @return A rectangle representing the union of all rectangles in <code>array</code>.
	 */
	public static Rectangle union(Rectangle[] array) {
		Rectangle result;

		if (array.length == 0) {
			return new Rectangle();
		}

		result = new Rectangle(array[0]);

		for (int i = 1; i < array.length; i++) {
			result.add(array[i]);
		}
		return result;
	}

	/**
	 * @param array An array of rectangles.
	 * @param bounds A rectangle to union with.
	 * @return A rectangle representing the union of all rectangles in <code>array</code> which
	 *         intersect with <code>bounds</code>, clipped to the area represented by
	 *         <code>bounds</code>.
	 */
	public static Rectangle unionIntersection(Rectangle[] array, Rectangle bounds) {
		Rectangle result = null;

		for (Rectangle element : array) {
			if (bounds.intersects(element)) {
				Rectangle clip = intersection(bounds, element);

				if (result == null) {
					result = clip;
				} else {
					result.add(clip);
				}
			}
		}
		return result == null ? new Rectangle() : result;
	}
}
