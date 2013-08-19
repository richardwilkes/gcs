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
 * 2005-2008 the Initial Developer. All Rights Reserved.
 *
 * Contributor(s):
 *
 * ***** END LICENSE BLOCK ***** */

// The class that will be run
#define MAIN_CLASS com/trollworks/gcs/app/GCS

// The maximum amount of RAM the VM will use for the app
#ifndef MAX_RAM
#define MAX_RAM 256M
#endif

#include <stdlib.h>
#include <stdio.h>
#include <string.h>
#include <limits.h>
#include <sys/stat.h>
#include <jni.h>
#include <sys/resource.h>
#include <dlfcn.h>
#include <pthread.h>
#include <unistd.h>
#include <CoreFoundation/CoreFoundation.h>

#define xstr(s)	str(s)
#define str(s)	#s

typedef jint (JNICALL *CreateJavaVM)(JavaVM **vm,JNIEnv **env,void *args);

extern char **environ;

static void ExitWithError(char *msg) {
	fprintf(stderr,"%s\n",msg);
	exit(1);
}

static void ExitIfNotZero(int result,char *msg) {
	if (result) {
		ExitWithError(msg);
	}
}

static void *ExitIfNull(void *result,char *msg) {
	if (!result) {
		ExitWithError(msg);
	}
	return result;
}

static char *CreateString1(char *format,char *embed) {
	char *ptr = (char *)malloc(strlen(format) + strlen(embed));

	sprintf(ptr,format,embed);
	return ptr;
}

static JavaVM *CreateVM(JNIEnv **vm_env,JavaVMInitArgs *vm_args,char *pathToAppDir) {
	JavaVM *vm;

	setenv("JAVA_JVM_VERSION","1.5",1);
	ExitIfNotZero(JNI_CreateJavaVM(&vm,(void **)vm_env,vm_args),"Unable to launch the Java 1.5 virtual machine.");
	return vm;
}

static char *GetPathToAppDir(char *argv0) {
	static char	base[PATH_MAX] = "";

	if (!*base) {
		char *	macTrailer	= "/Contents/MacOS";
		char *	pos;
		char	path[PATH_MAX * 2 + 1];
		int		i;
		int		j;

		// First, get a full path to the executable
		strcpy(path,argv0);
		if (*path != '.' && *path != '/') {
			pos = path;
			while (*pos && *pos != '/') {
				pos++;
			}
			if (*pos != '/') {
				// Search the command path
				char *cmdPath	= strdup(getenv("PATH"));
				char *cmd		= pos = cmdPath;

				while (*pos) {
					struct stat statBuf;

					while (*pos && *pos != ':') {
						pos++;
					}
					if (*pos) {
						*pos++ = 0;
					}
					strcpy(path,cmd);
					strcat(path,"/");
					strcat(path,argv0);
					if (stat(path,&statBuf) == 0) {
						break;
					}
					cmd = pos;
					if (!*pos) {
						strcpy(path,argv0);
					}
				}
				free(cmdPath);
			}
		}
		realpath(path,base);
		
		// Now, extract the parent directory
		pos = base + strlen(base) - 1;
		while (pos > base && *pos != '/') {
			pos--;
		}
		if (pos >= base) {
			*pos = 0;
		}

		// The app dir needs to be where the bundle was, not where the
		// executable was on Mac OS X.
		i	= strlen(base);
		j	= strlen(macTrailer);
		if (i > j && !strcmp(base + i - j,macTrailer)) {
			pos		= base + i - j;
			*pos--	= 0;
			while (pos > base && *pos != '/') {
				pos--;
			}
			if (pos >= base) {
				*pos = 0;
			}
		}
	}
	return base;
}

