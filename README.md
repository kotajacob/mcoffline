# mcoffline

Convert minecraft whitelist to offline-mode whitelist.

# Usage
```
mcoffline k0taa
mcoffline whitelist.json
```

There are two ways to use `mcoffline`. You can simply run it with a username and
it will print out the corresponding offline-mode UUID for that name.

You can instead run it with a path to a `whitelist.json` file. It will convert
the file, store the new offline-mode version as `whitelist.json.offline` in the
same directory. Then it will look for a `server.properties` file in that same
directory to discover your world name. Finally, it will enter the world
directory and create a copy of the `playerdata` folder named
`playerdata.offline` with each user renamed to the offline-mode version.

Once finished you can manually verify `whitelist.offline.json` and
`playerdata.offline`. If you like them you can backup your old online-mode
versions and rename the offline-mode versions to get rid of the suffix.

# Install
```
make all
sudo make install
```

# Uninstall
```
sudo make uninstall
```

# Author
Written and maintained by Dakota Walsh.
Up-to-date sources can be found at https://git.sr.ht/~kota/mcoffline/

# License
GNU GPL version 3 or later, see LICENSE.
Copyright 2022 Dakota Walsh
