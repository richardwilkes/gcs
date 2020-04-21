/*
 * Copyright (c) 1998-2020 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public License, version 2.0.
 * If a copy of the MPL was not distributed with this file, You can obtain one at
 * http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as defined by the
 * Mozilla Public License, version 2.0.
 */

package com.trollworks.gcs.services;

import java.io.DataOutputStream;
import java.io.IOException;
import java.io.InputStream;
import java.io.OutputStreamWriter;
import java.io.UnsupportedEncodingException;
import java.net.MalformedURLException;
import java.net.URL;
import java.net.URLConnection;
import java.net.URLEncoder;
import java.nio.charset.StandardCharsets;
import java.util.HashMap;
import java.util.Map;
import java.util.Scanner;

/** A class used for interacting with web services */
public class WebServiceClient {
    private String              mBaseUrl;
    private Map<String, String> mQueryParameters = new HashMap<>();
    private String              mBody;
    private byte[]              mByteBody;
    private String              mPath;
    private HttpMethodType      mMethod          = HttpMethodType.GET;

    /** Creates a WebServiceClient with a base url: ("http://example.com/") */
    public WebServiceClient(String baseUrl) {
        mBaseUrl = baseUrl;
    }

    /** Adds a querystring parameter: (?key=value) */
    public boolean addQueryParameter(String key, String value) {
        try {
            mQueryParameters.put(key, URLEncoder.encode(value, StandardCharsets.UTF_8.toString()));
        } catch (UnsupportedEncodingException exception) {
            return false;
        }
        return true;
    }

    /** Sets the body of the request */
    public void setBody(String body) {
        mBody = body;
    }

    /** Sets the path of the request: (/api/testmethod) */
    public void setPath(String path) {
        mPath = path;
    }

    /** Sets the http method type of the request (GET, POST, PUT, DELETE) */
    public void setHttpMethodType(HttpMethodType method) {
        mMethod = method;
    }

    /**
     * Sends the http request and returns string.
     */
    public String sendRequest(HttpMethodType method, String path, Map<String, String> queryParameters, String body) throws IOException, NotImplementedException {
        mBody = body;
        return sendRequest(method, path, queryParameters);
    }

    /**
     * Sends the http request and returns a string.
     */
    public String sendRequest(HttpMethodType method, String path, Map<String, String> queryParameters) throws IOException, NotImplementedException {
        if (queryParameters != null) {
            mQueryParameters = queryParameters;
        }
        return sendRequest(method, path);
    }

    /**
     * Sends the http request and returns a string.
     */
    public String sendRequest(HttpMethodType method, String path) throws IOException, NotImplementedException {
        mMethod = method;
        mPath = path;
        return sendRequest();
    }

    public String sendRequest(HttpMethodType method, String path, byte[] bytes) throws IOException, NotImplementedException {
        mBody = null;
        mByteBody = bytes;
        return sendRequest(method, path);
    }

    /**
     * Sends the http request and returns a string.
     */
    public String sendRequest() throws IOException, NotImplementedException {
        URLConnection connection = buildUrl().openConnection();
        connection.setRequestProperty("Accept-Charset", StandardCharsets.UTF_8.toString());
        if (mMethod == HttpMethodType.GET) {
            try (InputStream stream = connection.getInputStream()) {
                return convertStreamToString(stream);
            }
        }
        if (mMethod == HttpMethodType.POST) {
            connection.setDoOutput(true); // Triggers POST.
            connection.setRequestProperty("Content-Type", "application/json");
            if (mBody != null && !mBody.isEmpty()) {
                try (OutputStreamWriter writer = new OutputStreamWriter(connection.getOutputStream(), StandardCharsets.UTF_8)) {
                    writer.write(mBody);
                }
                try (InputStream stream = connection.getInputStream()) {
                    return convertStreamToString(stream);
                }
            }
            if (mByteBody != null) {
                try (DataOutputStream writer = new DataOutputStream(connection.getOutputStream())) {
                    writer.write(mByteBody);
                }
                try (InputStream stream = connection.getInputStream()) {
                    return convertStreamToString(stream);
                }
            }
        }
        throw new NotImplementedException("Method type not currently supported");
    }

    static String convertStreamToString(java.io.InputStream is) {
        try (Scanner s = new Scanner(is, StandardCharsets.UTF_8)) {
            s.useDelimiter("\\A");
            return s.hasNext() ? s.next() : "";
        }
    }

    private URL buildUrl() throws MalformedURLException {
        // sanitize strings
        mBaseUrl = ditchLeadingTrailingSlashes(mBaseUrl);
        mPath = ditchLeadingTrailingSlashes(mPath);
        // build the url
        StringBuilder url = new StringBuilder(mBaseUrl);
        url.append("/");
        url.append(mPath);
        url.append("/");
        // add query string variables
        if (mQueryParameters.isEmpty()) {
            return new URL(url.toString());
        }
        url.append("?");
        for (Map.Entry<String, String> entry : mQueryParameters.entrySet()) {
            url.append(entry.getKey());
            url.append("=");
            url.append(entry.getValue());
        }
        return new URL(url.toString());
    }

    private static String ditchLeadingTrailingSlashes(String str) {
        if (str.length() < 2) {
            return str;
        }
        if (str.endsWith("/")) {
            str = str.substring(0, str.length() - 1);
        }
        if (str.startsWith("/")) {
            str = str.substring(1);
        }
        return str;
    }
}
