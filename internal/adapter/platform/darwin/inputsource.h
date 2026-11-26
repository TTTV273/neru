//
//  inputsource.h
//  Neru
//
//  Copyright © 2025 Neru. All rights reserved.
//

#ifndef INPUTSOURCE_H
#define INPUTSOURCE_H

#import <Foundation/Foundation.h>

#pragma mark - Input Source

/// Start keeping the cached input source current. Primes the cache and observes
/// kTISNotifySelectedKeyboardInputSourceChanged on the main queue. Safe to call
/// from any thread, and more than once.
void NeruInputSourceStartObserving(void);

/// Return the cached identifier of the selected input source. Never touches
/// TIS, so it does not block and is safe from any thread.
/// @return A malloc'd UTF-8 string the caller frees, or NULL when not yet known
char *_Nullable NeruCurrentInputSourceID(void);

/// Queue a switch to the input source with the given identifier on the main
/// queue. The cache takes the new identifier at once, so a read straight after
/// sees the switch it asked for.
/// @param sourceID Input source identifier, copied before this returns
void NeruSelectInputSource(const char *_Nonnull sourceID);

#endif /* INPUTSOURCE_H */
