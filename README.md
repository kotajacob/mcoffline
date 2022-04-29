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
same directory and then the same for ops.json. Next, it will look for
a `server.properties` file in that same directory to discover your world name.
Finally, it will enter the world directory and create a copy of the `playerdata`
folder named `playerdata.offline` with each user renamed to the offline-mode
version and one after the other it will create offline suffixed versions of
stats and advancements aswell.

Once finished you can manually verify the files, backup the online versions and
finally rename the files and folders to remove the `.offline` suffix.

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
