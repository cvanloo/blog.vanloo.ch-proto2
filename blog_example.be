{author {name Colin van~Loo} {email contact@vanloo.ch}}
{title Reviewing the reMarkable}
{tags reMarkable review technology proprietary}
{abstract
}
{body

{paragraph
I have been using a reMarkable 2 for close to three years now.
The reMarkable is a well polished, high quality paper tablet.
Consistent with their advertised claims, writing on it really feels like
writing on real paper.
However, I don't think that the reMarkable is a safe space to keep your notes.
}

{section The Good Parts

{paragraph
Writing on a reMarkable feels really good.
Though you better leave a little margin to your notes.
At the screen edges the display has problems tracking the pen.
}

{paragraph
There are a variety of different pen(cils) and a highlighter, that change the
appearance of your scribbles.
My favorite is the mechanical pencil.
}

{paragraph
The user interface is clean and easy to use with a handful of gestures.
My only problem is, that sometimes while writing I accidentally touch the screen with my hand.
This the reMarkable interprets as a two finger tap, or the undo command.
}

{paragraph
The battery life is highly dependent on your usage.
If you just have your tablet laying around, sitting idle for weeks, it won't use any battery at all.
Even taking notes daily, I maybe charge it once every two weeks.
If I'm on my reMarkable for hours a day (during an intense phase of studying), I have to charge it more often.
}

}

{section The Bad Parts

{paragraph
To start with the most obvious one: the price.
If it weren't for the 30-day money back guarantee, I would have never bought a reMarkable.
However, after trying it for a month, returning it was the last thing on my mind.
}

{paragraph
With a reMarkable, you won't pay only once.
Here's what I think is the reMarkable's greatest anti-feature: The disposable pen tips.
Supposedly those tips provide for a more realistic feeling.
In practice, the tips are used up in no time, and refills are expensive.
}

{paragraph
The marker (that's what they call the pen you use to write) is rather fragile, an unacceptable deficiency considering how much it costs.
There is a thin piece of plastic to support the disposable tip.
If you're clumsy like me, you'll drop the marker at most three times before this
support has broken off and now your forced to buy a new marker entirely.
}

{paragraph
Furthermore, to really get the most out of your paper tablet, you're going to need a monthly {enquote Connect} subscription.
As an early adopter, I'm lucky that I got it for free for a lifetime.
{enquote Connect} synchronizes your files to the (Google) cloud.
From there, you can also access notes from the desktop ({sidenote Windows only \\ I got it to run with WINE once, but not since any of the more recent updates.}) and Android app.
}

{paragraph
There is a way around this, if you're technically inclined, that is.
Since the reMarkable runs on Linux, you can {mono ssh} into it, take your own backups using {mono rsync}, {sidenote possibly \\ I haven't tried that out yet.} even install a Syncthing service on it.
}

{paragraph
Anther annoyance will become apparent after maybe two or three years of note taking.
The reMarkable is equipped with laughably little storage space.
Seriously, it's not the sixties anymore.
Maybe the Apollo Guidance Computer ran on 2048 words of memory, but I'll be hard pressed to fit my notes into that.
}

{paragraph
Finally, {mono .rm} files are a proprietary format.
Aside from the reMarkable software, there is nothing else that can render your written thoughts.
You're locked in, switching to a different vendor would mean losing all of your notes.
If your reMarkable has an update, and all of a sudden isn't able to display your old notes anymore, there's {sidenote nothing \\ Everything is open source if you know how to reverse engineer, I know, shut up.} you can do.
If reMarkable, for some reason, decides that they don't wont to do any more business with you, your out of luck.
}

{paragraph
reMarkable files can be exported to PDF, but it's a slow process (only one notebook at a time).
Most PDF viewers don't even render the PDFs properly.
On the reMarkable a single page can be endlessly long and wide, but all the PDF viewers I tried ended up cutting parts off (Zathura, Edge) or only rendering a black page (Firefox).
I also had a problem that my reMarkable failed to generate PDFs for notebooks I had created a few software updates ago.
}

{subsection Don't Get a Type Folio

{paragraph
I can count the number of supported keyboard layouts on one hand.
You're running Linux, couldn't you have just made use of Linux's wide variety of keyboard layouts?
I'm told there's a way to modify the reMarkable and make your own keyboard layouts, but as of yet, I haven't figured out how.
}
}
}

{section A redeemable quality: It runs Linux

{paragraph
The reMarkable runs on Linux and it's surprisingly open about that.
In fact, as soon as it's plugged into a computer by USB, an SSH port is
automatically opened.
}

{paragraph
If you go cramming around in the settings, somewhere hidden beneath copyrights
and EULAs, in bold text, you can find some IP addresses and the root password.
}

{comment @todo: insert image here}

{paragraph
To get easier access in the future, you should probably {mono ssh-copy-id}
and add an SSH configuration:
}

{comment @todo: filename ~/.ssh/config}
{code \+
host remarkable
    Hostname 10.11.99.1
    User root
\+}

{paragraph
I use this to create backups:
}

{code \+
mkdir -p rm-backup-`date +%F`/files && cd $_/..
scp remarkable:~/.config/remarkable/xochitl.conf . # backup config
scp remarkable:/usr/bin/xochitl . # backup xochitl binary
rsync -aAXv remarkable:~/.local/share/remarkable/xochitl/ files/ # backup files
\+}
}

{section Features I'd Like to See

{paragraph
Aside from a pen that doesn't deplete, I'd find it useful to be able to insert
space in the middle of a page. (OneNote has a feature just like it.)
}

{paragraph
I often take notes during a lecture.
After the lecture I go over it once more, and extend on parts that are still
unclear to me.
Adding notes in between already crammed notes becomes quite awkward.
}

{paragraph
With its eInk display, the reMarkable also serves as a comfortable (for the
eyes) eReader.
Since the reMarkable is smaller than A4, a lot of PDFs (especially those
American scientific paper that use up half of the page for just margins) appear tiny.
It's possible to zoom in, but zoom is reset every time you flip a page.
Having to readjust the zoom all the time distracts from reading.
I've asked their customer support to change this behavior, they told me they've got no intentions to.
}
}
}

