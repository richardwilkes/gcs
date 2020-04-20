package com.lowagie.text.pdf;

import com.lowagie.text.ExceptionConverter;

import java.security.MessageDigest;

public class PdfEncryption {
	private static int seq;

	public static PdfObject createInfoId(byte id[]) {
		ByteBuffer buf = new ByteBuffer(90);
		buf.append('[').append('<');
		for (int k = 0; k < 16; ++k) {
			buf.appendHex(id[k]);
		}
		buf.append('>').append('<');
		id = createDocumentId();
		for (int k = 0; k < 16; ++k) {
			buf.appendHex(id[k]);
		}
		buf.append('>').append(']');
		return new PdfLiteral(buf.toByteArray());
	}

	public static byte[] createDocumentId() {
		MessageDigest md5;
		try {
			md5 = MessageDigest.getInstance("MD5");
		} catch (Exception e) {
			throw new ExceptionConverter(e);
		}
		long time = System.currentTimeMillis();
		long mem = Runtime.getRuntime().freeMemory();
		String s = time + "+" + mem + "+" + seq++;
		return md5.digest(s.getBytes());
	}
}
