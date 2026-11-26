//
//  inputsource_darwin.m
//  Neru
//
//  Copyright © 2025 Neru. All rights reserved.
//

#import "inputsource.h"
#import "nativelog.h"

#import <Carbon/Carbon.h>
#import <Cocoa/Cocoa.h>
#import <os/lock.h>

#pragma mark - Cache

// gCurrentSourceID is the selected input source as last seen. TIS is read and
// written only on the main queue, and Go reads this cache instead, because the
// mode handler asks under its lock and must not wait on the main thread.
static os_unfair_lock gSourceLock = OS_UNFAIR_LOCK_INIT;
static NSString *gCurrentSourceID = nil;
static id gSourceObserver = nil;

static void setCachedSourceID(NSString *sourceID) {
	os_unfair_lock_lock(&gSourceLock);
	gCurrentSourceID = [sourceID copy];
	os_unfair_lock_unlock(&gSourceLock);
}

/// Read the selected input source from TIS. Main queue only.
static NSString *readSelectedSourceID(void) {
	TISInputSourceRef source = TISCopyCurrentKeyboardInputSource();
	if (!source) {
		return nil;
	}

	CFStringRef sourceID = (CFStringRef)TISGetInputSourceProperty(source, kTISPropertyInputSourceID);
	NSString *result = sourceID ? [(__bridge NSString *)sourceID copy] : nil;
	CFRelease(source);

	return result;
}

#pragma mark - Observing

void NeruInputSourceStartObserving(void) {
	dispatch_async(dispatch_get_main_queue(), ^{
		setCachedSourceID(readSelectedSourceID());

		if (gSourceObserver) {
			return;
		}

		gSourceObserver = [[NSDistributedNotificationCenter defaultCenter]
		    addObserverForName:(__bridge NSString *)kTISNotifySelectedKeyboardInputSourceChanged
		                object:nil
		                 queue:[NSOperationQueue mainQueue]
		            usingBlock:^(NSNotification *_Nonnull note) {
			            (void)note;
			            setCachedSourceID(readSelectedSourceID());
		            }];
	});
}

char *NeruCurrentInputSourceID(void) {
	os_unfair_lock_lock(&gSourceLock);
	NSString *sourceID = gCurrentSourceID;
	os_unfair_lock_unlock(&gSourceLock);

	if (sourceID.length == 0) {
		return NULL;
	}

	return strdup(sourceID.UTF8String);
}

#pragma mark - Selecting

void NeruSelectInputSource(const char *sourceID) {
	if (!sourceID) {
		return;
	}

	NSString *target = [NSString stringWithUTF8String:sourceID];
	if (target.length == 0) {
		return;
	}

	setCachedSourceID(target);

	dispatch_async(dispatch_get_main_queue(), ^{
		NSDictionary *filter = @{(__bridge NSString *)kTISPropertyInputSourceID : target};
		CFArrayRef sources = TISCreateInputSourceList((__bridge CFDictionaryRef)filter, false);

		BOOL selected = NO;
		if (sources) {
			if (CFArrayGetCount(sources) > 0) {
				TISInputSourceRef source = (TISInputSourceRef)CFArrayGetValueAtIndex(sources, 0);
				selected = TISSelectInputSource(source) == noErr;
			}
			CFRelease(sources);
		}

		if (!selected) {
			NeruLog(NeruLogLevelWarn, @"Input source could not be selected", target);
			// The optimistic cache entry is wrong, so put the real one back.
			setCachedSourceID(readSelectedSourceID());
		}
	});
}
