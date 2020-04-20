package org.apache.commons.logging;

public class NoOpLog implements Log {
	public NoOpLog() {
	}

	@Override
	public void trace(Object message) {
		//
	}

	@Override
	public void trace(Object message, Throwable t) {
		//
	}

	@Override
	public void debug(Object message) {
		//
	}

	@Override
	public void debug(Object message, Throwable t) {
		//
	}

	@Override
	public void info(Object message) {
		//
	}

	@Override
	public void info(Object message, Throwable t) {
		//
	}

	@Override
	public void warn(Object message) {
		//
	}

	@Override
	public void warn(Object message, Throwable t) {
		//
	}

	@Override
	public void error(Object message) {
		//
	}

	@Override
	public void error(Object message, Throwable t) {
		//
	}

	@Override
	public void fatal(Object message) {
		//
	}

	@Override
	public void fatal(Object message, Throwable t) {
		//
	}

	@Override
	public final boolean isDebugEnabled() {
		return false;
	}

	@Override
	public final boolean isErrorEnabled() {
		return false;
	}

	@Override
	public final boolean isFatalEnabled() {
		return false;
	}

	@Override
	public final boolean isInfoEnabled() {
		return false;
	}

	@Override
	public final boolean isTraceEnabled() {
		return false;
	}

	@Override
	public final boolean isWarnEnabled() {
		return false;
	}
}
