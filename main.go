package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/data/validation"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"github.com/g0dsCookie/gopbo/pbo"
)

var workDir string

type ExtensionData struct {
	Name                         string      //Name of the extension
	ClearedName                  string      //Name of extension after removing unwanted characters
	Creator                      string      //Nickname of creator of the extension
	ClearedCreator               string      //Nickname of creator after removing unwanted characters
	Description                  string      //Description of the extension
	ClearedDescription           string      //Description of the extension after removing unwanted characters
	SongGroups                   []SongGroup //Songs in the extension
	GenerateGuitarCfg            bool
	GeneratePocketMusicPlayerCfg bool
	VolumeMultiplier             float64
}

var extensiondata ExtensionData

type SongStruct struct {
	Name                  string
	OriginalName          string  //Original name of the song
	ExtensionStrippedName string  //Name stripped of extension
	ClearedName           string  //Name cleared of unwanted characters
	Length                int     //Duration of song, in seconds
	Distance              int     //Hearing distance of the song
	Volume                float64 //Song volume, converted to Arma-ready value
	Ogg                   bool    //If song is already in .ogg format
	Path                  string  //song file path
	InternalPath          string  //path to song inside extension
	Pitch                 float64 //Pitch of the song
}

var songs []SongStruct

type SongGroup struct {
	GroupName        string
	ClearedGroupName string
	Songs            []SongStruct
}

//Clearing unwanted characters from user input
func clearUserInputCfg(input string) string {
	clearRegex := regexp.MustCompile(`-|&|@|\[|\]|\#|\%|\*|\^|\!|\'|\"|\.|,| |\(|\)`)
	return clearRegex.ReplaceAllString(input, "_")
}

//Strips any extension from file
func stripExtension(input string) string {
	return strings.Split(input, ".")[0]
}

func clearUserInput(input string) string {
	clearRegex := regexp.MustCompile(`-|&|@|\[|\]|\#|\%|\*|\^|\!|\'|\"|\.|,|\(|\)|\\|\/`)
	return clearRegex.ReplaceAllString(input, "")
}

func clearUserInputOnlyfloats(input string) string {
	clearRegex := regexp.MustCompile(`[^0-9.]`)
	return clearRegex.ReplaceAllString(input, "")
}

func createFolderStruct(addonPath string, addonName string) {
	os.MkdirAll(path.Join(addonPath, "Addons", addonName), os.ModePerm)
	os.MkdirAll(path.Join(addonPath, "Addons", addonName, "songs"), os.ModePerm)
}

func readSongs() {
	files, err := ioutil.ReadDir(path.Join(workDir, "songs"))
	if err != nil {
		log.Fatal(err)
	}
	for _, file := range files {
		song := readSongData(file.Name())
		songs = append(songs, song)
	}
}
func readSongData(songName string) SongStruct {
	song := SongStruct{}
	//Check if song is already in .ogg format
	fileExtension := strings.Split(songName, ".")
	if fileExtension[len(fileExtension)-1] == "ogg" {
		song.Ogg = true
	} else {
		song.Ogg = false
	}
	song.Name = songName
	song.ClearedName = clearUserInputCfg(stripExtension(songName))
	song.ExtensionStrippedName = stripExtension(songName)
	//Prepare song path for reading
	songName = filepath.Join(workDir, "songs", songName)
	songName = fmt.Sprintf("%s", songName)
	exePath := path.Join(workDir, "data", "ffmpeg.exe")
	//Read song data
	exeHandle := exec.Command(exePath, "-i", songName, "-filter:a", "volumedetect", "-f", "null", "/dev/null")
	output, _ := exeHandle.CombinedOutput()
	outputSplit := strings.Split(string(output), "\n")
	//Setup default song data
	song.Volume = 0
	song.Length = 0
	song.Distance = 200
	song.Path = songName
	//Find needed information in
	for _, line := range outputSplit {
		if strings.Contains(line, "mean_volume") {
			lineSplit := strings.Split(line, " ")
			meanVolume, _ := strconv.ParseFloat(lineSplit[4], 64)
			meanVolume = meanVolume * -1
			song.Volume = meanVolume / (3.142 - (meanVolume*0.042 + meanVolume*0.0012))
		}
		if strings.Contains(line, "Duration:") {
			lineSplit := strings.Split(line, " ")
			duration := lineSplit[3]
			durationSplit := strings.Split(duration, ":")
			durationParsed, _ := strconv.Atoi(durationSplit[1])
			song.Length = (durationParsed + 1) * 60
		}
	}
	return song
}

