// This file exposes compile-time configuration variables that can be edited by
// hand.

package main

// GluttonyRework makes the Gluttonous Bear from Advanced Spirits eat in pairs
// through a choice-of-two menu, without relying on a Gluttony status.
var GluttonyRework = false

// ResetMods makes it so starting a new game resets mod selection, instead of
// restoring the saved selection from the previous session.
var ResetMods = false

// VampiricHC implicitly sets Healing Combat when the Vampiric Bat is chosen,
// irrespectively of mod selection.
var VampiricHC = false
