# BluRay Playback Utilities

This is a set of utilities that enable playback of playlists on BluRay discs or decrypted ISOs when one can't use an actual BluRay player.  

Examples of situations where this might be true

* You want to use your BluRay images in media servers such as Plex, Emby or Jellyfin without remuxing them into an MKV.
* Your bluray drive is on one computer, but you want to play the movie on another device

## Utilities included

* **bluray-server**: This is an http server that enables you to specify an iso and a playlist and stream the m2ts contents of the BluRay playlists over http
 
* **m2ts-fs**: This is a fuse file system that allows you create _stub_ m2ts files that specify what iso and playlist they should represent.  When viewed within the file system, the contents of the file are the m2ts contents of the BluRay playlist 

* **bd-fs**: This is a fuse file system that enumerates all playlists on an ISO or bluray device enabling you to pick which one you want to play without any configuration.  When combined with MakeMKV's libmmbd, it can even enable one to use physical bluray discs in a drive as MakeMKV will handle the decryption on the fly

## Building the Utilities


```sh
go build ./cmd/bluray-server
go build ./cmd/bd-fs
go build ./cmd/m2ts-fs
```  
## bluray-server usage

you just run it ```./bluray-server``` and it will run an http server on port 8080.  One can configure which port it runs on with the ```-port``` flag

one can then fetch an m2ts stream by connecting to it and passing as url query parameters the full path to the iso and the playlist one wants to stream.

ex: ```http://localhost:8080/getm2ts?file=<url encoded filepath>&playlist=<playlist_number>```

## bd-fs

you run it with what device/iso you want to export, and where it should be mounted.

ex: ```./bd-fs <some iso> <some mount point> ```

One can use it with a physcal bluray device with a disc inside, by simply specifying the device instead of an iso, such as ```/dev/dvd```.  By itself, this wont work, as most BluRay discs are encrypted with libaacs.  However, if one has MakeMKV installed, one can use it to decryt the disc on the fly with its ```libmmbd``` library.

ex: ```LD_PRELOAD=/usr/lib/libmmbd.so.0 ./bd-fs /dev/dvd <some mount point>``` 

## m2ts-fs

you run it with what directory root you want to translate.  The fuse file system will hide all ```.iso``` files from view and turn every m2ts file it finds into a read only view of the playlist it defines.

the ```.m2ts``` stub files are simple yaml files

```yaml
file: <iso file relative to directory of .m2ts stub>
playlist: <playlist number to use from file specified>
``` 

therefore, one can populate the directory structure with these ```.m2ts``` stub files and the fuse file system will export them as playable m2ts files with the contents of the defined playlist

In order to make creation of the these files a little easier, another utility ```find-titles``` was written to enable autogeneration of m2ts files that fit user defined critiera.  With the utility, one can automatically create m2ts files thoughout the directory with a simple find command

ex: ```find . -name '*.iso' -exec ~/find-titles {} \;```  One can explore the option flags to ```find-titles``` to learn more about.

One you have a directory structure filled with```.m2ts``` stub files, one simply mounts it

ex: ```./m2ts-fs <directory to translate> <mount point>```  
