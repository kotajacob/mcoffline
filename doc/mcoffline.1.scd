mcoffline(1) [""["mcoffline-VERSION"]]

# NAME

mcoffline - Convert minecraft whitelist to offline-mode whitelist.

# SYNOPSIS

*mcoffline* [_USER_ | _FILE_]

# DESCRIPTION
```
mcoffline k0taa
mcoffline /path/to/whitelist.json
```

There are two ways to use *mcoffline*. You can simply run it with a username and
it will print out the corresponding offline-mode UUID for that name.

You can instead run it with a path to a *whitelist.json* file. It will convert
the file, store the new offline-mode version as *whitelist.offline.json* in the
same directory. Then it will look for a *server.properties* file in that same
directory to discover your world name. Finally, it will enter the world
directory and create a copy of the *playerdata* folder named *playerdata.offline*
with each user renamed to the offline-mode version.

Once finished you can manually verify *whitelist.offline.json*
and *playerdata.offline*. If you like them you can backup your old online-mode
versions and rename the offline-mode versions to get rid of the suffix.

# AUTHORS

Maintained by Dakota Walsh <kota at nilsu.org>.
Up-to-date sources can be found at https://git.sr.ht/~kota/mcoffline/

# COPYRIGHT

Copyright 2022 Dakota Walsh

License GPLv3+ GNU GPL version 3 or later <https://gnu.org/licenses/gpl.html>.
This is free software: you are free to change and redistribute it. There is NO
WARRANTY, to the extent permitted by law.
