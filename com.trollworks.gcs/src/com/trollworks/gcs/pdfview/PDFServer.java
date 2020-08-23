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

package com.trollworks.gcs.pdfview;

import com.sun.net.httpserver.HttpExchange;
import com.sun.net.httpserver.HttpServer;
import com.trollworks.gcs.utility.Log;
import com.trollworks.gcs.utility.PathUtils;
import com.trollworks.gcs.utility.text.Text;

import java.awt.Desktop;
import java.io.ByteArrayInputStream;
import java.io.ByteArrayOutputStream;
import java.io.IOException;
import java.io.InputStream;
import java.io.OutputStream;
import java.net.InetSocketAddress;
import java.net.URI;
import java.net.URISyntaxException;
import java.net.URLDecoder;
import java.nio.charset.StandardCharsets;
import java.nio.file.Files;
import java.nio.file.Path;
import java.time.Instant;
import java.time.ZoneId;
import java.time.format.DateTimeFormatter;
import java.util.Date;
import java.util.HashMap;
import java.util.Locale;
import java.util.Map;

public class PDFServer {
    private static       HttpServer          SERVER;
    private static final DateTimeFormatter   DATE_TIME_FORMATTER    = DateTimeFormatter.ofPattern("EEE, dd MMM yyyy HH:mm:ss z", Locale.ENGLISH).withZone(ZoneId.of("GMT"));
    private static final Path                ROOT_PATH              = Path.of("/");
    private static final Date                RESOURCE_LAST_MODIIFED = new Date();
    private static final Map<String, byte[]> CACHE                  = new HashMap<>();
    private static       int                 PORT;

    public static synchronized void showPDF(Path path, int page) throws IOException, URISyntaxException {
        if (SERVER == null) {
            HttpServer server = HttpServer.create(new InetSocketAddress(0), 0);
            server.createContext("/", PDFServer::handleRequest);
            server.start();
            SERVER = server;
            PORT = server.getAddress().getPort();
        }
        URI uri = new URI("http://127.0.0.1:" + PORT + "/web/viewer.html?file=" + encodeQueryParam(path.normalize().toAbsolutePath().toString()) + "#page=" + page);
        Desktop.getDesktop().browse(uri);
    }

    private static void handleRequest(HttpExchange httpExchange) throws IOException {
        switch (httpExchange.getRequestMethod()) {
        case "HEAD", "GET" -> {
            URI    requestURI  = httpExchange.getRequestURI();
            Path   p           = Path.of(requestURI.getPath()).normalize();
            String contentType = contentType(PathUtils.getExtension(p));
            if ("application/pdf".equals(contentType)) {
                p = Path.of(URLDecoder.decode(requestURI.getPath(), StandardCharsets.UTF_8)).normalize();
                if (!Files.isRegularFile(p) || !Files.isReadable(p)) {
                    notFound(httpExchange);
                    return;
                }
                long    size;
                Instant instant;
                try {
                    size = Files.size(p);
                    instant = Files.getLastModifiedTime(p).toInstant();
                } catch (IOException ex) {
                    notFound(httpExchange);
                    return;
                }
                int expiresInSeconds = 12 * 60 * 60; // 12 hours
                httpExchange.getResponseHeaders().add("Content-Type", contentType);
                httpExchange.getResponseHeaders().add("Content-Length", Long.toString(size));
                httpExchange.getResponseHeaders().add("Last-Modified", DATE_TIME_FORMATTER.format(instant));
                httpExchange.getResponseHeaders().add("Expires", DATE_TIME_FORMATTER.format(new Date().toInstant().plusSeconds(expiresInSeconds)));
                httpExchange.getResponseHeaders().add("Cache-Control", "max-age=" + expiresInSeconds);
                httpExchange.sendResponseHeaders(200, size);
                if ("HEAD".equals(httpExchange.getRequestMethod())) {
                    httpExchange.getResponseBody().close();
                } else {
                    try (OutputStream out = httpExchange.getResponseBody()) {
                        Files.copy(p, out);
                    }
                }
                return;
            }
            String path = p.toString();
            byte[] data;
            synchronized (CACHE) {
                data = CACHE.get(path);
                if (data == null) {
                    ByteArrayOutputStream out = new ByteArrayOutputStream();
                    try (InputStream in = PDFServer.class.getModule().getResourceAsStream("/pdfjs" + p.toString())) {
                        in.transferTo(out);
                    } catch (IOException ioe) {
                        out = null;
                    }
                    if (out != null) {
                        data = out.toByteArray();
                        CACHE.put(path, data);
                    }
                }
            }
            if (data == null) {
                notFound(httpExchange);
                return;
            }
            int expiresInSeconds = 12 * 60 * 60; // 12 hours
            httpExchange.getResponseHeaders().add("Content-Type", contentType);
            httpExchange.getResponseHeaders().add("Content-Length", Integer.toString(data.length));
            httpExchange.getResponseHeaders().add("Last-Modified", DATE_TIME_FORMATTER.format(RESOURCE_LAST_MODIIFED.toInstant()));
            httpExchange.getResponseHeaders().add("Expires", DATE_TIME_FORMATTER.format(new Date().toInstant().plusSeconds(expiresInSeconds)));
            httpExchange.getResponseHeaders().add("Cache-Control", "public, immutable, max-age=" + expiresInSeconds);
            httpExchange.sendResponseHeaders(200, data.length);
            OutputStream body = httpExchange.getResponseBody();
            if (!"HEAD".equals(httpExchange.getRequestMethod())) {
                try (InputStream in = new ByteArrayInputStream(data)) {
                    in.transferTo(body);
                }
            }
            body.close();
        }
        default -> methodNotAllowed(httpExchange);
        }
    }

