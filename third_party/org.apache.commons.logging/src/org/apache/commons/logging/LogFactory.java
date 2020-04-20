package org.apache.commons.logging;

public class LogFactory {
	public static final String		PRIORITY_KEY						= "priority";
	public static final String		TCCL_KEY							= "use_tccl";
	public static final String		FACTORY_PROPERTY					= "org.apache.commons.logging.LogFactory";
	public static final String		FACTORY_DEFAULT						= "org.apache.commons.logging.impl.LogFactoryImpl";
	public static final String		FACTORY_PROPERTIES					= "commons-logging.properties";
	public static final String		DIAGNOSTICS_DEST_PROPERTY			= "org.apache.commons.logging.diagnostics.dest";
	public static final String		HASHTABLE_IMPLEMENTATION_PROPERTY	= "org.apache.commons.logging.LogFactory.HashtableImpl";
	private static final LogFactory	instance							= new LogFactory();
	private static final NoOpLog	log									= new NoOpLog();

	protected LogFactory() {
	}

	@SuppressWarnings({ "static-method", "unused" })
	public Object getAttribute(String name) {
		return null;
	}

	@SuppressWarnings("static-method")
	public String[] getAttributeNames() {
		return new String[0];
	}

	@SuppressWarnings({ "static-method", "unused", "rawtypes" })
	public Log getInstance(Class clazz) throws LogConfigurationException {
		return log;
	}

	@SuppressWarnings({ "static-method", "unused" })
	public Log getInstance(String name) throws LogConfigurationException {
		return log;
	}

	public void release() {
		//
	}

	@SuppressWarnings("unused")
	public void removeAttribute(String name) {
		//
	}

	@SuppressWarnings("unused")
	public void setAttribute(String name, Object value) {
		//
	}

	public static LogFactory getFactory() throws LogConfigurationException {
		return instance;
	}

	@SuppressWarnings("rawtypes")
	public static Log getLog(Class clazz) throws LogConfigurationException {
		return getFactory().getInstance(clazz);
	}

	public static Log getLog(String name) throws LogConfigurationException {
		return getFactory().getInstance(name);
	}

	@SuppressWarnings("unused")
	public static void release(ClassLoader classLoader) {
		//
	}

	public static void releaseAll() {
		//
	}

	public static String objectId(Object o) {
		if (o == null) {
			return "null";
		}
		return o.getClass().getName() + "@" + System.identityHashCode(o);
	}
}