static int Start(int argc,char **argv) {
	int				arg_count		= 0;
	int				has_mx_arg		= 0;
	JavaVMOption *	vm_options;
	JavaVMInitArgs	vm_args;
	JNIEnv *		vm_env;
	JavaVM *		vm;
	jclass			stringClass;
	jclass			mainClass;
	jmethodID		method;
	jobjectArray	appArgs;
	int				i;
	int				j;
	char *			appDir;

	// Determine if any parameters are meant to be passed to the VM
	vm_args.nOptions = 2;
	for (i = 1; i < argc; i++) {
		if (!strncmp(argv[i],"-J",2)) {
			if (strlen(argv[i]) > 2) {
				vm_args.nOptions++;
				if (!strncmp(argv[i],"-J-Xmx",6)) {
					has_mx_arg = 1;
				}
			}
		} else if (!strncmp(argv[i],"-psn_",5)) {
			// Ignore this argument, as Mac OS X adds it as the first (and only)
			// argument when the app bundle has been double-clicked.
		} else {
			arg_count++;
		}
	}
	if (!has_mx_arg) {
		vm_args.nOptions++;
	}
	vm_options = (JavaVMOption *)calloc(vm_args.nOptions,sizeof(JavaVMOption));

	// Setup the VM options
	j								= 0;
	appDir							= GetPathToAppDir(argv[0]);
	vm_options[j++].optionString	= CreateString1("-Djava.class.path=%s/GURPS Character Sheet.app/Contents/Resources/Java/GCS.jar",appDir);
	vm_options[j++].optionString	= CreateString1("-Dapp.home=%s",appDir);
	if (!has_mx_arg) {
		vm_options[j++].optionString = "-Xmx" xstr(MAX_RAM);
	}
	for (i = 1; i < argc; i++) {
		if ((!strncmp(argv[i],"-J",2)) && strlen(argv[i]) > 2) {
			vm_options[j++].optionString = &argv[i][2];
		}
	}

	// Create the VM
	vm_args.version				= JNI_VERSION_1_4;
	vm_args.options				= vm_options;
	vm_args.ignoreUnrecognized	= JNI_TRUE;
	vm							= CreateVM(&vm_env,&vm_args,appDir);

	stringClass	= ExitIfNull((*vm_env)->FindClass(vm_env,"java/lang/String"),"The Java virtual machine is damaged.");
	mainClass	= ExitIfNull((*vm_env)->FindClass(vm_env,xstr(MAIN_CLASS)),"Unable to locate the main entry point.");
	method		= ExitIfNull((*vm_env)->GetStaticMethodID(vm_env,mainClass,"main","([Ljava/lang/String;)V"),"The GCS jar file is damaged.");
	appArgs		= (*vm_env)->NewObjectArray(vm_env,arg_count,stringClass,NULL);

	for (i = 1, j = 0; i < argc; i++) {
		if (strncmp(argv[i],"-J",2) && strncmp(argv[i],"-psn_",5)) {
			(*vm_env)->SetObjectArrayElement(vm_env,appArgs,j++,(*vm_env)->NewStringUTF(vm_env,argv[i]));
		}
	}

	(*vm_env)->CallStaticVoidMethod(vm_env,mainClass,method,appArgs);
	(*vm)->DetachCurrentThread(vm);
	(*vm)->DestroyJavaVM(vm);
	return 0;
}

static int		gArgCount;
static char **	gArgValues;

static void *startupJava(void *unused) {
	exit(Start(gArgCount,gArgValues));
}

/* Callback for dummy source used to make sure the CFRunLoop doesn't exit right
 * away.
 */
static void sourceCallBack(void *info) {
}

int main(int argc,char **argv) {
	size_t					stack_size		= 0;
	CFRunLoopSourceContext	sourceContext;
	struct rlimit			limit;
	int						rc;
	pthread_t				vmthread;
	pthread_attr_t			thread_attr;

	gArgCount	= argc;
	gArgValues	= argv;

	{
		long	pid			= (long)getpid();
		char	nameBuffer[32];
		char	iconBuffer[32];

		sprintf(nameBuffer,"APP_NAME_%ld",pid);
		setenv(nameBuffer,"GURPS Character Sheet",1);
		sprintf(iconBuffer,"APP_ICON_%ld",pid);
		setenv(iconBuffer,CreateString1("%s/GURPS Character Sheet.app/Contents/Resources/app.icns",GetPathToAppDir(argv[0])),1);
	}

	// Create a new pthread, copying the stack size of the primary pthread
	rc = getrlimit(RLIMIT_STACK,&limit);
	if (rc == 0) {
		if (limit.rlim_cur != 0LL) {
			stack_size = (size_t)limit.rlim_cur;
		}
	}
	pthread_attr_init(&thread_attr);
	pthread_attr_setscope(&thread_attr,PTHREAD_SCOPE_SYSTEM);
	pthread_attr_setdetachstate(&thread_attr,PTHREAD_CREATE_DETACHED);
	if (stack_size > 0) {
		pthread_attr_setstacksize(&thread_attr,stack_size);
	}

	// Start the thread that we will start the JVM on
	pthread_create(&vmthread,&thread_attr,startupJava,NULL);
	pthread_attr_destroy(&thread_attr);

	// Create a a sourceContext to be used by our source that makes sure the
	// CFRunLoop doesn't exit right away
	memset(&sourceContext,0,sizeof(sourceContext));
	sourceContext.perform = &sourceCallBack;
	
	// Use the constant kCFRunLoopCommonModes to add the source to the set of
	// objects monitored by all the common modes
	CFRunLoopAddSource(CFRunLoopGetCurrent(),CFRunLoopSourceCreate(NULL,0,&sourceContext),kCFRunLoopCommonModes); 
	
	// Park this thread in the runloop
	CFRunLoopRun();
	
	return 0;
}
