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

package com.trollworks.toolkit.window;

import com.trollworks.toolkit.utility.TKDebug;
import com.trollworks.toolkit.utility.TKGraphics;
import com.trollworks.toolkit.utility.TKRectUtils;
import com.trollworks.toolkit.utility.TKTiming;
import com.trollworks.toolkit.widget.TKPanel;

import java.awt.Color;
import java.awt.Component;
import java.awt.EventQueue;
import java.awt.Graphics;
import java.awt.Graphics2D;
import java.awt.GraphicsConfiguration;
import java.awt.Rectangle;
import java.awt.Window;
import java.awt.image.VolatileImage;
import java.util.ArrayList;
import java.util.HashSet;

/** Manages repainting native windows. */
public class TKRepaintManager implements Runnable {
	private TKBaseWindow		mBaseWindow;
	private Window				mWindow;
	private boolean				mPrinting;
	private boolean				mRepaintPending;
	private Rectangle[]			mRepaintClip;
	private VolatileImage		mOffscreen;
	private boolean				mIgnoreLazyUpdates;
	private boolean				mShowDrawingYellow;
	private ArrayList<Runnable>	mLazyUpdateTasks;
	private Rectangle[]			mLazyUpdateSavedClip;
	private HashSet<Component>	mInvalidComponents;
	private Runnable			mRunnableValidator;

	/**
	 * Creates a repaint manager for the specified base window.
	 * 
	 * @param window The window to manage repaint activity for.
	 */
	public TKRepaintManager(TKBaseWindow window) {
		mBaseWindow = window;
		mWindow = (Window) window;
		mRepaintClip = new Rectangle[0];
		mLazyUpdateSavedClip = new Rectangle[0];
		mLazyUpdateTasks = new ArrayList<Runnable>();
		mInvalidComponents = new HashSet<Component>();
		mRunnableValidator = new Runnable() {
			public void run() {
				validate();
			}
		};
	}

	/**
	 * Adds a component to the list of components that need validating.
	 * 
	 * @param comp The component to add.
	 */
	public void addInvalidComponent(Component comp) {
		synchronized (mInvalidComponents) {
			if (mInvalidComponents.isEmpty()) {
				mInvalidComponents.add(comp);
				// We need to add the runnable to both the lazy update task
				// as well as the event queue, just in case a repaint event
				// has already been placed in the event queue. If we just put
				// it on the event queue, we end up with flashing during
				// repaint, as we draw components that haven't been resized
				// yet. If we just put it in the lazy update queue, we don't
				// get correct painting during "live" manipulations, such as
				// column resizing in an outline.
				addLazyUpdateTask(mRunnableValidator);
				EventQueue.invokeLater(mRunnableValidator);
			} else {
				mInvalidComponents.add(comp);
			}
		}
	}

	/** Validates all components that have been marked as needing validation. */
	public void validate() {
		Component[] comps;

		synchronized (mInvalidComponents) {
			comps = mInvalidComponents.toArray(new Component[0]);
			mInvalidComponents.clear();
		}

		for (Component element : comps) {
			element.validate();
		}
	}

	/**
	 * Adds a task that will be executed just before the next repaint occurs.
	 * 
	 * @param task The task to run.
	 */
	public void addLazyUpdateTask(Runnable task) {
		synchronized (mLazyUpdateTasks) {
			mLazyUpdateTasks.add(task);
		}
	}

	/** Force a full repaint. */
	public void forceRepaint() {
		if (mOffscreen != null) {
			mOffscreen.flush();
			mOffscreen = null;
		}
		mWindow.repaint();
	}

	/** @return <code>true</code> if the window is being printed. */
	public boolean isPrinting() {
		return mPrinting;
	}

	/**
	 * Sets whether the window is being printed or not.
	 * 
	 * @param isPrinting Whether or not the window is being printed.
	 */
	public void setPrinting(boolean isPrinting) {
		mPrinting = isPrinting;
	}

	/**
	 * Paints the window.
	 * 
	 * @param graphics The graphics object to paint with.
	 */
	public void paint(Graphics graphics) {
		paint((Graphics2D) graphics, null);
	}

