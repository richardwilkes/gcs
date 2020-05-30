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

package com.trollworks.gcs.utility;

import java.io.BufferedReader;
import java.io.IOException;
import java.io.InputStreamReader;
import java.net.HttpURLConnection;
import java.net.URL;
import java.net.URLConnection;
import java.net.URLDecoder;
import java.nio.charset.StandardCharsets;
import java.util.HashMap;
import java.util.Map;

/** Utilities for working with URLs. */
public class UrlUtils {
    /**
     * @param uri The URI to setup a connection for.
     * @return A {@link URLConnection} configured with a 10 second timeout for connecting and
     *         reading data.
     */
    public static final URLConnection setupConnection(String uri) throws IOException {
        return setupConnection(new URL(uri));
    }

    /**
     * @param url The URL to setup a connection for.
     * @return A {@link URLConnection} configured with a 10 second timeout for connecting and
     *         reading data.
     */
    public static final URLConnection setupConnection(URL url) throws IOException {
        Map<String, Integer> visited = new HashMap<>();
        HttpURLConnection    conn;
        while (true) {
            if (visited.compute(url.toExternalForm(), (key, count) -> Integer.valueOf(count == null ? 1 : count.intValue() + 1)).intValue() > 3) {
                throw new IOException("Stuck in redirect loop");
            }
            conn = (HttpURLConnection) url.openConnection();
            conn.setConnectTimeout(10000);
            conn.setReadTimeout(10000);
            conn.setInstanceFollowRedirects(false);   // Make the logic below easier to detect redirections
            switch (conn.getResponseCode()) {
            case HttpURLConnection.HTTP_MOVED_PERM:
            case HttpURLConnection.HTTP_MOVED_TEMP:
            case 307:
                String location = URLDecoder.decode(conn.getHeaderField("Location"), StandardCharsets.UTF_8);
                url = new URL(url, location);  // Deal with relative URLs
                continue;
            }
            break;
        }
        return conn;
    }

    /**
     * @param uri The URI to retrieve.
     * @return The contents of what the URI points to.
     */
    public static final String get(String uri) throws IOException {
        return get(new URL(uri));
    }

    /**
     * @param url The URL to retrieve.
     * @return The contents of what the URL points to.
     */
    public static final String get(URL url) throws IOException {
        StringBuilder buffer = new StringBuilder();
        try (BufferedReader in = new BufferedReader(new InputStreamReader(setupConnection(url).getInputStream(), StandardCharsets.UTF_8))) {
            String line;
            while ((line = in.readLine()) != null) {
                if (buffer.length() > 0) {
                    buffer.append('\n');
                }
                buffer.append(line);
            }
        }
        return buffer.toString();
    }
}
