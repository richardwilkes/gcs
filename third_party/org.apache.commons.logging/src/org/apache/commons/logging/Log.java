package org.apache.commons.logging;

public interface Log {
	void debug(Object message);

	void debug(Object message, Throwable t);

	void error(Object message);

	void error(Object message, Throwable t);

	void fatal(Object message);

	void fatal(Object message, Throwable t);

	void info(Object message);

	void info(Object message, Throwable t);

	boolean isDebugEnabled();

	boolean isErrorEnabled();

	boolean isFatalEnabled();

	boolean isInfoEnabled();

	boolean isTraceEnabled();

	boolean isWarnEnabled();

	void trace(Object message);

	void trace(Object message, Throwable t);

	void warn(Object message);

	void warn(Object message, Throwable t);
}