func generateModCpp(addonName string, addonDescription string, authorName string) string {
	modCpp := fmt.Sprintf(`name = "%s";
tooltip = "%s";
overview = "%s";
author = "%s";
logo = "w_guitar_ca.paa";
picture = "w_guitar_ca.paa";`, addonName, addonDescription, addonDescription, authorName)
	return modCpp
}

func generatePboPrefix(addonName string) string {
	return addonName
}

func writePboPrefix(prefix string, addonPath string, addonName string) {
	byteData := []byte(prefix)
	ioutil.WriteFile(path.Join(addonPath, "Addons", addonName, "$PBOPREFIX$.txt"), byteData, os.ModePerm)
}

func writeModCpp(modCpp string, addonPath string) {
	byteData := []byte(modCpp)
	ioutil.WriteFile(path.Join(addonPath, "mod.cpp"), byteData, os.ModePerm)
}

func writeConfigCpp(configCpp string, addonPath string, addonName string) {
	byteData := []byte(configCpp)
	ioutil.WriteFile(path.Join(addonPath, "Addons", addonName, "config.cpp"), byteData, os.ModePerm)
}

func writeIconPaa(addonPath string) {
	iconPath := path.Join(workDir, "data", "w_guitar_ca.paa")
	iconData, err := ioutil.ReadFile(iconPath)
	if err != nil {
		log.Fatal(err)
	}
	ioutil.WriteFile(path.Join(addonPath, "w_guitar_ca.paa"), iconData, os.ModePerm)
}

func createPbo(addonPath string, addonName string) {
	pbo.Pack(filepath.Join(addonPath, "Addons", addonName), filepath.Join(addonPath, "Addons", addonName+".pbo"), true)
}

func cleanTempFiles(addonPath string, addonName string) {
	err := os.RemoveAll(filepath.Join(addonPath, "Addons", addonName))
	if err != nil {
		log.Fatal(err)
	}
}

func listSongs() {
	for _, song := range songs {
		fmt.Printf("Song: %s, clearedname: %s, volume: %f, length: %d, distance: %d, ogg: %t, path: %s\n", song.Name, song.ClearedName, song.Volume, song.Length, song.Distance, song.Ogg, song.Path)
	}
}

func convertSong(song *SongStruct, addonPath string, addonName string) {
	oggPath := path.Join(addonPath, "Addons", addonName, "songs", song.ClearedName+".ogg")
	exePath := path.Join(workDir, "data", "ffmpeg.exe")
	exeHandle := exec.Command(exePath, "-i", song.Path, "-c:a", "libvorbis", "-vn", oggPath)
	output, _ := exeHandle.CombinedOutput()
	fmt.Printf("FFmpeg.exe exited with output: %s", output)
	song.Path = oggPath
}

func convertSongs(addonPath string, addonName string, progressHandle *widget.ProgressBar) {
	numberOfSongs := len(songs)
	for i, _ := range songs {
		song := &songs[i]

		convertSong(song, addonPath, addonName)
		fmt.Printf("Converted song: %s\n", song.Name)

		progressHandle.SetValue(float64(i+1) / float64(numberOfSongs))
	}
}

func checkIfInputIsCorrect() {
	clearRegex := regexp.MustCompile(`-|&|@|\[|\]|\#|\%|\*|\^|\!|\'|\"|\.|,|\(|\)|\\|\/`)
	if clearRegex.MatchString(extensiondata.Name) {
		panic("Invalid characters in extension name")
	}
	if clearRegex.MatchString(extensiondata.Creator) {
		panic("Invalid characters in extension creator")
	}
	if clearRegex.MatchString(extensiondata.Description) {
		panic("Invalid characters in extension description")
	}
	if len(extensiondata.Name) < 1 {
		panic("Extension name should not be empty!")
	}
}

