// This file exposes compile-time configuration constants or variables that can
// be edited by hand before building the game.

package main

// GluttonyRework makes the Gluttonous Bear from Advanced Spirits eat in pairs
// through a choice-of-two menu, without relying on a Gluttony status.
const GluttonyRework = false

// ResetMods makes it so starting a new game resets mod selection, instead of
// restoring the saved selection from the previous session.
const ResetMods = false

// VampiricHC implicitly sets Healing Combat when the Vampiric Bat is chosen,
// irrespectively of mod selection.
const VampiricHC = false

// NoChaos changes chaos megabat to a screeching megabat with no chaos bite,
// and removes the rare comestible.
const NoChaos = false