    private static String encodeQueryParam(String str) {
        StringBuilder buffer = new StringBuilder();
        byte[]        bytes  = str.getBytes(StandardCharsets.UTF_8);
        int           max    = bytes.length;
        for (byte b : bytes) {
            if ((b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z') || (b >= '0' && b <= '9') || b == '-' || b == '_' || b == '.' || b == '~' || b == '/') {
                buffer.append((char) b);
            } else if (b == ' ') {
                buffer.append('+');
            } else {
                buffer.append('%');
                buffer.append(Text.HEX_DIGITS[b >> 4]);
                buffer.append(Text.HEX_DIGITS[b & 15]);
            }
        }
        return buffer.toString();
    }

    private static boolean shouldEscape(char ch) {
        return !((ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || (ch >= '0' && ch <= '9') || ch == '-' || ch == '_' || ch == '.' || ch == '~' || ch == '/');
    }

    private static String contentType(String extension) {
        return switch (extension.toLowerCase()) {
            case "css" -> "text/css; charset=UTF-8";
            case "html" -> "text/html; charset=UTF-8";
            case "js" -> "text/javascript; charset=UTF-8";
            case "png" -> "image/png";
            case "gif" -> "image/gif";
            case "svg" -> "image/svg+xml; charset=UTF-8";
            case "jpg", "jpeg" -> "image/jpeg";
            case "pdf" -> "application/pdf";
            case "map" -> "application/json; charset=UTF-8";
            case "bcmap", "cur" -> "application/octet-stream";
            default -> "text/plain; charset=UTF-8";
        };
    }


    private static Path resolveWebFile(URI uri) {
        try {
            Path in = Path.of(uri.getPath()).normalize();
            if (in.isAbsolute()) {
                in = ROOT_PATH.relativize(in);
            }
            return ROOT_PATH.resolve(in).normalize();
        } catch (Exception ex) {
            Log.error(ex);
            return null;
        }
    }

    private static void notFound(HttpExchange httpExchange) throws IOException {
        byte[] body = "404 Not Found".getBytes(StandardCharsets.UTF_8);
        httpExchange.getResponseHeaders().add("Content-Type", "text/plain; charset=UTF-8");
        respond(httpExchange, 404, body);
    }

    private static void methodNotAllowed(HttpExchange httpExchange) throws IOException {
        byte[] body = "405 Method Not Allowed".getBytes(StandardCharsets.UTF_8);
        httpExchange.getResponseHeaders().add("Content-Type", "text/plain; charset=UTF-8");
        respond(httpExchange, 405, body);
    }

    private static void respond(HttpExchange httpExchange, int code, byte[] body) throws IOException {
        boolean isHead = "HEAD".equals(httpExchange.getRequestMethod());
        int     size   = body != null ? body.length : 0;
        if (isHead) {
            httpExchange.getResponseHeaders().add("Content-Length", Integer.toString(size));
            size = -1;
        }
        httpExchange.sendResponseHeaders(code, size);
        try (OutputStream out = httpExchange.getResponseBody()) {
            if (!isHead && body != null) {
                out.write(body);
            }
        }
    }

    private static void respondWithFile(HttpExchange httpExchange, Path path) throws IOException {
        if (!Files.isRegularFile(path) || !Files.isReadable(path)) {
            notFound(httpExchange);
            return;
        }
        long    size;
        Instant instant;
        try {
            size = Files.size(path);
            instant = Files.getLastModifiedTime(path).toInstant();
        } catch (IOException ex) {
            notFound(httpExchange);
            return;
        }
        int expiresInSeconds = 12 * 60 * 60; // 12 hours
        httpExchange.getResponseHeaders().add("Content-Type", contentType(PathUtils.getExtension(path)));
        httpExchange.getResponseHeaders().add("Content-Length", Long.toString(size));
        httpExchange.getResponseHeaders().add("Last-Modified", DATE_TIME_FORMATTER.format(instant));
        httpExchange.getResponseHeaders().add("Expires", DATE_TIME_FORMATTER.format(new Date().toInstant().plusSeconds(expiresInSeconds)));
        httpExchange.getResponseHeaders().add("Cache-Control", "max-age=" + expiresInSeconds);
        httpExchange.sendResponseHeaders(200, size);
        if ("HEAD".equals(httpExchange.getRequestMethod())) {
            httpExchange.getResponseBody().close();
        } else {
            try (OutputStream out = httpExchange.getResponseBody()) {
                Files.copy(path, out);
            }
        }
    }
}
