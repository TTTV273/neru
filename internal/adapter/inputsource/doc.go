// Package inputsource reads and switches the system keyboard input source.
//
// The mode handler uses it to type in a plain Latin layout while a mode is
// open and to give the user's own input source (a Vietnamese Telex input
// method, say) back when Neru returns to idle.
//
// Platform slots: inputsource_darwin.go (TIS, through the darwin bridge) and
// inputsource_other.go, which answers "" and CodeNotSupported.
package inputsource
