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

package com.trollworks.gcs.expression;

import java.util.ArrayList;
import java.util.Collection;
import java.util.EmptyStackException;

/**
 * A stack of objects.
 *
 * @param <T> The type of objects within the stack.
 */
public class Stack<T> extends ArrayList<T> {
    /**
     * Pushes an item onto the top of the stack.
     *
     * @param item The item to be pushed onto the stack.
     * @return The {@code item} argument.
     */
    public T push(T item) {
        add(item);
        return item;
    }

    /**
     * Pushes the items in the specified {@link Collection} onto the top of the stack.
     *
     * @param items The items to be pushed onto the stack.
     */
    public void pushAll(Collection<T> items) {
        addAll(items);
    }

    /**
     * Removes the object at the top of the stack.
     *
     * @return The object at the top of the stack.
     * @throws EmptyStackException if the stack is empty.
     */
    public T pop() {
        int length = size();
        if (length == 0) {
            throw new EmptyStackException();
        }
        return remove(length - 1);
    }

    /**
     * Looks at the object at the top of the stack without removing it from the stack.
     *
     * @return The object at the top of the stack.
     * @throws EmptyStackException if the stack is empty.
     */
    public T peek() {
        int length = size();
        if (length == 0) {
            throw new EmptyStackException();
        }
        return get(length - 1);
    }
}