func createExtension(statusHandle *widget.Label, progressHandle *widget.ProgressBar) {
	checkIfInputIsCorrect()
	statusHandle.SetText("Creating folder structure")
	addonName := extensiondata.ClearedName
	addonPath := fmt.Sprintf("@%s", addonName)
	addonPath = path.Join(workDir, addonPath)
	createFolderStruct(addonPath, addonName)
	statusHandle.SetText("Generating mod.cpp")
	addonDescription := extensiondata.ClearedDescription
	addonAuthor := extensiondata.ClearedCreator
	modcpp := generateModCpp(addonName, addonDescription, addonAuthor)
	statusHandle.SetText("Writing mod.cpp")
	writeModCpp(modcpp, addonPath)
	statusHandle.SetText("Writing icon")
	writeIconPaa(addonPath)
	configcpp := generateConfigCpp()
	print(configcpp)
	statusHandle.SetText("Writing config.cpp")
	writeConfigCpp(configcpp, addonPath, addonName)
	statusHandle.SetText("Converting songs")
	convertSongs(addonPath, addonName, progressHandle)
	prefix := generatePboPrefix(addonName)
	statusHandle.SetText("Writing $PBOPREFIX$.txt")
	writePboPrefix(prefix, addonPath, addonName)
	statusHandle.SetText("Creating .pbo")
	createPbo(addonPath, addonName)
	statusHandle.SetText("Cleaning up")
	cleanTempFiles(addonPath, addonName)
	statusHandle.SetText("Done, your extension is ready!")
}

// UI===============================================================================================================