	/**
	 * Paints the window.
	 * 
	 * @param g2d The graphics object to paint with.
	 * @param clips The clipping rectangles.
	 */
	public void paint(Graphics2D g2d, Rectangle[] clips) {
		synchronized (mWindow.getTreeLock()) {
			Rectangle bounds = mBaseWindow.getLocalBounds(false);

			if (bounds.width > 0 && bounds.height > 0) {
				Rectangle clip;

				if (clips == null) {
					Rectangle realClip = g2d.getClipBounds();

					clips = new Rectangle[] { realClip == null ? bounds : realClip };
				}

				// For some unknown reason, on Mac OS X, the clips that get passed in
				// can sometimes be beyond the range of the window itself, causing
				// stretching of the graphics to occur. We trim them down here to
				// avoid this problem.
				for (int j = 0; j < clips.length; j++) {
					clips[j] = TKRectUtils.intersection(bounds, clips[j]);
				}

				clip = TKRectUtils.unionIntersection(clips, bounds);
				if (!clip.isEmpty()) {
					if (!isPrinting()) {
						synchronized (mLazyUpdateTasks) {
							if (!mIgnoreLazyUpdates && mLazyUpdateTasks.size() > 0) {
								if (mLazyUpdateSavedClip.length > 0) {
									Rectangle[] all = new Rectangle[mLazyUpdateSavedClip.length + clips.length];

									System.arraycopy(mLazyUpdateSavedClip, 0, all, 0, mLazyUpdateSavedClip.length);
									System.arraycopy(clips, 0, all, mLazyUpdateSavedClip.length, clips.length);
								} else {
									mLazyUpdateSavedClip = clips;
								}
								EventQueue.invokeLater(this);
								return;
							}
						}
					}

					while (true) {
						Throwable localThrowable = null;

						try {
							TKGraphics.configureGraphics(g2d);
							g2d.setClip(clip);
							if (!isPrinting() && TKGraphics.useDoubleBuffering()) {
								do {
									GraphicsConfiguration gc = mWindow.getGraphicsConfiguration();
									Graphics2D og2d;

									if (mOffscreen == null || bounds.width != mOffscreen.getWidth() || bounds.height != mOffscreen.getHeight() || mOffscreen.validate(gc) == VolatileImage.IMAGE_INCOMPATIBLE) {
										if (mOffscreen != null) {
											mOffscreen.flush();
										}
										mOffscreen = gc.createCompatibleVolatileImage(bounds.width, bounds.height);
									}

									og2d = TKGraphics.configureGraphics(mOffscreen.createGraphics());
									og2d.setClip(clip);

									try {
										paintChildren(og2d, clips);
									} catch (Exception paintException) {
										assert false : TKDebug.throwableToString(paintException);
									} finally {
										og2d.dispose();
									}

									for (Rectangle element : clips) {
										int x1 = element.x;
										int y1 = element.y;
										int x2 = x1 + element.width;
										int y2 = y1 + element.height;

										g2d.drawImage(mOffscreen, x1, y1, x2, y2, x1, y1, x2, y2, mWindow);
									}
								} while (mOffscreen.contentsLost());
							} else {
								paintChildren(g2d, clips);
							}
						} catch (Throwable throwable) {
							localThrowable = throwable;
						}
						if (localThrowable instanceof OutOfMemoryError) {
							if (!TKWindow.releaseAnEditFromEachWindow()) {
								// No undo buffers could be freed, so we're giving up.
								assert false : TKDebug.throwableToString(localThrowable);
								break;
							}
						} else {
							if (localThrowable != null) {
								assert false : TKDebug.throwableToString(localThrowable);
							}
							break;
						}
					}
				}
			}
		}
	}

	private void paintChildren(Graphics2D g2d, Rectangle[] clips) {
		for (int i = mWindow.getComponentCount() - 1; i >= 0; i--) {
			TKPanel comp = (TKPanel) mWindow.getComponent(i);

			if (comp.isVisible()) {
				Rectangle bounds = comp.getBounds();

				if (TKRectUtils.intersects(clips, bounds)) {
					Rectangle compClip = TKRectUtils.unionIntersection(clips, bounds);
					Graphics2D cg2d = TKGraphics.configureGraphics((Graphics2D) g2d.create(compClip.x, compClip.y, compClip.width, compClip.height));

					cg2d.translate(bounds.x - compClip.x, bounds.y - compClip.y);
					compClip.x -= bounds.x;
					compClip.y -= bounds.y;
					cg2d.setClip(compClip);

					TKRectUtils.offsetRectangles(clips, -bounds.x, -bounds.y);

					try {
						comp.paint(cg2d, clips);
					} catch (Exception exception) {
						assert false : TKDebug.throwableToString(exception);
					} finally {
						cg2d.dispose();
					}

					TKRectUtils.offsetRectangles(clips, bounds.x, bounds.y);
				}
			}
		}
	}

	/**
	 * Immediately paints the specified region in this repaint manager's base window and all of its
	 * descendants that overlap the region without waiting for the normal event loop repaint
	 * management.
	 * 
	 * @param bounds The area within the base window to repaint.
	 */
	public void paintImmediately(Rectangle bounds) {
		paintImmediately(new Rectangle[] { bounds });
	}

	/**
	 * Immediately paints the specified region in this repaint manager's base window and all of its
	 * descendants that overlap the region without waiting for the normal event loop repaint
	 * management.
	 * 
	 * @param bounds The areas within the base window to repaint.
	 */
	public void paintImmediately(Rectangle[] bounds) {
		if (mWindow.isShowing()) {
			Graphics2D g2d = (Graphics2D) mWindow.getGraphics();

			try {
				stopRepaint(bounds);
				prePaint(bounds, g2d);
				synchronized (mLazyUpdateTasks) {
					mIgnoreLazyUpdates = true;
				}
				paint(g2d, bounds);
				synchronized (mLazyUpdateTasks) {
					mIgnoreLazyUpdates = false;
				}
				postPaint(bounds, g2d);
			} catch (Exception exception) {
				assert false : TKDebug.throwableToString(exception);
			} finally {
				g2d.dispose();
			}
		}
	}

