package main

import (
	"bytes"
	"fmt"
	"text/template"
)

func generateConfigCpp() string {
	for i, _ := range songs {
		song := &songs[i]
		song.InternalPath = fmt.Sprintf(`\%s\songs\%s.ogg`, extensiondata.ClearedName, song.ClearedName)
		song.Pitch = 1
	}

	amountOfSongs := len(songs)
	songsPerGroup := 5
	amountOfGroups := amountOfSongs / songsPerGroup
	songGroups := make([]SongGroup, amountOfGroups)
	for i, _ := range songGroups {
		songGroups[i] = SongGroup{}
		songGroups[i].Songs = make([]SongStruct, 0)
		songGroups[i].GroupName = fmt.Sprintf("Song group %d", i)
		songGroups[i].ClearedGroupName = fmt.Sprintf("Song_group_%d", i)
	}
	for i, song := range songs {
		groupIndex := i / songsPerGroup
		song.Volume = song.Volume * extensiondata.VolumeMultiplier
		songGroups[groupIndex].Songs = append(songGroups[groupIndex].Songs, song)
	}

	extensiondata.SongGroups = songGroups
	tmpl, err := template.New("config").Parse(configCppTemplate)
	if err != nil {
		panic(err)
	}

	var tmplOutput bytes.Buffer

	err = tmpl.Execute(&tmplOutput, extensiondata)
	if err != nil {
		panic(err)
	}
	return tmplOutput.String()
}

const configCppTemplate = `
class CfgPatches
{
	class {{.ClearedName}}
	{
		units[]={};
		weapons[]={};
		requiredAddons[]={"A3_Characters_F_BLUFOR"};
	};
};

class CfgSounds
{
	sounds[] = {};

{{range .SongGroups}}
{{range .Songs}}
	class {{.ClearedName}}
	{
		name = "";
		sound[] = { "{{.InternalPath}}",{{.Volume}},{{.Pitch}},{{.Distance}} };
		titles[] = {};
	};
{{end}}
{{end}}
};

class CfgVehicles {
	{{if .GenerateGuitarCfg}}
	class Man;
	class CAManBase: Man {
		class ACE_SelfActions {
			class guitarActions
			{
				class guitarSongs
				{
					class {{.ClearedName}}_songs
					{
						displayName = "{{.Name}} extension songs";
{{range .SongGroups}}
						class {{.ClearedGroupName}}
						{
							displayName = "{{.GroupName}}";

						{{range .Songs}}
							class {{.ClearedName}}
							{
								displayname = "{{.ExtensionStrippedName}}";
								statement = "terminate guitar_script; guitar_script = ['{{.ClearedName}}',{{.Length}},{{.Distance}}] execVM '\ussr_guitar\scripts\playguitar.sqf'";
							};
						{{end}}
						};
{{end}}
					};

				};
			};
		};
	};
	{{end}}

	{{if .GeneratePocketMusicPlayerCfg}}
	class HubixPocketMusicPlayerObject {
		class ACE_Actions {
			class ACE_MainActions	
			{
				class HubixRadioSongs
				{
					class {{.ClearedName}}_songs
					{
						displayName = "{{.Name}} songs";
{{range .SongGroups}}
						class {{.ClearedGroupName}}
						{
							displayName = "{{.GroupName}}";

						{{range .Songs}}
							class {{.ClearedName}}
							{
								displayname = "{{.ExtensionStrippedName}}";	
								songlength = {{.Length}};
								statement = "[['{{.ClearedName}}',{{.Length}},{{.Distance}}, _target], '\HubixPocketMusicPlayer\scripts\playradio.sqf'] remoteExec ['execVM',2]";
							};
						{{end}}
						};	
{{end}}
					};
				};
				class HubixRadioShuffleSongs
				{
					condition=false;
					exceptions[] = {};
					displayname = Null;
{{range .SongGroups}}	
					{{range .Songs}}
						class {{.ClearedName}}
						{
							displayname = "{{.ExtensionStrippedName}}";
							songlength = {{.Length}};	
							statement = "[['{{.ClearedName}}',{{.Length}},{{.Distance}}, _target], '\HubixPocketMusicPlayer\scripts\playradio.sqf'] remoteExec ['execVM',2]";
						};
					{{end}}
{{end}}	
				};
			};
		};
	};

	{{end}}
};
`
