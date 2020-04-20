module org.apache.fontbox {
	requires org.apache.commons.logging;

	requires transitive java.desktop;

	exports org.apache.fontbox;
	exports org.apache.fontbox.afm;
	exports org.apache.fontbox.cff;
	exports org.apache.fontbox.cmap;
	exports org.apache.fontbox.encoding;
	exports org.apache.fontbox.pfb;
	exports org.apache.fontbox.ttf;
	exports org.apache.fontbox.type1;
	exports org.apache.fontbox.util;
	exports org.apache.fontbox.util.autodetect;
}