func main() {
	debug := true
	//Parse all song data before initiating UI
	workDir, _ = os.Getwd()
	fmt.Print("Parsing song information")
	readSongs()
	if debug {
		listSongs()
	}
	fmt.Print("Song parsing finished")
	//After gathering informations about songs initialize UI
	appHandle := app.New()
	window := appHandle.NewWindow("Music Extension Creator")
	size := fyne.NewSize(800, 600)
	window.Resize(size)
	statusLabel := widget.NewLabel("idle")
	progressBar := widget.NewProgressBar()
	var startButton *widget.Button
	startButton = widget.NewButton("Create extension", func() {
		startButton.Disable()
		createExtension(statusLabel, progressBar)

	})
	validator := validation.NewRegexp(`^[^-&@\[\]\#\%\*\^\!\'\"\.,\(\)\\\/]{3,}$`, "Invalid characters in input")
	statusBar := fyne.NewContainerWithLayout(layout.NewHBoxLayout(), widget.NewLabel("Status:"), statusLabel, layout.NewSpacer(), startButton)
	nameEntry := fyne.NewContainerWithLayout(layout.NewGridLayout(2), widget.NewLabel("Extension name:"), widget.NewEntry())
	nameEntry.Objects[1].(*widget.Entry).Validator = validator
	nameEntry.Objects[1].(*widget.Entry).OnChanged = func(value string) {
		extensiondata.Name = clearUserInput(value)
		extensiondata.ClearedName = clearUserInputCfg(value)
	}
	creatorEntry := fyne.NewContainerWithLayout(layout.NewGridLayout(2), widget.NewLabel("Extension creator:"), widget.NewEntry())
	creatorEntry.Objects[1].(*widget.Entry).Validator = validator
	creatorEntry.Objects[1].(*widget.Entry).OnChanged = func(value string) {
		extensiondata.Creator = clearUserInput(value)
		extensiondata.ClearedCreator = clearUserInputCfg(value)
	}
	descriptionEntry := fyne.NewContainerWithLayout(layout.NewGridLayout(2), widget.NewLabel("Extension description:"), widget.NewEntry())
	descriptionEntry.Objects[1].(*widget.Entry).Validator = validator
	descriptionEntry.Objects[1].(*widget.Entry).OnChanged = func(value string) {
		extensiondata.Description = clearUserInput(value)
		extensiondata.ClearedDescription = clearUserInputCfg(value)
	}
	generationOptions := widget.NewRadioGroup([]string{"Generate Guitar config", "Generate Pocket Music Player config", "Generate both configs"},
		func(value string) {
			if value == "Generate guitar config" {
				extensiondata.GenerateGuitarCfg = true
				extensiondata.GeneratePocketMusicPlayerCfg = false
			} else if value == "Generate pocket music player config" {
				extensiondata.GenerateGuitarCfg = false
				extensiondata.GeneratePocketMusicPlayerCfg = true
			} else {
				extensiondata.GenerateGuitarCfg = true
				extensiondata.GeneratePocketMusicPlayerCfg = true
			}
		})
	generationOptions.SetSelected("Generate both configs")

	floatValidator := validation.NewRegexp(`^[0-9.]*$`, "Invalid characters in input")
	volumeMultiplierEntry := fyne.NewContainerWithLayout(layout.NewGridLayout(2), widget.NewLabel("Volume multiplier for all songs, leave on default value if unsure:"), widget.NewEntry())
	volumeMultiplierEntry.Objects[1].(*widget.Entry).Validator = floatValidator
	volumeMultiplierEntry.Objects[1].(*widget.Entry).OnChanged = func(value string) {
		volumeMultiplier, err := strconv.ParseFloat(clearUserInputOnlyfloats(value), 64)
		if err != nil {
			fmt.Println("Error converting volume multiplier to float")
			return
		}
		extensiondata.VolumeMultiplier = volumeMultiplier
	}
	volumeMultiplierEntry.Objects[1].(*widget.Entry).SetText("1.0")

	topBar := fyne.NewContainerWithLayout(layout.NewVBoxLayout(), statusBar, progressBar, nameEntry, creatorEntry, descriptionEntry, generationOptions, volumeMultiplierEntry)
	songList := widget.NewList(
		func() int {
			return len(songs)
		},
		func() fyne.CanvasObject {
			volume := fyne.NewContainerWithLayout(layout.NewGridLayout(2), widget.NewLabel("Volume:"), widget.NewEntry())
			distance := fyne.NewContainerWithLayout(layout.NewGridLayout(2), widget.NewLabel("Distance:"), widget.NewEntry())
			item := fyne.NewContainerWithLayout(layout.NewVBoxLayout(), widget.NewLabel("Temp"), volume, distance)
			return item
		},
		func(id widget.ListItemID, item fyne.CanvasObject) {
			listItem := item.(*fyne.Container)
			listItem.Objects[0].(*widget.Label).SetText(songs[id].Name)
			listItem.Objects[1].(*fyne.Container).Objects[1].(*widget.Entry).SetText(fmt.Sprintf("%f", songs[id].Volume))
			listItem.Objects[1].(*fyne.Container).Objects[1].(*widget.Entry).OnChanged = func(value string) {
				volume, err := strconv.ParseFloat(value, 64)
				if err != nil {
					fmt.Println("Error converting volume to float")
					if value == "" {
						listItem.Objects[1].(*fyne.Container).Objects[1].(*widget.Entry).SetText("")
						songs[id].Volume = 0.0
					} else {
						listItem.Objects[1].(*fyne.Container).Objects[1].(*widget.Entry).SetText(fmt.Sprintf("%f", songs[id].Volume))
					}
				} else {
					songs[id].Volume = volume
				}
			}
			listItem.Objects[2].(*fyne.Container).Objects[1].(*widget.Entry).SetText(fmt.Sprintf("%d", songs[id].Distance))
			listItem.Objects[2].(*fyne.Container).Objects[1].(*widget.Entry).OnChanged = func(value string) {
				distance, err := strconv.Atoi(value)
				if err != nil {
					fmt.Println("Error converting distance to int")
					if value == "" {
						listItem.Objects[2].(*fyne.Container).Objects[1].(*widget.Entry).SetText("")
						songs[id].Distance = 0
					} else {
						listItem.Objects[2].(*fyne.Container).Objects[1].(*widget.Entry).SetText(fmt.Sprintf("%d", songs[id].Distance))
					}
				} else {
					songs[id].Distance = distance
				}
			}
		},
	)
	songListLabel := widget.NewLabel("Songs in extension:")
	songListContainer := fyne.NewContainerWithLayout(layout.NewBorderLayout(songListLabel, nil, nil, nil), songListLabel, songList)
	mainContainer := fyne.NewContainerWithLayout(layout.NewBorderLayout(topBar, nil, nil, nil), topBar, songListContainer)
	window.SetContent(mainContainer)
	window.ShowAndRun()
}
