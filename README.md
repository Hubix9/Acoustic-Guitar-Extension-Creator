# Music-Extension-Creator
Tool created to let users of acoustic guitar and pocket music player mods easily add songs

## How to download
In order to download the tool, head to the *Releases* tab:
https://github.com/Hubix9/Acoustic-Guitar-Extension-Creator/releases

## How to use

1. Extract Music Extension Creator folder from downloaded .zip
2. Enter that folder
3. Put your songs in the songs folder (.mp3 format is recommended)
4. Run Music_extension_creator.exe file
5. Enter or modify information in appropriate fields, you only need to edit extension name, description and author, rest you can leave on default values
6. Press "Create extension" button in the upper right corner
7. You addon will be created in the tool directory under name: @your_chosen_addon_name
8. Load that folder as mod in Arma 3 launcher
9. Enjoy!


## Known Issues
* If you encounter an issue with extension creation, you can fall back to **[Old version](https://github.com/Hubix9/Acoustic-Guitar-Extension-Creator/tree/master)** of this tool
* Make sure that filenames of songs don't contain special characters. Underscores, spaces, numbers and letters are ok, though it's best to keep them simple like: MySong1.mp3
* Tool will crash if songs folder contains .txt file or has no song files inside of it

## Credits
**[FFmpeg](https://www.ffmpeg.org/)** - used to convert songs to .ogg format and to get informations about them

**[gopbo](github.com/g0dsCookie/gopbo)** - used to pack mod files into .pbo file

**[fyne](https://github.com/fyne-io/fyne)** - ui framework used to build the application interface
