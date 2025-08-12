# v? ?

* New secondary spirit: Fire Salamander. It features a *fire retreat* ability
  that makes the player retreat up to two tiles while leaving a fire cloud on
  the previous position, as long as there's room for the retreat. The new
  spirit also confers +1 attack and a chance of burning foes on hit (like
  phoenixes do) and the upgraded version confers fire resistance (half duration
  of Fire effects).  As a drawback, it inflicts bad hearing on the player,
  making it harder to hear footsteps and impossible to hear creeping and
  wing-flapping noises. If the Fire Salamander is combined with the Barking
  Hound spirit, hearing becomes normal again (neither good nor bad). Fire
  Salamander is hence a spirit that offers very good offense capabilities and
  some mobility, but it can significantly impair stealth tactics.
* New unique monster: walking mushroom. It's less sturdy than a walking tree
  but releases lignifying spores on sight from afar, affecting other monsters
  too. Also, its melee hits are confusing.
* New unique monster: noisy imp. It performs music just for the player,
  attracting nearby monsters. It avoids fighting similarly to afraid monsters.
* New unique monster: teleporting ghost. Its hits teleport the player away.
* Improve “poison cloud” ability of the Venomous Viper spirit so that the
  poison clouds are created up to two free tiles away (before a monster or
  obstacle). This creates more varied possibilities for tactical movement,
  while still restricting safe movements in a similar but less extreme way, and
  providing similar defense from ranged monsters.
* When the player or a monster is both afraid and cornered with no possible
  escape, they now become berserk for a short while!
* Using “sprint” or “jump” abilities while imbalanced may now sometimes lead to
  a dazing fall.
* Space-distorting attacks now swap clouds too. While getting swapped into a
  cloud of fire or poison by foes was (very occasionally) fun, Temporal Cat
  players not being able to safely attack monsters on dangerous clouds was
  more annoying than fun. And in both cases, it could result in unexpected
  dangerous situations that would've required explicit examination to be
  avoided: when a monster is on a tile with a cloud, only the monster is drawn,
  and while that limitation could be solved for the tiles version with image
  layers, introducing that kind of visual clutter doesn't fit the simplicity of
  the game.
* Players with the Rampaging Boar spirit don't move into poison or fire clouds
  after pushing anymore. While sometimes funny, it had similar issues as
  swapping into unexpected clouds. Monster boars and dragons still move into
  dangerous clouds after pushing, because they're enraged and corrupted, so not
  as smart!
* The “sprint” and “focus” abilities now cancel each other. It's never really
  useful nor thematic to have both at the same time, so hydra players have now
  an instant sprint cancellation option (at the cost of a spirit charge).
* While sprinting, the frog's “jump” ability now unbalances foes without
  hurting them, but still dazes them, effectively combining the effects of the
  two kinds of jumps in a sensible way.
* Improve a bit monster distribution. Mostly, bias toward early-mid-late
  monsters depending on map level has been increased, making the various phases
  more distinct. Early and Mid game difficulty remains similar. Late game has
  now overall less monsters than before, but a bit more of dangerous ones, so
  it has become more stealth-friendly while still being challenging enough.
* Improve wandering monster destination selection. It's still quite random with
  some bias toward close locations, but there's some extra bias toward
  special vaults, so that interesting places are visited more often by
  monsters.
* Improvements in monster group formation. In particular, less groups of three
  monsters than before and more groups of two, so somewhat better coverage of
  the map in mid and late game, but smaller likeliness of large monster groups.
* Add a few new special vaults.
* Nadre explosions now properly blow up nearby walls too, like in Boohu and
  Harmonist.
* Fix spelling and formatting issues in a few log messages. Avoid case of
  duplicated "Saw *monster*" entries in timeline.
* Fix field of view sometimes not being updated just after cloud creation but
  later at the end of a turn, which could allow monsters to attack once from a
  distance before the update.
* Add concise information about current tile in the status bar which tells
  whether there's fog, an item, or a special kind of terrain under the player.
* Animation improvements. Monster deaths are now always animated when in view,
  independently of the cause. Hits between confused monsters are now animated
  too, making it clearer what's happening.
* Improve a few tiles (footsteps, translucent walls, food).

# v1.0.0-beta.1 2025-07-22

First public release of Shamogu.
