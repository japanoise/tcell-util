# tcell-util

[![GoDoc](https://pkg.go.dev/github.com/japanoise/tcell-util?status.svg)](https://pkg.go.dev/github.com/japanoise/tcell-util)

This is a simple utility library for tcell v2. It's designed to be a
more-or-less drop-in replacement of my earlier library
[termbox-util](https://github.com/japanoise/termbox-util) (in fact, the code is
basically the same, just changing termbox idioms for tcell idioms). It provides
sensible printing functions, input functions that take care of their own event
loop and give you cute readline-style keybindings, and a way to parse a key
event that gives you an Emacs-like keystring (e.g. `^C` (ev.Key() is
tcell.KeyCtrlC) -> `"C-c"`). Check out the documentation for more. Note that the
package is imported as `termutil`; if you really like the name `tcell-util`, or
you really hate yourself and are using the original alongside tcell-util (as
well as termbox alongside tcell; please don't do this), import it like this:

```go
import (
    "github.com/gdamore/tcell/v2"
    tcell-util "github.com/japanoise/tcell-util"
)
```

## Will you do a version for tcell v1?

Well, no. But this isn't a complicated library, so it's not likely to use too
many v2-isms. You could easily revert the commit where I added v2 support, or
just do something like:

    sed -i -e 's|github.com/gdamore/tcell/v2|github.com/gdamore/tcell|g' *.go

## Can I trust it?

Well, I've been dogfooding the original version pretty much since its inception,
because [Gomacs](https://github.com/japanoise/gomacs) is my usual "small
changes" text editor. Sure, I rarely add much to it, but the API that it has is
not going to change (because there's nothing wrong with it, and it'd be a
pain...) and I'm on GitHub all the time for work, so any open issues should be
seen to quickly (I may leave my own issues lying around, but if they annoy a
user other than myself, then the gloves come off... again, see Gomacs).

That being said, I'm pretty lazy. This is a "for fun" side project; I don't even
use Gomacs at work much these days because environment issues force me to use
the [inferior C version](https://github.com/japanoise/emsys). I also don't
really care about go modules; I will use them as much as they are needed to not
break my code and explode violently, and no more (if the pre-module days were
good enough for me, they're good enough for you!), but that should be fine; like
I said, the API is stable and extremely unlikely to change. If someone
complains, sure, I'll roll up a release at HEAD and be done with it.

It also has one major issue: It always assumes you're not using the combining
character madness (i.e. calls to `SetContent` always have `nil` as the fourth
argument). This will not change. Sorry, I don't want to complicate the API for
something that the vast majority of people will never use.

## Can't I just change the import lines and use the original with tcell?

Yes and no. I experimented with moving Gomacs over to tcell a while back, using
tcell's termbox shim; I noted some issue which I forget. It may or may not occur
in termbox-util, too. Sorry, I don't care enough to try it out.
