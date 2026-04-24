//
//  inputmethod.h
//  Neru
//
//  Copyright © 2025 Neru. All rights reserved.
//

#ifndef INPUTMETHOD_H
#define INPUTMETHOD_H

#import <Foundation/Foundation.h>

#pragma mark - Input Method

/// Returns the identifier of the currently active keyboard input source.
/// The caller is responsible for freeing the returned C string with free().
/// @return a heap-allocated C string (e.g. "com.apple.keylayout.ABC"), or NULL on failure.
const char* getInputSourceID(void);

/// Activates the keyboard input source with the given identifier.
/// @param sourceID the TIS input source identifier to switch to
/// @return 0 on success, -1 if the source was not found or could not be selected
int switchInputSourceByID(const char* sourceID);

#endif /* INPUTMETHOD_H */
