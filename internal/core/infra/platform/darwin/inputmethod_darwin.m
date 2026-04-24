//
//  inputmethod_darwin.m
//  Neru
//
//  Copyright © 2025 Neru. All rights reserved.
//

#import "inputmethod.h"

#import <Carbon/Carbon.h>
#import <Foundation/Foundation.h>

#pragma mark - Input Method

const char* getInputSourceID(void) {
	TISInputSourceRef source = TISCopyCurrentKeyboardInputSource();
	if (!source) {
		return NULL;
	}

	CFStringRef sourceID = (CFStringRef)TISGetInputSourceProperty(source, kTISPropertyInputSourceID);
	if (!sourceID) {
		CFRelease(source);
		return NULL;
	}

	// CFStringGetCStringPtr is fast but may return NULL for some encodings.
	// Fall back to CFStringGetCString in that case.
	char* result = NULL;
	const char* direct = CFStringGetCStringPtr(sourceID, kCFStringEncodingUTF8);
	if (direct) {
		result = strdup(direct);
	} else {
		CFIndex len = CFStringGetMaximumSizeForEncoding(CFStringGetLength(sourceID), kCFStringEncodingUTF8) + 1;
		result = (char*)malloc(len);
		if (result) {
			if (!CFStringGetCString(sourceID, result, len, kCFStringEncodingUTF8)) {
				free(result);
				result = NULL;
			}
		}
	}

	CFRelease(source);
	return result;
}

int switchInputSourceByID(const char* sourceID) {
	if (!sourceID) {
		return -1;
	}

	CFStringRef targetID = CFStringCreateWithCString(NULL, sourceID, kCFStringEncodingUTF8);
	if (!targetID) {
		return -1;
	}

	CFArrayRef sources = TISCreateInputSourceList(NULL, false);
	if (!sources) {
		CFRelease(targetID);
		return -1;
	}

	int result = -1;
	CFIndex count = CFArrayGetCount(sources);
	for (CFIndex i = 0; i < count; i++) {
		TISInputSourceRef src = (TISInputSourceRef)CFArrayGetValueAtIndex(sources, i);
		CFStringRef sID = (CFStringRef)TISGetInputSourceProperty(src, kTISPropertyInputSourceID);
		if (sID && CFStringCompare(sID, targetID, 0) == kCFCompareEqualTo) {
			// TISSelectInputSource requires the main thread (Carbon run loop).
			// Dispatch async so callers on non-main goroutines don't crash.
			CFRetain(src);
			dispatch_async(dispatch_get_main_queue(), ^{
				TISSelectInputSource(src);
				CFRelease(src);
			});
			result = 0;
			break;
		}
	}

	CFRelease(sources);
	CFRelease(targetID);
	return result;
}
