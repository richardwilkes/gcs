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
 * 2005-2007 the Initial Developer. All Rights Reserved.
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

#include <windows.h>
#include <conio.h>
#include <fcntl.h>
#include <stdlib.h>
#include <stdio.h>
#include <string.h>
#include <limits.h>
#include <sys/stat.h>
#include <jni.h>

#define xstr(s)	str(s)
#define str(s)	#s

typedef jint (JNICALL *CreateJavaVM)(JavaVM **vm,JNIEnv **env,void *args);

#define environ _environ

static void ExitWithError(char *msg) {
	MessageBox(NULL,msg,"Error",MB_OK);
	exit(1);
}

static void ExitIfNotZero(int result,char *msg) {
	if (result) {
		ExitWithError(msg);
	}
}

static void ExitIfZero(int result,char *msg) {
	if (!result) {
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

static char *CreateString2(char *format,char *embed1,char *embed2) {
	char *ptr = (char *)malloc(strlen(format) + strlen(embed1) + strlen(embed2));

	sprintf(ptr,format,embed1,embed2);
	return ptr;
}

static char JRE_version[MAX_PATH];
static char JRE_lib[MAX_PATH];

static int LocateSpecificJVM(void) {
	if (JRE_version[1] == '.' && (JRE_version[0] > '1' || (JRE_version[0] == '1' && JRE_version[2] >= '5'))) {
		unsigned long size = MAX_PATH;
		char path[MAX_PATH];
		HKEY key;
		int result;
		FILE *fp;

		sprintf(path,"SOFTWARE\\JavaSoft\\Java Runtime Environment\\%s",JRE_version);
		if (RegOpenKeyEx(HKEY_LOCAL_MACHINE,path,0,KEY_READ,&key) != ERROR_SUCCESS) {
			return 0;
		}
		result = RegQueryValueEx(key,"RuntimeLib",NULL,NULL,JRE_lib,&size);
		RegCloseKey(key);
		if (result != ERROR_SUCCESS) {
			return 0;
		}
		fp = fopen(JRE_lib, "rb");
		if (fp) {
			fclose(fp);
			return 1;
		}
	}
	return 0;
}	

static char *LocateJVM(void) {
	unsigned long size = MAX_PATH;
	HKEY key;
	int i = 0;

	if (RegOpenKeyEx(HKEY_LOCAL_MACHINE,"SOFTWARE\\JavaSoft\\Java Runtime Environment",0,KEY_READ,&key) != ERROR_SUCCESS) {
		ExitWithError("Unable to locate an installed Java Runtime Environment.");
	}
	if (RegQueryValueEx(key,"CurrentVersion",NULL,NULL,JRE_version,&size) != ERROR_SUCCESS) {
		RegCloseKey(key);
		ExitWithError("Unable to read the version of the current Java Runtime Environment.");
	}
	if (LocateSpecificJVM()) {
		RegCloseKey(key);
		return JRE_lib;
	}

	while (1) {
		char subKeyName[255];

		size = 255;
		if (RegEnumKeyEx(key,i++,subKeyName,&size,NULL,NULL,NULL,NULL) != ERROR_SUCCESS) {
			RegCloseKey(key);
			ExitWithError("Unable to locate a suitable Java Runtime Enviroment.");
		}
		strcpy(JRE_version,subKeyName);
		if (LocateSpecificJVM()) {
			RegCloseKey(key);
			return JRE_lib;
		}
	}
	return NULL; // Will never hit this line of code.
}

static JavaVM *CreateVM(JNIEnv **vm_env,JavaVMInitArgs *vm_args,char *pathToAppDir) {
	JavaVM *	vm;
	HINSTANCE	vmH;

	vmH = ExitIfNull(LoadLibrary(LocateJVM()),"Unable to load the Java Runtime Environment.");
	ExitIfNotZero((*(CreateJavaVM)GetProcAddress(vmH,"JNI_CreateJavaVM"))(&vm,vm_env,vm_args),"Unable to launch the Java Runtime Environment.");
	return vm;
}

static char *GetPathToAppDir(char *argv0) {
	static char	base[MAX_PATH] = "";

	if (!*base) {
		int	lastSlash	= -1;
		int	i			= 0;

		GetModuleFileName(NULL,base,MAX_PATH);
		while (base[i]) {
			if (base[i] == '\\' && (i < 1 || (i > 0 && base[i - 1] != ':'))) {
				lastSlash = i;
			}
			i++;
		}
		if (lastSlash >= 0) {
			base[lastSlash] = 0;
		}
	}
	return base;
}

int WINAPI WinMain(HINSTANCE inst,HINSTANCE prevInst,LPSTR cmdLine,int cmdShow) {
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
	for (i = 1; i < __argc; i++) {
		if (!strncmp(__argv[i],"-J",2) || !strncmp(__argv[i],"/J",2)) {
			if (strlen(__argv[i]) > 2) {
				vm_args.nOptions++;
				if (!strncmp(__argv[i],"-J-Xmx",6) || !strncmp(__argv[i],"/J-Xmx",6)) {
					has_mx_arg = 1;
				}
			}
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
	appDir							= GetPathToAppDir(__argv[0]);
	vm_options[j++].optionString	= CreateString1("-Djava.class.path=%s/GCS.jar",appDir);
	vm_options[j++].optionString	= CreateString1("-Dapp.home=%s",appDir);
	if (!has_mx_arg) {
		vm_options[j++].optionString = "-Xmx" xstr(MAX_RAM);
	}
	for (i = 1; i < __argc; i++) {
		if ((!strncmp(__argv[i],"-J",2) || !strncmp(__argv[i],"/J",2)) && strlen(__argv[i]) > 2) {
			vm_options[j++].optionString = &__argv[i][2];
		}
	}

	// Create the VM
	vm_args.version				= JNI_VERSION_1_4;
	vm_args.options				= vm_options;
	vm_args.ignoreUnrecognized	= JNI_TRUE;
	vm							= CreateVM(&vm_env,&vm_args,appDir);

	stringClass	= ExitIfNull((*vm_env)->FindClass(vm_env,"java/lang/String"),"The Java virtual machine is damaged.");
	mainClass	= ExitIfNull((*vm_env)->FindClass(vm_env,xstr(MAIN_CLASS)),"Unable to locate the main entry point.\nIs GCS.jar in the same directory as the program?");
	method		= ExitIfNull((*vm_env)->GetStaticMethodID(vm_env,mainClass,"main","([Ljava/lang/String;)V"),"The GCS jar file is damaged.");
	appArgs		= (*vm_env)->NewObjectArray(vm_env,arg_count,stringClass,NULL);

	for (i = 1, j = 0; i < __argc; i++) {
		if (strncmp(__argv[i],"-J",2) && strncmp(__argv[i],"/J",2)) {
			(*vm_env)->SetObjectArrayElement(vm_env,appArgs,j++,(*vm_env)->NewStringUTF(vm_env,__argv[i]));
		}
	}

	(*vm_env)->CallStaticVoidMethod(vm_env,mainClass,method,appArgs);
	(*vm)->DetachCurrentThread(vm);
	(*vm)->DestroyJavaVM(vm);
	return 0;
}