	/**
	 * Repaints the base window.
	 * 
	 * @param x The x coordinate of the area to repaint.
	 * @param y The y coordinate of the area to repaint.
	 * @param width The width of the area to repaint.
	 * @param height The height of the area to repaint.
	 */
	public void repaint(int x, int y, int width, int height) {
		repaint(new Rectangle(x, y, width, height));
	}

	/**
	 * Repaints the base window.
	 * 
	 * @param bounds The area to repaint.
	 */
	public void repaint(Rectangle bounds) {
		if (bounds != null) {
			bounds = TKRectUtils.intersection(mBaseWindow.getLocalBounds(false), bounds);
			if (!bounds.isEmpty()) {
				synchronized (this) {
					if (!((Window) mBaseWindow).isVisible()) {
						return;
					}
					if (mRepaintPending) {
						mRepaintClip = TKRectUtils.mergeInto(mRepaintClip, bounds, false);
					} else {
						mRepaintClip = new Rectangle[] { bounds };
						mRepaintPending = true;
						EventQueue.invokeLater(new Runnable() {
							public void run() {
								updateImmediatelyBypassingLazyUpdates();
							}
						});
					}
				}
			}
		}
	}

	/**
	 * Removes the specified region from any pending repaint.
	 * 
	 * @param bounds The area to remove from pending repaints.
	 */
	public void stopRepaint(Rectangle bounds) {
		synchronized (this) {
			if (mRepaintPending) {
				mRepaintClip = TKRectUtils.removeFrom(mRepaintClip, bounds);
			}
		}
	}

	/**
	 * Removes the specified region from any pending repaint.
	 * 
	 * @param bounds The areas to remove from pending repaints.
	 */
	public void stopRepaint(Rectangle[] bounds) {
		synchronized (this) {
			if (mRepaintPending) {
				mRepaintClip = TKRectUtils.removeFrom(mRepaintClip, bounds);
			}
		}
	}

	/** If there are any pending repaints, deal with them now. */
	public void updateImmediately() {
		synchronized (mLazyUpdateTasks) {
			mIgnoreLazyUpdates = true;
		}
		updateImmediatelyBypassingLazyUpdates();
		synchronized (mLazyUpdateTasks) {
			mIgnoreLazyUpdates = false;
		}
	}

	/**
	 * If there are any pending repaints, deal with them now, ignoring any lazy updates that may be
	 * present.
	 */
	public void updateImmediatelyBypassingLazyUpdates() {
		Rectangle[] bounds;

		synchronized (this) {
			bounds = new Rectangle[mRepaintClip.length];
			if (mRepaintClip.length > 0) {
				System.arraycopy(mRepaintClip, 0, bounds, 0, mRepaintClip.length);
			}
			mRepaintClip = new Rectangle[0];
			mRepaintPending = false;
		}

		if (bounds.length > 0 && mWindow.isShowing()) {
			Graphics2D g2d = (Graphics2D) mWindow.getGraphics();

			if (g2d != null) {
				try {
					prePaint(bounds, g2d);
					paint(g2d, bounds);
					postPaint(bounds, g2d);
				} catch (Exception exception) {
					assert false : TKDebug.throwableToString(exception);
				}
				g2d.dispose();
			}
		}
	}

	private void prePaint(Rectangle[] bounds, Graphics2D g2d) {
		if (TKGraphics.isShowDrawingOn()) {
			TKTiming timing = new TKTiming();

			g2d.setColor(mShowDrawingYellow ? Color.yellow : Color.cyan);
			mShowDrawingYellow = !mShowDrawingYellow;
			for (Rectangle element : bounds) {
				g2d.fill(element);
			}
			timing.delayUntilThenReset(200);
		}
	}

	private void postPaint(@SuppressWarnings("unused") Rectangle[] bounds, @SuppressWarnings("unused") Graphics2D g2d) {
		// Not currently used
	}

	public void run() {
		ArrayList<Runnable> list = null;
		Rectangle[] bounds = null;
		int i;

		synchronized (mLazyUpdateTasks) {
			if (mLazyUpdateSavedClip.length == 0 && mLazyUpdateTasks.size() > 0) {
				// Post ourselves again, since the update hasn't come in yet.
				EventQueue.invokeLater(this);
				return;
			}

			list = new ArrayList<Runnable>(mLazyUpdateTasks);
			bounds = new Rectangle[mLazyUpdateSavedClip.length];
			if (mLazyUpdateSavedClip.length > 0) {
				System.arraycopy(mLazyUpdateSavedClip, 0, bounds, 0, mLazyUpdateSavedClip.length);
			}
			mLazyUpdateSavedClip = new Rectangle[0];
			mLazyUpdateTasks.clear();
		}

		for (i = 0; i < bounds.length; i++) {
			repaint(bounds[i]);
		}

		for (Runnable runner : list) {
			try {
				runner.run();
			} catch (Exception exception) {
				assert false : TKDebug.throwableToString(exception);
			}
		}
	}
}
