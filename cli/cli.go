package main

import (
	"bufio"
	"log"
	"os"
	"strings"

	"github.com/codegangsta/cli"

	"github.com/jh-bate/d-data-cli/client"
	"github.com/jh-bate/d-data-cli/models"
)

var (
	api *client.Api
)

func main() {

	app := cli.NewApp()
	api = client.InitApi(client.NewStore())

	app.Name = "D-mate"
	app.Usage = "Allow you to interact with your diabates data locally"
	app.Version = "0.0.1"
	app.Author = "Jamie Bate"
	app.Email = "jamie.h.bate@gmail.com"

	app.Commands = []cli.Command{

		//e.g. import --type smbg --data "[{"value": 5.5, "time": "2015-05-16T10:42:57.539Z" }]"
		{
			Name:      "import",
			ShortName: "i",
			Usage:     "import diabetes data",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "d, data",
					Usage: "the data value(s) that you want to save e.g. [`{\"value\": 5.5, \"time\": \"2015-05-16T10:42:57.539Z\" }`, `{\"value\": 7.5, \"time\": \"2015-05-16T11:42:57.539Z\" }`, ...]",
				},
				cli.StringFlag{
					Name:  "t, type",
					Usage: "the type of event(s) you are adding e.g. smbg",
				},
			},
			Action: importData,
		},
		//e.g. export --type smbg
		{
			Name:      "export",
			ShortName: "e",
			Usage:     "export diabetes data",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "t, type",
					Usage: "the type of event(s) you are adding e.g. smbg",
				},
				cli.StringFlag{
					Name:  "f, file",
					Value: "_eventdata.json",
					Usage: "the name of the file you want to sve it too",
				},
			},
			Action: exportData,
		},
	}

	app.Run(os.Args)

}

func exportData(c *cli.Context) {

	if c.String("type") == "" {
		log.Fatal("Please specify the type data to export --type or -t flag.")
	}
	if c.String("file") == "" {
		log.Fatal("Please specify file name to save your exported data to --file or -f flag.")
	}

	file := c.String("file")
	typeOfData := c.String("type")

	log.Printf("export [%s] to [%s]", typeOfData, file)
	dataFile, err := os.Create("./" + file)
	if err != nil {
		//kill quick
		log.Panic(err.Error())
		return
	}
	defer dataFile.Close()

	if typeOfData == models.EventTypes.Smbg.String() {
		api.GetSmbgs(dataFile, "a_3455")
	}
	dataFile.Close()
	return
}

func importData(c *cli.Context) {

	if c.String("data") == "" {
		log.Fatal("Please specify the data to import --data or -d flag.")
	}
	if c.String("type") == "" {
		log.Fatal("Please specify the type data to import --type or -t flag.")
	}

	data := c.String("data")
	typeOfData := c.String("type")

	if typeOfData == models.EventTypes.Smbg.String() {
		f := bufio.NewWriter(os.Stdout)
		saved, err := api.SaveSmbgs2(strings.NewReader(data), f, "a_3455")
		if err != nil {
			log.Println(err, log.Ldate|log.Ltime|log.Lshortfile)
		}
		saved.EncodeAsJson(f)
	}

}
